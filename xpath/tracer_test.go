package xpath_test

import (
	"context"
	"strings"
	"testing"

	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"github.com/sdcio/yang-parser/xpath"
	"github.com/sdcio/yang-parser/xpath/grammars/expr"
	"github.com/sdcio/yang-parser/xpath/xutils"
)

type recordingTracer struct {
	events []xpath.TraceEvent
}

func (r *recordingTracer) TraceEvent(ev xpath.TraceEvent) {
	r.events = append(r.events, ev)
}

type stubEntry struct {
	path     *sdcpb.Path
	val      xpath.Datum
	children map[string]*stubEntry
	tracer   xpath.Tracer
}

func (s *stubEntry) GetValue() (xpath.Datum, error) {
	if s.tracer != nil {
		s.tracer.TraceEvent(xpath.TraceEvent{
			Kind: "adapter_get_value",
			Node: s.path.ToXPath(false),
		})
	}
	return s.val, nil
}

func (s *stubEntry) Navigate(p *sdcpb.Path) (xpath.Entry, error) {
	if s.tracer != nil {
		s.tracer.TraceEvent(xpath.TraceEvent{
			Kind: "adapter_navigate",
			Details: map[string]any{
				"target": p.ToXPath(false),
			},
		})
	}
	cur := s
	for _, elem := range p.GetElem() {
		next, ok := cur.children[elem.GetName()]
		if !ok {
			return &stubEntry{
				path:   &sdcpb.Path{},
				val:    xpath.NewNodesetDatum(nil),
				tracer: s.tracer,
			}, nil
		}
		cur = next
	}
	return cur, nil
}

func (s *stubEntry) Copy() xpath.Entry {
	cp := *s
	return &cp
}

func (s *stubEntry) FollowLeafRef() (xpath.Entry, error) {
	return s, nil
}

func (s *stubEntry) GetSdcpbPath() *sdcpb.Path {
	return s.path
}

func (s *stubEntry) BreadthSearch(_ context.Context, _ *sdcpb.Path) ([]xpath.Entry, error) {
	return nil, nil
}

func compileMustMachine(t *testing.T, exprStr string) *xpath.Machine {
	t.Helper()
	pb := xpath.NewProgBuilder(exprStr)
	lex := expr.NewExprLex(exprStr, pb, nil)
	lex.Parse()
	prog, err := lex.CreateProgram(exprStr)
	if err != nil {
		t.Fatalf("compile %q: %v", exprStr, err)
	}
	return xpath.NewMachine(exprStr, prog, "test:1")
}

func TestTracer_InterleavesInstructionsStackAndAdapterCalls(t *testing.T) {
	tracer := &recordingTracer{}
	mtuEntry := &stubEntry{
		path: &sdcpb.Path{
			IsRootBased: true,
			Elem: []*sdcpb.PathElem{
				sdcpb.NewPathElem("interface", map[string]string{"name": "eth0"}),
				sdcpb.NewPathElem("mtu", nil),
			},
		},
		val:    xpath.NewNumDatum(1500),
		tracer: tracer,
	}
	ifEntry := &stubEntry{
		path: &sdcpb.Path{
			IsRootBased: true,
			Elem: []*sdcpb.PathElem{
				sdcpb.NewPathElem("interface", map[string]string{"name": "eth0"}),
			},
		},
		val:      xpath.NewBoolDatum(true),
		children: map[string]*stubEntry{"mtu": mtuEntry},
		tracer:   tracer,
	}

	mach := compileMustMachine(t, "mtu > 9000")
	res := xpath.NewCtxFromCurrent(context.Background(), mach, ifEntry).
		SetTracer(tracer).
		Run()

	gotBool, err := res.GetBoolResult()
	if err != nil {
		t.Fatalf("GetBoolResult error: %v", err)
	}
	if gotBool {
		t.Fatalf("expected false, got true")
	}

	if len(tracer.events) == 0 {
		t.Fatalf("expected trace events, got none")
	}
	if first := tracer.events[0]; first.Kind != xpath.TraceEventVMStart || first.Expr != "mtu > 9000" || first.Node != "/interface[name=eth0]" {
		t.Fatalf("unexpected vm_start event: %+v", first)
	}
	last := tracer.events[len(tracer.events)-1]
	if last.Kind != xpath.TraceEventVMEnd || last.Result != "BOOL(false)" || last.Error != "" || last.Panicked {
		t.Fatalf("unexpected vm_end event: %+v", last)
	}

	evalStartIdx, navIdx, getValIdx, evalDoneIdx := -1, -1, -1, -1
	for i, ev := range tracer.events {
		switch {
		case ev.Kind == xpath.TraceEventVMInstr && ev.Instruction == "evalLocPath":
			evalStartIdx = i
			if ev.Node != "/interface[name=eth0]" {
				t.Fatalf("expected instruction Node=/interface[name=eth0], got %q", ev.Node)
			}
		case ev.Kind == "adapter_navigate":
			navIdx = i
		case ev.Kind == "adapter_get_value":
			getValIdx = i
		case ev.Kind == xpath.TraceEventVMInstrDone && ev.Instruction == "evalLocPath":
			evalDoneIdx = i
			if len(ev.Stack) != 1 || ev.Stack[0] != "NUMBER(1500)" {
				t.Fatalf("expected stack [NUMBER(1500)] after evalLocPath, got %+v", ev.Stack)
			}
		}
	}
	if !(evalStartIdx >= 0 && evalStartIdx < navIdx && navIdx < getValIdx && getValIdx < evalDoneIdx) {
		t.Fatalf("expected interleaved order evalLocPath start (%d) < navigate (%d) < get_value (%d) < evalLocPath done (%d); events=%+v",
			evalStartIdx, navIdx, getValIdx, evalDoneIdx, tracer.events)
	}
}

func TestTracer_CapturesPanicsAndFormatsDatumSliceAndNilNodeset(t *testing.T) {
	panicTracer := &recordingTracer{}
	root := &stubEntry{
		path: &sdcpb.Path{IsRootBased: true},
		val:  xpath.NewBoolDatum(true),
	}
	badMach := compileMustMachine(t, "count('abc') = 1")
	res := xpath.NewCtxFromCurrent(context.Background(), badMach, root).
		SetTracer(panicTracer).
		Run()
	if _, err := res.GetBoolResult(); err == nil {
		t.Fatalf("expected error from count('abc'), got nil")
	}
	last := panicTracer.events[len(panicTracer.events)-1]
	if last.Kind != xpath.TraceEventVMEnd || !last.Panicked || last.Error == "" {
		t.Fatalf("expected vm_end with Panicked=true and Error, got %+v", last)
	}

	dsTracer := &recordingTracer{}
	llEntry := &stubEntry{
		path: &sdcpb.Path{
			IsRootBased: true,
			Elem:        []*sdcpb.PathElem{sdcpb.NewPathElem("tags", nil)},
		},
		val: xpath.NewDatumSliceDatum([]xpath.Datum{
			xpath.NewLiteralDatum("alpha"),
			xpath.NewLiteralDatum("beta"),
		}),
	}
	listEntry := &stubEntry{
		path: &sdcpb.Path{
			IsRootBased: true,
			Elem:        []*sdcpb.PathElem{sdcpb.NewPathElem("items", nil)},
		},
		// Mirrors data-server's list GetValue() placeholder nodeset: nil XpathNode elements.
		val: xpath.NewNodesetDatum(make([]xutils.XpathNode, 2)),
	}
	parent := &stubEntry{
		path: &sdcpb.Path{IsRootBased: true},
		val:  xpath.NewBoolDatum(true),
		children: map[string]*stubEntry{
			"tags":  llEntry,
			"items": listEntry,
		},
	}
	mach := compileMustMachine(t, "tags = 'alpha' and boolean(items)")
	dsRes := xpath.NewCtxFromCurrent(context.Background(), mach, parent).
		SetDebug(true).
		SetTracer(dsTracer).
		Run()
	ok, err := dsRes.GetBoolResult()
	if err != nil || !ok {
		t.Fatalf("expected true without error, got ok=%v err=%v", ok, err)
	}
	dbg := dsRes.GetDebugOutput()
	if dbg == "" {
		t.Fatalf("expected SetDebug(true) output to still work alongside SetTracer")
	}
	if strings.Contains(dbg, "PANIC") || !strings.Contains(dbg, "DATUMSLICE") {
		t.Fatalf("expected clean debug output without String() panic and with DATUMSLICE, got:\n%s", dbg)
	}

	var sawDatumSlice, sawNilNodeset bool
	for _, ev := range dsTracer.events {
		for _, item := range ev.Stack {
			if item == `DATUMSLICE([LITERAL("alpha"), LITERAL("beta")])` {
				sawDatumSlice = true
			}
			if item == "NODESET(len=2, placeholder_nodes=2)" {
				sawNilNodeset = true
			}
		}
	}
	if !sawDatumSlice || !sawNilNodeset {
		t.Fatalf("expected structured stack snapshots for DatumSlice (%v) and nil-placeholder Nodeset (%v), events=%+v",
			sawDatumSlice, sawNilNodeset, dsTracer.events)
	}
}

func TestTracer_SnapshotsPathStackAndPredicateKeys(t *testing.T) {
	tracer := &recordingTracer{}
	mtuEntry := &stubEntry{
		path: &sdcpb.Path{
			IsRootBased: true,
			Elem: []*sdcpb.PathElem{
				sdcpb.NewPathElem("interface", map[string]string{"name": "eth0"}),
				sdcpb.NewPathElem("mtu", nil),
			},
		},
		val:    xpath.NewNumDatum(9000),
		tracer: tracer,
	}
	ifEntry := &stubEntry{
		path: &sdcpb.Path{
			IsRootBased: true,
			Elem:        []*sdcpb.PathElem{sdcpb.NewPathElem("interface", map[string]string{"name": "eth0"})},
		},
		val:      xpath.NewBoolDatum(true),
		children: map[string]*stubEntry{"mtu": mtuEntry},
		tracer:   tracer,
	}
	root := &stubEntry{
		path:     &sdcpb.Path{IsRootBased: true},
		val:      xpath.NewBoolDatum(true),
		children: map[string]*stubEntry{"interface": ifEntry},
		tracer:   tracer,
	}

	mach := compileMustMachine(t, "interface[name='eth0']/mtu = 9000")
	res := xpath.NewCtxFromCurrent(context.Background(), mach, root).
		SetTracer(tracer).
		Run()
	ok, err := res.GetBoolResult()
	if err != nil || !ok {
		t.Fatalf("expected true without error, got ok=%v err=%v", ok, err)
	}

	var sawPredKey, sawKeyedPath bool
	for _, ev := range tracer.events {
		for _, m := range ev.PredicateKeys {
			if m["name"] == "eth0" {
				sawPredKey = true
			}
		}
		for _, p := range ev.PathStack {
			if p == "interface[name=eth0]/mtu" {
				sawKeyedPath = true
			}
		}
	}
	if !sawPredKey || !sawKeyedPath {
		t.Fatalf("expected PredicateKeys[name=eth0] (%v) and PathStack[interface[name=eth0]/mtu] (%v); events=%+v",
			sawPredKey, sawKeyedPath, tracer.events)
	}
}



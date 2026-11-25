package xpath

import (
	"fmt"

	"github.com/sdcio/yang-parser/xpath/xutils"
)

// TraceEventKind classifies an event emitted during XPath machine execution.
type TraceEventKind string

const (
	TraceEventVMStart     TraceEventKind = "vm_start"
	TraceEventVMInstr     TraceEventKind = "vm_instr"
	TraceEventVMInstrDone TraceEventKind = "vm_instr_done"
	TraceEventVMEnd       TraceEventKind = "vm_end"
)

// TraceEvent captures a single VM step or caller-supplied adapter event in
// execution order. Stack is ordered top-of-stack first; PathStack is ordered
// bottom-to-top to match frame nesting.
type TraceEvent struct {
	Kind          TraceEventKind      `json:"kind"`
	Level         int                 `json:"level,omitempty"`
	Step          int                 `json:"step,omitempty"`
	Expr          string              `json:"expr,omitempty"`
	Node          string              `json:"node,omitempty"`
	Instruction   string              `json:"instruction,omitempty"`
	Stack         []string            `json:"stack,omitempty"`
	PathStack     []string            `json:"path_stack,omitempty"`
	PredicateKeys []map[string]string `json:"predicate_keys,omitempty"`
	Result        string              `json:"result,omitempty"`
	Error         string              `json:"error,omitempty"`
	Panicked      bool                `json:"panicked,omitempty"`
	Details       map[string]any      `json:"details,omitempty"`
}

// Tracer receives ordered TraceEvents from an XPath context when attached via
// SetTracer. Nil tracers incur no allocation or formatting overhead.
type Tracer interface {
	TraceEvent(ev TraceEvent)
}

// FormatDatum returns a single-line, type-tagged representation of d suitable
// for structured traces, including nil-element nodesets and DatumSlices.
func FormatDatum(d Datum) string {
	if d == nil {
		return "<nil>"
	}
	return d.traceString()
}

func formatStackable(s stackable) string {
	if d, ok := s.(Datum); ok {
		return FormatDatum(d)
	}
	return fmt.Sprintf("%s", s)
}

func (ctx *context) buildInstrTraceEvent(kind TraceEventKind, step int, instrName string) TraceEvent {
	return TraceEvent{
		Kind:          kind,
		Level:         ctx.level,
		Step:          step,
		Node:          ctx.traceNode(),
		Instruction:   instrName,
		Stack:         ctx.snapshotStack(),
		PathStack:     ctx.snapshotPathStack(),
		PredicateKeys: ctx.snapshotPredicateKeys(),
	}
}

func (ctx *context) traceNode() string {
	if ctx.current != nil {
		if p := ctx.current.GetSdcpbPath(); p != nil {
			s := p.ToXPath(false)
			if s == "" {
				return "/"
			}
			return s
		}
	}
	if ctx.node != nil {
		return xutils.NodeString(ctx.node)
	}
	return ""
}

func (ctx *context) snapshotStack() []string {
	if len(ctx.stack) == 0 {
		return nil
	}
	out := make([]string, len(ctx.stack))
	for i := range ctx.stack {
		out[i] = formatStackable(ctx.stack[len(ctx.stack)-1-i])
	}
	return out
}

func (ctx *context) snapshotPathStack() []string {
	if ctx.actualPathStack == nil || len(ctx.actualPathStack.stack) == 0 {
		return nil
	}
	out := make([]string, len(ctx.actualPathStack.stack))
	for i, p := range ctx.actualPathStack.stack {
		if p == nil {
			out[i] = "<nil>"
			continue
		}
		s := p.ToXPath(false)
		if s == "" && p.GetIsRootBased() {
			s = "/"
		}
		out[i] = s
	}
	return out
}

func (ctx *context) snapshotPredicateKeys() []map[string]string {
	if ctx.predicatePathElemStack == nil || len(ctx.predicatePathElemStack.stack) == 0 {
		return nil
	}
	out := make([]map[string]string, len(ctx.predicatePathElemStack.stack))
	for i, m := range ctx.predicatePathElemStack.stack {
		cp := make(map[string]string, len(m))
		for k, v := range m {
			cp[k] = v
		}
		out[i] = cp
	}
	return out
}

// Copyright (c) 2017-2019, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package parse

type NodeType int

func (t NodeType) Type() NodeType { return t }
func (t NodeType) String() string { return nodeNames[t] }

func (t NodeType) IsDataNode() bool {
	return (t > NodeDataDef) && (t < NodeDataDefEnd)
}

func (t NodeType) IsDataOrCaseNode() bool {
	return (t > NodeDataDef) && (t <= NodeCase)
}

func (t NodeType) IsTypeRestriction() bool {
	return (t > NodeTypeRestrictionStart && t < NodeTypeRestrictionEnd)
}

func (t NodeType) IsConfigdNode() bool {
	return (t > NodeConfigdStart) && (t < NodeConfigdStop)
}

func (t NodeType) IsOpdDefNode() bool {
	return (t > NodeOpdDef) && (t < NodeOpdDefEnd)
}

func (t NodeType) IsOpdExtension() bool {
	return (t > NodeOpdExtensionStart) && (t < NodeOpdExtensionEnd)
}

func (t NodeType) IsExtensionNode() bool {
	return t == NodeUnknown || t.IsConfigdNode() || t.IsOpdExtension()
}

func (t NodeType) IsDeviateNode() bool {
	return (t > NodeDeviate) && (t < NodeDeviateEnd)
}

const (
	NodeUnknown NodeType = iota //Unknown node (yang extension)
	NodeModule
	NodeImport
	NodeInclude
	NodeRevision
	NodeSubmodule
	NodeBelongsTo
	NodeTypedef
	//Special purpose nodes according to RFC
	NodeDataDef
	NodeContainer
	NodeLeaf
	NodeLeafList
	NodeList
	NodeChoice
	NodeUses
	NodeAnyxml
	NodeDataDefEnd
	//Additional node for augment data-def
	NodeCase
	NodeGrouping
	NodeMust
	NodeRpc
	NodeInput
	NodeOutput
	NodeNotification
	NodeAugment
	NodeIdentity
	NodeExtension
	NodeArgument
	NodeFeature
	NodeDeviation
	NodeDeviate
	NodeDeviateAdd
	NodeDeviateDelete
	NodeDeviateReplace
	NodeDeviateNotSupported
	NodeDeviateEnd
	// Start of Type Restrictions
	NodeTypeRestrictionStart
	NodeTyp
	NodeRange
	NodeLength
	NodePattern
	NodeEnum
	NodeBit
	NodePath
	NodeFractionDigits
	NodeRequireInstance
	NodeTypeRestrictionEnd
	// End of Type Restrictions
	NodeContact
	NodeDescription
	NodeNamespace
	NodeOrganization
	NodePrefix
	NodeReference
	NodeYangVersion
	NodeRevisionDate
	NodeDefault
	NodeStatus
	NodeUnits
	NodeConfig
	NodeIfFeature
	NodePresence
	NodeWhen
	NodeErrorAppTag
	NodeErrorMessage
	NodeMandatory
	NodeMinElements
	NodeMaxElements
	NodeOrderedBy
	NodeKey
	NodeUnique
	NodeRefine
	NodeBase
	NodeYinElement
	NodeValue
	NodePosition
	NodeConfigdStart
	NodeConfigdHelp
	NodeConfigdValidate
	NodeConfigdNormalize
	NodeConfigdSyntax
	NodeConfigdPriority
	NodeConfigdAllowed
	NodeConfigdBegin
	NodeConfigdEnd
	NodeConfigdCreate
	NodeConfigdDelete
	NodeConfigdUpdate
	NodeConfigdSubst
	NodeConfigdSecret
	NodeConfigdErrMsg
	NodeConfigdPHelp
	NodeConfigdCallRpc
	NodeConfigdGetState
	NodeConfigdDeferActions
	NodeConfigdMust
	NodeConfigdStop
	NodeOpdDef
	NodeOpdArgument
	NodeOpdAugment
	NodeOpdCommand
	NodeOpdOption
	NodeOpdDefEnd
	NodeOpdInherit
	NodeOpdRepeatable
	NodeOpdPassOpcArgs
	NodeOpdPrivileged
	NodeOpdLocal
	NodeOpdSecret
	// OPD extensions which can extend non-opd nodes
	NodeOpdExtensionStart
	NodeOpdHelp
	NodeOpdAllowed
	NodeOpdOnEnter
	NodeOpdPatternHelp
	NodeOpdExtensionEnd
	NodeTypeIndexSize // MUST BE LAST. NOT A VALID NODE TYPE
)

var nodeNames = [...]string{
	NodeUnknown:             "unknown",
	NodeModule:              "module",
	NodeImport:              "import",
	NodeInclude:             "include",
	NodeRevision:            "revision",
	NodeSubmodule:           "submodule",
	NodeBelongsTo:           "belongs-to",
	NodeTypedef:             "typedef",
	NodeTyp:                 "type",
	NodeContainer:           "container",
	NodeMust:                "must",
	NodeLeaf:                "leaf",
	NodeLeafList:            "leaf-list",
	NodeList:                "list",
	NodeChoice:              "choice",
	NodeCase:                "case",
	NodeDataDef:             "data definition",
	NodeDataDefEnd:          "data definition end",
	NodeAnyxml:              "anyxml",
	NodeGrouping:            "grouping",
	NodeUses:                "uses",
	NodeRpc:                 "rpc",
	NodeInput:               "input",
	NodeOutput:              "output",
	NodeNotification:        "notification",
	NodeAugment:             "augment",
	NodeIdentity:            "identity",
	NodeExtension:           "extension",
	NodeArgument:            "argument",
	NodeFeature:             "feature",
	NodeDeviation:           "deviation",
	NodeDeviate:             "deviate",
	NodeDeviateAdd:          "deviate-add",
	NodeDeviateDelete:       "deviate-delete",
	NodeDeviateReplace:      "deviate-replace",
	NodeDeviateNotSupported: "deviate-not-supported",
	NodeRange:               "range",
	NodeLength:              "length",
	NodePattern:             "pattern",
	NodeEnum:                "enum",
	NodeBit:                 "bit",
	NodeContact:             "contact",
	NodeDescription:         "description",
	NodeNamespace:           "namespace",
	NodeOrganization:        "organization",
	NodePrefix:              "prefix",
	NodeReference:           "reference",
	NodeYangVersion:         "yang-version",
	NodeRevisionDate:        "revision-date",
	NodeDefault:             "default",
	NodeStatus:              "status",
	NodeUnits:               "units",
	NodePath:                "path",
	NodeRequireInstance:     "require-instance",
	NodeConfig:              "config",
	NodeIfFeature:           "if-feature",
	NodePresence:            "presence",
	NodeWhen:                "when",
	NodeErrorAppTag:         "error-app-tag",
	NodeErrorMessage:        "error-message",
	NodeMandatory:           "mandatory",
	NodeMinElements:         "min-elements",
	NodeMaxElements:         "max-elements",
	NodeOrderedBy:           "ordered-by",
	NodeKey:                 "key",
	NodeUnique:              "unique",
	NodeRefine:              "refine",
	NodeBase:                "base",
	NodeYinElement:          "yin",
	NodeValue:               "value",
	NodePosition:            "position",
	NodeFractionDigits:      "fraction-digits",
	NodeConfigdStart:        "configd start",
	NodeConfigdHelp:         "configd:help",
	NodeConfigdValidate:     "configd:validate",
	NodeConfigdNormalize:    "configd:normalize",
	NodeConfigdSyntax:       "configd:syntax",
	NodeConfigdPriority:     "configd:priority",
	NodeConfigdAllowed:      "configd:allowed",
	NodeConfigdBegin:        "configd:begin",
	NodeConfigdEnd:          "configd:end",
	NodeConfigdCreate:       "configd:create",
	NodeConfigdDelete:       "configd:delete",
	NodeConfigdUpdate:       "configd:update",
	NodeConfigdSubst:        "configd:subst",
	NodeConfigdSecret:       "configd:secret",
	NodeConfigdErrMsg:       "configd:error-message",
	NodeConfigdPHelp:        "configd:pattern-help",
	NodeConfigdCallRpc:      "configd:call-rpc",
	NodeConfigdGetState:     "configd:get-state",
	NodeConfigdDeferActions: "configd:defer-actions",
	NodeConfigdMust:         "configd:must",
	NodeConfigdStop:         "configd stop",
	NodeOpdDef:              "opd definition",
	NodeOpdArgument:         "opd:argument",
	NodeOpdAugment:          "opd:augment",
	NodeOpdCommand:          "opd:command",
	NodeOpdOption:           "opd:option",
	NodeOpdDefEnd:           "opd definition end",
	NodeOpdOnEnter:          "opd:on-enter",
	NodeOpdInherit:          "opd:inherit",
	NodeOpdRepeatable:       "opd:repeatable",
	NodeOpdPassOpcArgs:      "opd:pass-opc-args",
	NodeOpdPrivileged:       "opd:privileged",
	NodeOpdLocal:            "opd:local",
	NodeOpdSecret:           "opd:secret",
	NodeOpdExtensionStart:   "opd extension start",
	NodeOpdHelp:             "opd:help",
	NodeOpdAllowed:          "opd:allowed",
	NodeOpdPatternHelp:      "opd:pattern-help",
	NodeOpdExtensionEnd:     "opd extension end",
}

var nodeTypeMap map[string]NodeType

func init() {
	nodeTypeMap = make(map[string]NodeType, NodeTypeIndexSize)
	for i, v := range nodeNames {
		nodeTypeMap[v] = NodeType(i)
	}
}

func NodeTypeFromName(name, arg string) NodeType {
	if ntype, ok := nodeTypeMap[name]; ok {
		if ntype == NodeDeviate {
			switch arg {
			case "not-supported":
				ntype = NodeDeviateNotSupported
			case "add":
				ntype = NodeDeviateAdd
			case "delete":
				ntype = NodeDeviateDelete
			case "replace":
				ntype = NodeDeviateReplace
			default:
				ntype = NodeDeviate
			}
		}
		return ntype
	}

	return NodeUnknown
}

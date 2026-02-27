package op

import (
	"slices"

	"github.com/sekiguchi-nagisa/guniset/set"
)

type Node interface {
	Eval(*EvalContext) set.UniSet
}

type RangeNode struct { // FF..U+1234
	runeRange set.RuneRange
}

func (i *RangeNode) Eval(*EvalContext) set.UniSet {
	builder := set.UniSetBuilder{}
	builder.AddRange(i.runeRange)
	return builder.Build()
}

type GeneralCategoryNode struct { // cat:Lu,Lo
	properties []GeneralCategory
}

func NewGeneralCategoryNode(properties []GeneralCategory) *GeneralCategoryNode {
	node := GeneralCategoryNode{}
	node.properties = properties[0:]
	slices.Sort(node.properties)
	node.properties = slices.Compact(node.properties)
	return &node
}

func (g *GeneralCategoryNode) Eval(context *EvalContext) set.UniSet {
	builder := set.UniSetBuilder{}
	for _, property := range g.properties {
		if s, ok := context.CateMap[property]; ok {
			builder.AddSet(s)
		} else if comb := property.Combinations(); len(comb) > 0 {
			for _, c := range comb {
				if s, ok = context.CateMap[c]; ok {
					builder.AddSet(s)
				}
			}
		}
	}
	return builder.Build()
}

type EastAsianWidthNode struct { // eaw:W,F
	properties []EastAsianWidth
}

func NewEastAsianWidthNode(properties []EastAsianWidth) *EastAsianWidthNode {
	node := EastAsianWidthNode{}
	node.properties = properties[0:]
	slices.Sort(node.properties)
	node.properties = slices.Compact(node.properties)
	return &node
}

func (e *EastAsianWidthNode) Eval(context *EvalContext) set.UniSet {
	builder := set.UniSetBuilder{}
	for _, property := range e.properties {
		if s, ok := context.EawMap[property]; ok {
			builder.AddSet(s)
		} else if property == EAW_N {
			builder.AddSet(context.FillEawN())
		}
	}
	return builder.Build()
}

type ScriptNode struct { // sc:Common
	properties []Script
	extension  bool
}

func NewScriptNode(properties []Script) *ScriptNode {
	node := ScriptNode{extension: false}
	node.properties = properties[0:]
	slices.Sort(node.properties)
	node.properties = slices.Compact(node.properties)
	return &node
}

func NewScriptXNode(properties []Script) *ScriptNode {
	node := NewScriptNode(properties)
	node.extension = true
	return node
}

func (e *ScriptNode) Eval(context *EvalContext) set.UniSet {
	builder := set.UniSetBuilder{}
	for _, property := range e.properties {
		if e.extension {
			if s, ok := context.ScriptXMap[property]; ok {
				builder.AddSet(s)
			}
		} else {
			if s, ok := context.ScriptMap[property]; ok {
				builder.AddSet(s)
			} else if property == context.DefRecord.ScriptDef.Unknown() {
				builder.AddSet(context.FillScriptUnknown())
			}
		}
	}
	return builder.Build()
}

type PropertyNode[T ~int] struct {
	properties []T
	callback   func(*EvalContext, T) (*set.UniSet, bool)
}

func NewPropertyNode[T ~int](properties []T, callback func(*EvalContext, T) (*set.UniSet, bool)) *PropertyNode[T] {
	node := PropertyNode[T]{properties: properties, callback: callback}
	node.properties = properties[0:]
	slices.Sort(node.properties)
	node.properties = slices.Compact(node.properties)
	return &node
}

func (p *PropertyNode[T]) Eval(context *EvalContext) set.UniSet {
	builder := set.UniSetBuilder{}
	for _, property := range p.properties {
		if s, ok := p.callback(context, property); ok {
			builder.AddSet(s)
		}
	}
	return builder.Build()
}

type CompNode struct { // ! SET
	node Node
}

func (c *CompNode) Eval(context *EvalContext) set.UniSet {
	negate := true
	node := c.node
	for target, ok := node.(*CompNode); ok; target, ok = target.node.(*CompNode) {
		negate = !negate
		node = target.node
	}
	uniSet := node.Eval(context)
	if negate {
		tmp := set.NewUniSetAll()
		tmp.RemoveSet(&uniSet)
		uniSet = tmp
	}
	return uniSet
}

type UnionNode struct { // SET + SET
	left  Node
	right Node
}

func (u *UnionNode) Eval(context *EvalContext) set.UniSet {
	leftSet := u.left.Eval(context)
	rightSet := u.right.Eval(context)
	leftSet.AddSet(&rightSet)
	return leftSet
}

type DiffNode struct { // SET - SET
	left  Node
	right Node
}

func (d *DiffNode) Eval(context *EvalContext) set.UniSet {
	leftSet := d.left.Eval(context)
	rightSet := d.right.Eval(context)
	leftSet.RemoveSet(&rightSet)
	return leftSet
}

type IntersectNode struct { // SET * SET
	left  Node
	right Node
}

func (i *IntersectNode) Eval(context *EvalContext) set.UniSet {
	leftSet := i.left.Eval(context)
	rightSet := i.right.Eval(context)
	newSet := leftSet.AndSet(&rightSet)
	return newSet
}

type CaseFoldNode struct { // @fold(SET)
	node Node
}

func (c *CaseFoldNode) Eval(context *EvalContext) set.UniSet {
	retSet := c.node.Eval(context)
	if len(context.CaseFoldingMap) == 0 {
		return retSet
	}
	builder := set.UniSetBuilder{}
	for r := range retSet.Iter {
		if to, ok := context.CaseFoldingMap[r]; ok {
			builder.Add(to)
		} else {
			builder.Add(r)
		}
	}
	return builder.Build()
}

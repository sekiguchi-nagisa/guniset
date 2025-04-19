package op

import (
	"github.com/sekiguchi-nagisa/guniset/set"
	"slices"
)

type Node interface {
	Eval(*EvalContext) set.UniSet
}

type IntervalNode struct { // FF..U+1234
	interval set.RuneInterval
}

func (i *IntervalNode) Eval(*EvalContext) set.UniSet {
	builder := set.UniSetBuilder{}
	builder.AddInterval(i.interval)
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
		if s, ok := context.catSet[property]; ok {
			builder.AddSet(s)
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
		if s, ok := context.eawSet[property]; ok {
			builder.AddSet(s)
		} else if property == EAW_N {
			builder.AddSet(context.FillEawN())
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

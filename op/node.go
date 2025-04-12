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
	uniSet := set.NewUniSet()
	uniSet.AddInterval(i.interval)
	return uniSet
}

type GeneralCategoryNode struct { // cat:Lu,Lo
	properties []GeneralCategory
}

func NewGeneralCategoryNode(properties []GeneralCategory) *GeneralCategoryNode {
	node := GeneralCategoryNode{}
	copy(node.properties, properties)
	slices.Sort(node.properties)
	node.properties = slices.Compact(node.properties)
	return &node
}

func (g *GeneralCategoryNode) Eval(context *EvalContext) set.UniSet {
	uniSet := set.NewUniSet()
	for _, property := range g.properties {
		if s, ok := context.catSet[property]; ok {
			uniSet.AddSet(&s)
		}
	}
	return uniSet
}

type EastAsianWidthNode struct { // eaw:W,F
	properties []EastAsianWidth
}

func NewEastAsianWidthNode(properties []EastAsianWidth) *EastAsianWidthNode {
	node := EastAsianWidthNode{}
	copy(node.properties, properties)
	slices.Sort(node.properties)
	node.properties = slices.Compact(node.properties)
	return &node
}

func (e *EastAsianWidthNode) Eval(context *EvalContext) set.UniSet {
	uniSet := set.NewUniSet()
	for _, property := range e.properties {
		if s, ok := context.eawSet[property]; ok {
			uniSet.AddSet(&s)
		}
	}
	return uniSet
}

type UnionNode struct { // SET + SET
	left  Node
	right Node
}

func (u *UnionNode) Eval(context *EvalContext) set.UniSet {
	leftNode := u.left.Eval(context)
	rightNode := u.right.Eval(context)
	leftNode.AddSet(&rightNode)
	return leftNode
}

type DiffNode struct { // SET - SET
	left  Node
	right Node
}

func (d *DiffNode) Eval(context *EvalContext) set.UniSet {
	leftNode := d.left.Eval(context)
	rightNode := d.right.Eval(context)
	leftNode.RemoveSet(&rightNode)
	return leftNode
}

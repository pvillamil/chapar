package widgets

import (
	"image"
	"strings"
	"time"

	"gioui.org/io/pointer"
	"gioui.org/op"

	"gioui.org/x/component"

	"gioui.org/io/input"

	"gioui.org/op/clip"

	"gioui.org/op/paint"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type TreeView struct {
	nodes []*TreeViewNode
	list  *widget.List

	filterText    string
	filteredNodes []*TreeViewNode
}

type TreeViewNode struct {
	collapsed bool
	clickable *widget.Clickable

	Children   []*TreeViewNode
	Text       string
	Identifier string

	lastClickAt time.Time
	order       int

	menuClickable   *widget.Clickable
	menuContextArea component.ContextArea
	menu            component.MenuState
	menuOptions     []string

	onDoubleClick func(tr *TreeViewNode)
}

func NewTreeView() *TreeView {
	return &TreeView{
		list: &widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
	}
}

func (t *TreeView) AddRootNode(text string, collapsed bool) {
	t.AddNode(NewNode(text, collapsed), nil)
}

func NewNode(text string, collapsed bool) *TreeViewNode {
	return &TreeViewNode{
		Text:          text,
		collapsed:     collapsed,
		clickable:     &widget.Clickable{},
		menuClickable: &widget.Clickable{},
		order:         1,
		menuContextArea: component.ContextArea{
			Activation:       pointer.ButtonPrimary,
			AbsolutePosition: true,
		},
		menuOptions: []string{"Delete", "Duplicate"},
	}
}

func (tr *TreeViewNode) OnDoubleClick(f func(tr *TreeViewNode)) {
	tr.onDoubleClick = f
}

func (tr *TreeViewNode) AddChild(node *TreeViewNode) {
	tr.Children = append(tr.Children, node)
}

func (tr *TreeViewNode) SetIdentifier(identifier string) {
	tr.Identifier = identifier
}

func (t *TreeView) AddNode(node *TreeViewNode, parent *TreeViewNode) {
	if parent == nil {
		t.nodes = append(t.nodes, node)
		return
	}

	parent.Children = append(parent.Children, node)
}

func (t *TreeView) Filter(text string) {
	t.filterText = text

	if text == "" {
		t.filteredNodes = make([]*TreeViewNode, 0)
		return
	}

	var items = make([]*TreeViewNode, 0)
	for _, item := range t.nodes {
		if strings.Contains(item.Text, text) {
			items = append(items, item)
		}

		for _, child := range item.Children {
			if strings.Contains(child.Text, text) {
				items = append(items, child)
			}
		}
	}

	t.filteredNodes = items
}

func (t *TreeView) childLayout(theme *material.Theme, gtx layout.Context, node *TreeViewNode) layout.Dimensions {
	background := theme.Palette.Bg
	for node.clickable.Clicked(gtx) {
		node.collapsed = !node.collapsed
	}

	return node.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Background{}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, 0).Push(gtx.Ops).Pop()
				if gtx.Source == (input.Source{}) {
					background = Disabled(theme.Palette.Bg)
				} else if node.clickable.Hovered() || gtx.Focused(node.clickable) {
					background = Hovered(theme.Palette.Bg)
				}

				paint.Fill(gtx.Ops, background)
				return layout.Dimensions{Size: gtx.Constraints.Min}
			},
			func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8), Left: unit.Dp(48)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return material.Label(theme, theme.TextSize, node.Text).Layout(gtx)
				})
			},
		)
	})
}

func (t *TreeView) parentLayout(gtx layout.Context, theme *material.Theme, node *TreeViewNode) layout.Dimensions {
	background := theme.Palette.Bg
	for node.clickable.Clicked(gtx) {
		// is this a double click?
		if time.Since(node.lastClickAt) < 500*time.Millisecond {
			if node.onDoubleClick != nil {
				node.onDoubleClick(node)
			}
		} else {
			node.lastClickAt = time.Now()
			if node.Children == nil {
				continue
			}

			node.collapsed = !node.collapsed
		}
	}

	pr := node.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Background{}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, 0).Push(gtx.Ops).Pop()
				if gtx.Source == (input.Source{}) {
					background = Disabled(theme.Palette.Bg)
				} else if node.clickable.Hovered() || gtx.Focused(node.clickable) {
					background = Hovered(theme.Palette.Bg)
				}
				paint.Fill(gtx.Ops, background)
				return layout.Dimensions{Size: gtx.Constraints.Min}
			},
			func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8), Left: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if node.Children == nil {
								s := gtx.Constraints.Min
								s.X = gtx.Dp(unit.Dp(16))
								return layout.Dimensions{Size: s}
							}

							gtx.Constraints.Min.X = gtx.Dp(16)
							if !node.collapsed {
								return ExpandIcon.Layout(gtx, theme.ContrastFg)
							}
							return ForwardIcon.Layout(gtx, theme.ContrastFg)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return material.Label(theme, theme.TextSize, node.Text).Layout(gtx)
									})
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									if !node.clickable.Hovered() {
										return layout.Dimensions{}
									}

									//ib := &IconButton{
									//	Icon:                 MoreVertIcon,
									//	Color:                theme.ContrastFg,
									//	Size:                 unit.Dp(16),
									//	BackgroundColor:      Hovered(theme.Palette.Bg),
									//	BackgroundColorHover: theme.Palette.Bg,
									//	Clickable:            node.menuClickable,
									//}

									iconBtn := layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										//return ib.Layout(gtx, theme)
										gtx.Constraints.Min.X = gtx.Dp(16)
										return MoreVertIcon.Layout(gtx, theme.ContrastFg)
									})

									menuMicro := op.Record(gtx.Ops)
									node.menu.Options = node.menu.Options[:0]
									for _, opt := range node.menuOptions {
										opt := opt
										node.menu.Options = append(node.menu.Options, func(gtx layout.Context) layout.Dimensions {
											dim := component.MenuItem(theme, &widget.Clickable{}, opt).Layout(gtx)
											return dim
										})
									}
									menuDim := component.Menu(theme, &node.menu).Layout(gtx)
									menuMacroCall := menuMicro.Stop()

									return layout.Stack{}.Layout(gtx,
										layout.Stacked(func(gtx layout.Context) layout.Dimensions {
											return iconBtn
										}),
										layout.Expanded(func(gtx layout.Context) layout.Dimensions {
											return node.menuContextArea.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
												offset := layout.Inset{
													Top:  unit.Dp(float32(iconBtn.Size.Y)/gtx.Metric.PxPerDp + 1),
													Left: unit.Dp(4),
												}
												return offset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
													gtx.Constraints.Min = image.Point{}
													menuMacroCall.Add(gtx.Ops)
													return menuDim
												})
											})
										}),
									)
								}),
							)
						}),
					)
				})
			},
		)
	})

	if node.collapsed {
		return pr
	}

	var children []layout.FlexChild
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return pr
	}))
	for _, child := range node.Children {
		child := child
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.childLayout(theme, gtx, child)
		}))
	}

	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle, Spacing: layout.SpaceEnd}.Layout(gtx, children...)

}

func (t *TreeView) Layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	nodes := t.nodes
	if t.filterText != "" {
		nodes = t.filteredNodes
	}

	if len(nodes) == 0 {
		return layout.Center.Layout(gtx, material.Label(theme, unit.Sp(14), "No items").Layout)
	}

	return material.List(theme, t.list).Layout(gtx, len(nodes), func(gtx layout.Context, index int) layout.Dimensions {
		return t.parentLayout(gtx, theme, nodes[index])
	})
}

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package xray

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
)

// Section represents an xray renderer.
type Section struct {
	render.Base
}

// Render renders an xray node.
func (s *Section) Render(ctx context.Context, ns string, o any) error {
	section, ok := o.(render.Section)
	if !ok {
		return fmt.Errorf("expected Section, but got %T", o)
	}
	root := NewTreeNode(client.NewGVR(section.GVR), section.Title)
	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}
	s.outcomeRefs(root, section)
	parent.Add(root)

	return nil
}

func (*Section) outcomeRefs(parent *TreeNode, section render.Section) {
	for k, issues := range section.Outcome {
		p := NewTreeNode(client.NewGVR(section.GVR), cleanse(k))
		parent.Add(p)
		for _, issue := range issues {
			msg := colorize(cleanse(issue.Message), issue.Level)
			c := NewTreeNode(client.NewGVR(fmt.Sprintf("issue_%d", issue.Level)), msg)
			if issue.Group == "__root__" {
				p.Add(c)
				continue
			}
			if pa := p.Find(client.NewGVR(issue.GVR), issue.Group); pa != nil {
				pa.Add(c)
				continue
			}
			pa := NewTreeNode(client.NewGVR(issue.GVR), issue.Group)
			pa.Add(c)
			p.Add(pa)
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func colorize(s string, l render.Level) string {
	c := "green"
	// nolint:exhaustive
	switch l {
	case render.ErrorLevel:
		c = "red"
	case render.WarnLevel:
		c = "orange"
	case render.InfoLevel:
		c = "blue"
	}
	return fmt.Sprintf("[%s::]%s", c, s)
}

func cleanse(s string) string {
	s = strings.ReplaceAll(s, "[", "(")
	s = strings.ReplaceAll(s, "]", ")")
	s = strings.ReplaceAll(s, "/", "::")
	return s
}

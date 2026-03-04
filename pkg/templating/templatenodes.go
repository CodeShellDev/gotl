package templating

import (
	"fmt"
	"strings"
	"text/template"
	"text/template/parse"
)

// Apply a template function to every field `{{ .VAR }}` => `{{ funcName ( .VAR ) }}`
func ApplyTemplateFunc(templt *template.Template, funcName string) {
	WalkTemplate(templt, func(node parse.Node) {
		cmd, ok := node.(*parse.CommandNode)
		if !ok {
			return
		}

		for i, arg := range cmd.Args {
			field, ok := arg.(*parse.FieldNode)
			if !ok {
				continue
			}

			cmd.Args[i] = &parse.PipeNode{
				NodeType: parse.NodePipe,
				Cmds: []*parse.CommandNode{
					{
						Args: []parse.Node{
							// add function as node
							&parse.IdentifierNode{
								NodeType: parse.NodeIdentifier,
								Ident:    funcName,
							},
							field,
						},
					},
				},
			}
		}
	})
}

// Transform template fields with transform function (example: `{{ .VAR.IABLE }}` => `{{ .var.iable }}`)
func TransformTemplateFields(templt *template.Template, transform func(fieldName string) string) {
	WalkTemplate(templt, func(node parse.Node) {
		field, ok := node.(*parse.FieldNode)

		if !ok {
			return
		}

		if len(field.Ident) == 0 {
			return
		}

		name := strings.Join(field.Ident, ".")

		fmt.Println(name)

		newName := transform(name)

		parts := strings.Split(newName, ".")

		field.Ident = parts
	})
}

// Walk template nodes and apply fn on them
func WalkTemplate(tmpl *template.Template, fn func(node parse.Node)) {
	type queueItem struct {
		node parse.Node
	}

	queue := []queueItem{}

	for _, t := range tmpl.Templates() {
		if t.Tree != nil && t.Tree.Root != nil {
			queue = append(queue, queueItem{node: t.Tree.Root})
		}
	}

	i := 0

	for len(queue) > 0 && i <= 100 {
		i++

		// get next
		item := queue[0]
		queue = queue[1:]

		if item.node == nil {
			continue
		}

		fmt.Println(item.node)

		fn(item.node)

		switch n := item.node.(type) {
		case *parse.ListNode:
			for _, child := range n.Nodes {
				queue = append(queue, queueItem{node: child})
			}

		case *parse.ActionNode:
			if n.Pipe != nil {
				queue = append(queue, queueItem{node: n.Pipe})
			}

		case *parse.TemplateNode:
			if n.Pipe != nil {
				queue = append(queue, queueItem{node: n.Pipe})
			}

		case *parse.IfNode:
			if n.Pipe != nil {
				queue = append(queue, queueItem{node: n.Pipe})
			}
			if n.List != nil {
				queue = append(queue, queueItem{node: n.List})
			}
			if n.ElseList != nil {
				queue = append(queue, queueItem{node: n.ElseList})
			}

		case *parse.RangeNode:
			if n.Pipe != nil {
				queue = append(queue, queueItem{node: n.Pipe})
			}
			if n.List != nil {
				queue = append(queue, queueItem{node: n.List})
			}
			if n.ElseList != nil {
				queue = append(queue, queueItem{node: n.ElseList})
			}

		case *parse.WithNode:
			if n.Pipe != nil {
				queue = append(queue, queueItem{node: n.Pipe})
			}
			if n.List != nil {
				queue = append(queue, queueItem{node: n.List})
			}
			if n.ElseList != nil {
				queue = append(queue, queueItem{node: n.ElseList})
			}

		case *parse.PipeNode:
			for _, decl := range n.Decl {
				queue = append(queue, queueItem{node: decl})
			}
			for _, cmd := range n.Cmds {
				queue = append(queue, queueItem{node: cmd})
			}

		case *parse.CommandNode:
			for _, arg := range n.Args {
				queue = append(queue, queueItem{node: arg})
			}

		// no children
		case
			*parse.CommentNode,
			*parse.TextNode,
			*parse.IdentifierNode,
			*parse.VariableNode,
			*parse.FieldNode,
			*parse.DotNode,
			*parse.StringNode,
			*parse.NumberNode,
			*parse.BoolNode,
			*parse.NilNode:
			continue
		}
	}
}
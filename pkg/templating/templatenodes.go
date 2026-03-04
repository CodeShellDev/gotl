package templating

import (
	"fmt"
	"strings"
	"text/template"
	"text/template/parse"
)

// Apply a template function to every field `{{ .VAR }}` => `{{ funcName ( .VAR ) }}`
func ApplyTemplateFunc(templt *template.Template, funcName string) error {
	return WalkTemplate(templt, func(node parse.Node) {
		cmd, ok := node.(*parse.CommandNode)

		if !ok {
			return
		}

		for i, arg := range cmd.Args {
			field, ok := arg.(*parse.FieldNode)

			if !ok {
				continue
			}

			cmd.Args[i] = &parse.CommandNode{
				NodeType: parse.NodeCommand,
				Args: []parse.Node{
					// add function call as node
					&parse.IdentifierNode{
						NodeType: parse.NodeIdentifier,
						Ident: funcName,
					},
					field,
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

// Recursively walk template nodes and apply fn on them
func WalkTemplate(tmpl *template.Template, fn func(node parse.Node)) error {
	for _, t := range tmpl.Templates() {
		if t.Tree != nil && t.Tree.Root != nil {
			walkNode(t.Tree.Root, fn)
		}
	}

	return nil
}

func walkNode(node parse.Node, fn func(node parse.Node)) {
	if node == nil {
		return
	}

	fn(node)

	switch n := node.(type) {
	case *parse.ListNode:
		for _, child := range n.Nodes {
			walkNode(child, fn)
		}

	case *parse.ActionNode:
		walkNode(n.Pipe, fn)

	case *parse.CommentNode:
		// no children

	case *parse.TextNode:
		// no children

	case *parse.TemplateNode:
		walkNode(n.Pipe, fn)

	case *parse.IfNode:
		walkNode(n.Pipe, fn)
		walkNode(n.List, fn)
		walkNode(n.ElseList, fn)

	case *parse.RangeNode:
		walkNode(n.Pipe, fn)
		walkNode(n.List, fn)
		walkNode(n.ElseList, fn)

	case *parse.WithNode:
		walkNode(n.Pipe, fn)
		walkNode(n.List, fn)
		walkNode(n.ElseList, fn)

	case *parse.PipeNode:
		for _, decl := range n.Decl {
			walkNode(decl, fn)
		}
		for _, cmd := range n.Cmds {
			walkNode(cmd, fn)
		}

	case *parse.CommandNode:
		for _, arg := range n.Args {
			walkNode(arg, fn)
		}

	case *parse.IdentifierNode:
		// no children

	case *parse.VariableNode:
		// no children

	case *parse.FieldNode:
		// no children

	case *parse.DotNode:
		// no children

	case *parse.StringNode:
		// no children

	case *parse.NumberNode:
		// no children

	case *parse.BoolNode:
		// no children

	case *parse.NilNode:
		// no children
	}
}
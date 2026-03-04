package templating

import (
	"fmt"
	"reflect"
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

// Recursively walk template nodes and apply fn on them
func WalkTemplate(tmpl *template.Template, fn func(node parse.Node)) error {
	visited := map[uintptr]struct{}{}

	for _, t := range tmpl.Templates() {
		if t.Tree != nil && t.Tree.Root != nil {
			walkNode(t.Tree.Root, fn, visited)
		}
	}

	return nil
}

func walkNode(node parse.Node, fn func(node parse.Node), visited map[uintptr]struct{}) {
	if node == nil {
		return
	}

	ptr := reflect.ValueOf(node).Pointer()

	_, exists := visited[ptr]
	if exists {
		return
	}

	// mark as visited
	visited[ptr] = struct{}{}

	fn(node)

	switch n := node.(type) {
	case *parse.ListNode:
		for _, child := range n.Nodes {
			walkNode(child, fn, visited)
		}

	case *parse.ActionNode:
		walkNode(n.Pipe, fn, visited)

	case *parse.TemplateNode:
		walkNode(n.Pipe, fn, visited)

	case *parse.IfNode:
		walkNode(n.Pipe, fn, visited)
		walkNode(n.List, fn, visited)
		walkNode(n.ElseList, fn, visited)

	case *parse.RangeNode:
		walkNode(n.Pipe, fn, visited)
		walkNode(n.List, fn, visited)
		walkNode(n.ElseList, fn, visited)

	case *parse.WithNode:
		walkNode(n.Pipe, fn, visited)
		walkNode(n.List, fn, visited)
		walkNode(n.ElseList, fn, visited)

	case *parse.PipeNode:
		for _, decl := range n.Decl {
			walkNode(decl, fn, visited)
		}
		for _, cmd := range n.Cmds {
			walkNode(cmd, fn, visited)
		}

	case *parse.CommandNode:
		for _, arg := range n.Args {
			walkNode(arg, fn, visited)
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
		return
	}
}
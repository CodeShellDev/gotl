package templating

import (
	"strings"
	"text/template"
	"text/template/parse"
)

// Apply a template function to every field `{{ .VAR }}` => `{{ funcName ( .VAR ) }}`
func ApplyTemplateFunc(t *template.Template, funcName string) {
	WalkTemplate(t, func(node parse.Node) bool {
		action, ok := node.(*parse.ActionNode)
		if !ok || action.Pipe == nil {
			return false
		}

		for _, cmd := range action.Pipe.Cmds {
			for i, arg := range cmd.Args {

				switch v := arg.(type) {
				case *parse.FieldNode:
					cmd.Args[i] = wrapInFunc(funcName, v)

				case *parse.ChainNode:
					cmd.Args[i] = wrapInFunc(funcName, v)
				}
			}
		}

		return true
	})
}

func wrapInFunc(funcName string, expr parse.Node) parse.Node {
	return &parse.PipeNode{
		NodeType: parse.NodePipe,
		Cmds: []*parse.CommandNode{
			{
				NodeType: parse.NodeCommand,
				Args: []parse.Node{
					&parse.IdentifierNode{
						NodeType: parse.NodeIdentifier,
						Ident: funcName,
					},
					expr,
				},
			},
		},
	}
}

// Transform template fields with transform function (example: `{{ .VAR.IABLE }}` => `{{ .var.iable }}`)
func TransformTemplateFields(templt *template.Template, transform func(fieldName string) string) {
	WalkTemplate(templt, func(node parse.Node) bool {
		field, ok := node.(*parse.FieldNode)

		if !ok {
			return false
		}

		if len(field.Ident) == 0 {
			return false
		}

		name := strings.Join(field.Ident, ".")

		newName := transform(name)

		parts := strings.Split(newName, ".")

		field.Ident = parts

		return true
	})
}

// Recursively walk template nodes and apply fn on them
func WalkTemplate(tmpl *template.Template, fn func(node parse.Node) bool) {
	for _, t := range tmpl.Templates() {
		if t.Tree != nil && t.Tree.Root != nil {
			walkNode(t.Tree.Root, fn)
		}
	}
}

func walkNode(node parse.Node, fn func(node parse.Node) bool) {
	if node == nil {
		return
	}

	if fn(node) {
		return
	}

	switch n := node.(type) {
	case *parse.ListNode:
		for _, child := range n.Nodes {
			walkNode(child, fn)
		}

	case *parse.ActionNode:
		walkNode(n.Pipe, fn)
		
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
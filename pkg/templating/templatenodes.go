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

			// wrap the FieldNode in a PipeNode calling funcName
			cmd.Args[i] = &parse.PipeNode{
				NodeType: parse.NodePipe,
				Cmds: []*parse.CommandNode{
					{
						Args: []parse.Node{
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
	visitedTemplates := make(map[string]struct{})

	return walkTemplateSafe(tmpl, fn, visitedTemplates)
}

func walkTemplateSafe(tmpl *template.Template, fn func(node parse.Node), visited map[string]struct{}) error {
	if tmpl == nil {
		return nil
	}

	_, ok := visited[tmpl.Name()]
	
	if ok {
		return nil
	}

	visited[tmpl.Name()] = struct{}{}

	if tmpl.Tree != nil && tmpl.Tree.Root != nil {
		walkNodeSafe(tmpl.Tree.Root, fn)
	}

	for _, t := range tmpl.Templates() {
		if t != tmpl {
			walkTemplateSafe(t, fn, visited)
		}
	}

	return nil
}

func walkNodeSafe(node parse.Node, fn func(node parse.Node)) {
	if node == nil {
		return
	}

	fn(node)

	switch n := node.(type) {
	case *parse.ListNode:
		// snapshot to prevent walking into newly created ones
		children := append([]parse.Node(nil), n.Nodes...)

		for _, child := range children {
			walkNodeSafe(child, fn)
		}

	case *parse.ActionNode:
		walkNodeSafe(n.Pipe, fn)

	case *parse.TemplateNode:
		walkNodeSafe(n.Pipe, fn)

	case *parse.IfNode:
		walkNodeSafe(n.Pipe, fn)
		walkNodeSafe(n.List, fn)
		walkNodeSafe(n.ElseList, fn)

	case *parse.RangeNode:
		walkNodeSafe(n.Pipe, fn)
		walkNodeSafe(n.List, fn)
		walkNodeSafe(n.ElseList, fn)

	case *parse.WithNode:
		walkNodeSafe(n.Pipe, fn)
		walkNodeSafe(n.List, fn)
		walkNodeSafe(n.ElseList, fn)

	case *parse.PipeNode:
		decls := append([]*parse.VariableNode(nil), n.Decl...)
		cmds := append([]*parse.CommandNode(nil), n.Cmds...)

		for _, decl := range decls {
			walkNodeSafe(decl, fn)
		}
		
		for _, cmd := range cmds {
			walkNodeSafe(cmd, fn)
		}

	case *parse.CommandNode:
		args := append([]parse.Node(nil), n.Args...)
		for _, arg := range args {
			walkNodeSafe(arg, fn)
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
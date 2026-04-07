package templating

import (
	"fmt"
	"strings"
	"text/template"
	"text/template/parse"
)

type Target struct {
	Parent   *parse.PipeNode
	CmdIndex int
	ArgIndex int
	Node     parse.Node
}

// Apply a template function to every field `{{ .VAR }}` => `{{ funcName ( .VAR ) }}`
func ApplyTemplateFunc(t *template.Template, funcName string) {
	WalkTemplate(t, func(node parse.Node) bool {
		action, ok := node.(*parse.ActionNode)

		fmt.Println(action)

		if !ok || action.Pipe == nil {
			return false
		}

		pipe := action.Pipe

		fmt.Println(pipe)

		// only handle simple expressions: {{ .VAR }}
		if len(pipe.Cmds) != 1 {
			return false
		}

		cmd := pipe.Cmds[0]

		// must be exactly one argument
		if len(cmd.Args) != 1 {
			return false
		}

		field, ok := cmd.Args[0].(*parse.FieldNode)

		fmt.Println(field)

		if !ok {
			return false
		}

		cmd.Args = []parse.Node{
			// add function as node
			&parse.IdentifierNode{
				NodeType: parse.NodeIdentifier,
				Ident:    funcName,
			},
			field,
		}

		fmt.Println(cmd)

		return true
	})
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
		if n == nil {
        	return
    	}

		for _, child := range n.Nodes {
			walkNode(child, fn)
		}

	case *parse.ActionNode:
		if n == nil {
        	return
    	}

		walkNode(n.Pipe, fn)
		
	case *parse.TemplateNode:
		if n == nil {
        	return
    	}
		
		walkNode(n.Pipe, fn)

	case *parse.IfNode:
		if n == nil {
        	return
    	}

		walkNode(n.Pipe, fn)
		walkNode(n.List, fn)
		walkNode(n.ElseList, fn)

	case *parse.RangeNode:
		if n == nil {
        	return
    	}

		walkNode(n.Pipe, fn)
		walkNode(n.List, fn)
		walkNode(n.ElseList, fn)

	case *parse.WithNode:
		if n == nil {
        	return
    	}

		walkNode(n.Pipe, fn)
		walkNode(n.List, fn)
		walkNode(n.ElseList, fn)

	case *parse.PipeNode:
		if n == nil {
        	return
    	}

		for _, decl := range n.Decl {
			walkNode(decl, fn)
		}
		for _, cmd := range n.Cmds {
			walkNode(cmd, fn)
		}

	case *parse.CommandNode:
		if n == nil {
        	return
    	}

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
package shellquest

import (
	"fmt"
	"strings"
)

// Node is a file or a directory in the pretend filesystem.
//
// This is not a simulation of a real filesystem and does not try to be. It has
// no permissions, no links and no sizes, because none of those are the thing
// being taught. What it does have is the shape a real one has: a tree, a
// working directory, paths that can be absolute or relative, and files with
// something inside them.
type Node struct {
	Name    string
	IsDir   bool
	Content string

	Children []*Node
	parent   *Node
}

// Dir builds a directory node.
func Dir(name string, children ...*Node) *Node {
	return &Node{Name: name, IsDir: true, Children: children}
}

// File builds a file node. A name starting with a dot is hidden from a plain
// ls, exactly as it would be on the real machine.
func File(name, content string) *Node {
	return &Node{Name: name, Content: content}
}

// NewRoot builds the root directory and wires up every parent pointer, which is
// what makes `cd ..` work.
func NewRoot(children ...*Node) *Node {
	root := Dir("", children...)
	link(root, nil)
	return root
}

func link(n *Node, parent *Node) {
	n.parent = parent
	for _, c := range n.Children {
		link(c, n)
	}
}

// Hidden reports whether a plain ls should skip this node.
func (n *Node) Hidden() bool { return strings.HasPrefix(n.Name, ".") }

// Path is the absolute path to this node.
func (n *Node) Path() string {
	if n.parent == nil {
		return "/"
	}
	var parts []string
	for cur := n; cur != nil && cur.parent != nil; cur = cur.parent {
		parts = append(parts, cur.Name)
	}
	// Collected leaf-first; reverse into path order.
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return "/" + strings.Join(parts, "/")
}

// child finds a directly contained node by name.
func (n *Node) child(name string) (*Node, bool) {
	for _, c := range n.Children {
		if c.Name == name {
			return c, true
		}
	}
	return nil, false
}

// resolve walks a path from cwd, or from root when it starts with a slash.
//
// The errors it returns are the ones a real shell gives, near enough, because
// learning to read "No such file or directory" is part of the point.
func resolve(root, cwd *Node, path string) (*Node, error) {
	if path == "" {
		return cwd, nil
	}

	cur := cwd
	if strings.HasPrefix(path, "/") {
		cur = root
	}

	for _, seg := range strings.Split(path, "/") {
		switch seg {
		case "", ".":
			continue
		case "..":
			if cur.parent != nil {
				cur = cur.parent
			}
			continue
		}

		if !cur.IsDir {
			return nil, fmt.Errorf("%s: Not a directory", cur.Name)
		}
		next, ok := cur.child(seg)
		if !ok {
			return nil, fmt.Errorf("%s: No such file or directory", path)
		}
		cur = next
	}
	return cur, nil
}

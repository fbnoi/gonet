package kernel

import (
	"fmt"
	"regexp"
	"strings"

	"fbnoi.com/goutil/collection"
)

type RouteNode struct {
	parent   *RouteNode
	children *collection.LinkedList

	root bool
	leaf bool

	path     string
	fullPath string
	pattern  string
	wildcard bool

	handler map[string]*Handler
}

// search the given path, if find, return matched node and a params,
// if not, return nill, nil.
func (routeNode *RouteNode) search(path string, params func() *Params) (*RouteNode, *Params) {
	//split given path to a slice and make a slice of collection.Iterator with the capacity
	//of the len of path slice. each path item will go through a list of entry, if match then go to
	//find in this entry's children list. if there is no match in current list, go upper to
	//find another entry and so on. if no match at all, notfound.
	l := strings.Split(path, "/")
	p := make([]collection.Iterator, len(l))
	k := routeNode
	var ps *Params
	// beautyPath always start with '/' and the begin of the path must be "".
	// search from root entry's children and skip the begin of path slice.
	// it means that root must match with the begin of path slice.
	at := 1
	it := k.children.GetIterator()
	for at >= 1 && at < len(l) {
		find := false
		for it.HasNext() {
			child := it.Next().(*RouteNode)

			// if find fit node
			if child.fit(l[at]) {
				// push Iterator to same position
				p[at] = it
				find = true
				k = child
				//save the params if necessery
				if child.wildcard {
					if params != nil {
						if ps == nil {
							ps = params()
						}
						*ps = append(*ps, Param{key: child.path, value: l[at]})
					}
				}
				// then we need to go further
				it = k.children.GetIterator()
				at++
				// but if end, we can not
				if at >= len(l) {
					break
				}
			}
		}
		// this is a blind alley, we need go back to find way out in another entry.
		if !find {
			// if the upper entry is a wildcard, the last param should be removed
			if ps != nil && len(*ps) > 0 {
				*ps = (*ps)[:len(*ps)-1]
			}
			at--
			it = p[at]
		}
	}

	// if it return to the root of the tree, or the matched node is not a leaf,
	// there is no matched node.
	if at == 0 || !k.leaf {
		return nil, nil
	} else {
		return k, ps
	}
}

// addPath add a new path to this node's children, then return the leaf node.
func (routeNode *RouteNode) addPath(method, path string, handler *Handler) *RouteNode {
	node := routeNode.pave(path)

	it := node.parent.children.GetIterator()
	for it.HasNext() {
		sbling := it.Next().(*RouteNode)
		if node != sbling && sbling.leaf && sbling.wildcard {
			panic(fmt.Sprintf("route [%s] conflict with [%s]", path, sbling.fullPath))
		}
	}
	node.setHandlers(method, handler)
	return node
}

// pave search the matched node in the tree and then return the matched node.
// if no matched node exist, find the last matched node and add a new path, then return
// the last added node.
func (routeNode *RouteNode) pave(path string) *RouteNode {
	path = strings.TrimLeft(path, "/")
	if path == "" && !routeNode.root {
		return routeNode
	}
	l := strings.Split(path, "/")
	currentnode := routeNode
	len := len(l)
	i := 0
	for i = 0; i < len; i++ {
		matched := false
		it := currentnode.children.GetIterator()
		pathcut, _, wildcard := resolvePathNode(l[i])
		for it.HasNext() {
			node := it.Next().(*RouteNode)
			if pathcut == node.path && wildcard == node.wildcard {
				matched = true
				if i >= len-1 {
					return node
				}
				currentnode = node
				break
			}
		}
		if !matched {
			break
		}
	}
	for ; i < len; i++ {
		currentnode = currentnode.addChildNode(l[i])
	}
	return currentnode
}

// addChildNode add a node to child list
// if th path is wildcard, add to end, else add to begining.
func (routeNode *RouteNode) addChildNode(path string) *RouteNode {
	node := &RouteNode{
		root:     false,
		children: &collection.LinkedList{},
		parent:   routeNode,
	}
	node.path, node.pattern, node.wildcard = resolvePathNode(path)
	var fullPath string
	if node.wildcard {
		fullPath = fmt.Sprintf("%s/:%s", routeNode.fullPath, path)
	} else {
		fullPath = fmt.Sprintf("%s/%s", routeNode.fullPath, path)
	}
	node.fullPath = fullPath
	if node.wildcard {
		routeNode.children.Add(node)
		return node
	}
	routeNode.children.AddFirst(node)
	return node
}

// fit compare the given path with the node path, if match then return true,
// else return false
func (routeNode *RouteNode) fit(path string) bool {
	if routeNode.wildcard {
		if routeNode.pattern != "" {
			fit, _ := regexp.MatchString(routeNode.pattern, path)
			return fit
		}
		return true
	}
	return routeNode.path == path
}

// getHandlers return the handler stored in the node.
// if handler exist, return (handler, true), else return (nil, false)
func (routeNode *RouteNode) getHandlers(method string) (handler *Handler, ok bool) {
	if routeNode.handler == nil {
		return nil, false
	}
	handler, ok = routeNode.handler[method]
	return
}

// setHandlers store a handler with the given method, if the handler in given method alreay exist,
// panic
func (routeNode *RouteNode) setHandlers(method string, handler *Handler) *RouteNode {
	routeNode.leaf = true
	if routeNode.handler == nil {
		routeNode.handler = make(map[string]*Handler)
	}
	if _, ok := routeNode.handler[method]; ok {
		panic(fmt.Sprintf("method [%s] on route [%s] already exist", method, routeNode.fullPath))
	}
	routeNode.handler[method] = handler
	return routeNode
}

// resolvePathNode resolve the given path string, if it is a static path,
// return (name, nil, false), if it is a wildcard, return (name, nil, true),
// if it is a wildcard with a pattern, return (name, pattern, true), if the pattern
// compiled failed, panic
func resolvePathNode(pathnode string) (string, string, bool) {
	if !strings.HasPrefix(pathnode, ":") {
		return pathnode, "", false
	}
	pathnode = strings.TrimPrefix(pathnode, ":")
	if !strings.Contains(pathnode, "(") {
		return pathnode, "", true
	}
	path := pathnode[:strings.Index(pathnode, "(")]
	rgxStr := pathnode[strings.Index(pathnode, "(")+1 : len(pathnode)-1]
	_, err := regexp.Compile(rgxStr)
	if err != nil {
		panic(fmt.Sprintf("resolve path node [%s] faild, error: %s", pathnode, err))
	}
	return path, rgxStr, true
}

package gee

import (
	"log"
	"net/http"
	"strings"
)

type router struct {
	handlers map[string] HandlerFunc

	// add
	roots    map[string]*node
}

func newRouter() *router {
	return &router{
		handlers: make(map[string]HandlerFunc),

		// add
		roots: make(map[string]*node),
	}
}

// Only one * is allowed
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			// 添加一个后, 在判断
			if item[0] == '*' { // 如果是 *xxx开头的就退出
				break
			}
		}
	}
	return parts
}


// 添加router
func (r *router) addRouter(method string, pattern string, handlerFunc HandlerFunc)  {

	log.Printf("Route %4s - %s", method, pattern)

	// 添加的时候就拆解
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	// 方法GET/POST
	_, ok := r.roots[method]
	if !ok { // 如果没有 初始化一个
		r.roots[method] = &node{}
	}
	// 插入
	r.roots[method].insert(pattern, parts, 0)

	r.handlers[key] = handlerFunc

	//key := method + "-" + pattern
	//r.handlers[key] = handlerFunc
}

// 获取routes
func (r *router) getRouter(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil
	}
	//log.Println(searchParts)
	n := root.search(searchParts, 0)

	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}



// 处理router
func (r *router) handler(c *Context){
	//key := c.Method + "-" + c.Path
	//if handle, ok := r.handlers[key]; ok {
	//	handle(c)
	//} else {
	//	c.String(http.StatusNotFound, "404 NOT FOUNDD %q", c.Path)
	//}
	//
	n, params := r.getRouter(c.Method, c.Path)
	//log.Println(n, params, c.Method, c.Path)
	if n != nil {

		c.Params = params
		key := c.Method + "-" + n.pattern

		// 中间件, 加入到队列后面
		c.handlers = append(c.handlers, r.handlers[key])


		r.handlers[key](c)
	} else {
		// 也加到后面
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})

		// c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}

	// 调用方法循环执行, index默认是-1
	c.Next()
}

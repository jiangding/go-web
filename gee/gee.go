package gee

import (
	"log"
	"net/http"
	path "path/filepath"
	"strings"
	"text/template"
)

type HandlerFunc func(c *Context)

type RouterGroup struct {
	prefix string
	middlewares []HandlerFunc // support middleware
	parent *RouterGroup  // 支持嵌套
	engine *Engine       // 所有的组共享一个engine实例
}


type Engine struct {
	//router map[string]HandlerFunc
	router *router

	//将Engine作为最顶层的分组，也就是说Engine拥有RouterGroup所有的能力。
	*RouterGroup

	// 一个 engine存了所有的router group
	groups []*RouterGroup // store all groups


	htmlTemplates *template.Template // for html render
	funcMap       template.FuncMap   // for html render

}

func New() *Engine {

	engine := &Engine{router: newRouter()}
	// engine 里面有个互相嵌套的group
	engine.RouterGroup = &RouterGroup{engine: engine}

	//
	engine.groups = []*RouterGroup{engine.RouterGroup}

	return engine

	//return &Engine{
	//	router:newRouter(),
	//
	//}
}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)

	return newGroup
}

func (engine *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, engine)
}

// SetFuncMap html
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}
// LoadHTMLGlob html
func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}



func (group *RouterGroup) addRouter(method string, comp string, handleFunc HandlerFunc) {
	pattern := group.prefix + comp
	// log.Printf("Route %4s - %s", method, pattern)
	// 调用了group.engine.router.addRoute来实现了路由的映射
	group.engine.router.addRouter(method, pattern, handleFunc)
}

//func (engine *Engine) addRouter(method string, pattern string, handlerFunc HandlerFunc)  {
//	//key := method + "-" + pattern
//	//engine.router.handlers[key] = handlerFunc
//	engine.router.addRouter(method, pattern, handlerFunc)
//}

func (group *RouterGroup) GET(pattern string, handleFunc HandlerFunc) {
	group.addRouter("GET", pattern, handleFunc)
}

func (group *RouterGroup) POST(pattern string, handleFunc HandlerFunc) {
	group.addRouter("POST", pattern, handleFunc)
}

//func (engine *Engine) GET(pattern string, handlerFunc HandlerFunc)  {
//	engine.addRouter("GET", pattern, handlerFunc)
//}
//func (engine *Engine) POST(pattern string, handlerFunc HandlerFunc)  {
//	engine.addRouter("POST", pattern, handlerFunc)
//}

// 接口实现方法才能处理请求
func (engine *Engine) ServeHTTP( w http.ResponseWriter, req *http.Request) {

	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		// log.Printf("current group: %#v",group)
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	// log.Printf("current c mid: %#v",middlewares)

	c := NewContext(w, req)

	// 把匹配到的中间件添加到当前context中
	c.handlers = middlewares

	// template
	c.engine = engine

	engine.router.handler(c)
}

// Use is defined to add middleware to the group
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}


// 模板 static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Params["filepath"]
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// Static  serve static files 暴露给用户的。用户可以将磁盘上的某个文件夹root映射到路由relativePath
func (group *RouterGroup) Static(relativePath string, root string){
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")

	log.Println("lallal",urlPattern)
	// Register GET handlers
	group.GET(urlPattern, handler)
}

//r := gee.New()
//r.Static("/assets", "/usr/geektutu/blog/static")
//或相对路径 r.Static("/assets", "./static")



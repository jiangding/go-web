package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	Writer http.ResponseWriter
	Req *http.Request

	// 请求参数
	Method string
	Path string

	StatusCode int


	Params map[string]string


	// middleware
	handlers []HandlerFunc
	index    int // index是记录当前执行到第几个中间件


	// template
	engine *Engine  // engine pointer

}

func NewContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req: req,
		Method:req.Method,
		Path: req.URL.Path,

		index: -1,
	}
}

func (c *Context) Next() {
	// 当在中间件中调用Next方法时，控制权交给了下一个中间件，直到调用到最后一个中间件，
	// 然后再从后往前，调用每个中间件在Next方法之后定义的部分
	c.index++ // index默认是-1, 一调用Next()就从0开始执行，直到完毕
	s := len(c.handlers) // handlerFunc数量
	for ; c.index < s; c.index++ { // 循环调用
		c.handlers[c.index](c)
	}
}
func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

func (c *Context) PostForm(key string) string {
	// FormValue返回key为键查询r.Form字段得到结果[]string切片的第一个值
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string){
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)

	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, name string, data interface{})  {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)

	//c.Writer.Write([]byte(html))
	// template
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}


//
func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}





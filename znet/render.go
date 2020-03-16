/*
 * @Author: seekwe
 * @Date:   2019-05-23 19:16:32
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-29 15:40:30
 */

package znet

import (
	"encoding/json"
	"fmt"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
	"html/template"
	"io"
	"net/http"
)

type (
	render interface {
		Render(*Context, int) error
		Content() (content string)
	}
	renderByte struct {
		Data []byte
	}
	renderString struct {
		Format string
		Data   []interface{}
	}
	renderJSON struct {
		Data interface{}
	}
	renderFile struct {
		Status interface{}
		Data   interface{}
	}
	renderHTML struct {
		Template *template.Template
		Name     string
		Data     interface{}
	}
	Api struct {
		Code int         `json:"code" example:"200"`
		Msg  string      `json:"msg"`
		Data interface{} `json:"data"`
	}
	Data map[string]interface{}
)

var (
	plainContentType = "text/plain; charset=utf-8"
	htmlContentType  = "text/html; charset=utf-8"
	jsonContentType  = "application/json; charset=utf-8"
)

func (c *Context) render(code int, r render) {
	c.Info.Mutex.Lock()
	c.Info.Code = code
	c.Info.render = r
	c.Info.StopHandle = true
	c.Info.Mutex.Unlock()
}

func (r *renderByte) Content() (content string) {
	return zstring.Bytes2String(r.Data)
}

func (r *renderByte) Render(c *Context, code int) (err error) {
	w := c.Writer
	c.SetStatus(code)
	c.SetHeader("Content-Type", plainContentType)
	_, err = w.Write(r.Data)
	return
}

func (r *renderString) Content() (content string) {
	if len(r.Data) > 0 {
		content = fmt.Sprintf(r.Format, r.Data...)
	} else {
		content = r.Format
	}
	return
}

func (r *renderString) Render(c *Context, code int) (err error) {
	w := c.Writer
	c.SetStatus(code)
	c.SetHeader("Content-Type", plainContentType)
	if len(r.Data) > 0 {
		_, err = fmt.Fprintf(w, r.Format, r.Data...)
	} else {
		_, err = io.WriteString(w, r.Format)
	}
	return
}

func (r *renderJSON) Content() (content string) {
	j, err := json.Marshal(r.Data)
	if err != nil {
		return
	}
	content = zstring.Bytes2String(j)
	return
}

func (r *renderFile) Content() (content string) {
	content = r.Data.(string)
	return
}

func (r *renderFile) Render(c *Context, _ int) error {
	http.ServeFile(c.Writer, c.Request, r.Data.(string))
	return nil
}

func (r *renderJSON) Render(c *Context, code int) error {
	w := c.Writer
	c.SetStatus(code)
	c.SetHeader("Content-Type", jsonContentType)
	jsonBytes, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	_, _ = w.Write(jsonBytes)
	return nil
}

func (r *renderHTML) Content() (content string) {
	if r.Name != "" {
		var t *template.Template
		var err error
		t, err = templateParse(r.Name)
		if err != nil {
			return
		}
		content = t.Name()
	} else {
		content = fmt.Sprint(r.Data)
	}
	return
}

func (r *renderHTML) Render(c *Context, code int) (err error) {
	w := c.Writer
	c.SetStatus(code)
	c.SetHeader("Content-Type", htmlContentType)
	if r.Name != "" {
		var t *template.Template
		t, err = templateParse(r.Name)
		if err != nil {
			return
		}
		err = t.Execute(w, r.Data)
	} else {
		_, err = fmt.Fprint(c.Writer, r.Data)
	}
	return
}

func (c *Context) Byte(code int, value []byte) {
	c.render(code, &renderByte{Data: value})
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.render(code, &renderString{Format: format, Data: values})
}

func (c *Context) File(filepath string) {
	c.render(zutil.IfVal(zfile.FileExist(filepath), 200, 404).(int), &renderFile{Data: filepath})
}

func (c *Context) JSON(code int, values interface{}) {
	c.render(code, &renderJSON{Data: values})
}

// ResJSON ResJSON
func (c *Context) ResJSON(code int, msg string, data interface{}) {
	httpState := code
	if code < 300 && code >= 200 {
		httpState = http.StatusOK
	}
	c.render(httpState, &renderJSON{Data: Api{Code: code, Data: data, Msg: msg}})
}

func (c *Context) HTML(code int, html string) {
	c.render(code, &renderHTML{
		Name: "",
		Data: html,
	})
}

func (c *Context) Template(code int, name string, data ...interface{}) {
	var _data interface{}
	if len(data) > 0 {
		_data = data[0]
	}
	c.render(code, &renderHTML{
		Name: name,
		Data: _data,
	})
}

// Abort Abort
func (c *Context) Abort(code ...int) {
	c.Info.Mutex.Lock()
	c.Info.StopHandle = true
	c.Info.Mutex.Unlock()
	if len(code) > 0 {
		c.SetStatus(code[0])
	}
}

// Redirect Redirect
func (c *Context) Redirect(link string, statusCode ...int) {
	c.Writer.Header().Set("Location", c.CompletionLink(link))
	code := http.StatusFound
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	c.SetStatus(code)
}

func (c *Context) SetStatus(code int) *Context {
	c.Info.Mutex.Lock()
	c.Info.Code = code
	c.Info.Mutex.Unlock()
	return c
}

func (c *Context) PrevStatus() (code int) {
	c.Info.Mutex.Lock()
	code = c.Info.Code
	c.Info.Mutex.Unlock()
	return
}

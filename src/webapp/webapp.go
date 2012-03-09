package webapp
import (
    "net"
    "net/http"
    "net/http/fcgi"
    "strconv"
    "log"
    "html/template"
    "fmt"
    "regexp"
    "compress/gzip"
    "strings"
	"time"
    )

const (
    ErrNotFound         = "Page Not Found"
    ErrAccessDenied     = "Access Denied"
    ErrInternalServerError = "Internal Server Error"
    )


type AbsApp interface {
    Run()
    AddHandleFunc(string, func(http.ResponseWriter, *http.Request))
    SetStaticPath(string, string)
}

type App struct {
    Port int;
    StaticPath string;
    Handler RootHandler;
}

type ContextInfo struct {
    StartTime time.Time
    During int64
    URL string
    HttpCode int
    Message string
    UseGZip bool
}

type Context struct {
    Writer http.ResponseWriter
    Request *http.Request
    Application *App
    Info ContextInfo
	Headers map[string]string
}

var EmailPattern, URLPattern *regexp.Regexp

type RenderHandler func (*App, map[string]string) (string, error)

type RootHandler struct {
    HandleFunc func(ctx *Context)
    Application *App
}

func (h RootHandler) ServeHTTP (w http.ResponseWriter, req * http.Request) {
    c := new(Context);
    c.Writer = w
    c.Request = req
    c.Application = h.Application
    h.HandleFunc(c)
}

func init() {
    EmailPattern = regexp.MustCompile("^[0-9a-zA-Z_.\\-+]+@[0-9a-zA-Z_.]+\\.[0-9a-zA-Z_.]+$")
    URLPattern = regexp.MustCompile("^[0-9a-zA-Z]+://.*$")
}

func (ctx * Context) Redirect(url string, code int) {
    ctx.Info.Message = "Redirect"
    ctx.Info.HttpCode = http.StatusFound
    ctx.Application.AccessLog(ctx)
    http.Redirect(ctx.Writer, ctx.Request, url, code)
}

func (ctx * Context) Error(msg string, code int) {
    ctx.Info.Message = msg
    ctx.Info.HttpCode = code
    ctx.Application.ErrorLog(ctx)
    http.Error(ctx.Writer, msg, code)
}

func (ctx * Context) SetHeader(key string, val string) {
	if ctx.Headers == nil {
		ctx.Headers = make(map[string]string)
	}
	ctx.Headers[key] = val;
}

func (ctx * Context) Execute(tpl * template.Template, data interface{}) error {
	var err error
    ctx.Info.Message = "OK"
    ctx.Info.HttpCode = 200
    ctx.Application.AccessLog(ctx)
    ctx.Writer.Header().Set("Content-Type", "text/html")
    ctx.Writer.Header().Set("Connection", "keep-alive")
    ctx.Writer.Header().Set("Cache-Control", "must-revalidate, max-age=300")

	// overlay headers
	if ctx.Headers != nil {
		for k, v := range ctx.Headers {
			ctx.Writer.Header().Set(k, v)
		}
	}

	// compress ?
    if ctx.Info.UseGZip {
		gw := gzip.NewWriter(ctx.Writer)
        ctx.Writer.Header().Set("Content-Encoding", "gzip")
        err = tpl.Execute(gw, data)
        gw.Close()
    } else {
        err = tpl.Execute(ctx.Writer, data)
    }
	return err
}

func (app *App) Run(port int) {
    app.Port = port
    err := http.ListenAndServe(":" + strconv.Itoa(app.Port), nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

func (app *App) RunCGI(port int) {
    app.Port = port
    l, err := net.Listen("tcp", "127.0.0.1:" + strconv.Itoa(port))
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
    fcgi.Serve(l, app.Handler)
}

func (app *App) SetHandler(url string, handleFunc func(*Context)) {
    app.Handler.HandleFunc = handleFunc
    app.Handler.Application = app
    http.Handle(url, app.Handler)
    return
}

func (app *App) SetStaticPath(url string, path string) {
    app.StaticPath = path;
    http.Handle(url, http.StripPrefix(url[:len(url)-1], http.FileServer(http.Dir(path))))
    return
}

func (app *App) ErrorLog(ctx *Context) {
    log.Printf("[ERR] host: '%s', request: '%s %s', proto: '%s', ua: '%s', code: %d, remote: '%s', message: '%s'\n", ctx.Request.Host, ctx.Request.Method, ctx.Request.URL.Path, ctx.Request.Proto, ctx.Request.UserAgent(), ctx.Info.HttpCode, ctx.Request.RemoteAddr, ctx.Info.Message)
}

func (app *App) AccessLog(ctx *Context) {
    log.Printf("[ACC] host: '%s', request: '%s %s', proto: '%s', ua: %s'', code: %d, remote: '%s', message: '%s'\n", ctx.Request.Host, ctx.Request.Method, ctx.Request.URL.Path, ctx.Request.Proto, ctx.Request.UserAgent(), ctx.Info.HttpCode, ctx.Request.RemoteAddr, ctx.Info.Message)
}

func (app *App) Log(tag string, msg string) {
    log.Printf("[%s] %s\n", tag, msg)
}

func CheckForm(re *regexp.Regexp, in string) bool {
    return re.Match([]byte(in))
}
func CheckEmailForm(in string) bool {
    return EmailPattern.Match([]byte(in))
}
func TransformTags(in string) string {
    in = strings.Replace(in, "<", "&lt;", -1)
    in = strings.Replace(in, ">", "&gt;", -1)
    return in
}
func CheckURLForm(in string) bool {
    return URLPattern.Match([]byte(in))
}




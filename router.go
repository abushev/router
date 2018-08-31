package router

import (
	"fmt"
	"net/http"
	"strings"
)

type HandlerFunc func(http.ResponseWriter, *http.Request, []string)
type ActionStruct struct {
	Fn HandlerFunc
}

type Router struct {
	Level   int
	Offset  int
	Routes  map[string]ActionStruct
	Dstatus *Dstatus
}

func (rtr *Router) Handler(w http.ResponseWriter, r *http.Request) {
	uriParts := strings.Split(r.URL.Path[1:], "/")
	if rtr.Level < len(uriParts) {
		if handle, ok := rtr.Routes[uriParts[rtr.Level]]; ok {
			if rtr.Dstatus != nil {
				rtr.Dstatus.HandleWrapper(handle.Fn)(w, r, uriParts[rtr.Offset:])
			} else {
				handle.Fn(w, r, uriParts[rtr.Offset:])
			}
			return
		} else if uriParts[rtr.Level] == "dstatus" && rtr.Dstatus != nil {
			rtr.Dstatus.HandleWrapper(rtr.Dstatus.Show)(w, r, uriParts[rtr.Offset:])
		}

	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

func (rtr *Router) SubHandler(w http.ResponseWriter, r *http.Request, uriParts []string) {
	if rtr.Level < len(uriParts) {
		if handle, ok := rtr.Routes[uriParts[rtr.Level]]; ok {
			handle.Fn(w, r, uriParts)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "router.Handler uri parts: %+v", uriParts)
}

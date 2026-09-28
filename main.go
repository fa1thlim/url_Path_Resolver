package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// route — один маршрут: регулярка + обработчик
type route struct {
	re      *regexp.Regexp
	handler http.HandlerFunc
}

// regexResolver — наш роутер
type regexResolver struct {
	routes []route // ВАЖНО: slice, а не map — порядок сохраняется
}

func newPathResolver() *regexResolver {
	return &regexResolver{}
}

func main() {
	rr := newPathResolver()

	// ВАЖНО: специфичные маршруты регистрируем РАНЬШЕ общих
	rr.Add(`GET /hello`, helloHandler)
	rr.Add(`(GET|HEAD) /goodbye(/?[A-Za-z0-9]*)?`, goodbyeHandler)

	if err := http.ListenAndServe(":8080", rr); err != nil {
		panic(err)
	}
}

// Add — регистрирует маршрут. Паникует сразу, если regexp кривой.
func (r *regexResolver) Add(pattern string, handler http.HandlerFunc) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		// лучше упасть сразу при старте, чем потом в рантайме
		panic(fmt.Sprintf("bad route pattern %q: %v", pattern, err))
	}
	r.routes = append(r.routes, route{re: re, handler: handler})
}

// ServeHTTP — вызывается на каждый запрос
func (r *regexResolver) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Собираем строку вида "GET /hello" — её и будем матчить
	check := req.Method + " " + req.URL.Path

	// Идём по маршрутам ПО ПОРЯДКУ (slice)
	for _, rt := range r.routes {
		if rt.re.MatchString(check) {
			rt.handler(w, req)
			return
		}
	}

	// Ничего не совпало
	http.NotFound(w, req)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	value := r.URL.Query().Get("value")
	if name == "" {
		name = "World"
	}
	if value == "" {
		value = "default"
	}
	fmt.Fprintf(w, "Hello, %s! It's your value: %v", name, value)
}

func goodbyeHandler(w http.ResponseWriter, r *http.Request) {
	// путь /goodbye/John → делим по "/" → ["", "goodbye", "John"]
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	name := ""
	if len(parts) >= 2 {
		name = parts[len(parts)-1]
	}
	if name == "" {
		name = "World"
	}
	fmt.Fprintf(w, "Goodbye, %s!", name)
}

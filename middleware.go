package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/gorilla/context"
	"golang.org/x/text/language"
)

type middleware func(http.HandlerFunc) http.HandlerFunc

var matcher = language.NewMatcher([]language.Tag{
	language.English,
	language.AmericanEnglish,
	language.Spanish,
})

// https://gist.github.com/arxdsilva/7392013cbba7a7090cbcd120b7f5ca31
func inSet(a, b []string) []string {
	for i := len(a) - 1; i >= 0; i-- {
		for _, vD := range b {
			if a[i] == vD {
				a = append(a[:i], a[i+1:]...)
				break
			}
		}
	}
	return a
}

func langCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang, _ := r.Cookie("lang")
		accept := r.Header.Get("Accept-Language")
		tag, _ := language.MatchStrings(matcher, lang.String(), accept)
		if tag.String() == "en-US" {
			context.Set(r, "lang", "en")
		} else {
			context.Set(r, "lang", tag.String())
		}
		next.ServeHTTP(w, r)
	})
}

type includeCheck struct {
	validParams []string
}

func (icw *includeCheck) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key := vars["key"]
		v := r.URL.Query()
		include := v.Get("include")
		if key == "" && include != "" {
			http.Error(w, "HTTP 400: include not supported on this endpoint", 400)
			return
		}
		if include != "" {
			includeParams := strings.Split(include, ",")
			invalidElements := inSet(includeParams, icw.validParams)
			if len(invalidElements) > 0 {
				var invalids strings.Builder
				for _, elem := range invalidElements {
					invalids.WriteString(elem)
					invalids.WriteString(" ")
				}
				http.Error(w, fmt.Sprintf("HTTP 400: invalid includes parameters: %s", invalids.String()), 400)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

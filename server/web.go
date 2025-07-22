package server

import (
	"embed"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/thoas/stats"
)

// ApiRequestObserver measures the time it takes to process a request
var ApiRequestObserver = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Help: "Observation of http request duration",
		Name: "api_http_requests",
	}, []string{"method", "endpoint", "code"})

func init() {
	prometheus.MustRegister(ApiRequestObserver)
}

func Start(logger *log.Entry, host string, root embed.FS) error {

	m := stats.New()
	router := chi.NewMux()
	// router.Use(mm.ApiRequestInstrumentationHandler)
	router.Use(m.Handler)

	// Only debug when play mode is activated
	/*if c.Play {
		// Trace path
		router.Use(debug)

		// Mount pprof
		// router.Mount("/debug", middleware.Profiler())
	}*/

	router.Method(http.MethodGet, "/metrics", promhttp.Handler())
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "404 - "+AppName+" v1")
		fmt.Fprintf(w, "Time: %s\n", time.Now().String())
		w.Write([]byte(r.URL.Path))
	})

	// Set Web
	Web(root, router)

	router.Route("/"+BASE, func(sub chi.Router) {
		sub.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", HeaderText)
			fmt.Fprintf(w, "%s v1", AppName)
		})

		//Route(r, sub)

	})

	//  Introduce Version
	logger.Infof("HTTP - start  %s", host)
	return http.ListenAndServe(host, router)
}

func debug(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func Web(root embed.FS, sub chi.Router) {
	/*
	   	// Starting the file server
	   		workDir, _ := os.Getwd()
	   		filesDir := filepath.Join(workDir, "/",  h.Assets)//path to executable
	       	fileServer(h, sub, "/"+lvsvc.BASE+"/web", http.Dir(filesDir))
	*/
	// Starting the embedded file server
	fileServer(sub, "/"+BASE+"/web", http.FS(root))
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(router chi.Router, path string, root http.FileSystem) {

	if strings.ContainsAny(path, "{}*") {
		//http.Error(w, "Forbidden... FileServer does not permit URL parameters.", http.StatusForbidden)
		router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Forbidden... Invalid parameters", http.StatusForbidden)
		})
		return
	}

	assetsFs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		router.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	router.Get(path, func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println("url____: " + r.RequestURI + ", path = " + path)
		// draft.Dbg(r.RequestURI)
		// Check for pages that do not need authentication
		if !isPermittedRoute(r.RequestURI) { // block directories/files from public access using isPermittedRoute(uri)
			//http.Error(w, "<b>Please login to access this resource</b>", http.StatusForbidden)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("<h1><b style='color: red'>Please login to access this resource</b></h1>"))
			return
		}

		assetsFs.ServeHTTP(w, r)
	})
}

func isPermittedRoute(url string) bool {

	exclude := []string{
		//fmt.Sprintf("/%s/login",  lvsvc.BASE),
		//fmt.Sprintf("/%s/send",   lvsvc.BASE),
		//fmt.Sprintf("/%s/logout", lvsvc.BASE),
		//http://localhost:8084/paytfare/web/intro.html
		//fmt.Sprintf("/%s/signup.html", lvsvc.BASE+"/web"),
		fmt.Sprintf("/%s/", BASE+"/auth"),
		fmt.Sprintf("/%s/", BASE+"/web/html/tmpl"),
	}
	// Remove Certain Pages
	for _, v := range exclude {
		if strings.HasPrefix(url, v) {
			return false
		}
	}
	return true
}

/*
func Route(h *resource.Resource, sub chi.Router) {

	sub.Route("/auth", func(r chi.Router) {

		// Use Authentication
		r.Use(Auth(h))


		r.Post(lvsvc.ApiAuthFindHymn, operations.FindHymnBook(h))
		r.Get(lvsvc.ApiAuthListResources, operations.GetList(h))

		////////////Wrong Http Method

		r.Get(lvsvc.ApiAuthFindHymn, helper.UsePOST(h))
		r.Post(lvsvc.ApiAuthListResources, helper.UseGET(h))


	})

}*/

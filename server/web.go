package server

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"com.itis.apps/gotermchat/cmd"
	"github.com/apex/log"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
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

func Start(server *Server, logger *log.Entry, host string, root embed.FS) error {

	var onConnect = func(server *Server, ws *websocket.Conn, req *http.Request, response http.ResponseWriter) {
		data := req.FormValue("data")

		k, err := cmd.NewKryptik(ExchangeKeysSecret, cmd.ModeCBC) //base64.RawURLEncoding.DecodeString(base64Str)
		if err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte("...Error loading password decryptor!"))
			ws.Close()
			return
		}

		jsonData, err := k.Decrypt(data)
		if err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte("...Error decrypting credentials!..."+fmt.Errorf("...err: %v\n", err).Error()))
			ws.Close()
			return
		}

		var config ClientConfig
		err = cmd.DecodeItem(jsonData, &config)

		if err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte("...Error decoding client credentials"))
			ws.Close()
			return
		}

		pwd, err := k.Decrypt(config.Password) //from client terminal app
		if err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte("...Error decrypting password!..."+fmt.Errorf("...err: %v\n", err).Error()))
			ws.Close()
			return
		}

		if len(pwd) < 6 {
			ws.WriteMessage(websocket.TextMessage, []byte("...Decrypted Password too short!"))
			ws.Close()
			return
		}

		defer func() {
			err := ws.Close()
			if err != nil {
				server.ErrCh <- err
			}
		}()

		var u *cmd.User
		/**
		Register this person
		*/
		if config.Reg {

			if len(strings.Trim(config.Phone, " ")) < 7 {
				ws.WriteMessage(websocket.TextMessage, []byte("...Registration Failed. Bad phone."))
				ws.Close()
				return
			}
			if len(strings.Trim(config.Username, " ")) < 3 {
				ws.WriteMessage(websocket.TextMessage, []byte("...Registration Failed. Username too short."))
				ws.Close()
				return
			}

			user := new(cmd.User)
			user.ID = cmd.GenUlid()
			user.Phone = config.Phone
			user.Name = config.Username
			user.Password = config.Password
			user.RegTime = time.Now()

			server.GetUserManager().CreateOrUpdateUser(*user)
			ws.WriteMessage(websocket.TextMessage, []byte("...Connected!"))
			u = user

		} else {

			if len(strings.Trim(config.Phone, " ")) >= 7 {
				user, err := server.GetUserManager().ShowUser(config.Phone)
				if err != nil {
					ws.WriteMessage(websocket.TextMessage, []byte("...Login Failed. Bad credentials."))
					ws.Close()
					return
				}

				pswd, err := k.Decrypt(user.Password)
				if err != nil {
					ws.WriteMessage(websocket.TextMessage, []byte("...Error decrypting password from db!..."+fmt.Errorf("...err: %v\n", err).Error()))
					ws.Close()
					return
				}

				if pswd != pwd { //do passwords match?
					ws.WriteMessage(websocket.TextMessage, []byte("...Login Failed. Incorrect credentials."))
					ws.Close()
					return
				}
				//valid user---allow login via phone
				var resp cmd.LoginResponse = cmd.LoginResponse{
					User:    user,
					Message: "...Login successful!! via phone",
				}
				if rspJsn, err := cmd.EncodeStruct(resp); err == nil {
					ws.WriteMessage(websocket.TextMessage, []byte(rspJsn))
				} else {
					ws.WriteMessage(websocket.TextMessage, []byte("...login success, but error occurred"))
				}

				u = &user

			} else if len(strings.Trim(config.Username, " ")) >= 3 { //usernames should be at least 3 characters long
				user, err := server.GetUserManager().ShowUserByUserName(config.Username)
				if err != nil {
					ws.WriteMessage(websocket.TextMessage, []byte("...Login Failed!! Bad credentials."))
					ws.Close()
					return
				}
				pswd, err := k.Decrypt(user.Password)
				if err != nil {
					ws.WriteMessage(websocket.TextMessage, []byte("...Error decrypting password from db!..."+fmt.Errorf("...err: %v\n", err).Error()))
					ws.Close()
					return
				}
				if pswd != pwd {
					ws.WriteMessage(websocket.TextMessage, []byte(".................Login Failed!! Incorrect  credentials."))
					ws.Close()
					return
				}
				//valid user---allow login via username
				var resp cmd.LoginResponse = cmd.LoginResponse{
					User:    user,
					Message: "...Login successful!! via username",
				}
				if rspJsn, err := cmd.EncodeStruct(resp); err == nil {
					ws.WriteMessage(websocket.TextMessage, []byte(rspJsn))
				} else {
					ws.WriteMessage(websocket.TextMessage, []byte("...login success, but error occurred"))
				}

				u = &user
			}
		}

		client := NewClient(u, ws, server)
		client.Conn = ws
		client.MsgChan = make(chan *Message, ChannelBufSize)

		server.Add(client)
		client.Listen()
	}

	var rootHandler = func(w http.ResponseWriter, r *http.Request) {
		content, err := os.ReadFile("index.html")
		if err != nil {
			fmt.Println("Could not open file.", err)
		}
		fmt.Fprintf(w, "%s", content)
	}

	var wsHandler = func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("host: ", r.Host, "wsHandler called")
		/*log.Print(r.Host)
		log.Println("----------------------")
		log.Println(r.Header.Get("Origin"))
		if r.Header.Get("Origin") != "http://"+r.Host {
			http.Error(w, "Origin not allowed", 403)
			return
		}*/

		var upgrader = websocket.Upgrader{
			ReadBufferSize:    1024,
			WriteBufferSize:   1024,
			EnableCompression: true,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			msg := fmt.Sprintf("Could not open websocket connection: %v", err)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		socketTimeOut := 30 * time.Minute
		err = conn.SetReadDeadline(time.Now().Add(socketTimeOut))
		if err != nil {
			msg := fmt.Sprintf("Could not set read deadline on socket %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		err = conn.SetWriteDeadline(time.Now().Add(socketTimeOut))
		if err != nil {
			msg := fmt.Sprintf("Could not set write deadline on socket %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		// websocket handler
		go onConnect(server, conn, r, w)

	}

	fmt.Println("url: ", host)
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

	router.HandleFunc("/ws/imaxine-that", wsHandler)
	router.HandleFunc("/", rootHandler)

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

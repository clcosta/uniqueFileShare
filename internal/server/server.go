package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/clcosta/uniqueShareFile/pkg/background"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type Server struct {
	Logger *slog.Logger
	Addr   string
}

func NewServer(addr string, logger *slog.Logger) *Server {
	return &Server{
		Logger: logger,
		Addr:   addr,
	}
}

func (s *Server) RunServer(db *gorm.DB, worker *background.Worker) {

	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler((http.StripPrefix("/static/", http.FileServer(http.Dir("./public/")))))
	router.HandleFunc("/expired", LoggerMiddleware(s.Logger, expiredPageHandler)).Methods("GET")
	router.HandleFunc("/success/{link}", LoggerMiddleware(s.Logger, successUploadPageHandler(db))).Methods("GET")
	router.HandleFunc("/upload", LoggerMiddleware(s.Logger, uploadFormHandler(db, worker))).Methods("POST")
	router.HandleFunc("/download/{link}", LoggerMiddleware(s.Logger, downloadPageHandler)).Methods("GET")
	router.HandleFunc("/d/{link}", LoggerMiddleware(s.Logger, downloadFileHandler(db, worker))).Methods("GET")
	router.HandleFunc("/stats/{link}", LoggerMiddleware(s.Logger, statsHandler(db))).Methods("GET")
	router.HandleFunc("/", LoggerMiddleware(s.Logger, homePageHandler)).Methods("GET")

	s.Logger.Info(fmt.Sprintf("server started on %s", s.Addr))
	if err := http.ListenAndServe(s.Addr, router); err != nil {
		s.Logger.Error("failed to start http server", "error", err)
	}
}

package api

import (
	"comments/pkg/storage"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

// API handles HTTP requests and routes.
type API struct {
	r  *mux.Router
	db storage.Storage
}

// New creates and initializes a new API instance.
func New(db storage.Storage) *API {
	api := API{}
	api.r = mux.NewRouter()
	api.db = db
	api.endpoints()
	return &api
}

// Router returns the request router.
func (api *API) Router() *mux.Router {
	return api.r
}

// Registration of API methods in the request router.
func (api *API) endpoints() {
	api.r.Use(api.loggingMiddleWare)
	api.r.HandleFunc("/comments/{n}", api.commentsByNewsHandler).Methods(http.MethodGet)
	api.r.HandleFunc("/comments", api.addCommentHandler).Methods(http.MethodPost)
}

// commentsByNewsHandler - returns the comments by news id.
func (api *API) commentsByNewsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	newsID := mux.Vars(r)["n"]
	comments, err := api.db.CommentsByNews(r.Context(), newsID)
	if err != nil {
		slog.Error("commentsByNewsHandler: failed to get comments", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(toDTOs(comments))
	if err != nil {
		slog.Error("commentsByNewsHandler: failed to encode JSON", "err", err)
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}

// addCommentHandler - creates a new comment.
func (api *API) addCommentHandler(w http.ResponseWriter, r *http.Request) {
	var c storage.Comment
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		slog.Error("addCommentHandler: failed to decode JSON", "err", err)
		http.Error(w, "failed to decode response", http.StatusBadRequest)
		return
	}

	comment, err := api.db.AddComment(r.Context(), c)
	if err != nil {
		slog.Error("addCommentHandler: failed to add comment", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(comment)
	if err != nil {
		slog.Error("addCommentHandler: failed to encode JSON", "err", err)
		http.Error(w, "failed to encode responce", http.StatusBadRequest)
		return
	}
}

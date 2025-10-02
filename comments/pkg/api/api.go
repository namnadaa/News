package api

import (
	"comments/pkg/storage"
	"encoding/json"
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
	api.r.HandleFunc("/comments/{n}", api.commentsByNewsHandler).Methods(http.MethodGet)
	api.r.HandleFunc("/comments", api.addCommentHandler).Methods(http.MethodPost)
}

// commentsByNewsHandler - returns the comments by news id.
func (api *API) commentsByNewsHandler(w http.ResponseWriter, r *http.Request) {
	newsID := mux.Vars(r)["n"]

	comments, err := api.db.CommentsByNews(newsID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(comments) == 0 {
		http.Error(w, "no comments found", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(comments)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}

// addCommentHandler - creates a new comment.
func (api *API) addCommentHandler(w http.ResponseWriter, r *http.Request) {
	var c storage.Comment
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		http.Error(w, "failed to decode response", http.StatusBadRequest)
		return
	}

	comment, err := api.db.AddComment(c)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(comment)
	if err != nil {
		http.Error(w, "failed to encode responce", http.StatusBadRequest)
		return
	}
}

package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"news/pkg/storage"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
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
	//api.r.Use(api.headersMiddleWare)
	api.r.Use(api.loggingMiddleWare)
	api.r.HandleFunc("/news/{id}", api.postHandler).Methods(http.MethodGet)
	api.r.HandleFunc("/news", api.postsHandler).Methods(http.MethodGet)
	api.r.HandleFunc("/news", api.addPostHandler).Methods(http.MethodPost)
	api.r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./webapp"))))
}

// postHandler - returns the post by id.
func (api *API) postHandler(w http.ResponseWriter, r *http.Request) {
	rawID := mux.Vars(r)["id"]
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}

	post, err := api.db.Post(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Warn("postHandler: post not found", "id", id, "err", err)
			http.Error(w, "post not found", http.StatusNotFound)
		} else {
			slog.Error("postHandler: failed to get post", "id", id, "err", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	err = json.NewEncoder(w).Encode(post)
	if err != nil {
		slog.Error("postHandler: failed to encode JSON", "err", err)
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}

// postsHandler - returns all posts.
func (api *API) postsHandler(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "invalid limit format", http.StatusBadRequest)
			return
		}
		if l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(w, "invalid offset format", http.StatusBadRequest)
			return
		}
		if o > 0 {
			offset = o
		}
	}

	posts, err := api.db.Posts(limit, offset)
	if err != nil {
		slog.Error("postsHandler: failed to get posts", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(posts)
	if err != nil {
		slog.Error("postsHandler: failed to encode JSON", "err", err)
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}

// addPostHandler - creates a new post.
func (api *API) addPostHandler(w http.ResponseWriter, r *http.Request) {
	var p storage.Post
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		slog.Error("addPostHandler: failed to decode JSON", "err", err)
		http.Error(w, "failed to decode response", http.StatusBadRequest)
		return
	}

	post, err := api.db.AddPost(p)
	if err != nil {
		slog.Error("addPostHandler: failed to add post", "err", err)
		http.Error(w, "failed to create post", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(post)
	if err != nil {
		slog.Error("addPostHandler: failed to encode JSON", "err", err)
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}

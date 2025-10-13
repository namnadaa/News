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
	api.r.Use(api.requestIDMiddleware)
	api.r.Use(api.loggingMiddleware)
	api.r.HandleFunc("/news/filter", api.filterHandler).Methods(http.MethodGet)
	api.r.HandleFunc("/news/new/{id}", api.postHandler).Methods(http.MethodGet)
	api.r.HandleFunc("/news/{n}", api.postsHandler).Methods(http.MethodGet)
	api.r.HandleFunc("/news", api.postsHandler).Methods(http.MethodGet)
	api.r.HandleFunc("/news", api.addPostHandler).Methods(http.MethodPost)
	api.r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./webapp"))))
}

// filterHandler - returns posts filtered by title substring.
func (api *API) filterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	requestID := getRequestID(r.Context())

	search := r.URL.Query().Get("s")
	if search == "" {
		http.Error(w, "missing search parameter 's'", http.StatusBadRequest)
		return
	}

	posts, err := api.db.SearchPosts(search)
	if err != nil {
		slog.Error("filterHandler: failed to search posts", "err", err, "request_id", requestID)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(posts) == 0 {
		http.Error(w, "no posts found", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(toDTOs(posts))
	if err != nil {
		slog.Error("filterHandler: failed to encode JSON", "err", err, "request_id", requestID)
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}

// postHandler - returns the post by id.
func (api *API) postHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	requestID := getRequestID(r.Context())

	rawID := mux.Vars(r)["id"]
	id, err := strconv.Atoi(rawID)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}

	post, err := api.db.Post(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Warn("postHandler: post not found", "id", id, "err", err, "request_id", requestID)
			http.Error(w, "post not found", http.StatusNotFound)
		} else {
			slog.Error("postHandler: failed to get post", "id", id, "err", err, "request_id", requestID)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	err = json.NewEncoder(w).Encode(toDTO(post))
	if err != nil {
		slog.Error("postHandler: failed to encode JSON", "err", err, "request_id", requestID)
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}

// postsHandler returns n posts and a paginated list of news posts.
func (api *API) postsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	requestID := getRequestID(r.Context())

	vars := mux.Vars(r)
	if rawN, ok := vars["n"]; ok {
		n, err := strconv.Atoi(rawN)
		if err != nil || n <= 0 {
			http.Error(w, "invalid n format", http.StatusBadRequest)
			return
		}

		posts, err := api.db.Posts(n)
		if err != nil {
			slog.Error("postsHandler: failed to get posts", "err", err, "request_id", requestID)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(toDTOs(posts))
		if err != nil {
			slog.Error("postsHandler: failed to encode JSON", "err", err, "request_id", requestID)
			http.Error(w, "failed to encode response", http.StatusBadRequest)
			return
		}
		return
	}

	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p <= 0 {
			http.Error(w, "invalid page format", http.StatusBadRequest)
			return
		}
		page = p
	}

	const perPage = 15
	posts, pagination, err := api.db.GetPostsPaginated(page, perPage)
	if err != nil {
		slog.Error("postsHandler: failed to fetch posts", "err", err, "request_id", requestID)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	resp := struct {
		News       []postDTO
		Pagination storage.Pagination
	}{
		News:       toDTOs(posts),
		Pagination: pagination,
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		slog.Error("postsHandler: failed to encode JSON", "err", err, "request_id", requestID)
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}

// addPostHandler - creates a new post.
func (api *API) addPostHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r.Context())

	var p storage.Post
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		slog.Error("addPostHandler: failed to decode JSON", "err", err, "request_id", requestID)
		http.Error(w, "failed to decode response", http.StatusBadRequest)
		return
	}

	post, err := api.db.AddPost(p)
	if err != nil {
		slog.Error("addPostHandler: failed to add post", "err", err, "request_id", requestID)
		http.Error(w, "failed to create post", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(post)
	if err != nil {
		slog.Error("addPostHandler: failed to encode JSON", "err", err, "request_id", requestID)
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}

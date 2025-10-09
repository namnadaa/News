package api

import "news/pkg/storage"

// postDTO represents a data transfer object for a single news post.
type postDTO struct {
	ID      int
	Title   string
	Content string
	PubTime int64
	Link    string
}

// NewsListResponse represents a response structure containing a list of posts and pagination info.
type NewsListResponse struct {
	News       []postDTO
	Pagination storage.Pagination
}

// toDTO converts a storage.Post entity to a postDTO.
func toDTO(p storage.Post) postDTO {
	return postDTO{
		ID:      p.ID,
		Title:   p.Title,
		Content: p.Content,
		PubTime: p.PubTime.Unix(),
		Link:    p.Link,
	}
}

// toDTOs converts a slice of storage.Post entities to a slice of postDTOs.
func toDTOs(p []storage.Post) []postDTO {
	out := make([]postDTO, len(p))
	for i := range p {
		out[i] = toDTO(p[i])
	}
	return out
}

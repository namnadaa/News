package api

import "news/pkg/storage"

type postDTO struct {
	ID      int
	Title   string
	Content string
	PubTime int64
	Link    string
}

func toDTO(p storage.Post) postDTO {
	return postDTO{
		ID:      p.ID,
		Title:   p.Title,
		Content: p.Content,
		PubTime: p.PubTime.Unix(),
		Link:    p.Link,
	}
}

func toDTOs(p []storage.Post) []postDTO {
	out := make([]postDTO, len(p))
	for i := range p {
		out[i] = toDTO(p[i])
	}
	return out
}

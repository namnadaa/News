package api

import (
	"comments/pkg/storage"
)

type commentDTO struct {
	ID        string
	NewsID    string
	ParentID  string
	Author    string
	Content   string
	CreatedAt int64
}

func toDTO(p storage.Comment) commentDTO {
	return commentDTO{
		ID:        p.ID,
		NewsID:    p.NewsID,
		ParentID:  p.ParentID,
		Author:    p.Author,
		Content:   p.Content,
		CreatedAt: p.CreatedAt.Unix(),
	}
}

func toDTOs(p []storage.Comment) []commentDTO {
	out := make([]commentDTO, len(p))
	for i := range p {
		out[i] = toDTO(p[i])
	}
	return out
}

package postgres

import (
	"github.com/danielh839/simple-forum/codegen/db"
	"github.com/danielh839/simple-forum/internal/model"
)

func convertMany[Record, Model any](records []Record, convert func(r Record) Model) []Model {
	models := make([]Model, len(records))
	for i, record := range records {
		models[i] = convert(record)
	}
	return models
}

func convertRef[T any](t T) *T {
	return &t
}

func createUserModel(record db.User) model.User {
	return model.User{
		ID:        int(record.ID),
		Username:  record.Username,
		Password:  record.Password,
		CreatedAt: record.CreatedAt,
	}
}

func createThreadModel(record db.Thread) model.Thread {
	title := ""
	if record.Title.Valid {
		title = record.Title.String
	}

	return model.Thread{
		ID:        record.ID,
		RootID:    record.RootID,
		ParentID:  record.ParentID,
		AuthorID:  int(record.AuthorID),
		CreatedAt: record.CreatedAt,
		Title:     title,
		Content:   record.Content,
	}
}

func createDetailedThreadModel(record db.GetThreadsRow) model.ThreadExtended {
	thread := createThreadModel(db.Thread{
		ID:        record.ID,
		RootID:    record.RootID,
		ParentID:  record.ParentID,
		AuthorID:  record.AuthorID,
		CreatedAt: record.CreatedAt,
		Title:     record.Title,
		Content:   record.Content,
	})

	var vote int = 0
	if record.HasUpvoted {
		vote = 1
	} else if record.HasDownvoted {
		vote = -1
	}

	return model.ThreadExtended{
		Thread:        thread,
		Author:        record.Username,
		UpvoteCount:   int(record.UpvoteCount),
		DownvoteCount: int(record.DownvoteCount),
		CommentCount:  int(record.CommentCount),
		Vote:          vote,
	}
}

func createVoteModel(record db.Vote) model.Vote {
	return model.Vote{
		ID:        int(record.ID),
		VoterID:   int(record.VoterID),
		ThreadID:  record.ThreadID,
		Upvote:    record.Upvote,
		CreatedAt: record.CreatedAt,
	}
}

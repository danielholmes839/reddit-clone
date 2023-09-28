package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        int
	Username  string
	Password  []byte
	CreatedAt time.Time
}

// Threads are posts and comments
type Thread struct {
	ID        uuid.UUID
	RootID    uuid.UUID     // this can match the thread id
	ParentID  uuid.NullUUID // the direct parent thread. if nil then this is the root/post
	AuthorID  int           // user id
	CreatedAt time.Time
	Title     string // empty string for comments
	Content   string
}

type ThreadExtended struct {
	Thread
	// author username
	Author string

	// for the user who requested the thread.
	Vote int // 0 = not voted, 1 = upvoted, -1 = downvoted

	// stats
	UpvoteCount   int
	DownvoteCount int
	CommentCount  int
}

type ThreadQuery struct {
	ReaderID         int
	ReaderIDValid    bool
	AuthorID         int
	AuthorIDValid    bool
	RootID           uuid.UUID
	RootIDValid      bool
	RootThreadsOnly  bool
	ChildThreadsOnly bool
}

func NewThreadQuery() ThreadQuery {
	return ThreadQuery{}
}

func (q ThreadQuery) WithAuthor(authorID int) ThreadQuery {
	q.AuthorID = authorID
	q.AuthorIDValid = true
	return q
}

func (q ThreadQuery) WithReader(readerID int) ThreadQuery {
	q.ReaderID = readerID
	q.ReaderIDValid = true
	return q
}

func (q ThreadQuery) WithRoot(rootID uuid.UUID) ThreadQuery {
	q.RootID = rootID
	q.RootIDValid = true
	return q
}

func (q ThreadQuery) WithPostsOnly() ThreadQuery {
	q.RootThreadsOnly = true
	return q
}

func (q ThreadQuery) WithCommentsOnly() ThreadQuery {
	q.ChildThreadsOnly = true
	return q
}

type Vote struct {
	ID        int
	VoterID   int // user id
	ThreadID  uuid.UUID
	Upvote    bool
	CreatedAt time.Time
}

type CreatePost struct {
	AuthorID int
	Title    string
	Content  string
}

type CreateReply struct {
	AuthorID int
	ParentID uuid.UUID // the thread that the reply is for
	Content  string
}

type CreateVote struct {
	VoterID  int
	ThreadID uuid.UUID
	Upvote   bool
}

package postgres

import (
	"context"
	"database/sql"
	"log"

	"github.com/danielh839/simple-forum/codegen/db"
	"github.com/danielh839/simple-forum/internal/model"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
	Query *db.Queries // sqlc access
}

func NewDB(credentials Credentials) *DB {
	// connect to postgres
	conn, err := sql.Open("postgres", credentials.String())
	if err != nil {
		log.Fatal(err)
	}

	if err := conn.Ping(); err != nil {
		log.Fatal(err)
	}

	return &DB{
		DB:    conn,
		Query: db.New(conn),
	}
}

func (pdb *DB) CreatePost(ctx context.Context, input model.CreatePost) (model.Thread, error) {
	// we need to generate the uuid ahead of time so that we can use it as the root id as well
	postID := uuid.New()

	// create the post
	post, err := pdb.Query.CreatePost(ctx, db.CreatePostParams{
		ID:       postID,
		RootID:   postID,
		AuthorID: int32(input.AuthorID),
		Title:    sql.NullString{String: input.Title, Valid: true},
		Content:  input.Content,
	})
	if err != nil {
		return model.Thread{}, err
	}

	return createThreadModel(post), nil
}

func (pdb *DB) CreateComment(ctx context.Context, input model.CreateReply) (model.Thread, error) {
	// get the parent thread
	parent, err := pdb.Query.GetThread(ctx, input.ParentID)
	if err != nil {
		return model.Thread{}, err
	}

	// create the reply
	reply, err := pdb.Query.CreateComment(ctx, db.CreateCommentParams{
		RootID:   parent.RootID, // use the parent's root id
		ParentID: uuid.NullUUID{UUID: parent.ID, Valid: true},
		AuthorID: int32(input.AuthorID),
		Content:  input.Content,
	})
	if err != nil {
		return model.Thread{}, err
	}

	return createThreadModel(reply), nil
}

func (pdb *DB) CreateUser(ctx context.Context, username, password string) (model.User, error) {
	user, err := pdb.Query.CreateUser(ctx, db.CreateUserParams{
		Username: username,
		Password: []byte(password),
	})
	if err != nil {
		return model.User{}, err
	}
	return createUserModel(user), nil
}

func (pdb *DB) CreateVote(ctx context.Context, input model.CreateVote) (model.Vote, error) {
	record, err := pdb.Query.CreateVote(ctx, db.CreateVoteParams{
		VoterID:  int32(input.VoterID),
		ThreadID: input.ThreadID,
		Upvote:   input.Upvote,
	})
	if err != nil {
		return model.Vote{}, err
	}
	return createVoteModel(record), err
}

func (pdb *DB) DeleteVote(ctx context.Context, voterID int, threadID uuid.UUID) error {
	return pdb.Query.DeleteVote(ctx, db.DeleteVoteParams{
		VoterID:  int32(voterID),
		ThreadID: threadID,
	})
}

// GetThreadsExtended
func (pdb *DB) GetThreadsExtended(ctx context.Context, query model.ThreadQuery) []model.ThreadExtended {
	// query the threads
	records, err := pdb.Query.GetThreads(ctx, db.GetThreadsParams{
		ReaderID:         sql.NullInt32{Int32: int32(query.ReaderID), Valid: query.ReaderIDValid},
		AuthorID:         sql.NullInt32{Int32: int32(query.AuthorID), Valid: query.AuthorIDValid},
		RootID:           uuid.NullUUID{UUID: query.RootID, Valid: query.RootIDValid},
		RootThreadsOnly:  query.RootThreadsOnly,
		ChildThreadsOnly: query.ChildThreadsOnly,
	})

	if err != nil {
		return []model.ThreadExtended{}
	}

	// convert to []model.ThreadExtended
	return convertMany(records, createDetailedThreadModel)
}

// GetThreadsExtended query all threads (root/post and comments) with the same root ID
func (pdb *DB) GetThreadExtended(ctx context.Context, readerID int, threadID uuid.UUID) (model.ThreadExtended, error) {
	res, err := pdb.Query.GetThreadExtended(ctx, db.GetThreadExtendedParams{ReaderID: sql.NullInt32{Int32: int32(readerID), Valid: true}, ID: threadID})
	if err != nil {
		return model.ThreadExtended{}, nil
	}
	return createDetailedThreadModel(db.GetThreadsRow(res)), nil
}

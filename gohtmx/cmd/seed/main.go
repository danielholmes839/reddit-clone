package main

import (
	"context"
	"io"
	"os"

	"github.com/danielh839/simple-forum/internal/model"
	"github.com/danielh839/simple-forum/internal/postgres"
)

func main() {
	// open the file containing the sql table schemas
	schema, _ := os.Open("./codegen/schema.sql")
	defer schema.Close()
	query, err := io.ReadAll(schema)

	if err != nil {
		panic(err)
	}

	// create the tables
	db := postgres.NewDB(postgres.DefaultCredentials)
	_, err = db.Exec(string(query))

	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	dhh, err := db.CreateUser(ctx, "dhh", "password")
	if err != nil {
		panic(err)
	}

	primeagen, err := db.CreateUser(ctx, "primeagen", "password")
	if err != nil {
		panic(err)
	}

	post, err := db.CreatePost(ctx, model.CreatePost{
		AuthorID: primeagen.ID,
		Title:    "Reaction: Turbo 8 is dropping Typescript",
		Content:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi hendrerit eros a faucibus porttitor",
	})

	post, err = db.CreatePost(ctx, model.CreatePost{
		AuthorID: dhh.ID,
		Title:    "Turbo 8 is dropping Typescript",
		Content:  "Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos",
	})

	post, err = db.CreatePost(ctx, model.CreatePost{
		AuthorID: primeagen.ID,
		Title:    "Reaction: Python Is BAD For Beginners",
		Content:  "Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos",
	})

	if err != nil {
		panic(err)
	}

	_, err = db.CreateComment(ctx, model.CreateComment{
		AuthorID: primeagen.ID,
		ParentID: post.ID,
		Content:  "Stop the cap",
	})

	if err != nil {
		panic(err)
	}

	db.CreateVote(ctx, model.CreateVote{
		VoterID:  primeagen.ID,
		ThreadID: post.ID,
		Upvote:   false,
	})

	db.CreateVote(ctx, model.CreateVote{
		VoterID:  dhh.ID,
		ThreadID: post.ID,
		Upvote:   true,
	})

	// threads := db.GetPosts(ctx, dhh.ID)
	// fmt.Println("all posts", threads)

	// threads = db.GetThreadsExtended(ctx, dhh.ID, post.RootID)
	// fmt.Println("all threads in post", threads)

	// threads = db.GetUserComments(ctx, dhh.ID, dhh.ID)
	// fmt.Println("all dhh comments", threads)

	// threads = db.GetUserPosts(ctx, dhh.ID, dhh.ID)
	// fmt.Println("all dhh posts", threads)

	// threads = db.GetUserComments(ctx, dhh.ID, primeagen.ID)
	// fmt.Println("all primeagen comments", threads)
}

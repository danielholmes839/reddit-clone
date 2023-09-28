package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/danielh839/simple-forum/internal/model"
	"github.com/danielh839/simple-forum/internal/postgres"
	"github.com/fsnotify/fsnotify"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
)

type Head struct {
	Title string
}

type HomePage struct {
	Head
	Username string
	Posts    []model.ThreadExtended
}

type ThreadPage struct {
	Head
	Post model.ThreadExtended
}

type LoginFormValues struct {
	Username string
	Password string
}

type LoginForm struct {
	Values LoginFormValues
	Err    error
}

type LoginPage struct {
	Head
	LoginForm
}

func ParseLoginFormValues(r *http.Request) (LoginFormValues, error) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	values := LoginFormValues{Username: username, Password: password}

	if len([]rune(username)) < 3 {
		return values, errors.New("username must be at least 3 characters long")
	}

	if len([]rune(username)) > 20 {
		return values, errors.New("username cannot be longer than 20 characters")
	}

	if len([]rune(password)) < 3 {
		return values, errors.New("password must be at least 3 characters long")
	}

	if len([]rune(password)) > 40 {
		return values, errors.New("password cannot be longer than 40 characters")
	}

	return values, nil
}

func Sub(a, b int) int {
	return a - b
}

func ThreadTimestamp(t time.Time) string {
	if time.Since(t) < time.Minute {
		return "Just now"
	}

	if dur := time.Since(t); dur < 60*time.Minute {
		return fmt.Sprintf("%d minutes ago", int(dur.Minutes()))
	}

	if dur := time.Since(t); dur < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(dur.Hours()))
	}

	return t.Format("2006-01-02")
}

var funcMap = template.FuncMap{
	"Sub":       Sub,
	"Timestamp": ThreadTimestamp,
}

func DefaultTemplates() (*template.Template, error) {
	templateFiles, err := filepath.Glob("./assets/templates/*.html")
	if err != nil {
		return nil, err
	}

	// parse templates
	templates, err := template.New("").Funcs(funcMap).ParseFiles(templateFiles...)
	if err != nil {
		return nil, err
	}
	return templates, nil
}

func RefreshTemplates(refresh func() (*template.Template, error)) func() *template.Template {
	watcher, err := fsnotify.NewWatcher()
	watcher.Add("./assets/templates")
	if err != nil {
		log.Fatal(err)
	}

	mu := sync.RWMutex{}

	activeTemplate, err := refresh()
	if err != nil {
		log.Fatal(err)
	}

	getActiveTemplate := func() *template.Template {
		mu.RLock()
		defer mu.RUnlock()
		return activeTemplate
	}

	refreshActiveTemplate := func() {
		mu.Lock()
		defer mu.Unlock()

		log.Print("refreshing templates")
		newTemplate, err := refresh()
		if err != nil {
			log.Fatal(err)
		}

		activeTemplate = newTemplate
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					refreshActiveTemplate()
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	return getActiveTemplate
}

func main() {
	// parse templates
	templates := RefreshTemplates(DefaultTemplates)
	db := postgres.NewDB(postgres.DefaultCredentials)

	userID := 1

	r := chi.NewRouter()
	r.Use(middleware.DefaultLogger)

	// handle login form submissions
	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		values, err := ParseLoginFormValues(r)

		if err != nil {
			// re-execute the login-form template with current values
			// and error message.
			templates().ExecuteTemplate(w, "login-form.html", LoginForm{
				Values: values,
				Err:    err,
			})
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "auth",
			Value:    values.Username,
			Secure:   true,
			SameSite: http.SameSiteDefaultMode,
			HttpOnly: true,
			MaxAge:   0,
		})

		w.Header().Set("HX-Redirect", "/home")
	})

	// login page
	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		_ = templates().ExecuteTemplate(w, "login.html", LoginPage{
			Head:      Head{Title: "Login"},
			LoginForm: LoginForm{}, // empty by default
		})
	})

	r.Get("/home", func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie("auth")

		query := model.NewThreadQuery().WithReader(userID).WithPostsOnly()
		posts := db.GetThreadsExtended(r.Context(), query)

		_ = templates().ExecuteTemplate(w, "home.html", HomePage{
			Head:     Head{Title: "Home"},
			Username: cookie.Value,
			Posts:    posts,
		})
	})

	r.Post("/thread/{thread_id}/vote", func(w http.ResponseWriter, r *http.Request) {
		upvote := r.URL.Query().Get("direction") == "up"
		threadId := uuid.Must(uuid.Parse(chi.URLParam(r, "thread_id")))

		_, err := db.CreateVote(r.Context(), model.CreateVote{
			VoterID:  userID,
			ThreadID: threadId,
			Upvote:   upvote,
		})

		if err != nil {
			fmt.Println(err, upvote, threadId.String())
		}

		thread, err := db.GetThreadExtended(r.Context(), userID, threadId)
		if err != nil {
			log.Fatal(err)
		}

		_ = templates().ExecuteTemplate(w, "post-preview.html", thread)
	})

	r.Delete("/thread/{thread_id}/vote", func(w http.ResponseWriter, r *http.Request) {
		threadID := uuid.Must(uuid.Parse(chi.URLParam(r, "thread_id")))

		err := db.DeleteVote(r.Context(), 1, threadID)

		thread, err := db.GetThreadExtended(r.Context(), 1, threadID)
		if err != nil {
			log.Fatal(err)
		}
		_ = templates().ExecuteTemplate(w, "post-preview.html", thread)
	})

	r.Get("/thread/{thread_id}", func(w http.ResponseWriter, r *http.Request) {
		threadID := uuid.Must(uuid.Parse(chi.URLParam(r, "thread_id")))
		post, _ := db.GetThreadExtended(r.Context(), userID, threadID)

		// query := model.NewThreadQuery().WithReader(userID)
		// conversation := db.GetThreadsExtended(r.Context(), query)

		post, err := db.GetThreadExtended(r.Context(), userID, threadID)
		if err != nil {
			log.Fatal(err)
		}

		err = templates().ExecuteTemplate(w, "post.html", ThreadPage{
			Head: Head{Title: fmt.Sprintf("Post: %s", post.Title)},
			Post: post,
		})
		fmt.Println(err)
	})

	log.Print("listening...")
	http.ListenAndServe("localhost:3000", r)
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string
	Title     string
	Done      bool
	UpdatedAt time.Time
}

type Clock interface {
	Now() time.Time
}

type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List() []Task
	SetDone(id string, done bool) (Task, error)
}

type InMemoryTaskRepo struct {
	mu      sync.RWMutex
	tasks   map[string]Task
	clock   Clock
	counter int
}

func NewInMemoryTaskRepo(clock Clock) *InMemoryTaskRepo {
	return &InMemoryTaskRepo{
		tasks: make(map[string]Task),
		clock: clock,
	}
}

func (r *InMemoryTaskRepo) Create(title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.counter++
	id := fmt.Sprintf("%d", r.counter)

	task := Task{
		ID:        id,
		Title:     title,
		Done:      false,
		UpdatedAt: r.clock.Now(),
	}
	r.tasks[id] = task
	return task, nil
}

func (r *InMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, ok := r.tasks[id]
	return task, ok
}

func (r *InMemoryTaskRepo) List() []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		list = append(list, t)
	}

	sort.Slice(list, func(i, j int) bool {
		if !list[i].UpdatedAt.Equal(list[j].UpdatedAt) {
			return list[i].UpdatedAt.After(list[j].UpdatedAt)
		}
		return list[i].ID < list[j].ID
	})

	return list
}

func (r *InMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, errors.New("not found")
	}

	task.Done = done
	task.UpdatedAt = r.clock.Now()
	r.tasks[id] = task
	return task, nil
}

type HTTPHandler struct {
	repo *InMemoryTaskRepo
}

func NewHTTPHandler(repo *InMemoryTaskRepo) *HTTPHandler {
	return &HTTPHandler{
		repo: repo,
	}
}

func (s *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	switch {
	case method == http.MethodPost && path == "/tasks":
		s.handleCreate(w, r)
	case method == http.MethodGet && path == "/tasks":
		s.handleList(w, r)
	case method == http.MethodGet && strings.HasPrefix(path, "/tasks/"):
		s.handleGet(w, r)
	case method == http.MethodPatch && strings.HasPrefix(path, "/tasks/"):
		s.handlePatch(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *HTTPHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(body.Title)
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	task, _ := s.repo.Create(title)
	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(task)
	if err != nil {
		return
	}
}

func (s *HTTPHandler) handleList(w http.ResponseWriter, r *http.Request) {
	tasks := s.repo.List()
	err := json.NewEncoder(w).Encode(tasks)
	if err != nil {
		return
	}
}

func (s *HTTPHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	task, ok := s.repo.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	err := json.NewEncoder(w).Encode(task)
	if err != nil {
		return
	}
}

func (s *HTTPHandler) handlePatch(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/tasks/")

	var body struct {
		Done *bool `json:"done"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&body); err != nil || body.Done == nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	task, err := s.repo.SetDone(id, *body.Done)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(task)
	if err != nil {
		return
	}
}

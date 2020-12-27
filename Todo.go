package main

type TodoStatus uint

const (
  NOT_STARTED TodoStatus = iota
  IN_PROGRESS
  DONE
)

type Todo struct {
  Id string `json:"id"`
  Name string `json:"name"`
  Status TodoStatus `json:"status"`
}

func newTodo(id string, name string, status TodoStatus) Todo {
  return Todo{
    Id:     id,
    Name:   name,
    Status: status,
  }
}

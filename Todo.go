package main

type TodoStatus uint

const (
  NOT_STARTED TodoStatus = iota
  IN_PROGRESS
  DONE
)

type Todo struct {
  Id int `json:"id"`
  Name string `json:"name"`
  Status TodoStatus `json:"status"`
}

func newTodo(id int, name string) Todo {
  return Todo{
    Id:     id,
    Name:   name,
    Status: NOT_STARTED,
  }
}

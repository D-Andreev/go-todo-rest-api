package main

type TodoStatus uint

const (
  NOT_STARTED TodoStatus = iota
  IN_PROGRESS
  DONE
)

type Todo struct {
  id int
  name string
  status TodoStatus
}

func newTodo(id int, name string) Todo {
  return Todo{
    id:     id,
    name:   name,
    status: NOT_STARTED,
  }
}

package main

import (
  "encoding/json"
  "fmt"
  "github.com/pkg/errors"
  "io"
  "io/ioutil"
  "net/http"
  "os"
  "strconv"
  "strings"
)

var todos = []Todo{
  newTodo(1, "Go shopping"),
  newTodo(2, "Clean room"),
}

var statusNames = map[TodoStatus]string{
  NOT_STARTED: "Not started",
  IN_PROGRESS: "In Progress",
  DONE: "Done",
}

type TodoBody struct {
  Name string `json:"name"`
  Status TodoStatus `json:"status"`
}

func checkErr(e error) {
  if e != nil {
    panic(e)
  }
}

func main() {
  http.HandleFunc("/todo/", todoController)
  port := getPort()

  fmt.Println("Server listening at: ", port)
  err := http.ListenAndServe(":" + port, nil)
  checkErr(err)
}

func getPort() string {
  port := os.Getenv("PORT")
  if port == "" {
    port = "8090"
  }

  return port
}

func deserializeTodoBody(b []byte, parseBody *TodoBody) error {
  err := json.Unmarshal(b, &parseBody)
  if err != nil {
    return err
  }

  return nil
}

func todoController(w http.ResponseWriter, req *http.Request) {
  if req.Method == "POST" {
    err := addTodo(req.Body)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    w.WriteHeader(http.StatusOK)
  } else if req.Method == "GET" {
    todos := serializeTodos()
    w.Header().Set("Content-Type", "application/json")
    w.Write(todos)
  } else if req.Method == "PUT" {
    todoBS := updateTodo(w, req)
    w.Header().Set("Content-Type", "application/json")
    w.Write(todoBS)
    w.WriteHeader(http.StatusOK)
  }
}

func updateTodo(w http.ResponseWriter, req *http.Request) []byte {
  splitPath := strings.Split(req.URL.Path, "/")
  id, err := strconv.Atoi(splitPath[len(splitPath) - 1])
  if err != nil {
    http.Error(w, "Invalid todo id", http.StatusBadRequest)
    return nil
  }

  idx := findTodoIdx(id)
  if idx == -1 {
    http.Error(w, "Todo does not exists", http.StatusBadRequest)
    return nil
  }

  b, err := ioutil.ReadAll(req.Body)
  if err != nil {
    http.Error(w, "Invalid body", http.StatusBadRequest)
    return nil
  }
  var parsedBody TodoBody
  e := deserializeTodoBody(b, &parsedBody)
  if e != nil {
    http.Error(w, "Invalid todo update data", http.StatusBadRequest)
    return nil
  }
  todos[idx].Name = parsedBody.Name
  todos[idx].Status = parsedBody.Status
  todoBS, err := serializeTodo(todos[idx])
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return nil
  }

  return todoBS
}

func findTodoIdx(id int) int {
  idx := -1
  for i, todo := range todos {
    if todo.Id == id {
      idx = i
    }
  }

  return idx
}

func addTodo(body io.ReadCloser) error {
  b, err := ioutil.ReadAll(body)
  checkErr(err)

  var parseBody TodoBody
  e := deserializeTodoBody(b, &parseBody)
  if e != nil {
    return errors.New("Invalid todo data")
  }
  todos = append(todos, newTodo(len(todos) + 1, parseBody.Name))
  return nil
}

func serializeTodos() []byte {
  b, err := json.Marshal(&todos)
  checkErr(err)

  return b
}

func serializeTodo(todo Todo) ([]byte, error) {
  b, err := json.Marshal(&todo)
  if err != nil {
    return []byte{}, err
}

  return b, nil
}

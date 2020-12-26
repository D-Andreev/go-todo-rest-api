package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "os"
)

var todos = []Todo{
  newTodo(1, "Go shopping"),
  newTodo(2, "Clean room"),
}

type TodoBody struct {
  Name string `json:"name"`
}

func checkErr(e error) {
  if e != nil {
    panic(e)
  }
}

func main() {
  http.HandleFunc("/todo", todoController)
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
    b, err := ioutil.ReadAll(req.Body)
    checkErr(err)

    var parseBody TodoBody
    e := deserializeTodoBody(b, &parseBody)
    if e != nil {
      http.Error(w, "Invalid todo data", 500)
      return
    }
    todos = append(todos, newTodo(len(todos) + 1, parseBody.Name))
    w.WriteHeader(200)
  }
}

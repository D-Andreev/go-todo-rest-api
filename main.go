package main

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/pkg/errors"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
  "io"
  "io/ioutil"
  "net/http"
  "os"
  "strings"
)

const DbName = "go_todos"
const CollectionName = "todos"

var client *mongo.Client
var ctx context.Context

type mongoTodo struct {
  Id primitive.ObjectID `bson:"_id"`
  Name string `bson:"name"`
  Status TodoStatus `bson:"status"`
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
  connectToMongo()
  http.HandleFunc("/todo/", todoController)
  port := getPort()

  fmt.Println("Server listening at: ", port)
  err := http.ListenAndServe(":"+port, nil)
  checkErr(err)
}

func connectToMongo() {
  mongoUrl := getMongoUrl()
  var err error
  clientOptions := options.Client().ApplyURI(mongoUrl)
  client, err = mongo.Connect(context.TODO(), clientOptions)
  if err != nil {
    checkErr(err)
  }
  err = client.Ping(context.TODO(), nil)
  if err != nil {
    checkErr(err)
  }

  fmt.Println("Connected to MongoDB!")
}

func getMongoUrl() string {
  url := os.Getenv("DB")
  if url == "" {
    url = "mongodb://localhost:27017"
  }

  return url
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
    todos := selectTodos()
    todosBs := serializeTodos(todos)
    w.Header().Set("Content-Type", "application/json")
    w.Write(todosBs)
  } else if req.Method == "PUT" {
    todoBS := updateTodo(w, req)
    w.Header().Set("Content-Type", "application/json")
    w.Write(todoBS)
    w.WriteHeader(http.StatusOK)
  } else if req.Method == "DELETE" {
    err := deleteTodo(req.URL.Path)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    w.WriteHeader(http.StatusOK)
  }
}

func deleteTodo(path string) error {
  id := getIdFromUrl(path)
  collection := client.Database(DbName).Collection(CollectionName)
  objID, err := primitive.ObjectIDFromHex(id)
  if err != nil {
    return err
  }
  filter := bson.M{"_id": bson.M{"$eq": objID}}

  _, err = collection.DeleteOne(ctx, filter)
  return err
}

func selectTodos() []Todo {
  var todos []Todo
  collection := client.Database(DbName).Collection(CollectionName)
  cur, err := collection.Find(ctx, bson.D{})
  if err != nil {
    checkErr(err)
  }
  defer cur.Close(ctx)
  for cur.Next(ctx) {
    var result mongoTodo
    err := cur.Decode(&result)
    if err != nil {
      checkErr(err)
    }
    todo := newTodo(result.Id.Hex(), result.Name, result.Status)
    todos = append(todos, todo)
  }
  if err := cur.Err(); err != nil {
    checkErr(err)
  }

  return todos
}

func getIdFromUrl(path string) string {
  splitPath := strings.Split(path, "/")
  return splitPath[len(splitPath) - 1]
}

func updateTodo(w http.ResponseWriter, req *http.Request) []byte {
  id := getIdFromUrl(req.URL.Path)

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
  collection := client.Database(DbName).Collection(CollectionName)
  objID, err := primitive.ObjectIDFromHex(id)
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return nil
  }
  filter := bson.M{"_id": bson.M{"$eq": objID}}
  update := bson.M{"$set": bson.M{
    "name": parsedBody.Name,
    "status": parsedBody.Status,
  }}

  _, err = collection.UpdateOne(ctx, filter, update)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return nil
  }
  todoBS, err := serializeTodo(newTodo(id, parsedBody.Name, parsedBody.Status))
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return nil
  }

  return todoBS
}

func addTodo(body io.ReadCloser) error {
  b, err := ioutil.ReadAll(body)
  checkErr(err)

  var parseBody TodoBody
  e := deserializeTodoBody(b, &parseBody)
  if e != nil {
    return errors.New("Invalid todo data")
  }

  collection := client.Database(DbName).Collection(CollectionName)
  _, err = collection.InsertOne(ctx, bson.M{
    "name": parseBody.Name,
    "status": parseBody.Status,
  })
  return err
}

func serializeTodos(todos []Todo) []byte {
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

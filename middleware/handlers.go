package middleware

import (
    "database/sql"
    "encoding/json" 
    "fmt"
    "go-postgres/models" 
    "log"
    "net/http" 
    "os"     
    "strconv"  
    "github.com/gorilla/mux" 
    "github.com/joho/godotenv" 
    _ "github.com/lib/pq"     
)

// response format
type response struct {
    ID      int64  `json:"id,omitempty"`
    Message string `json:"message,omitempty"`
}

// create connection with postgres db
func createConnection() *sql.DB {
    err := godotenv.Load(".env")

    if err != nil {
        log.Fatalf("Error loading .env file")
    }
	
    db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))

    if err != nil {
        panic(err)
    }

    err = db.Ping()

    if err != nil {
        panic(err)
    }

    fmt.Println("Successfully connected!")
    return db
}

// CreateUser create a user in the postgres db
func CreateUser(w http.ResponseWriter, r *http.Request) {
  
    w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    var user models.User

    err := json.NewDecoder(r.Body).Decode(&user)
	fmt.Println("checkccc",user)

    if err != nil {
        log.Fatalf("Unable to decode the request body.  %v", err)
    }

    insertID := insertUser(user)

    res := response{
        ID:      insertID,
        Message: "User created successfully",
    }

    json.NewEncoder(w).Encode(res)
}

// GetUser will return a single user by its id
func GetUser(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    params := mux.Vars(r)

    id, err := strconv.Atoi(params["id"])

    if err != nil {
        log.Fatalf("Unable to convert the string into int.  %v", err)
    }

    user, err := getUser(int64(id))

    if err != nil {
        log.Fatalf("Unable to get user. %v", err)
    }

    json.NewEncoder(w).Encode(user)
}

// GetAllUser will return all the users
func GetAllUser(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    users, err := getAllUsers()

    if err != nil {
        log.Fatalf("Unable to get all user. %v", err)
    }

    json.NewEncoder(w).Encode(users)
}

// UpdateUser update user's detail in the postgres db
func UpdateUser(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "PUT")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    params := mux.Vars(r)

    id, err := strconv.Atoi(params["id"])

    if err != nil {
        log.Fatalf("Unable to convert the string into int.  %v", err)
    }

    var user models.User

    err = json.NewDecoder(r.Body).Decode(&user)

    if err != nil {
        log.Fatalf("Unable to decode the request body.  %v", err)
    }

    updatedRows := updateUser(int64(id), user)

    msg := fmt.Sprintf("User updated successfully. Total rows/record affected %v", updatedRows)

    res := response{
        ID:      int64(id),
        Message: msg,
    }

    json.NewEncoder(w).Encode(res)
}

// DeleteUser delete user's detail in the postgres db
func DeleteUser(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "DELETE")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    params := mux.Vars(r)

    id, err := strconv.Atoi(params["id"])

    if err != nil {
        log.Fatalf("Unable to convert the string into int.  %v", err)
    }

    deletedRows := deleteUser(int64(id))

    msg := fmt.Sprintf("User updated successfully. Total rows/record affected %v", deletedRows)

    res := response{
        ID:      int64(id),
        Message: msg,
    }

    json.NewEncoder(w).Encode(res)
}

// insert one user in the DB
func insertUser(user models.User) int64 {

    db := createConnection()

    defer db.Close()

    sqlStatement := `INSERT INTO users (name, location, age) VALUES ($1, $2, $3) RETURNING userid`

    var id int64

    err := db.QueryRow(sqlStatement, user.Name, user.Location, user.Age).Scan(&id)

    if err != nil {
        log.Fatalf("Unable to execute the query. %v", err)
    }

    fmt.Printf("Inserted a single record %v", id)

    return id
}

// get one user from the DB by its userid
func getUser(id int64) (models.User, error) {
    db := createConnection()

    defer db.Close()

    var user models.User

    sqlStatement := `SELECT * FROM users WHERE userid=$1`

    row := db.QueryRow(sqlStatement, id)

    err := row.Scan(&user.ID, &user.Name, &user.Age, &user.Location)

    switch err {
    case sql.ErrNoRows:
        fmt.Println("No rows were returned!")
        return user, nil
    case nil:
        return user, nil
    default:
        log.Fatalf("Unable to scan the row. %v", err)
    }

    return user, err
}

// get one user from the DB by its userid
func getAllUsers() ([]models.User, error) {
    db := createConnection()

    defer db.Close()

    var users []models.User

    sqlStatement := `SELECT * FROM users`

    rows, err := db.Query(sqlStatement)

    if err != nil {
        log.Fatalf("Unable to execute the query. %v", err)
    }

    defer rows.Close()

    for rows.Next() {
        var user models.User

        err = rows.Scan(&user.ID, &user.Name, &user.Age, &user.Location)

        if err != nil {
            log.Fatalf("Unable to scan the row. %v", err)
        }

        users = append(users, user)

    }

    return users, err
}

// update user in the DB
func updateUser(id int64, user models.User) int64 {

    db := createConnection()

    defer db.Close()

    sqlStatement := `UPDATE users SET name=$2, location=$3, age=$4 WHERE userid=$1`

    res, err := db.Exec(sqlStatement, id, user.Name, user.Location, user.Age)

    if err != nil {
        log.Fatalf("Unable to execute the query. %v", err)
    }

    rowsAffected, err := res.RowsAffected()

    if err != nil {
        log.Fatalf("Error while checking the affected rows. %v", err)
    }

    fmt.Printf("Total rows/record affected %v", rowsAffected)

    return rowsAffected
}

// delete user in the DB
func deleteUser(id int64) int64 {

    db := createConnection()

    defer db.Close()

    sqlStatement := `DELETE FROM users WHERE userid=$1`

    res, err := db.Exec(sqlStatement, id)

    if err != nil {
        log.Fatalf("Unable to execute the query. %v", err)
    }

    rowsAffected, err := res.RowsAffected()

    if err != nil {
        log.Fatalf("Error while checking the affected rows. %v", err)
    }

    fmt.Printf("Total rows/record affected %v", rowsAffected)

    return rowsAffected
}
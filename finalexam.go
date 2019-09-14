package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"net/http"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

type Customer struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

var myDB *sql.DB

func authMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token != "token2019" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort() // required here, otherwise c.next continues.
		return
	}

	c.Next()
}

func CreateCustomerTable() {
	var err error

	createTb := `
	CREATE TABLE IF NOT EXISTS customers (
		id SERIAL PRIMARY KEY, 
		name TEXT, 
		email TEXT,
		status TEXT
	);
	`
	_, err = myDB.Exec(createTb)
	if err != nil {
		log.Println("Can't create table", err)
	}

	fmt.Println("Create table success")
}

func createCustomerHandler(c *gin.Context) {
	var cust Customer

	err := c.ShouldBindJSON(&cust)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Error Binding Message " + err.Error()})
		return
	}

	iString := "INSERT INTO customers (name, email, status) values ($1, $2, $3) RETURNING id"
	row := myDB.QueryRow("INSERT INTO customers (name, email, status) values ($1, $2, $3) RETURNING id", cust.Name, cust.Email, cust.Status)
	var id int
	err = row.Scan(&id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Error Insertion String: " + iString + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "name": cust.Name, "email": cust.Email, "status": cust.Status})
}

func getCustomerHandler(c *gin.Context) {
	strID := c.Param("id")
	rowID, err := strconv.Atoi(strID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Can't convert customer ID: " + strID + " -" + err.Error()})
		return
	}

	qString := "SELECT id, name, email, status FROM customers WHERE id=$1"
	stmt, err := myDB.Prepare(qString)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Can't prepare query one row statement" + qString + " - " + err.Error()})
		return
	}

	row := stmt.QueryRow(rowID)
	var cust Customer

	err = row.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Status)
	switch err {
	case sql.ErrNoRows:
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "No rows found with id: " + strID + " - " + err.Error()})
	case nil:
		c.JSON(http.StatusOK, cust)
	default:
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Can't scan row into variable id: " + strID + " - " + err.Error()})
	}
}

func getAllCustomersHandler(c *gin.Context) {
	qString := "SELECT id, name, email, status FROM customers"
	stmt, err := myDB.Prepare(qString)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Can't creat Prepare statement " + qString + " - " + err.Error()})
		return
	}

	rows, err := stmt.Query()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Can't query all customers" + err.Error()})
		return
	}

	var custRows []Customer
	var countRows int

	for rows.Next() {
		var cust Customer

		err := rows.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Status)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"status": "Can't scan next row into variable " + err.Error()})
			return
		}
		custRows = append(custRows, cust)
		countRows++
	}

	if countRows < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "No data found "})
		return
	}
	c.JSON(http.StatusOK, custRows)
}

func updateCustomerHandler(c *gin.Context) {
	var cust Customer

	strID := c.Param("id")
	rowID, err := strconv.Atoi(strID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Can't convert customer ID: " + strID + " -" + err.Error()})
		return
	}

	err = c.ShouldBindJSON(&cust)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Error Binding Message " + err.Error()})
		return
	}

	uString := "UPDATE customers SET name=$2,email=$3,status=$4 WHERE id=$1"
	stmt, err := myDB.Prepare(uString)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Can't prepare update statement " + uString + err.Error()})
		return
	}

	_, err = stmt.Exec(strID, cust.Name, cust.Email, cust.Status)
	if err != nil {
		log.Println("error execute update ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Can't update " + uString + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": rowID, "name": cust.Name, "email": cust.Email, "status": cust.Status})
}

func deleteCustomerHandler(c *gin.Context) {
	strID := c.Param("id")

	dString := "DELETE FROM todos WHERE id=$1"
	stmt, err := myDB.Prepare("DELETE FROM todos WHERE id=$1")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Can't prepare delete statement " + dString + err.Error()})
		return
	}

	_, err = stmt.Exec(strID)
	if err != nil {
		log.Println("error execute delete ", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Can't delete " + dString + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})
}

func DeleteTodo(db *sql.DB, rowID int) {
	stmt, err := db.Prepare("DELETE FROM todos WHERE id=$1; ")
	if err != nil {
		log.Fatal("can't prepare statement delete ", err)
	}

	res, err := stmt.Exec(rowID)
	if err != nil {
		log.Fatal("error execute delete ", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		log.Fatal("error getting number of deleted rows")
	}
	fmt.Println(count, "rows deleted")
}

func main() {
	var err error

	mydbEnv := os.Getenv("DATABASE_URL")
	myDB, err = sql.Open("postgres", mydbEnv)

	if err != nil {
		log.Println("Connect to database error", err)
		return
	}
	defer myDB.Close()

	CreateCustomerTable()

	r := gin.Default()
	r.Use(authMiddleware)

	r.POST("/customers", createCustomerHandler)
	r.GET("/customers/:id", getCustomerHandler)
	r.GET("/customers", getAllCustomersHandler)
	r.PUT("/customers/:id", updateCustomerHandler)
	r.DELETE("/customers/:id", deleteCustomerHandler)

	r.Run(":2019")
}

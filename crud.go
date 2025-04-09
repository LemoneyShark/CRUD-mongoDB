package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Employee โครงสร้างข้อมูลพนักงาน
type Employee struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username string             `json:"username" bson:"username"`
	Password int                `json:"password" bson:"password"`
	Skills   []string           `json:"skills" bson:"skills"`
}

// ตัวแปรกลางสำหรับการเชื่อมต่อ MongoDB
var collection *mongo.Collection
var ctx = context.TODO()

func main() {
	// โหลดไฟล์ .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// เชื่อมต่อกับ MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// ทดสอบการเชื่อมต่อ
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Cannot connect to MongoDB:", err)
	}
	fmt.Println("Connected to MongoDB!")

	// เลือกฐานข้อมูลและคอลเลกชัน
	collection = client.Database("company").Collection("employee")

	// สร้าง router
	router := mux.NewRouter()

	// กำหนด routes
	router.HandleFunc("/api/employees", getEmployees).Methods("GET")
	router.HandleFunc("/api/employees", createEmployee).Methods("POST")
	router.HandleFunc("/api/employees/{id}", getEmployee).Methods("GET")
	router.HandleFunc("/api/employees/{id}", updateEmployee).Methods("PUT")
	router.HandleFunc("/api/employees/{id}", deleteEmployee).Methods("DELETE")

	// เริ่มเซิร์ฟเวอร์
	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// getEmployees ดึงข้อมูลพนักงานทั้งหมด
func getEmployees(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// ค้นหาเอกสารทั้งหมด
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Error finding employees"}`))
		return
	}
	defer cursor.Close(ctx)

	var employees []Employee
	if err = cursor.All(ctx, &employees); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Error decoding employees"}`))
		return
	}

	json.NewEncoder(w).Encode(employees)
}

// getEmployee ดึงข้อมูลพนักงานตาม ID
func getEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Invalid ID format"}`))
		return
	}

	var employee Employee
	err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&employee)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Employee not found"}`))
		return
	}

	json.NewEncoder(w).Encode(employee)
}

// createEmployee เพิ่มข้อมูลพนักงานใหม่
func createEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var employee Employee
	// แปลงข้อมูล JSON จาก request body เป็น struct
	err := json.NewDecoder(r.Body).Decode(&employee)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Invalid request body"}`))
		return
	}

	// สร้าง ID ใหม่ถ้าไม่มี
	if employee.ID.IsZero() {
		employee.ID = primitive.NewObjectID()
	}

	// เพิ่มข้อมูลใน MongoDB
	result, err := collection.InsertOne(ctx, employee)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Error inserting employee"}`))
		return
	}

	// สร้าง response กลับไป
	response := struct {
		ID string `json:"id"`
	}{
		ID: result.InsertedID.(primitive.ObjectID).Hex(),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// updateEmployee อัปเดตข้อมูลพนักงาน
func updateEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Invalid ID format"}`))
		return
	}

	var employee Employee
	err = json.NewDecoder(r.Body).Decode(&employee)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Invalid request body"}`))
		return
	}

	// อัปเดตเอกสาร
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"username": employee.Username,
		"password": employee.Password,
		"skills":   employee.Skills,
	}}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Error updating employee"}`))
		return
	}

	if result.MatchedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Employee not found"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Employee updated successfully"}`))
}

// deleteEmployee ลบข้อมูลพนักงาน
func deleteEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Invalid ID format"}`))
		return
	}

	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Error deleting employee"}`))
		return
	}

	if result.DeletedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Employee not found"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Employee deleted successfully"}`))
}
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// โหลดไฟล์ .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// เชื่อมต่อ MongoDB
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
	db := client.Database("company")
	collection := db.Collection("employee")

	// เรียกใช้ฟังก์ชัน
	update(ctx, collection)
}

// สร้างโครงสร้างสำหรับข้อมูล
type Employee struct {
	Username string   `bson:"username"`
	Password int      `bson:"password"`
	Skills   []string `bson:"skills"`
}

// CREATE - เพิ่มข้อมูลใหม่
/*func create(ctx context.Context, collection *mongo.Collection) {
	employees := []interface{}{
		Employee{
			Username: "PP",
			Password: 1122,
			Skills:   []string{"Python", "FastAPI"},
		},
		Employee{
			Username: "Kong",
			Password: 3322,
			Skills:   []string{"JavaScript", "React"},
		},
	}

	insertResult, err := collection.InsertMany(ctx, employees)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted documents with IDs:", insertResult.InsertedIDs)
}*/

// READ - อ่านข้อมูล
func read(ctx context.Context, collection *mongo.Collection) {
	fmt.Println("\n📥 READ: ค้นหาคนชื่อ PP")
	
	var result Employee
	err := collection.FindOne(ctx, bson.M{"username": "PP"}).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Found document: %+v\n", result)
}

// UPDATE - แก้ไขข้อมูล
func update(ctx context.Context, collection *mongo.Collection) {
	fmt.Println("\n🛠️ UPDATE: แก้ password ของ Kong เป็น 9999")
	
	// แสดงข้อมูลก่อนอัปเดต
	fmt.Println("ข้อมูลก่อนอัปเดต:")
	var beforeUpdate Employee
	err := collection.FindOne(ctx, bson.M{"username": "Kong"}).Decode(&beforeUpdate)
	if err != nil {
		log.Printf("ไม่พบเอกสาร Kong ก่อนอัปเดต: %v", err)
	} else {
		fmt.Printf("%+v\n", beforeUpdate)
	}
	
	// ทำการอัปเดต
	filter := bson.M{"username": "Kong"}
	update := bson.M{"$set": bson.M{"password": 9999}}
	
	updateResult, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("เอกสารที่ถูกค้นพบ: %d\n", updateResult.MatchedCount)
	fmt.Printf("เอกสารที่ถูกอัปเดต: %d\n", updateResult.ModifiedCount)
	
	// แสดงข้อมูลหลังอัปเดต
	fmt.Println("ข้อมูลหลังอัปเดต:")
	var afterUpdate Employee
	err = collection.FindOne(ctx, bson.M{"username": "Kong"}).Decode(&afterUpdate)
	if err != nil {
		log.Printf("ไม่พบเอกสาร Kong หลังอัปเดต: %v", err)
	} else {
		fmt.Printf("%+v\n", afterUpdate)
	}
	
	// แสดงข้อมูลทั้งหมดในคอลเลกชัน
	fmt.Println("\nเอกสารทั้งหมดในคอลเลกชัน:")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)
	
	var employees []Employee
	if err = cursor.All(ctx, &employees); err != nil {
		log.Fatal(err)
	}
	
	for _, emp := range employees {
		fmt.Printf("%+v\n", emp)
	}
}

// DELETE - ลบข้อมูล
/*func delete(ctx context.Context, collection *mongo.Collection) {
	fmt.Println("\n🗑️ DELETE: ลบคนชื่อ PP ออก")
	
	deleteResult, err := collection.DeleteOne(ctx, bson.M{"username": "PP"})
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Deleted %d document(s)\n", deleteResult.DeletedCount)
}*/
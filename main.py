from pymongo import MongoClient
from dotenv import load_dotenv
import os

load_dotenv()

# เชื่อมต่อ MongoDB
mongo_uri = os.getenv("MONGO_URI")
client = MongoClient(mongo_uri)
db = client["company"]
collection = db["employee"]


def main():
    update()

def create():
# --------- CREATE (เพิ่มข้อมูลใหม่) ------------
    data = [
    {"username": "PP", "password": 1122, "skills": ["Python", "FastAPI"]},
    {"username": "Kong", "password": 3322 , "skills": ["JavaScript", "React"]}
    ]
    collection.insert_many(data)

# --------- READ (อ่านข้อมูล) --------------------
def read():
    print("\n📥 READ: ค้นหาคนชื่อ PP")
    pp = collection.find_one({"username": "PP"})
    print(pp)

# --------- UPDATE (แก้ไขข้อมูล) -----------------
def update():
    print("\n🛠️ UPDATE: แก้ password ของ Kong เป็น 9999")
    
    # แสดงข้อมูลก่อนอัปเดต
    print("ข้อมูลก่อนอัปเดต:")
    kong_before = collection.find_one({"username": "Kong"})
    print(kong_before)
    
    # ทำการอัปเดตและเก็บผลลัพธ์
    result = collection.update_one({"username": "Kong"}, {"$set": {"password": 2222}})
    
    # แสดงจำนวนเอกสารที่ถูกอัปเดต
    print(f"เอกสารที่ถูกค้นพบ: {result.matched_count}")
    print(f"เอกสารที่ถูกอัปเดต: {result.modified_count}")
    
    # แสดงข้อมูลหลังอัปเดต
    print("ข้อมูลหลังอัปเดต:")
    kong_after = collection.find_one({"username": "Kong"})
    print(kong_after)

# --------- DELETE (ลบข้อมูล) --------------------
def delete():
    print("\n🗑️ DELETE: ลบคนชื่อ PP ออก")
    collection.delete_one({"username": "PP"})

main()
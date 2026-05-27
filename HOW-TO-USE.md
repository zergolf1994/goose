# 📘 How To Use Goose

ขั้นตอนการใช้งาน Goose — MongoDB ODM สำหรับ Go แบบ step-by-step

---

## สารบัญ

- [1. ติดตั้ง](#1-ติดตั้ง)
- [2. เชื่อมต่อ MongoDB](#2-เชื่อมต่อ-mongodb)
- [3. สร้าง Model](#3-สร้าง-model)
- [4. CRUD Operations](#4-crud-operations)
- [5. Query Builder](#5-query-builder)
- [6. Indexes](#6-indexes)
- [7. Struct Tags](#7-struct-tags)
- [8. ตัวอย่างแบบเต็ม](#8-ตัวอย่างแบบเต็ม)

---

## 1. ติดตั้ง

```bash
go get github.com/zergolf1994/goose
```

---

## 2. เชื่อมต่อ MongoDB

### วิธีที่ 1: ใช้ Connection URI

```go
import "github.com/zergolf1994/goose"

func main() {
    // Goose จะ parse ชื่อ database จาก URI ให้อัตโนมัติ
    err := goose.Connect("mongodb://localhost:27017/mydb")
    if err != nil {
        log.Fatal(err)
    }
    defer goose.Close()
}
```

### วิธีที่ 2: ใช้ Database ที่มีอยู่แล้ว

```go
import (
    "github.com/zergolf1994/goose"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    // เชื่อมต่อเอง
    client, _ := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    db := client.Database("mydb")

    // ส่ง database ให้ goose
    goose.SetDB(db)
}
```

### Access client/database ที่เชื่อมต่อแล้ว

```go
client := goose.Client()   // *mongo.Client
db     := goose.DB()        // *mongo.Database
col    := goose.Collection("users")  // *mongo.Collection
```

---

## 3. สร้าง Model

### 3.1 Define Struct

```go
type Media struct {
    goose.BaseModel `bson:",inline"`     // _id, slug, createdAt, updatedAt
    Type    string  `bson:"type"    json:"type"    goose:"default:video"`
    FileID  string  `bson:"fileId"  json:"fileId"  goose:"required,index"`
    Status  string  `bson:"status"  json:"status"  goose:"default:waiting"`
}
```

> **📌 สำคัญ:** ต้อง embed `goose.BaseModel` ด้วย `bson:",inline"` เสมอ
> เพื่อให้ `_id`, `slug`, `createdAt`, `updatedAt` ทำงานอัตโนมัติ

### 3.2 สร้าง Model Instance

```go
var MediaModel = goose.NewModel[Media]("medias")
```

พารามิเตอร์คือชื่อ collection ใน MongoDB

---

## 4. CRUD Operations

### 📝 Create (สร้างข้อมูล)

```go
ctx := context.Background()

// วิธีที่ 1: ใช้ New() + กำหนดค่า + Create()
media := MediaModel.New()      // _id, slug, timestamps ถูกสร้างอัตโนมัติ
media.Type = "video"
media.FileID = "file_abc123"
result, err := MediaModel.Create(ctx, media)

// วิธีที่ 2: สร้าง struct เอง (defaults จาก goose tag จะถูก apply ตอน Create)
media2 := &Media{
    Type:   "audio",
    FileID: "file_xyz789",
}
MediaModel.Create(ctx, media2)

// Insert หลายรายการ
docs := []*Media{
    {Type: "video", FileID: "f1"},
    {Type: "audio", FileID: "f2"},
}
MediaModel.InsertMany(ctx, docs)
```

### 🔍 Read (อ่านข้อมูล)

```go
// หาด้วย filter
media, err := MediaModel.FindOne(ctx, bson.M{"type": "video"})

// หาด้วย _id
media, err := MediaModel.FindByID(ctx, "uuid-string-here")

// หาด้วย slug
media, err := MediaModel.FindBySlug(ctx, "aB3-cD5_eF7")

// หาหลายรายการ
results, err := MediaModel.Find(ctx, bson.M{"status": "waiting"})

// หาทั้งหมด
all, err := MediaModel.Find(ctx, bson.M{})
```

### ✏️ Update (แก้ไขข้อมูล)

```go
// อัปเดต 1 รายการ (updatedAt อัตโนมัติ)
MediaModel.UpdateOne(ctx,
    bson.M{"_id": id},
    bson.M{"$set": bson.M{"type": "audio"}},
)

// อัปเดตด้วย _id (shorthand)
MediaModel.UpdateByID(ctx, id, bson.M{
    "$set": bson.M{"status": "completed"},
})

// อัปเดตหลายรายการ
MediaModel.UpdateMany(ctx,
    bson.M{"status": "waiting"},
    bson.M{"$set": bson.M{"status": "processing"}},
)

// อัปเดตแบบ raw (ไม่ inject updatedAt) — ใช้กับ $push, $pull, $unset
MediaModel.UpdateOneRaw(ctx,
    bson.M{"_id": id},
    bson.M{"$push": bson.M{"tags": "new-tag"}},
)
```

> **💡 หมายเหตุ:** `UpdateOne`, `UpdateByID`, `UpdateMany` จะเพิ่ม `updatedAt: time.Now()` ใน `$set` ให้อัตโนมัติ
> ถ้าไม่ต้องการ ให้ใช้ `UpdateOneRaw`

### 🗑️ Delete (ลบข้อมูล)

```go
// ลบ 1 รายการ
MediaModel.DeleteOne(ctx, bson.M{"_id": id})

// ลบด้วย _id (shorthand)
MediaModel.DeleteByID(ctx, id)

// ลบหลายรายการ
MediaModel.DeleteMany(ctx, bson.M{"status": "expired"})
```

### 🔢 Count & Exists

```go
// นับจำนวน
count, err := MediaModel.CountDocuments(ctx, bson.M{"type": "video"})

// เช็คว่ามีข้อมูลหรือไม่
exists, err := MediaModel.Exists(ctx, bson.M{"fileId": "abc"})
if exists {
    fmt.Println("Found!")
}
```

### 📊 Aggregation

```go
cursor, err := MediaModel.Aggregate(ctx, mongo.Pipeline{
    {{"$match", bson.M{"type": "video"}}},
    {{"$group", bson.M{
        "_id":   "$status",
        "count": bson.M{"$sum": 1},
    }}},
})
defer cursor.Close(ctx)

var results []bson.M
cursor.All(ctx, &results)
```

---

## 5. Query Builder

Query Builder ช่วยให้เขียน query แบบ chain ได้สะดวก:

### พื้นฐาน

```go
// หาหลายรายการ
results, err := MediaModel.Query(bson.M{"type": "video"}).
    SortDesc("createdAt").
    Limit(10).
    Exec(ctx)

// หา 1 รายการ
doc, err := MediaModel.Query(bson.M{"slug": "abc"}).One(ctx)

// นับจำนวน
count, err := MediaModel.Query(bson.M{"type": "video"}).Count(ctx)
```

### Sort (เรียงลำดับ)

```go
// แบบระบุทิศทาง
.Sort("createdAt", -1)   // -1 = DESC
.Sort("name", 1)         // 1 = ASC

// แบบ shorthand
.SortDesc("createdAt")   // เรียงจากใหม่ไปเก่า
.SortAsc("name")         // เรียงจาก A-Z

// Sort หลาย field
.SortDesc("createdAt").SortAsc("name")
```

### Pagination (แบ่งหน้า)

```go
// แบบ manual
.Skip(20).Limit(10)

// แบบใช้ Page helper (1-indexed)
.Page(1, 10)   // หน้า 1, แสดง 10 รายการ
.Page(2, 10)   // หน้า 2, skip 10 แสดง 10 รายการ
.Page(3, 25)   // หน้า 3, skip 50 แสดง 25 รายการ
```

### Projection (เลือก field)

```go
// เลือกเฉพาะ field ที่ต้องการ
.Select("name", "status", "createdAt")

// ซ่อน field ที่ไม่ต้องการ
.Exclude("password", "secret")
```

### ตัวอย่างรวม

```go
// ค้นหา video ที่ status = waiting, เรียงจากใหม่ไปเก่า, หน้า 2
results, err := MediaModel.Query(bson.M{
    "type":   "video",
    "status": "waiting",
}).
    SortDesc("createdAt").
    Page(2, 20).
    Select("name", "type", "status", "createdAt").
    Exec(ctx)
```

---

## 6. Indexes

### Auto Index จาก Struct Tags

```go
type User struct {
    goose.BaseModel `bson:",inline"`
    Email  string   `bson:"email"  goose:"unique"`   // unique index
    Status string   `bson:"status" goose:"index"`     // regular index
}

var UserModel = goose.NewModel[User]("users")

// เรียกตอน startup (หลัง Connect)
UserModel.EnsureIndexes(ctx)
```

### Compound Index

```go
// Regular compound index
MediaModel.EnsureCompoundIndex(ctx, []string{"type", "status"}, false)

// Unique compound index
MediaModel.EnsureCompoundIndex(ctx, []string{"userId", "email"}, true)
```

---

## 7. Struct Tags

### `goose` Tag ทั้งหมดที่รองรับ

| Tag | คำอธิบาย | ใช้กับ Type |
|---|---|---|
| `default:uuid` | สร้าง UUID อัตโนมัติ | `string` |
| `default:random(N)` | สร้าง random slug ความยาว N | `string` |
| `default:now` | ใส่เวลาปัจจุบัน | `time.Time` |
| `default:<value>` | ค่า default แบบ literal | `string`, `int`, `float64`, `bool` |
| `required` | ระบุว่า field นี้จำเป็น (metadata) | ทุก type |
| `unique` | สร้าง unique index | ทุก type |
| `index` | สร้าง regular index | ทุก type |
| `ref:<collection>` | อ้างอิง collection อื่น (metadata) | `string` |

### ใช้หลาย tag พร้อมกัน

```go
type File struct {
    goose.BaseModel `bson:",inline"`

    // UUID + unique
    Code string `bson:"code" goose:"default:uuid,unique"`

    // Random slug + unique
    Slug string `bson:"slug" goose:"default:random(8),unique"`

    // Literal default + index
    Status string `bson:"status" goose:"default:pending,index"`

    // Required + reference
    OwnerID string `bson:"ownerId" goose:"required,ref:users"`

    // Default int
    Priority int `bson:"priority" goose:"default:0"`

    // Default bool
    Active bool `bson:"active" goose:"default:true"`

    // Default time
    StartAt time.Time `bson:"startAt" goose:"default:now"`
}
```

### Inspect Schema (Debug)

```go
// ดู schema ของ model
fmt.Println(goose.DescribeSchema[File]())

// Output:
//   _id: default=uuid
//   slug: default=random(11) unique
//   createdAt: default=now
//   updatedAt: default=now
//   code: default=uuid unique
//   status: default=pending index
//   ownerId: required ref=users
```

---

## 8. ตัวอย่างแบบเต็ม

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/zergolf1994/goose"
    "go.mongodb.org/mongo-driver/bson"
)

// ── Define Models ──

type User struct {
    goose.BaseModel `bson:",inline"`
    Name   string   `bson:"name"   json:"name"   goose:"required"`
    Email  string   `bson:"email"  json:"email"  goose:"required,unique"`
    Role   string   `bson:"role"   json:"role"   goose:"default:member"`
}

type Post struct {
    goose.BaseModel `bson:",inline"`
    Title    string `bson:"title"    json:"title"    goose:"required"`
    Content  string `bson:"content"  json:"content"`
    AuthorID string `bson:"authorId" json:"authorId" goose:"required,index,ref:users"`
    Status   string `bson:"status"   json:"status"   goose:"default:draft,index"`
}

var (
    UserModel = goose.NewModel[User]("users")
    PostModel = goose.NewModel[Post]("posts")
)

func main() {
    ctx := context.Background()

    // ── Connect ──
    if err := goose.Connect("mongodb://localhost:27017/blog"); err != nil {
        log.Fatal(err)
    }
    defer goose.Close()

    // ── Ensure Indexes ──
    UserModel.EnsureIndexes(ctx)
    PostModel.EnsureIndexes(ctx)

    // ── Create User ──
    user := UserModel.New()
    user.Name = "สมชาย"
    user.Email = "somchai@example.com"
    UserModel.Create(ctx, user)
    fmt.Printf("Created user: %s (slug: %s)\n", user.ID, user.Slug)

    // ── Create Posts ──
    for i := 1; i <= 5; i++ {
        post := PostModel.New()
        post.Title = fmt.Sprintf("บทความที่ %d", i)
        post.Content = "เนื้อหา..."
        post.AuthorID = user.ID
        PostModel.Create(ctx, post)
    }

    // ── Query: หาบทความของ user, เรียงจากใหม่สุด, หน้าแรก ──
    posts, _ := PostModel.Query(bson.M{"authorId": user.ID}).
        SortDesc("createdAt").
        Page(1, 3).
        Select("title", "status", "createdAt").
        Exec(ctx)

    fmt.Printf("\nPosts (page 1, 3 per page):\n")
    for _, p := range posts {
        fmt.Printf("  - %s [%s] %s\n", p.Title, p.Status, p.CreatedAt.Format(time.RFC3339))
    }

    // ── Count ──
    total, _ := PostModel.CountDocuments(ctx, bson.M{"authorId": user.ID})
    fmt.Printf("\nTotal posts: %d\n", total)

    // ── Update ──
    if len(posts) > 0 {
        PostModel.UpdateByID(ctx, posts[0].ID, bson.M{
            "$set": bson.M{"status": "published"},
        })
        fmt.Printf("\nPublished: %s\n", posts[0].Title)
    }

    // ── Find updated ──
    published, _ := PostModel.Query(bson.M{
        "authorId": user.ID,
        "status":   "published",
    }).Exec(ctx)
    fmt.Printf("Published posts: %d\n", len(published))

    // ── Exists ──
    hasAdmin, _ := UserModel.Exists(ctx, bson.M{"role": "admin"})
    fmt.Printf("Has admin: %v\n", hasAdmin)
}
```

---

## 💡 Tips & Best Practices

1. **เรียก `EnsureIndexes` ตอน startup** — ทำครั้งเดียวหลัง `Connect()`
2. **ใช้ `New()` เพื่อสร้าง doc** — defaults ทั้งหมดจะถูก apply ให้
3. **ใช้ Query Builder แทน `Find`** — อ่านง่ายกว่าและรองรับ pagination
4. **ใช้ `UpdateOneRaw` สำหรับ `$push`, `$pull`, `$unset`** — เพราะไม่ต้องการ auto `updatedAt`
5. **ใช้ `bson:",inline"` กับ `BaseModel` เสมอ** — ไม่งั้น fields จะ nested ผิด

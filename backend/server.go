package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"fmt"
	"database/sql"
	"net/http"
	"syscall"
	"time"
	"net/smtp"
)

type DB struct{
	db *sql.DB
}

type CheckinRecord struct {
	Name string `json:"name" form:"name" binding:"required"`
	CheckinDate string `json:"checkinDate" form: "checkinDate" binding:"required"`
}

type CheckinImage struct {
	CheckinDate     string    `json:"checkinDate"`
	Name  			string    `json:"name"`
	ImgUrl			string	  `json:"imgUrl"`
}

type ReturnImage struct {
	CheckinDate     string    `json:"checkinDate"`
	ImgUrl			string	  `json:"imgUrl"`
}

type User struct {
	name	string
	email	string
}

func CheckIn(db *DB) gin.HandlerFunc{

	fn := func(c *gin.Context) {

		code := 200
		msg := "success"
		DB_NAME := "checkin"

		checkinDate := c.Query("checkinDate")
		name := c.Query("name")

		_, err := db.db.Exec(fmt.Sprintf(`
			INSERT INTO %s(checkin_date, name) VALUES
				('%s', '%s')
		`, DB_NAME, checkinDate, name))

		if err != nil {
			log.Print(err)
			code = 400
			msg = "failed"
		}

		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg": msg,
		})
	}
	return gin.HandlerFunc(fn)
}


func GetCheckinRecord(db *DB) gin.HandlerFunc{

	fn := func(c *gin.Context) {

		code := 200
		DB_NAME := "checkin"

		var cnt int
		_ = db.db.QueryRow(fmt.Sprintf(`
			SELECT COUNT(*) FROM %s
		`, DB_NAME)).Scan(&cnt)

		rows, err := db.db.Query(fmt.Sprintf(`
			SELECT name, 
			TO_CHAR(checkin_date, 'YYYY-MM-DD') AS formatted_date
			FROM %s
		`, DB_NAME))
		if err != nil {
			log.Print(err)
			code = 400
		}
		defer rows.Close()

		var (
			name string
			checkinDate string
		)

		checkinRecordsList := make([]CheckinRecord, cnt)
		iter := 0
		for rows.Next() {
			err := rows.Scan(&name, &checkinDate)
			checkinRecord := CheckinRecord{
				Name: name,
				CheckinDate: checkinDate,
			}
			checkinRecordsList[iter] = checkinRecord
			iter = iter + 1
			if err != nil {
				log.Print(err)
				code = 400
			}
		}

		//fmt.Println("ok this is the record", checkinRecordsList)

		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"checkinRecord": checkinRecordsList,
		})
	}
	return gin.HandlerFunc(fn)
}

func UploadImages(c *gin.Context) {
	SAVE_DST := "./images"

	code := 200
	file,_ := c.FormFile("file")
	err := c.SaveUploadedFile(file, fmt.Sprintf("%s/%s", SAVE_DST, file.Filename))
	if err != nil{
		log.Print(err)
		code = 400
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg": fmt.Sprintf("'%s' uploaded!", file.Filename),
	})
}

func UploadImagesDB(db *DB) gin.HandlerFunc{
	fn := func(c *gin.Context) {
		code := 200
		IMG_DB_NAME := "checkin_imgs"

		var checkinImage CheckinImage
		err := c.BindJSON(&checkinImage)
		if err != nil {
			log.Print("Bind json error ")
			log.Print(err)
		}

		checkinDate, name, imgUrl := checkinImage.CheckinDate, checkinImage.Name, checkinImage.ImgUrl

		_, err = db.db.Exec(fmt.Sprintf(`
			INSERT INTO %s(checkin_date, name, img_url) VALUES
				('%s', '%s', '%s')
		`, IMG_DB_NAME, checkinDate, name, imgUrl))

		if err != nil {
			log.Print(err)
			code = 400
		}

		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg": fmt.Sprintf("'%s' uploaded to db", imgUrl),
		})
	}
	return gin.HandlerFunc(fn)
}

func GetImages(db *DB) gin.HandlerFunc{
	fn := func(c *gin.Context) {
		code := 200
		name := c.DefaultQuery("name", "empty")
		IMG_DB_NAME := "checkin_imgs"

		rows, err := db.db.Query(fmt.Sprintf(`
			SELECT img_url, checkin_date
			FROM %s
			WHERE name = '%s'
			ORDER BY checkin_date DESC
		`, IMG_DB_NAME, name))

		if err != nil {
			log.Print(err)
		}

		returnImages := make([]ReturnImage, 0)

		var tempImage string
		var tempDate time.Time

		for rows.Next() {
			err := rows.Scan(&tempImage, &tempDate)
			if err != nil {
				log.Print(err)
				code = 400
			}
			temp := ReturnImage{
				CheckinDate : tempDate.Format("2006-01-02"),
				ImgUrl: tempImage,
			}
			returnImages = append(returnImages, temp)
		}

		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg": "success",
			"images": returnImages,
		})
	}
	return gin.HandlerFunc(fn)
}

func sendEmail(recipients []User) {
	host_name := "smtp.gmail.com"
	auth := smtp.PlainAuth("", "howardchu95@gmail.com", "pmdf eplw fome sunx", host_name)
	from := "From: Sberm 打卡网站"

	ct := time.Now()
	date := ct.Format("2006年01月02日")

	for _, recipient := range recipients {
		fmt.Println("Sending email notification to:", recipient.name)
		// Sprintf不能和\r\n一起使用
		subject := fmt.Sprintf("Subject: 您%s没有打卡",date)
		mime := "MIME-version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\n\r\n"
		body := fmt.Sprintf(`<html>
<body style="background-color: #B4B4B4;">
	<div style="background-color: white; border-radius: 20px; margin: 2rem 2rem;">
		<p><span>%s 先生/女士，您今天(%s)还没有在学习打卡网站</span>
		<span>sberm.cn/checkin</span>
		<span>上打卡，请您在空闲时补打卡，谢谢。</span></p>
	<div>
</body>
</html>`, recipient.name, ct.Format("2006-01月02日")[5:])
		msg := []byte(from + "\r\n" + subject + "\r\n" + mime + body)

		err := smtp.SendMail(host_name+":587", auth, from, []string{recipient.email}, msg)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func EmailNotif(db *DB) {

	// 发邮件的时间
	timeToSendEmail := map[string]int {
		"hour": 21,
		"minute": 30,
		"second": 0,
	}

	for {
		ct := time.Now()
		h, m, s := ct.Clock()
		date := ct.Format("2006-01-02")

		// 若到点，开始发邮件
		if h == timeToSendEmail["hour"] &&
		   m == timeToSendEmail["minute"] &&
		   s == timeToSendEmail["second"] {
		
			// 取所有users和他们的email
			IMG_DB_NAME := "users"
			rows, err := db.db.Query(fmt.Sprintf(`
				SELECT name, email 
				FROM %s
			`, IMG_DB_NAME))
			if err != nil {
				log.Print(err)
			}
			var users []User
			for rows.Next() {
				var name string
				var email string
				err := rows.Scan(&name, &email)
				if err != nil {
					log.Print(err)
				}
				users = append(users, User {
					name: name,
					email: email,
				})
			}

			// 查找需要发邮件提醒的用户(未打卡用户)
			dbName := "checkin"
			var recipients []User
			checked := false
			for _,user := range users {
				err := db.db.QueryRow(fmt.Sprintf(`
					SELECT EXISTS(SELECT * FROM %s
					WHERE checkin_date = '%s'
					AND name = '%s'
					LIMIT 1)
				`, dbName, date, user.name)).Scan(&checked)
				if err != nil {
					log.Print(err)
				}
				if !checked {
				   recipients = append(recipients, user)
				}
			}
			sendEmail(recipients)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func ConnectDB(db *DB) {
	connStr := "user=root dbname=root password=112358 sslmode=disable"
	db_, err := sql.Open("postgres", connStr)
	db.db =db_
	if err != nil {
		log.Fatal(err)
	}
}

func StartRouters(db *DB) {
	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20  // 8 MiB
	r.GET("/checkin-checkin", CheckIn(db))
	r.GET("/checkin-record", GetCheckinRecord(db))
	r.POST("/checkin-upload-imgs", UploadImages)
	r.POST("/checkin-upload-imgs-db", UploadImagesDB(db))
	r.GET("/checkin-get-imgs", GetImages(db))
	r.Run("127.0.0.1:14175");
}

func main() {
	syscall.Chdir("/root/hw/checkin")
	var db *DB = new(DB)
	ConnectDB(db)
	go EmailNotif(db)
	StartRouters(db)
	defer db.db.Close()
}

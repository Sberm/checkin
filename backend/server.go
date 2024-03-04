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
	ImgUrls			[]string  `json:"imgUrls"`
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
	form, _ := c.MultipartForm()
	files := form.File["file"]

	for _, file := range files {
		log.Println(file.Filename)

		// Upload the file to specific dst.
		err := c.SaveUploadedFile(file, fmt.Sprintf("%s/%s", SAVE_DST, file.Filename))
		if err != nil{
			log.Print(err)
			code = 400
			break
		}
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg": fmt.Sprintf("'%d' files uploaded", len(files)),
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

		checkinDate, name, imgUrls := checkinImage.CheckinDate, checkinImage.Name, checkinImage.ImgUrls

		sql := fmt.Sprintf("INSERT INTO %s (checkin_date, name, img_url) VALUES", IMG_DB_NAME)

		for i, url := range imgUrls {
			if (i != len(imgUrls) - 1) {
				sql += fmt.Sprintf(`('%s', '%s', '%s'),`, checkinDate, name, url)
			} else {
				sql += fmt.Sprintf(`('%s', '%s', '%s');`, checkinDate, name, url)
			}
		}

		_, err = db.db.Exec(sql)

		if err != nil {
			log.Print(err)
			code = 400
		}

		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg": fmt.Sprintf("%d images uploaded to db", len(imgUrls)),
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

func sendEmail(recipients []User, consec []int) {
	host_name := "smtp.gmail.com"
	auth := smtp.PlainAuth("", "howardchu95@gmail.com", "pmdf eplw fome sunx", host_name)
	from := "From: Sberm 打卡网站"

	fmt_str := "2006年01月02日"
	ct := time.Now()
	date := ct.Format(fmt_str)[8:]
	var body string
	for i, recipient := range recipients {

		// log
		fmt.Sprintln("Sending email notification to: %s, miss check-in %d days", recipient.name, consec[i])

		subject := fmt.Sprintf("Subject: 您%s没有打卡",date)
		mime := "MIME-version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\n\r\n"
		if (consec[i] == 1) {
			body = fmt.Sprintf(`
				<p><span>&nbsp;&nbsp;&nbsp;&nbsp;</span><span class="hl">%s</span><span>先生/女士，您今天(</span><span class="hl">%s</span>)<span>还没有在学习打卡网站</span>
				<span>sberm.cn/checkin</span>
				<span>上打卡，请您在空闲时补打卡，谢谢。</span></p>`, recipient.name, date)
		} else {
			body = fmt.Sprintf(`
				<p><span>&nbsp;&nbsp;&nbsp;&nbsp;</span><span class="hl">%s</span><span>先生/女士，您截止至(</span><span class="hl">%s</span>)<span>已经连续</span><span class="hl">%d</span><span>天没有在学习打卡网站</span>
				<span>sberm.cn/checkin</span>
				<span>上打卡，请您在空闲时补打卡，谢谢。</span></p>`, recipient.name, date, consec[i])
		}

		html_txt := fmt.Sprintf(`
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<style>
.box {
	background-color: white;
	border-radius: 20px;
    padding: 0.6rem;
}
.bg {
	background-color: #ECECEC;
	border-radius: 20px;
	padding: 2rem;
}
.hl {
	color: #1bc700;
}
.hd {
	margin-bottom: 1rem;
}
.hdt {
	font-size: 25px;
	font-weight: bold;
}
</style>
</head>
<body>
<div class="bg">
<div class="box hd">
	<p class="hdt"><span class="hl">Sberm打卡网站</span> ✅</p>
</div>
<div class="box">
%s
</div>
</div>
</body>
</html>
`, body)

		msg := []byte(from + "\r\n" + subject + "\r\n" + mime + html_txt)

		err := smtp.SendMail(host_name+":587", auth, from, []string{recipient.email}, msg)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func testEmail(db *DB) {

	fmt.Println("Email test")

	ct := time.Now()
	date := ct.Format("2006-01-02")

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
	var consec []int
	checked := false
	for _,user := range users {
		if user.name != "朱皓炜" {
			continue;
		}
		err := db.db.QueryRow(fmt.Sprintf(`
			SELECT EXISTS(SELECT * FROM %s
			WHERE checkin_date = '%s'
			AND name = '%s'
			LIMIT 1)
		`, dbName, date, user.name)).Scan(&checked)
		if err != nil {
			log.Print(err)
		}
		// if a user didn't check in
		if !checked {
			var last_checkin time.Time
			err := db.db.QueryRow(fmt.Sprintf(`
				SELECT MAX(checkin_date) FROM %s
				WHERE checkin_date <= '%s'
				AND name = '%s'
			`, dbName, date, user.name)).Scan(&last_checkin)
			var y2 int
			var m2 time.Month 
			var d2 int
			if err == sql.ErrNoRows {
				y2, m2, d2 = 2023, time.Month(9), 24
			} else if err != nil {
				log.Print(err)
			} else {
				y2, m2, d2 = last_checkin.Date()
			}

			// set time in date to 00:00
			y1, m1, d1 := ct.Date()
			date1 := time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
			date2 := time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC)
			diff := int(date1.Sub(date2).Hours()) / 24

			recipients = append(recipients, user)
			consec = append(consec, diff)
		}
		sendEmail(recipients, consec)
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
			var consec []int
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
				// if a user didn't check in
				if !checked {
					var last_checkin time.Time
					err := db.db.QueryRow(fmt.Sprintf(`
						SELECT MAX(checkin_date) FROM %s
						WHERE checkin_date <= '%s'
						AND name = '%s'
					`, dbName, date, user.name)).Scan(&last_checkin)
					var y2 int
					var m2 time.Month 
					var d2 int
					if err == sql.ErrNoRows {
						y2, m2, d2 = 2023, time.Month(9), 24
					} else if err != nil {
						log.Print(err)
					} else {
						y2, m2, d2 = last_checkin.Date()
					}

					// set time in date to 00:00
					y1, m1, d1 := ct.Date()
					date1 := time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
					date2 := time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC)
					diff := int(date1.Sub(date2).Hours()) / 24

					recipients = append(recipients, user)
					consec = append(consec, diff)
				}
			}
			sendEmail(recipients, consec)
			// make sure it doesn't send email twice
			time.Sleep(1 * time.Second)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func Cors() gin.HandlerFunc {
	fn := func (c* gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		if c.Request.Method == "OPTIONS"  {
			c.JSON(http.StatusOK, "")
		}
		c.Next()
	}
	return gin.HandlerFunc(fn)
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
	r.Use(Cors())
	r.Run("127.0.0.1:14175");
}

func main() {
	syscall.Chdir("/root/hw/checkin")

	var db *DB = new(DB)
	ConnectDB(db)

	go EmailNotif(db)
	// test 
	// testEmail(db)

	StartRouters(db)
	defer db.db.Close()
}

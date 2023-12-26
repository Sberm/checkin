package main

import (
	   "net/smtp"
	   "log"
	   "fmt"
)

var (
	from       = ""
	recipients = []string{"1007273067@qq.com"}
)

func main() {
	 // Set up authentication information.
	 host_name := "smtp.gmail.com"
	 auth := smtp.PlainAuth("", "howardchu95@gmail.com", "pmdf eplw fome sunx", host_name)
	 date := "12月11日"

	 from := "From: Sberm 打卡网站"
	 // Sprintf不能和\r\n一起使用
	 subject := fmt.Sprintf("Subject: 您%s没有打卡",date)
	 mime := "MIME-version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\n\r\n"
	 body := `
<html>
<body style="background-color: #B4B4B4;">
	<div style="background-color: white; border-radius: 20px; margin: 2rem 2rem;">
		<p><span>陈梓恒先生/女士，您在11月29日当天没有在学习打卡网站</span>
		<span>sberm.cn/checkin</span>
		<span>上打卡。请您在空闲时补打卡.</span></p>
	<div>
</body>
</html>`
	 msg := []byte(from + "\r\n" + subject + "\r\n" + mime + body)

	 //fmt.Println(string(msg))

	 err := smtp.SendMail(host_name+":587", auth, from, recipients, msg)
	 if err != nil {
	 	log.Fatal(err)
	 }
}

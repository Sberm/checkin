import "./checkin.css"
import React, {useEffect, useState} from "react"
import * as MD5 from "./md5_min.js"

let name
const today = new Date(Date.now())
const offset = today.getTimezoneOffset()
const checkinDate = new Date(today.getTime() - (offset*60*1000)).toISOString().split('T')[0]
let imagePointer = ""

async function fetchImages(setImages) {
	try {
		const url = `https://sberm.cn/checkin-get-imgs?name=${name}`
		fetch(url, {
			headers: {
				Accept: 'application/json',
			},
			mode: "cors"
		})
		.then(response => {
			return response.json()
		})
		.then(data => {
			//setImages([...data.images.slice(0, imageIndex - 1)])
			setImages(cut(data.images, 0, 30))
		})
	} catch(err) {
		console.log(err)
	}
}

// (index 0, 30 days) -- (index 1, 30 days)
function cut(data, index, days) {
	// 一页显示30天内的照片
	let i = 0
	let imageIndex = 0
	let tempDate = new Date(data.length ? data[0].checkinDate : today)
	tempDate.setHours(1,1,4,5)
	let left = 0
	for (let k = 0; k <= index; k++) {
		left = imageIndex
		while (i < days && imageIndex < data.length) {
			const d = new Date(data[imageIndex].checkinDate)
			d.setHours(0,0,0,0)
			if (compareDate(d, tempDate) !== 0) {
				i += 1
				tempDate.setTime(d.getTime())
			}
			imageIndex += 1
		}
		i = 0
	}
	return [...data.slice(left, imageIndex - 1)]
}

async function uploadImages(formData, setPercent) {

	const p = new Promise(function (resolve, reject) {
		const url = "https://sberm.cn/checkin-upload-imgs"
		let xhr = new XMLHttpRequest()

		xhr.onload = () => {
			const res = JSON.parse(xhr.response)
			resolve(res);
		}

		xhr.upload.addEventListener("progress", (event) => {
			let complete = (event.loaded / event.total * 100 | 0)
			setPercent(complete)
		})

		xhr.upload.addEventListener("load", (event) => {
		});

		xhr.open("POST", url, true)
		xhr.send(formData)
	})

	let r = await p;

	return r.code
}

async function uploadImagesDB(imgUrls) {
	try {
		const imgData = {
			checkinDate: checkinDate,
			name: name,
			imgUrls: imgUrls,
		}

		const url = "https://sberm.cn/checkin-upload-imgs-db"
		const responseR = await fetch(url, {
			method: "POST",
			headers: {
				Accept: 'application/json',
			},
			mode: "cors",
			body: JSON.stringify(imgData)
		})

		const response = await responseR.json()
		return response.code

	} catch(err) {
		console.log(err)
	}
}

async function checkin() {
	try {
		const url = `https://sberm.cn/checkin-checkin?checkinDate=${checkinDate}&name=${name}`

		const responseR = await fetch(url, {
			headers: {
				Accept: 'application/json',
			},
			mode: "cors"
		})

		const response = await responseR.json()
		//console.log("checkin return data:", response)
		return response.code
	} catch(err) {
		console.log(err)
	}
}

function CheckinButton() {
	return (
		<button id = "upload-button" onClick={
			async () => {
				let codeB = await checkin() === 200 ? true : false
				if(codeB) {
					alert("签到成功")
					window.location.reload();
				} else {
					alert("签到失败")
				}
			}
		}>签到</button>
	)
}

async function md5FromFile() {
	let reader = new FileReader();
	let fileStr;
	const FileStrStart = 1145;
	const FileStrEnd = 1919;
	const readFilePromise = new Promise((resolve) => {
		const func = () => {
			fileStr = reader.result.substring(FileStrStart > reader.result.length ? reader.result.length : FileStrStart, FileStrEnd > reader.result.length ? reader.result.length : FileStrEnd);
			reader.removeEventListener("load", func);
			resolve();
		}
		reader.addEventListener(
			"load",
			func,
			false);
	})
	reader.readAsText(document.getElementById("checkin-image").files[0]);
	await readFilePromise;
	return MD5.hex_md5(fileStr);
}

function UploadButton(props) {
	return (
		<button id = "upload-button" onClick={
			async () => {
				const reg = /.jpg|.jpeg|.png|.bmp|.ico|.gif|.svg$/i
				if (document.getElementById("checkin-image").files.length === 0 ) {
					alert("请选择图片")
					return
				}

				let formData = new FormData()
				let imgUrls = []
				const files = document.getElementById("checkin-image").files
				for (let i = 0;i < files.length; i++) {
					const file = files[i];
					if (file.name.match(reg) == null) {
						alert("请选择图片")
						return
					}

					const md5Suffix = await md5FromFile();

					const fileName = file.name;
					const lastDot = fileName.lastIndexOf(".");
					const name = fileName.substring(0, lastDot); 
					const ext = fileName.substring(lastDot+1);

					// write image file
					const finalName = `${name}-${md5Suffix}.${ext}`
					const imgFile = new File([file], finalName, {
						type: file.type,
						lastModified: file.lastModified,
					});
					formData.append('file', imgFile);
					imgUrls.push(finalName)
				}

				props.setUploading(true)
				props.setUploadStatus("上传中")

				let a = await uploadImages(formData, props.setPercent) === 200 ? true : false
				let b = await uploadImagesDB(imgUrls) === 200 ? true : false
				let c = await checkin() === 200 ? true : false
				let code = a && b && c

				if(code) {
					props.setUploadStatus("上传成功")
					const getImages = async () => {
						await fetchImages(props.setImages)
					}
					getImages()

					// 清空文件input栏
					document.getElementById("checkin-image").value = null
				} else {
					props.setUploadStatus("上传失败")
				}

				setTimeout(() => props.setUploading(false), 3000)
			}
		}>上传图片</button>
	)
}

function PopUp(props) {
	return (
		<>
			<div className="pop-up">
				<a href="javascript:void()">
					<div id="quit-button" onClick={
							 () => {
								 props.setPop(false)
							 }
						 }><img src="/images/delete-button.png" />
					</div>
				</a>
				<img src={imagePointer}/>
			</div>
		</>
	)
}

function ProgressBar(props) {
	const progress = props.uploadPercent
	const statuss = props.uploadStatus
	return (
		<>
			<div style={{maxHeight: "50px", width: "100%", position: "fixed", top: "20px", display: "flex", justifyContent: "center", alignItems: "center"}}>
				<div className="pbarw" style={{padding: "0 16px", backgroundColor: "white", borderRadius: "20px", display: "flex", alignItems: "center", justifyContent: "center", gap: "10px"}}>
					<span style={{width: "fix-content", fontWeight: "bold"}}>{statuss}</span>
					<div className="progress-bar">
							<div style={{borderRadius: "20px", height: "9px", width: `${progress}%`, backgroundColor: "#58DB43"}}></div>
					</div>
					<span>{`${progress}%`}</span>
				</div>
			</div>
		</>
	)
}

function compareDate(d1, d2) {
	const d1t = d1.getTime()
	const d2t = d2.getTime()
	if (d1 < d2) {return -1}
	else if (d1 > d2) {return 1}
	else {return 0}
}

export default function PageRoot() {
	let [pop, setPop] = useState(false)
	let [images, setImages] = useState([])
	// progress bar
	let [uploading, setUploading] = useState(false)
	let [uploadPercent, setPercent] = useState(0)
	let [uploadStatus, setUploadStatus] = useState("上传中")

	useEffect(() => {
		document.title = '点击签到';
		const urlParams = new URLSearchParams(window.location.search);
		name = urlParams.get('name');

		const getImages = async () => {
			await fetchImages(setImages)
		}
		getImages()

		return () => {}
	}, [])

	// 初始化临时日期
	let tempDate = new Date(images.length ? images[0].checkinDate : today)
	tempDate.setHours(1,1,4,5)

	return (
		<>
			{uploading && <ProgressBar uploadPercent={uploadPercent} uploadStatus={uploadStatus}/>}

			<h1>点击签到</h1>
			<div className="checkin-background">
				<a href="javascript:void()">
					<input type="file" id="checkin-image" name="checkin-image" multiple accept="image/*" /><br/>
				</a>
				<CheckinButton />&nbsp;&nbsp;
			 	<UploadButton setUploading={setUploading} setPercent={setPercent} setUploadStatus={setUploadStatus} setImages={setImages}/>

				<p>历史截图:</p>
				<div className="checkin-imgs" key={images}>
					{images.map((image) => {
						let imgUrl = ""
						if(image.imgUrl.slice(0, 4) !== "http") {
							imgUrl = "/images/"+image.imgUrl
						}
						let dateTitle = false
						const d = new Date(image.checkinDate)
						d.setHours(0,0,0,0)
						if (compareDate(d, tempDate) != 0) {
							dateTitle = true
							tempDate.setTime(d.getTime())
						}
						return (
							<>
								{dateTitle && <p className="checkin-date-header date">{`${d.getMonth()+1}月${d.getDate()}日`}</p>}
								<div className="checkin-imgs-container">
									<img src = {imgUrl} alt="check in pictures" onClick={() => {
										imagePointer = imgUrl
										setPop(true)
									}}/>
								</div>
							</>
						)
					})}
				</div>
			</div>
			{pop && <PopUp setPop={setPop}/>}
		</>
	)
}

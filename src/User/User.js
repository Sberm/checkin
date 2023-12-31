import "./checkin.css"
import React, {useEffect, useState} from "react"
import * as MD5 from "./md5_min.js"

let name
const today = new Date(Date.now())
const offset = today.getTimezoneOffset()
const checkinDate = new Date(today.getTime() - (offset*60*1000)).toISOString().split('T')[0]
let imagePointer = ""
let imgFile;

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
			if (compareDate(d, tempDate) != 0) {
				i += 1
				tempDate.setTime(d.getTime())
			}
			imageIndex += 1
		}
		i = 0
	}
	//console.log("pd",[...data.slice(left, imageIndex - 1)])
	return [...data.slice(left, imageIndex - 1)]
}

async function uploadImages() {
	try {
		const url = `https://sberm.cn/checkin-upload-imgs`
		//const file = document.getElementById("checkin-image").files[0]
		const formData = new FormData()
		formData.append("file", imgFile)

		const responseR = await fetch(url, {
			method: "POST",
			headers: {
				Accept: 'application/json',
			},
			mode: "cors",
			body: formData
		})

		const response = await responseR.json()
		return response.code
	} catch(err) {
		console.log(err)
	}
}

async function uploadImagesDB() {
	try {
		const imgUrl = imgFile.name;
		
		const imgData = {
			checkinDate: checkinDate,
			name: name,
			imgUrl: imgUrl
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
				const file = document.getElementById("checkin-image").files[0];
				if (file.name.match(reg) == null) {
					alert("请选择图片")
					return
				}

				const md5Suffix = await md5FromFile();
				//console.log("md5Suffix,",md5Suffix);

				const fileName = file.name;
				const lastDot = fileName.lastIndexOf(".");
				const name = fileName.substring(0, lastDot); 
				const ext = fileName.substring(lastDot+1);
				
				// write image file
				imgFile = new File([file], `${name}-${md5Suffix}.${ext}`, {
					type: file.type,
					lastModified: file.lastModified,
				});

				let codeB = await uploadImages() === 200 ? true : false
				codeB = codeB && await uploadImagesDB() === 200 ? true : false
				codeB = codeB && await checkin() === 200 ? true : false

				if(codeB) {
					alert("上传成功")
					window.location.reload();
				} else {
					alert("上传失败")
				}
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
			<h1>点击签到</h1>
			<div className="checkin-background">
				<a href="javascript:void()">
					<input type="file" id="checkin-image" name="checkin-image" accept="image/*" /><br/>
				</a>
				<CheckinButton />&nbsp;&nbsp;
			 	<UploadButton />
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

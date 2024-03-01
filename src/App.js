import "./checkin.css"
import React, {useEffect, useState} from "react"

const USERNAMES = [
	"秦子晋",
	"张俊仪",
	"蔡博锐",
	"朱皓炜",
	"申恒瑜",
	"杨嘉卓",
	"陈梓恒"
]

let checkinForm = []

let EPOCH = new Date("2023-09-24")
EPOCH.setHours(0,0,0,0)

let userNameIndexMap = new Map()
USERNAMES.forEach((userName,index) => {
	userNameIndexMap.set(userName, index)
})

async function fetchRecords(setRecords) {
	try {
		fetch("https://sberm.cn/checkin-record", {
			headers: {
				Accept: 'application/json',
			},
			mode: "cors"
		})
		.then(response => {
			return response.json()
		})
		.then(data => {
			setRecords(data)
		})
	} catch(err) {
		console.log(err)
	}
}

function binarySearch(date, low, high, checkinForm) {
	if (low > high) {
		return -1
	}
	const middle = Math.floor((low + high) / 2);

	if (date > checkinForm[middle].date) {
		return binarySearch(date, low, middle - 1, checkinForm)
	} else if (date < checkinForm[middle].date) {
		return binarySearch(date, middle + 1, high, checkinForm)
	} else {
		return middle;
	}
}

export default function Main() {

	const [records, setRecords] = useState([])

	useEffect(() => {
		document.title = '学习打卡';
		const getRecords = async () => {
			await fetchRecords(setRecords);
		}

		getRecords()

		let today = new Date(Date.now())
		today.setHours(0,0,0,0)

		let USERS = Array(USERNAMES.length).fill(false)

		let temp = new Date(today.getTime())
		const daysInBetween = (temp - EPOCH) / (1000 * 60 * 60 * 24) + 1
		checkinForm = Array.from(Array(daysInBetween), () => ({
			date: null,
			users: structuredClone(USERS)
		}));

		for (let index = 0; index < daysInBetween; index++) {
			temp.setDate(today.getDate() - index)
			checkinForm[index].date = temp
			temp = new Date()
			temp.setHours(0, 0, 0, 0)
			if (temp < EPOCH) 
				break
		}

		return () => {}
	}, [])
	

	let checkinRecords = records.checkinRecord;

	// TODO: indexing is fine, binary search is unnecessary
	if (checkinRecords) {
		for (let index = 0; index < checkinRecords.length; index++) {
			let date = new Date(checkinRecords[index].checkinDate)
			date.setHours(0,0,0,0)
			let name = checkinRecords[index].name
			let indexCheckin = binarySearch(date, 0, checkinForm.length - 1, checkinForm)
			if (indexCheckin !== -1) {
				checkinForm[indexCheckin].users[userNameIndexMap.get(name)] = true;
			}
		}
	}

	return (
		<>
			<h1>学习打卡</h1>
			<CheckInApp checkinRecords={checkinForm} />
		</>
	)
}

function CheckInApp(props) {
	let checkinRecords = props.checkinRecords
	return (
		<>
		<div className="checkin-background">
				{checkinRecords.map((checkinRecord) => {
					let tempLink = "";
					return (
					<>
						<div className="checkin-container">
							<div className={
								(() => {
									let today = new Date(Date.now())
									today.setHours(0,0,0,0)
									if (!(checkinRecord.date < today || checkinRecord.date > today)) {
										return "checkin-date today"
									} else {
										return "checkin-date"
									}
								})()}>{((date = checkinRecord.date) => {

								let today = new Date(Date.now())
								today.setHours(0,0,0,0)

								const offset = date.getTimezoneOffset()
								const dateStr = new Date(date.getTime() - (offset*60*1000)).toISOString().split('T')[0]
								return dateStr
							})()}</div>
							<div className="checkin-names">
								{checkinRecord.users.map((val, index) => 
									<div className={
										(() => {
											if (val === false){
												tempLink = `/user?name=${USERNAMES[index]}`
												return "unsigned"
											} else {
												tempLink = `/user?name=${USERNAMES[index]}`
												return "signed"
											}
										})()
									}><a href={tempLink} onClick={(e) => {
										if (tempLink === 0) {
											e.preventDefault()
										}
									}}><span>{USERNAMES[index]}</span></a></div>)}
							</div>
						</div>
					</>)
				})}
		</div>
		</>
	);
}


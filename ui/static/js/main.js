
const convertDates = () => {
    const dates = document.getElementsByTagName("time")
    for (let i = 0; i < dates.length; i++) {
        const e = dates.item(i)
        const d = e.attributes.getNamedItem("datetime").value
        const dt = new Date(Date.parse(d))
        const localtime = dt.toLocaleString("en-US", { "year": "numeric", "month": "short", "day": "numeric", "hour": "numeric", "minute": "numeric", "timeZoneName": "short"})
        e.innerText = localtime
    }
}

convertDates()

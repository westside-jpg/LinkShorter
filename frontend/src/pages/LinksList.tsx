import {useEffect, useState} from "react"
import LinkCard from "../components/LinkCard.tsx";

function LinksList() {
    const [results, setResults] = useState<{id?: number, short?: string, original?: string, views?: number}[]>([])
    const [message, setMessage] = useState("")
    const [loading, setLoading] = useState(true)
    const [openQRIndex, setOpenQRIndex] = useState<number | null>(null)

    const showLinks = async () => {
        try {
            const response = await fetch("http://localhost:8080/my-links", {
                method: "GET",
                credentials: "include"
            })

            const data = await response.json()
            console.log(data)

            if (response.ok) {
                setResults(data["results"])
                if (data["message"]) {
                    setMessage(data["message"])
                } else {
                    setMessage("")
                }
                setLoading(false)
            } else {
                setMessage(data["message"])
                setLoading(false)
            }
        } catch (err) {
            setMessage("Ошибка соединения с сервером")
            setLoading(false)
        }
    }

    useEffect(() => {
        void showLinks()
    }, [])

    return (
        <div className="flex flex-col items-center min-h-screen gap-6 pt-20">

            <p className="text-4xl font-bold text-black">
                <span className="text-blue-600">Мои </span>
                <span className="text-black">ссылки</span>
            </p>

            <div className="flex flex-col gap-3 w-150">
                {results.length > 0 && !loading && results.map((item, index) => (
                    <LinkCard
                        key={item.id}
                        id={item.id}
                        short={item.short}
                        original={item.original}
                        views={item.views}
                        isQROpen={openQRIndex === index}
                        onQRToggle={() => setOpenQRIndex(openQRIndex === index ? null : index)}
                        onDelete={() => setResults(prev => prev.filter(r => r.id !== item.id))}
                    />
                ))}

                {message && !loading &&
                    <div className="border-2 border-red-600 bg-red-50
                     rounded-xl px-4 py-3 flex flex-col gap-1 text-center">
                            <p className="text-red-600">
                                {message}
                            </p>
                    </div>
                }

                {results.length == 0 && !loading && !message &&
                    <div className="border-2 border-red-600 bg-red-50
                     rounded-xl px-4 py-3 flex flex-col gap-1 text-center">
                        <p className="text-red-600">
                            Вы удалили все свои ссылки
                        </p>
                    </div>
                }
            </div>

            {loading && <p>Загрузка...</p>}
        </div>
    )
}

export default LinksList
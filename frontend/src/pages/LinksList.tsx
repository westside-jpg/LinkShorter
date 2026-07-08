import {useEffect, useState} from "react"
import { toast } from "sonner"

function LinksList() {
    const [results, setResults] = useState<{short?: string, original?: string, views?: number}[]>([])
    const [message, setMessage] = useState("")
    const [loading, setLoading] = useState(true)
    const [copiedIndex, setCopiedIndex] = useState<number | null>(null)
    const [copySuccess, setCopySuccess] = useState(false)

    useEffect(() => {
        if (copiedIndex === null) return

        const timer = setTimeout(() => {
            setCopiedIndex(null)
        }, 2000)

        return () => clearTimeout(timer)
    }, [copiedIndex])

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
                            <div key={index} className="relative border-2 border-blue-600 bg-blue-50 rounded-xl px-4 py-3 flex flex-col gap-1">
                                <>
                                    <div className="absolute top-3 right-3 w-5 h-5">
                                        <img src="/copy.svg"
                                             alt="Скопировать"
                                             className={`absolute inset-0 w-5 h-5 cursor-pointer transition-all duration-300
                                             ${copiedIndex === index ? "opacity-0" : "opacity-50 hover:opacity-100"}`}
                                             onClick={() => {
                                                 navigator.clipboard.writeText(`${item.short}`)
                                                     .then(() => {
                                                         setCopiedIndex(index)
                                                         setCopySuccess(true)
                                                         toast.success("Ссылка скопирована")
                                                     })
                                                     .catch(() => {
                                                         setCopiedIndex(index)
                                                         setCopySuccess(false)
                                                         toast.error("Ошибка копирования")
                                                     })
                                             }}/>
                                        <img
                                            src={copySuccess ? "/ok.svg" : "/decline.svg"}
                                            alt={copySuccess ? "Ссылка скопирована" : "Ошибка копирования"}
                                            className={`absolute inset-0 w-5 h-5 pointer-events-none transition-all duration-300
                                            ${copiedIndex === index ? "opacity-70" : "opacity-0"}`}
                                        />
                                    </div>
                                    <p className="text-blue-600 font-bold">
                                        <a href={`http://${item.short}`} className="text-blue-600 hover:underline"
                                           target="_blank">
                                            {item.short}
                                        </a>
                                    </p>
                                    <p className="text-gray-400 text-sm">
                                        <a href={item.original} className="text-gray-400 hover:underline"
                                           target="_blank">
                                            {item.original}
                                        </a>
                                    </p>
                                    <div className="flex flex-row gap-2 items-center">
                                        <img src="/views.svg" alt="Просмотры ссылки" className="w-5 h-5"/>
                                        <p className="translate-y-[1.3px]">{item.views}</p>
                                    </div>

                                </>
                            </div>
                ))}

                {results.length == 0 && !loading &&
                    <div
                        className="border-2 border-red-600 bg-red-50 rounded-xl px-4 py-3 flex flex-col gap-1 text-center">
                        <>
                            <p className="text-red-600">
                                {message}
                            </p>
                        </>
                    </div>
                }
            </div>

            {loading && <p>Загрузка...</p>}
        </div>
    )
}

export default LinksList
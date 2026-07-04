import {useEffect, useState} from "react";

function LinksList() {
    const [results, setResults] = useState<{short?: string, original?: string}[]>([])
    const [message, setMessage] = useState("")
    const [loading, setLoading] = useState(true)

    const showLinks = async () => {
        try {
            const response = await fetch("http://localhost:8080/my-links", {
                method: "GET",
                credentials: "include"
            })

            const data = await response.json()

            if (response.ok) {
                setResults(data["results"])
                setMessage("")
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
                            <div key={index} className="border-2 border-blue-600 bg-blue-50 rounded-xl px-4 py-3 flex flex-col gap-1">
                                <>
                                    <p className="text-blue-600 font-bold">
                                        <a href={`http://${item.short}`} className="text-blue-600 hover:underline" target="_blank">
                                            {item.short}
                                        </a>
                                    </p>
                                    <p className="text-gray-400 text-sm">
                                        <a href={item.original} className="text-gray-400 hover:underline" target="_blank">
                                            {item.original}
                                        </a>
                                    </p>
                                </>
                            </div>
                ))}

                {results.length == 0 && !loading &&
                    <div className="border-2 border-red-600 bg-red-50 rounded-xl px-4 py-3 flex flex-col gap-1 text-center">
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
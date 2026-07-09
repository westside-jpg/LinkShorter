import {useEffect, useState} from "react"
import LinkCard from "../components/LinkCard.tsx";

function LinksList() {
    const [results, setResults] = useState<{id?: number, short?: string, original?: string, views?: number, created_at?: string}[]>([])
    const [message, setMessage] = useState("")
    const [loading, setLoading] = useState(true)
    const [openQRIndex, setOpenQRIndex] = useState<number | null>(null)

    const buttonHoverActiveStyle = `hover:bg-blue-600 hover:border-blue-600 hover:text-white
        active:bg-blue-500 active:border-blue-500 active:text-white hover:shadow-lg hover:scale-105
        active:scale-110 hover:shadow-blue-500/50 transition-all duration-200 cursor-pointer`

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

            <div className="flex w-150 h-8 gap-2">
                <input
                    className="flex-1 border-2 border-black rounded-lg pl-2 focus:outline-none"
                    placeholder="Введите название тэга..."
                />
                <div className="flex flex-1 gap-2">

                    <div className="relative flex-1">
                        <div className={`flex items-center justify-center gap-2 h-8 border-2 border-black rounded-lg
                        ${buttonHoverActiveStyle} group`}>
                            <img src="/sort.svg" alt="Сортировка" className="h-5 w-5 group-hover:invert" />
                            <p>Сортировка</p>
                        </div>
                        <div className="absolute top-full mt-2 left-0 w-full bg-white border-2 border-black rounded-lg p-3 z-10">
                            <p>Здесь будет содержимое сортировки</p>
                        </div>
                    </div>

                    <div className="relative flex-1">
                        <div className={`flex items-center justify-center gap-2 h-8 border-2 border-black rounded-lg
                        ${buttonHoverActiveStyle} group`}>
                            <img src="/filter.svg" alt="Фильтрация" className="h-4 w-4 group-hover:invert" />
                            <p>Фильтрация</p>
                        </div>
                        <div className="absolute top-full mt-2 left-0 w-full bg-white border-2 border-black rounded-lg p-3 z-10">
                            <p>Здесь будет содержимое фильтрации</p>
                        </div>
                    </div>

                </div>
            </div>

            <div className="flex flex-col gap-3 w-150">
                {results.length > 0 && !loading && results.map((item, index) => (
                    <LinkCard
                        key={item.id}
                        id={item.id}
                        short={item.short}
                        original={item.original}
                        views={item.views}
                        created_at={item.created_at}
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
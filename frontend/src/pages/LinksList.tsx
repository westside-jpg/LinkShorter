import {useEffect, useState} from "react"
import LinkCard from "../components/LinkCard.tsx";

function LinksList() {
    const [results, setResults] = useState<{id?: number, short?: string, original?: string, views?: number, tag?: string, created_at?: string}[]>([])
    const [message, setMessage] = useState("")
    const [loading, setLoading] = useState(true)
    const [openQRIndex, setOpenQRIndex] = useState<number | null>(null)
    const [selectedSortButton, setSelectedSortButton] = useState(1)
    const [isSortOpen, setIsSortOpen] = useState(false)
    const [isFilterOpen, setIsFilterOpen] = useState(false)

    const buttonHoverActiveStyle = `hover:bg-blue-600 hover:border-blue-600 hover:text-white
        active:bg-blue-500 active:border-blue-500 active:text-white hover:shadow-lg hover:scale-105
        active:scale-110 hover:shadow-blue-500/50 transition-all duration-200 cursor-pointer`
    const selectedButtonStyle = `bg-blue-600 border-blue-600 text-white shadow-lg shadow-blue-500/50
    pointer-events-none`

    useEffect(() => {
        void SortLinks("date", "desc", 1)
    }, [])

    const SortLinks = async (sortBy: string, order: string, selectedSortNumber: number)=> {
        setLoading(true)
        setSelectedSortButton(0)
        try {
            const response = await fetch(`http://localhost:8080/my-links?sort=${sortBy}&order=${order}`, {
                method: "GET",
                credentials: "include"
            })

            const data = await response.json()

            if (response.ok) {
                setSelectedSortButton(selectedSortNumber)
                setResults(data["results"])
                setMessage("")
            } else {
                setResults([])
                setMessage(data["message"])
            }
        } catch (err) {
            setMessage("Ошибка соединения с сервером")
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="flex flex-col items-center min-h-screen gap-6 pt-20">

            <p className="text-4xl font-bold text-black">
                <span className="text-blue-600">Мои </span>
                <span className="text-black">ссылки</span>
            </p>

            {!message && !(results.length == 0) && (<div className="flex w-150 h-8 gap-2">
                <input
                    className="flex-1 border-2 border-black rounded-lg pl-2 focus:outline-none"
                    placeholder="Введите название тэга..."
                />
                <div className="flex flex-1 gap-2">

                    <div className="relative flex-1">
                        <div className={`flex items-center justify-center z-20 gap-2 h-8 border-2 border-black rounded-lg
                        ${buttonHoverActiveStyle} group`}
                             onClick={() => { setIsSortOpen(!isSortOpen) }}>
                            <img src="/sort.svg" alt="Сортировка" className="h-5 w-5 group-hover:invert" />
                            <p>Сортировка</p>
                        </div>

                        <div className={`absolute top-full flex flex-col gap-1 mt-2
                         left-0 w-full bg-white border-2 border-black rounded-lg p-3 z-10
                         transform transition-all duration-300
                        ${isSortOpen ? "opacity-100 translate-y-0 pointer-events-auto"
                            : "opacity-0 -translate-y-2 pointer-events-none"}`}>
                            <p className="text-center">По убыванию</p>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedSortButton == 1 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     void SortLinks("date", "desc", 1)
                                 }}>Даты</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedSortButton == 2 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     void SortLinks("views", "desc", 2)
                                 }}>Просмотров</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedSortButton == 3 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     void SortLinks("tag", "desc", 3)
                                 }}>Тэга</div>
                            <p className="text-center pt-3">По возрастанию</p>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedSortButton == 4 ? selectedButtonStyle : ""}`}
                            onClick={() => {
                                void SortLinks("date", "asc", 4)
                            }}>Даты</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedSortButton == 5 ? selectedButtonStyle : ""}`}
                            onClick={() => {
                                void SortLinks("views", "asc", 5)
                            }}>Просмотров</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedSortButton == 6 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     void SortLinks("tag", "asc", 6)
                                 }}>Тэга</div>
                        </div>
                    </div>

                    <div className="relative flex-1">
                        <div className={`flex items-center justify-center z-20 gap-2 h-8 border-2 border-black rounded-lg
                        ${buttonHoverActiveStyle} group`}
                             onClick={() => { setIsFilterOpen(!isFilterOpen) }}>
                            <img src="/filter.svg" alt="Фильтрация" className="h-4 w-4 group-hover:invert" />
                            <p>Фильтрация</p>
                        </div>

                        <div className={`absolute top-full flex flex-col gap-1 mt-2
                         left-0 w-full bg-white border-2 border-black rounded-lg p-3 z-10
                         transform transition-all duration-300
                        ${isFilterOpen ? "opacity-100 translate-y-0 pointer-events-auto"
                            : "opacity-0 -translate-y-2 pointer-events-none"}`}>
                            <p className="text-center">Диапазон</p>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}`}>Все время</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}`}>Неделя</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}`}>Месяц</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}`}>Год</div>
                            <p className="text-center pt-3">Просмотры</p>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}`}>0+</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}`}>10+</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}`}>100+</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}`}>1000+</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}`}>10000+</div>
                            <p className="text-center pt-3">Тэги</p>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}`}>Все</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}`}>С тэгами</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}`}>Без тэгов</div>
                        </div>
                    </div>

                </div>
            </div>)}

            <div className="flex flex-col gap-3 w-150">
                {results.length > 0 && !loading && results.map((item, index) => (
                    <LinkCard
                        key={item.id}
                        id={item.id}
                        short={item.short}
                        original={item.original}
                        views={item.views}
                        tag={item.tag}
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
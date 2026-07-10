import {useEffect, useState} from "react"
import LinkCard from "../components/LinkCard.tsx";

function LinksList() {
    const [results, setResults] = useState<{id?: number, short?: string, original?: string, views?: number, tag?: string, created_at?: string}[]>([])
    const [message, setMessage] = useState("")
    const [loading, setLoading] = useState(true)
    const [openQRIndex, setOpenQRIndex] = useState<number | null>(null)
    const [isSortOpen, setIsSortOpen] = useState(false)
    const [isFilterOpen, setIsFilterOpen] = useState(false)
    const [searchInput, setSearchInput] = useState("")
    const [query, setQuery] = useState({
        search: "",
        sort: "date",
        order: "desc",
        period: "all",
        views: "0",
        tags: "all"
    })
    const [selectedButtons, setSelectedButtons] = useState({
        sortButton: 1,
        filterRangeButton: 1,
        filterViewsButton: 1,
        filterTagsButton: 1
    })

    const buttonHoverActiveStyle = `hover:bg-blue-600 hover:border-blue-600 hover:text-white
        active:bg-blue-500 active:border-blue-500 active:text-white hover:shadow-lg hover:scale-105
        active:scale-110 hover:shadow-blue-500/50 transition-all duration-200 cursor-pointer`
    const selectedButtonStyle = `bg-blue-600 border-blue-600 text-white shadow-lg shadow-blue-500/50
    pointer-events-none`

    // Для мгновенного изменения категорий
    // при изменении фильтров/сортировки
    useEffect(() => {
        void SortLinks()
    }, [query])

    // Для дебаунса при вводе тэга в поиске
    // и вызывании функции сортировки только
    // при паузе в 0.3 секунды в написании
    useEffect(() => {
        const timer = setTimeout(() => {
            setQuery(prev => ({
                ...prev,
                search: searchInput
            }))
        }, 300)

        return () => clearTimeout(timer)
    }, [searchInput])

    const SortLinks = async ()=> {
        setLoading(true)
        try {
            const response = await fetch(`http://localhost:8080/my-links?search=${query.search}&sort=${query.sort}&order=${query.order}&period=${query.period}&views=${query.views}&tags=${query.tags}`, {
                method: "GET",
                credentials: "include"
            })

            const data = await response.json()

            if (response.ok) {
                setResults(data["results"])
                if (data["message"]) {
                    setMessage(data["message"])
                } else if (data["results"].length > 0) {
                    setMessage("")
                }
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

            <div className="flex w-150 h-8 gap-2">
                <input
                    className="flex-1 border-2 border-black rounded-lg pl-2 focus:outline-none"
                    placeholder="Введите название тэга..."
                    value={searchInput}
                    onChange={e => { setSearchInput(e.target.value) }}
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
                            ${selectedButtons.sortButton == 1 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         sortButton: 1,
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         sort: "date",
                                         order: "desc",
                                     }))
                                 }}>Даты</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.sortButton == 2 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         sortButton: 2,
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         sort: "views",
                                         order: "desc",
                                     }))
                                 }}>Просмотров</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.sortButton == 3 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         sortButton: 3,
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         sort: "tag",
                                         order: "desc",
                                     }))
                                 }}>Тэга</div>
                            <p className="text-center pt-3">По возрастанию</p>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.sortButton == 4 ? selectedButtonStyle : ""}`}
                            onClick={() => {
                                setSelectedButtons(prev => ({
                                    ...prev,
                                    sortButton: 4,
                                }))
                                setQuery(prev => ({
                                    ...prev,
                                    sort: "date",
                                    order: "asc",
                                }))
                            }}>Даты</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.sortButton == 5 ? selectedButtonStyle : ""}`}
                            onClick={() => {
                                setSelectedButtons(prev => ({
                                    ...prev,
                                    sortButton: 5,
                                }))
                                setQuery(prev => ({
                                    ...prev,
                                    sort: "views",
                                    order: "asc",
                                }))
                            }}>Просмотров</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.sortButton == 6 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         sortButton: 6,
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         sort: "tag",
                                         order: "asc",
                                     }))
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
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.filterRangeButton == 1 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         filterRangeButton: 1
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         period: "all"
                                     }))
                                 }}>Все время</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.filterRangeButton == 2 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         filterRangeButton: 2
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         period: "week"
                                     }))
                                 }}>Неделя</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.filterRangeButton == 3 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         filterRangeButton: 3
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         period: "month"
                                     }))
                                 }}>Месяц</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.filterRangeButton == 4 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         filterRangeButton: 4
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         period: "year"
                                     }))
                                 }}>Год</div>
                            <p className="text-center pt-3">Просмотры</p>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.filterViewsButton == 1 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         filterViewsButton: 1
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         views: "0"
                                     }))
                                 }}>0+</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.filterViewsButton == 2 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         filterViewsButton: 2
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         views: "10"
                                     }))
                                 }}>10+</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.filterViewsButton == 3 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         filterViewsButton: 3
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         views: "100"
                                     }))
                                 }}>100+</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.filterViewsButton == 4 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         filterViewsButton: 4
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         views: "1000"
                                     }))
                                 }}>1000+</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.filterViewsButton == 5 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         filterViewsButton: 5
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         views: "10000"
                                     }))
                                 }}>10000+</div>
                            <p className="text-center pt-3">Тэги</p>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.filterTagsButton == 1 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         filterTagsButton: 1
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         tags: "all"
                                     }))
                                 }}>Все</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.filterTagsButton == 2 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         filterTagsButton: 2
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         tags: "with"
                                     }))
                                 }}>С тэгами</div>
                            <div className={`text-sm border-2 rounded-lg pl-1.5
                            ${buttonHoverActiveStyle}
                            ${selectedButtons.filterTagsButton == 3 ? selectedButtonStyle : ""}`}
                                 onClick={() => {
                                     setSelectedButtons(prev => ({
                                         ...prev,
                                         filterTagsButton: 3
                                     }))
                                     setQuery(prev => ({
                                         ...prev,
                                         tags: "without"
                                     }))
                                 }}>Без тэгов</div>
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
                        tag={item.tag}
                        created_at={item.created_at}
                        isQROpen={openQRIndex === index}
                        onQRToggle={() => setOpenQRIndex(openQRIndex === index ? null : index)}
                        onDelete={() => setResults(prev => prev.filter(r => r.id !== item.id))}
                        onChange={() => { void SortLinks() }}
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
            </div>

            {loading && <p>Загрузка...</p>}
        </div>
    )
}

export default LinksList
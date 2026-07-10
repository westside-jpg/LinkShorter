import {useEffect, useRef, useState} from "react"
import {toast} from "sonner"

interface LinkCardProps {
    id?:         number
    short?:      string
    original?:   string
    views?:      number
    tag?:        string
    error?:      string
    created_at?: string
    isQROpen:    boolean
    onQRToggle:  () => void
    onDelete?:   () => void
    onChange?:   () => void
}

function LinkCard({id, short, original, views, error, created_at, isQROpen, onQRToggle, onDelete, tag, onChange}: LinkCardProps) {
    const [copied, setCopied] = useState(false)
    const [copySuccess, setCopySuccess] = useState(false)
    const [showQRImage, setShowQRImage] = useState(false)
    const [showTrashConfirmation, setShowTrashConfirmation] = useState(false)
    const [isDeleting, setIsDeleting] = useState(false)
    const [tagInputActive, setTagInputActive] = useState(false)
    const [currentTag, setCurrentTag] = useState(tag ?? "") // реальный тэг
    const [tagValue, setTagValue] = useState(tag ?? "") // для изменения тэга

    const code = short?.split("/").pop()

    // Для автоматического переключения курсора на инпут
    // при нажатии на кнопку изменения тэга
    const inputRef = useRef<HTMLInputElement>(null)

    useEffect(() => {
        if (tagInputActive) {
            inputRef.current?.focus()
        }
    }, [tagInputActive])

    // для изменения иконки при копировании
    useEffect(() => {
        if (!copied) return

        const timer = setTimeout(() => {
            setCopied(false)
        }, 2000)

        return () => clearTimeout(timer)
    }, [copied])

    // для задержки скрытия картинки, чтобы была плавная анимация
    useEffect(() => {
        if (isQROpen) {
            setShowQRImage(true)
        } else {
            const timer = setTimeout(() => setShowQRImage(false), 200)
            return () => clearTimeout(timer)
        }
    }, [isQROpen])

    function formatDateTime(isoString: any) {
        const date = new Date(isoString)

        const day = String(date.getDate()).padStart(2, '0')
        const month = String(date.getMonth() + 1).padStart(2, '0')
        const year = date.getFullYear()

        const hours = String(date.getHours()).padStart(2, '0')
        const minutes = String(date.getMinutes()).padStart(2, '0')

        return `${hours}:${minutes} | ${day}.${month}.${year}`
    }

    const handleCopy = () => {
        navigator.clipboard.writeText(short ?? "")
            .then(() => {
                setCopied(true)
                setCopySuccess(true)
                toast.success("Ссылка скопирована")
            })
            .catch(() => {
                setCopied(true)
                setCopySuccess(false)
                toast.error("Ошибка копирования")
            })
    }

    const handleDownload = async () => {
        const imageUrl = `http://localhost:8080/api/qr/${code}`
        try {
            const response = await fetch(imageUrl)
            const blob = await response.blob()

            const url = window.URL.createObjectURL(blob)
            const link = document.createElement("a")
            link.href = url

            link.download = `qr-${code}.png`

            document.body.appendChild(link)
            link.click()
            document.body.removeChild(link)
            window.URL.revokeObjectURL(url)
        } catch (err) {
            toast.error("Не удалось загрузить")
        }
    }

    const handleDelete = async () => {
        try {
            const response = await fetch(`http://localhost:8080/delete-link/${id}`, {
                method: "DELETE",
                credentials: "include"
            })

            if (response.ok) {
                toast.success("Ссылка удалена")
                setShowTrashConfirmation(false)
                setIsDeleting(true)
                if (onDelete) {
                    setTimeout(() => onDelete(), 300) // для плавного удаления из родителя
                }
                setTimeout(() => onChange?.(), 300) // для пересортировки после анимаций
            } else {
                const data = await response.json()
                toast.error(data["error"])
            }
        } catch (err) {
            toast.error("Ошибка соединения с сервером")
        }
    }

    const handleTag = async () => {
        try {
            const response = await fetch(`http://localhost:8080/my-links/add-tag`, {
                method: "PATCH",
                credentials: "include",
                headers: {"Content-Type": "application/json"},
                body: JSON.stringify({
                    id: id, // айди ссылки
                    tag: tagValue
                })
            })

            const data = await response.json()

            if (response.ok) {
                setCurrentTag(tagValue)
                setTagInputActive(false)
                setTimeout(() => onChange?.(), 300) // для пересортировки после анимаций
            } else {
                toast.error(data["error"])
            }
        } catch (err) {
            toast.error("Ошибка соединения с сервером")
        }
    }

    if (error) {
        return (
            <div className="border-2 rounded-xl px-4 py-3
                 border-red-500 bg-red-50">
                <p className="text-red-500">
                    {error}
                </p>
            </div>
        )
    }

    return (
        <div className={`relative border-2 rounded-xl px-4 py-3 flex flex-col
         gap-1 border-blue-600 bg-blue-50 transition-all duration-300
         ${isDeleting ? "opacity-0 scale-95 -translate-y-2" : ""}`}>

            {id !== undefined && (<div className="absolute top-3 right-3 w-5 h-5">
                <img src="/trash.svg" alt="Удалить ссылку"
                     className={`absolute inset-0 w-5 h-5 cursor-pointer
                        opacity-50 hover:opacity-100 transition-all duration-300`}
                     onClick={() => { setShowTrashConfirmation(true) }}
                />
            </div>)}

            {showTrashConfirmation && (
                <div
                    className="fixed inset-0 z-9 cursor-alias"
                    onClick={() => { setShowTrashConfirmation(false) }}
                />
            )}

            {id !== undefined && (<div className={`absolute flex flex-col top-10 -right-28.75 mt-2 items-center justify-center
            bg-white border-2 border-black rounded-xl p-3 shadow-lg z-10
            transition-all duration-200
            ${showTrashConfirmation ? "opacity-100 translate-y-0 pointer-events-auto"
                : "opacity-0 -translate-y-2 pointer-events-none"}`}>
                <p>Вы точно хотите удалить ссылку?</p>
                <div className="flex gap-2">
                    <button className="flex-1 border-2 rounded-lg px-1 mt-2 w-30
                    hover:scale-105 hover:bg-red-600 hover:border-red-600
                    active:scale-110 active:bg-red-500 active:border-red-500
                    hover:shadow-lg hover:shadow-red-500/50 hover:text-white
                    transition-all duration-200 cursor-pointer active:text-white"
                    onClick={() => { void handleDelete() }}>Удалить</button>
                    <button className="flex-1 border-2 rounded-lg px-1 mt-2 w-30
                    hover:scale-105 hover:bg-gray-600 hover:border-gray-600
                    active:scale-110 active:bg-gray-500 active:border-gray-500
                    hover:shadow-lg hover:shadow-gray-500/50 hover:text-white
                    transition-all duration-200 cursor-pointer active:text-white"
                    onClick={() => { setShowTrashConfirmation(false) }}>Отмена</button>
                </div>
            </div>)}
            

            <div className={`absolute top-3 w-5 h-5
            ${id !== undefined ? "right-10" : "right-3"}`}>
                <img src="/copy.svg" alt="Скопировать"
                     onClick={handleCopy}
                     className={`absolute inset-0 w-5 h-5 cursor-pointer
                        transition-all duration-300
                        ${copied ? "opacity-0" : "opacity-50 hover:opacity-100"}`}
                />

                <img src={copySuccess ? "/ok.svg" : "/decline.svg"}
                     alt={copySuccess ? "Ссылка скопирована" : "Ошибка копирования"}
                     className={`absolute inset-0 w-5 h-5 pointer-events-none
                        transition-all duration-300
                        ${copied ? "opacity-70" : "opacity-0"}`}
                />
            </div>

            <div className={`absolute top-3 w-5 h-5
            ${id !== undefined ? "right-17" : "right-10"}`}>
                <img src="/qr.svg" alt="Создать QR-код"
                     className={`absolute inset-0 w-5 h-5 cursor-pointer
                        opacity-50 hover:opacity-100 transition-all duration-300`}
                     onClick={onQRToggle}
                />
            </div>

            {id !== undefined && (<div className={`absolute top-3 w-5 h-5 right-24`}>
                <img src="/tag.svg" alt="Создать тэг"
                     className={`absolute inset-0 w-5 h-5 cursor-pointer
                        opacity-50 hover:opacity-100 transition-all duration-300`}
                     onClick={() => {
                         setTagInputActive(!tagInputActive)
                         if (tagInputActive) {
                             setTimeout(() => {
                                 setTagValue(currentTag)
                             }, 300)
                         }
                     }}
                />
            </div>)}

            {id !== undefined && <p className="absolute bottom-2 right-3
            text-sm text-gray-400">{formatDateTime(created_at)}</p>}

            <div className="flex items-center gap-2">
                <p className="text-blue-600 font-bold">
                    <a href={`http://${short}`}
                        target="_blank"
                        className="hover:underline">
                        {short}
                    </a>
                </p>

                <div className="relative h-6 min-w-48">
                    <div className={`absolute inset-0 flex items-center gap-2 transition-all duration-300
                        ${!tagInputActive ? "opacity-100 translate-y-0" 
                            : "opacity-0 -translate-y-2 pointer-events-none"}`}>

                        {currentTag && (
                            <>
                                <span className="text-gray-400">|</span>
                                <span className="text-gray-600 font-bold">
                                    {currentTag}
                                </span>
                            </>
                        )}
                    </div>

                    <div className={`absolute inset-0 flex items-center gap-2 transition-all duration-300
                        ${tagInputActive ? "opacity-100 translate-y-0" 
                            : "opacity-0 translate-y-2 pointer-events-none"}`}>
                        <input className="border-[1.5px] border-gray-400 h-5 pl-1 rounded-md
                            focus:outline-none text-gray-600"
                            placeholder="Введите тэг..."
                            maxLength={25}
                            ref={inputRef}
                            value={tagValue}
                            onChange={e => setTagValue(e.target.value)}/>

                        <button className="border-[1.5px] border-gray-400 h-5 w-10 shrink-0 rounded-md group
                            hover:bg-gray-200 hover:border-gray-400
                            active:bg-gray-300 active:border-gray-400
                            hover:shadow-lg hover:scale-105
                            active:scale-110 hover:shadow-gray-500/50
                            transition-all duration-200 cursor-pointer"
                            onClick={() => { void handleTag() }}>
                            <img alt="Подтвердить" src="/ok.svg"
                                 className="h-3 w-3 mx-auto group-hover:invert"/>
                        </button>
                    </div>
                </div>
            </div>

            <p className="text-gray-400 text-sm">
                <a href={original}
                   target="_blank"
                   className="hover:underline"
                >
                    {original}
                </a>
            </p>

            <div className={`absolute flex flex-col top-10 -right-3.75 mt-2 items-center justify-center
            bg-white border-2 border-black rounded-xl p-3 shadow-lg z-10
            transition-all duration-200
            ${isQROpen ? "opacity-100 translate-y-0 pointer-events-auto" 
                : "opacity-0 -translate-y-2 pointer-events-none"}`}>

                {showQRImage && <img alt="QR-код на ссылку" src={`http://localhost:8080/api/qr/${code}`} className="w-40 h-40"/>}

                <div className="whitespace-pre-line text-center items-center text-sm text-gray-400">
                    {short}
                </div>

                <div className="border-2 border-black rounded-xl mt-2
                 hover:scale-105 hover:bg-blue-600 hover:border-blue-600
                 active:scale-110 active:bg-blue-500 active:border-blue-500
                 hover:shadow-lg hover:shadow-blue-500/50
                 transition-all duration-200 cursor-pointer"
                 onClick={handleDownload}>
                    <img src="/download.svg" alt="Скачать"
                         className="cursor-pointer py-2 px-2 h-9 w-9 hover:invert">
                    </img>
                </div>

            </div>

            {/* прозрачный оверлей для закрытия
             окна кьюара при клике вне зоны окна */}
            {isQROpen && (
                <div
                    className="fixed inset-0 z-9 cursor-alias"
                    onClick={onQRToggle}
                />
            )}

            {views !== undefined && (
                <div className="flex flex-row gap-2 items-center">
                    <img
                        src="/views.svg"
                        alt="Просмотры"
                        className="w-5 h-5"
                    />
                    <p className="translate-y-[1.3px]">
                        {views}
                    </p>
                </div>
            )}

        </div>
    )
}

export default LinkCard
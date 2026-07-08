import {useEffect, useState} from "react"
import {toast} from "sonner"

interface LinkCardProps {
    short?: string
    original?: string
    views?: number
    error?: string
    isQROpen: boolean
    onQRToggle: () => void
}

function LinkCard({short, original, views, error, isQROpen, onQRToggle}: LinkCardProps) {
    const [copied, setCopied] = useState(false)
    const [copySuccess, setCopySuccess] = useState(false)
    const [showQRImage, setShowQRImage] = useState(false)

    const code = short?.split("/").pop()

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
        <div className="relative border-2 rounded-xl
        px-4 py-3 flex flex-col gap-1 border-blue-600 bg-blue-50">

            <div className="absolute top-3 right-3 w-5 h-5">
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

            <div className="absolute top-3 right-10 w-5 h-5">
                <img src="/qr.svg" alt="Создать QR-код"
                     className={`absolute inset-0 w-5 h-5 cursor-pointer
                        opacity-50 hover:opacity-100 transition-all duration-300`}
                     onClick={onQRToggle}
                />
            </div>


            <p className="text-blue-600 font-bold">
                <a href={`http://${short}`}
                   target="_blank"
                   className="hover:underline"
                >
                    {short}
                </a>
            </p>


            <p className="text-gray-400 text-sm">
                <a href={original}
                   target="_blank"
                   className="hover:underline"
                >
                    {original}
                </a>
            </p>

            <div className={`absolute flex flex-col top-10 -right-10 mt-2 items-center justify-center
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
                    className="fixed inset-0 z-9"
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
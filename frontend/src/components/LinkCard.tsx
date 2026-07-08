import {useEffect, useState} from "react"
import {toast} from "sonner"

interface LinkCardProps {
    short?: string
    original?: string
    views?: number
    error?: string
}

function LinkCard({short, original, views, error}: LinkCardProps) {
    const [copied, setCopied] = useState(false)
    const [copySuccess, setCopySuccess] = useState(false)

    useEffect(() => {
        if (!copied) return

        const timer = setTimeout(() => {
            setCopied(false)
        }, 2000)

        return () => clearTimeout(timer)
    }, [copied])


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
        px-4 py-3 flex flex-col gap-1
        border-blue-600 bg-blue-50">

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
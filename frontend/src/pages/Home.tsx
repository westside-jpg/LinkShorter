import {useEffect, useState} from "react";
import LinkCard from "../components/LinkCard.tsx";
import {useAuth} from "../context/AuthContext.tsx";

function Home() {
    const [inputURL, setInputURL] = useState("")
    const [results, setResults] =
        useState<{type: "success" | "error", short?: string, original?: string, message?: string}[]>([])
    const [novalue, setNoValue] = useState("")
    const [verifiedMessage, setVerifiedMessage] = useState("")
    const [openQRIndex, setOpenQRIndex] = useState<number | null>(null)
    const [customEnabled, setCustomEnabled] = useState(false) // для тумблера
    const [customLinkValue, setCustomLinkValue] = useState("")

    const { user } = useAuth()

    useEffect(() => {
        const params = new URLSearchParams(window.location.search)
        if (params.get("verified") === "true") {
            setVerifiedMessage("Почта успешно подтверждена")
            window.history.replaceState({}, "", "/")
        }
    }, [])

    const isValidURL = (url: string) => {
        try {
            const parsed = new URL(url)
            return (parsed.protocol === "http:" || parsed.protocol === "https:")
                && parsed.hostname.includes(".")
        } catch {
            return false
        }
    }

    const handleSubmit = async () => {
        if (!inputURL) {
            setNoValue("Сначала введите ссылку")
            return
        }

        let url = inputURL.trim()
        if (!url.startsWith("http://") && !url.startsWith("https://")) {
            url = "https://" + url
        }

        if (!(isValidURL(url))) {
            setNoValue("Некорректная ссылка")
            return
        }

        setNoValue("")
        setInputURL(url)

        try {
            const response = await fetch("http://localhost:8080/api/create-link", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ url: url }),
                    credentials: "include"
                }
            )

            const data = await response.json()

            if (response.ok) {
                setResults(prev => [{ type: "success", short: data["short-link"], original: url }, ...prev])
                setInputURL("")
            } else {
                setResults(prev => [{ type: "error", message: data["error"] }, ...prev])
            }
        } catch (err) {
            setResults(prev => [{ type: "error", message: "Не удалось связаться с сервером" }, ...prev])
        }

    }

    const handleCustomLink = async () => {
        if (!inputURL) {
            setNoValue("Сначала введите ссылку")
            return
        }

        let url = inputURL.trim()
        if (!url.startsWith("http://") && !url.startsWith("https://")) {
            url = "https://" + url
        }

        if (!(isValidURL(url))) {
            setNoValue("Некорректная ссылка")
            return
        }

        let custom = customLinkValue.trim()

        if (custom == "") {
            setNoValue("Введите название короткой ссылки")
            return
        }

        if (custom.length > 25) {
            setNoValue("Название ссылки не должно превышать 25 символов")
            return
        }

        const regex = /^[\p{L}\p{N}_-]+$/u
        if (!regex.test(custom)) {
            setNoValue("Название ссылки может содержать только буквы, цифры, '_' и '-'")
            return
        }

        setNoValue("")
        setInputURL(url)

        try {

            const response = await fetch("http://localhost:8080/api/create-link/custom", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        url: url,
                        custom: custom,
                    }),
                    credentials: "include"
                }
            )

            const data = await response.json()

            if (response.ok) {
                setResults(prev => [{ type: "success", short: data["short-link"], original: url }, ...prev])
                setInputURL("")
                setCustomLinkValue("")
            } else {
                setResults(prev => [{ type: "error", message: data["error"] }, ...prev])
            }
        } catch (err) {
            setResults(prev => [{ type: "error", message: "Не удалось связаться с сервером" }, ...prev])
        }
    }

    return (
        <>
            {verifiedMessage && (<div className="flex justify-center py-4">
                <div className="bg-green-100 border-2 border-green-600 text-green-700 rounded-xl px-6 py-2
                dark:bg-green-900 dark:border-green-600 dark:text-green-100">
                    {verifiedMessage}
                </div>
            </div>)}

            <div className="flex flex-col items-center min-h-screen gap-6 pt-20">


                <p className="text-5xl">
                    <span className="text-blue-600
                    dark:text-blue-400">Link</span>
                    <span className="text-black
                    dark:text-zinc-200">Shorter</span>
                </p>

                { !user && (
                    <div className="w-100 rounded-2xl border-2 border-gray-300 bg-gray-100 px-5 py-4 text-left
                    dark:bg-zinc-800 dark:border-zinc-700">
                        <p className="text-sm font-medium leading-3 text-gray-600
                        dark:text-zinc-200">
                            Некоторые возможности сервиса доступны только авторизованным пользователям
                        </p>

                        <p className="mt-3 text-sm leading-3 text-gray-400">
                            Без аккаунта вы можете создавать короткие ссылки,
                            однако они не сохраняются между сессиями и не имеют
                            статистики просмотров. После регистрации становятся
                            доступны кастомные ссылки, хранение истории, поиск,
                            сортировка, фильтрация, теги, QR-коды и другие функции
                            для удобного управления ссылками
                        </p>
                    </div>
                )}
                { user && (<div className="flex flex-col">
                    <div className="flex flex-row gap-2 w-149">
                        <button
                            onClick={() => setCustomEnabled(!customEnabled)}
                            className={`relative w-11 h-6 rounded-full
                            transition-all duration-300 cursor-pointer
                            ${customEnabled ? "bg-blue-500 dark:bg-blue-400" : "bg-gray-300 dark:bg-zinc-500"}`}>

                            <div className={`absolute top-1 left-1
                            w-4 h-4 rounded-full bg-white shadow-md
                            transition-all duration-300
                            ${customEnabled ? "translate-x-5" : ""}`}/>
                        </button>

                        <p className="text-lg">Кастомная ссылка</p>
                    </div>
                    <div className={`overflow-hidden transition-all duration-300
                        ${customEnabled ? "max-h-20 opacity-100 mt-2" : "max-h-0 opacity-0"}`}>
                        <input
                            className="border-2 w-150 h-10 text-lg border-black rounded-lg pl-4 focus:outline-none
                            dark:placeholder-zinc-400 placeholder-zinc-500
                            dark:border-gray-400 dark:text-zinc-200"
                            placeholder="Введите название короткой ссылки..."
                            maxLength={25}
                            value={customLinkValue}
                            onChange={e => { setCustomLinkValue(e.target.value) }}
                        />
                        <p className="text-gray-400 pl-1 pt-1
                        dark:text-zinc-500">Ваша ссылка: localhost:8080/{customLinkValue}</p>
                    </div>
                </div>)}

                <div className="flex flex-col gap-4 w-150 items-center">
                    <textarea
                        className="rounded-xl px-4 py-2 text-black border-2 w-150 h-30 text-2xl
                        focus:outline-none focus:scale-101 resize-none
                        transition-all duration-300 placeholder-zinc-500
                        dark:placeholder-zinc-400
                        dark:border-gray-400 dark:text-zinc-200"
                        placeholder="Вставьте Вашу ссылку..."
                        value={inputURL}
                        onChange={e => setInputURL(e.target.value)}
                        onKeyDown={e => {
                            if (e.key === "Enter" && !e.shiftKey) {
                                e.preventDefault()
                                if (!customEnabled) {
                                    void handleSubmit()
                                } else if (customEnabled) {
                                    void handleCustomLink()
                                }
                        }}}
                    />
                    {novalue && <p className="text-red-500 dark:text-red-400">{novalue}</p>}
                    <button className="rounded-xl w-70 px-4 py-2 text-xl text-black font-bold border-2
                     hover:scale-105 hover:bg-blue-600 hover:border-blue-600 hover:text-white
                     active:scale-110 active:bg-blue-500 active:border-blue-500 active:text-white
                     hover:shadow-lg hover:shadow-blue-500/50
                     transition-all duration-200 cursor-pointer

                     dark:border-zinc-600 dark:text-zinc-200
                     dark:hover:bg-blue-400 dark:hover:border-blue-400
                     dark:active:bg-blue-300 dark:active:border-blue-300
                     dark:hover:shadow-lg dark:hover:shadow-blue-400/50"
                            onClick={ () => {
                                if (!customEnabled) {
                                void handleSubmit()
                            } else if (customEnabled) {
                                void handleCustomLink()
                            }
                            }}>
                        Зашортить
                    </button>
                </div>

                <div className="flex flex-col gap-3 w-150">
                    {results.map((item, index) => (
                        <LinkCard
                            key={index}
                            short={item.short}
                            original={item.original}
                            error={item.message}
                            isQROpen={openQRIndex === index}
                            onQRToggle={() => setOpenQRIndex(openQRIndex === index ? null : index)}
                        />
                    ))}
                </div>


            </div>
        </>
    )
}

export default Home
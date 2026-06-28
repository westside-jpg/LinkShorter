import {useState} from "react";

function Home() {
    const [inputURL, setInputURL] = useState("")
    const [results, setResults] =
        useState<{type: "success" | "error", short?: string, original?: string, message?: string}[]>([])
    const [novalue, setNoValue] = useState("")

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

        const request = await fetch("http://localhost:8080/create-link", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ url: url })
            }
        )

        const data = await request.json()

        if (request.ok) {
            setResults(prev => [{ type: "success", short: data["short-link"], original: url }, ...prev])
            setInputURL("")
        } else {
            setResults(prev => [{ type: "error", message: data["error"] }, ...prev])
        }
    }

    return (
        <div className="flex flex-col items-center min-h-screen gap-6 pt-20">

            <p className="text-5xl">
                <span className="text-blue-600">Link</span>
                <span className="text-black">Shorter</span>
            </p>

            <div className="flex flex-col gap-4 w-150 items-center">
                <textarea
                    className="rounded-xl px-4 py-2 text-black border-2 w-150 h-30 text-2xl
                    focus:outline-none focus:border-black focus:scale-101 resize-none
                    transition-all duration-200"
                    placeholder="Вставьте Вашу ссылку..."
                    value={inputURL}
                    onChange={e => setInputURL(e.target.value)}
                    onKeyDown={e => {
                        if (e.key === "Enter" && !e.shiftKey) {
                            e.preventDefault()
                            void handleSubmit()
                        }
                    }}
                />
                {novalue && <p className="text-red-500">{novalue}</p>}
                <button className="rounded-xl w-70 px-4 py-2 text-xl text-black font-bold border-2
                 hover:scale-105 hover:bg-blue-600 hover:border-blue-600 hover:text-white
                 active:scale-110 active:bg-blue-500 active:border-blue-500 active:text-white
                 hover:shadow-lg hover:shadow-blue-500/50
                 transition-all duration-200 cursor-pointer"
                        onClick={handleSubmit}>
                    Зашортить
                </button>
            </div>

            <div className="flex flex-col gap-3 w-150">
                {results.map((item, index) => (
                    <div key={index} className={`border-2 rounded-xl px-4 py-3 flex flex-col gap-1
                        ${item.type === "error" ? "border-red-500" : "border-blue-600"}`}>
                        {item.type === "success" ? (
                            <>
                                <p className="text-blue-600 font-bold">
                                    <a href={`http://${item.short}`}
                                       className="text-blue-600 hover:underline"
                                       target="_blank">{item.short}</a>
                                </p>
                                <p className="text-gray-400 text-sm">
                                    <a href={item.original}
                                       className="text-gray-400 hover:underline"
                                       target="_blank">{item.original}</a>
                                </p>
                            </>
                        ) : (
                            <p className="text-red-500">{item.message}</p>
                        )}
                    </div>
                ))}
            </div>


        </div>
    )
}

export default Home
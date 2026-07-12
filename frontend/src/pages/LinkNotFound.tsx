import { Link } from "react-router-dom";

function LinkNotFound() {
    return (
        <div className="flex flex-col items-center justify-center min-h-screen px-6">

            <img
                src="/broken-link.svg"
                alt="Ссылка не найдена"
                className="w-24 h-24 opacity-80
                dark:invert"
            />

            <h1 className="mt-6 text-4xl font-bold
            dark:text-zinc-100">
                Ссылка не найдена
            </h1>

            <p className="mt-2 mb-7 max-w-xl text-center leading-4 text-gray-500 whitespace-pre-line
            dark:text-zinc-400">
                Возможно, ссылка была удалена, срок её {"\n"}
                действия истёк или она никогда не существовала
            </p>

            <Link
                to="/"
                className="rounded-xl w-70 px-4 py-2 text-xl text-center text-black font-bold border-2
                     hover:scale-105 hover:bg-blue-600 hover:border-blue-600 hover:text-white
                     active:scale-110 active:bg-blue-500 active:border-blue-500 active:text-white
                     hover:shadow-lg hover:shadow-blue-500/50
                     transition-all duration-200 cursor-pointer

                     dark:text-gray-400"
            >
                На главную
            </Link>

        </div>
    )
}

export default LinkNotFound
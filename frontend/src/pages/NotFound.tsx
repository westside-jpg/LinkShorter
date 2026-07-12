import { Link } from "react-router-dom"

function NotFound() {

    return (
        <div className="min-h-screen flex flex-col items-center justify-center gap-5">
            <h1 className="text-7xl font-bold
            dark:text-zinc-100">404</h1>
            <p className="text-gray-500 text-xl
            dark:text-zinc-400">
                Упс! Кажется такой страницы не существует
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

export default NotFound
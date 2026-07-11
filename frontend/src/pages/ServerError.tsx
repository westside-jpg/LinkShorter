import { Link } from "react-router-dom";

function ServerError() {
    return (
        <div className="flex flex-col items-center justify-center -mt-8.5 min-h-screen px-6">

            <img
                src="/server-error.svg"
                alt="Ссылка не найдена"
                className="w-56 h-56 opacity-80"
            />

            <h1 className="-mt-12.5 text-4xl font-bold">
                Ошибка сервера
            </h1>

            <p className="mt-2 mb-7 max-w-xl text-center leading-4 text-gray-500 whitespace-pre-line">
                Во время обработки запроса произошла {"\n"}
                непредвиденная ошибка. Попробуйте {"\n"}
                обновить страницу немного позже
            </p>

            <Link
                to="/"
                className="rounded-xl w-70 px-4 py-2 text-xl text-center text-black font-bold border-2
                     hover:scale-105 hover:bg-blue-600 hover:border-blue-600 hover:text-white
                     active:scale-110 active:bg-blue-500 active:border-blue-500 active:text-white
                     hover:shadow-lg hover:shadow-blue-500/50
                     transition-all duration-200 cursor-pointer"
            >
                На главную
            </Link>

        </div>
    )
}

export default ServerError
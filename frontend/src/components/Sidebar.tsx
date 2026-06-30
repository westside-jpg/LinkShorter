import {useState} from "react";

type SidebarProps = {
    isOpen: boolean
}

function Sidebar({ isOpen }: SidebarProps) {
    // Для показа нужных элементов
    const [isRegistration, setIsRegistration] = useState(true)
    const [isLogin, setIsLogin] = useState(false)
    const [isMenu, setIsMenu] = useState(false)
    const [isLoading, setIsLoading] = useState(false)

    // Для принятия значений и возврата ошибок/подтверждений
    const [username, setUsername] = useState("")
    const [email, setEmail] = useState("")
    const [loginInput, setLoginInput] = useState("")
    const [password, setPassword] = useState("")
    const [errors, setErrors] = useState<string[]>([])
    const [verify, setVerify] = useState("")

    const buttonClass = `flex-1 border-2 rounded-xl text-center px-4 py-1.5
    transition-[background-color,border-color,color,box-shadow] duration-200
    ${errors.length > 0 ? "mt-0" : "mt-5"}
    ${isLoading
        ? "bg-gray-300 border-gray-300 text-gray-500 cursor-not-allowed"
        : "hover:bg-blue-600 hover:border-blue-600 hover:text-white active:bg-blue-500 active:border-blue-500 active:text-white hover:shadow-lg hover:shadow-blue-500/50 cursor-pointer"
    }`

    const loginTabClass = `flex-1 border-2 rounded-xl px-4 py-1.5 text-center
    transition-[background-color,border-color,color,box-shadow] duration-200
    ${isLoading
        ? "bg-gray-300 border-gray-300 text-gray-500 cursor-not-allowed"
        : isLogin
            ? "bg-blue-600 border-blue-600 text-white shadow-lg shadow-blue-500/50 cursor-default"
            : "bg-white border-black text-black shadow-none cursor-pointer hover:bg-blue-600 hover:border-blue-600 hover:text-white active:bg-blue-500 active:border-blue-500 active:text-white hover:shadow-lg hover:shadow-blue-500/50"
    }`

    const registrationTabClass = `flex-1 border-2 rounded-xl px-4 py-1.5 text-center
    transition-[background-color,border-color,color,box-shadow] duration-200
    ${isLoading
        ? "bg-gray-300 border-gray-300 text-gray-500 cursor-not-allowed"
        : isRegistration
            ? "bg-blue-600 border-blue-600 text-white shadow-lg shadow-blue-500/50 cursor-default"
            : "bg-white border-black text-black shadow-none cursor-pointer hover:bg-blue-600 hover:border-blue-600 hover:text-white active:bg-blue-500 active:border-blue-500 active:text-white hover:shadow-lg hover:shadow-blue-500/50"
    }`

    // const sleep = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

    const handleRegistration = async () => {
        if (!username || !email || !password) {
            setErrors(["Сначала заполните все поля"])
            return
        }

        setIsLoading(true)
        try {
            const request = await fetch("http://localhost:8080/api/registration", {
                    method: "POST",
                    headers: {"Content-Type": "application/json"},
                    body: JSON.stringify({
                        username: username,
                        email: email,
                        password: password
                    }),
                    credentials: "include"
                }
            )

            const data = await request.json()

            if (request.ok) {
                setIsLoading(false)
                setErrors([])
                setIsRegistration(false)
                setIsLogin(false)
                setIsMenu(true)
                setVerify("Вам необходимо подтвердить почту по ссылке из почтового ящика, иначе Ваш аккаунт будет удален (дата и время удаления аккаунта в UTC)")
            } else {
                setIsLoading(false)
                setErrors(data["errors"])
                return
            }
        } catch (err) {
            setErrors(["Не удалось связаться с сервером"])
        } finally {
            setIsLoading(false)
        }
    }

    const handleLogin = async () => {
        if (!loginInput || !password) {
            setErrors(["Сначала заполните все поля"])
            return
        }

        setIsLoading(true)
        try {
            const request = await fetch("http://localhost:8080/api/login", {
                    method: "POST",
                    headers: {"Content-Type": "application/json"},
                    body: JSON.stringify({
                        loginInput: loginInput,
                        password: password
                    }),
                    credentials: "include"
                }
            )

            const data = await request.json()

            if (request.ok) {
                setIsLoading(false)
                setErrors([])
                setIsRegistration(false)
                setIsLogin(false)
                setIsMenu(true)
                if (data["is_verified"]) {
                    setVerify("")
                } else {
                    setVerify("Вам необходимо подтвердить почту по ссылке из почтового ящика, иначе Ваш аккаунт будет удален (дата и время удаления аккаунта в UTC)")
                }
            } else {
                setIsLoading(false)
                setErrors(data["errors"])
                return
            }
        } catch (err) {
            setErrors(["Не удалось связаться с сервером"])
        } finally {
            setIsLoading(false)
        }
    }

    return (
        <div className={`fixed top-0 right-0 h-full w-90 bg-white shadow-xl shadow-white rounded-tl-4xl rounded-bl-4xl z-50
            transform transition-transform duration-300
            ${isOpen ? "translate-x-0" : "translate-x-full"}`}>

            <div className="flex flex-col gap-2 pt-10 px-6 m-0 py-0">
                <p className="text-4xl flex-auto self-center font-bold mb-2">
                    <span className="text-blue-600">Link</span>
                    <span>Shorter</span>
                </p>
            </div>

            {isMenu && verify && <p className="text-orange-400 self-center text-center px-4">{verify}</p>}

            {!isMenu && (
                <div className="flex flex-row gap-2 pt-8 px-6">
                    <button className={registrationTabClass}
                            disabled={isLoading}
                            onClick={() => {
                                if (isLoading) return
                                if (isRegistration) return
                                setIsRegistration(true)
                                setIsLogin(false)
                                setIsMenu(false)
                                setPassword("")
                                setErrors([])
                            }}>
                        Регистрация
                    </button>
                    <button className={loginTabClass}
                            disabled={isLoading}
                            onClick={() => {
                                if (isLoading) return
                                if (isLogin) return
                                setIsLogin(true)
                                setIsRegistration(false)
                                setIsMenu(false)
                                setPassword("")
                                setErrors([])
                            }}>
                        Вход
                    </button>
                </div>
            )}

            {isRegistration && (
                <div className="flex flex-col gap-2 pt-5 px-6">
                    <div>
                        <span className="ml-2">Придумайте логин</span>
                        <input
                            type="text"
                            className="border-2 rounded-xl px-4 py-1 w-full focus:outline-none"
                            value={username}
                            placeholder="Максимум 30 символов"
                            maxLength={30}
                            onChange={e => setUsername(e.target.value)}
                        />
                    </div>
                    <div>
                        <span className="ml-2">Введите почту</span>
                        <input
                            type="email"
                            className="border-2 rounded-xl px-4 py-1 w-full focus:outline-none"
                            value={email}
                            placeholder="example@gmail.com"
                            onChange={e => setEmail(e.target.value)}
                        />
                    </div>
                    <div>
                        <span className="ml-2">Придумайте пароль</span>
                        <input
                            type="password"
                            className="border-2 rounded-xl px-4 py-1 w-full focus:outline-none"
                            value={password}
                            placeholder="Минимум 8 символов"
                            onChange={e => setPassword(e.target.value)}
                        />
                    </div>
                    {errors.length > 0 && errors.map((e, i) => (
                        <p key={i} className="text-red-500 text-sm text-center">{e}</p>
                    ))}
                    <button className={buttonClass} disabled={isLoading} onClick={() => { void handleRegistration() }}>
                        {isLoading ? "Подождите..." : "Зарегистрироваться"}
                    </button>
                </div>
            )}

            {isLogin && (
                <div className="flex flex-col gap-2 pt-5 px-6">
                    <div>
                        <span className="ml-2">Введите логин или почту</span>
                        <input
                            type="text"
                            className="border-2 rounded-xl px-4 py-1 w-full focus:outline-none"
                            value={loginInput}
                            onChange={e => setLoginInput(e.target.value)}
                        />
                    </div>
                    <div>
                        <span className="ml-2">Введите пароль</span>
                        <input
                            type="password"
                            className="border-2 rounded-xl px-4 py-1 w-full focus:outline-none"
                            value={password}
                            onChange={e => setPassword(e.target.value)}
                        />
                    </div>
                    {errors.length > 0 && errors.map((e, i) => (
                        <p key={i} className="text-red-500 text-sm text-center">{e}</p>
                    ))}
                    <button className={`flex-1 border-2 rounded-xl text-center px-4 py-1.5
              hover:bg-blue-600 hover:border-blue-600 hover:text-white
              active:bg-blue-500 active:border-blue-500 active:text-white
                hover:shadow-lg hover:shadow-blue-500/50
                transition-[background-color,border-color,color,box-shadow] duration-200 cursor-pointer
                            ${errors.length > 0 ? "mt-0" : "mt-5" }`}
                 onClick={() => { void handleLogin() }}>
                        {isLoading ? "Подождите..." : "Войти"}
                    </button>
                </div>
            )}

            {isMenu && (
                <div className="flex flex-col gap-2 pt-5 px-6">

                    <button className="border-2 rounded-xl px-4 py-1.5 text-left
               hover:bg-blue-600 hover:border-blue-600 hover:text-white
               active:bg-blue-500 active:border-blue-500 active:text-white
                 hover:shadow-lg hover:shadow-blue-500/50
                 transition-all duration-200 cursor-pointer">
                        Главная
                    </button>
                    <button className="border-2 rounded-xl px-4 py-1.5 text-left
                hover:bg-blue-600 hover:border-blue-600 hover:text-white
                active:bg-blue-500 active:border-blue-500 active:text-white
                  hover:shadow-lg hover:shadow-blue-500/50
                  transition-all duration-200 cursor-pointer">
                        Мои ссылки
                    </button>
                </div>
            )}


        </div>
    )
}

export default Sidebar
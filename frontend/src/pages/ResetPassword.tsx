import {useEffect, useState} from "react"
import {Link} from "react-router-dom";
import { toast } from "sonner"

function ResetPassword() {
    const [step, setStep] = useState(1)
    const [email, setEmail] = useState("")
    const [lastEmail, setLastEmail] = useState("")
    const [code, setCode] = useState("")
    const [couldResend, setCouldResend] = useState(false)
    const [visible, setVisible] = useState(true)
    const [isLoading, setIsLoading] = useState(false)
    const [timeWait, setTimeWait] = useState(60)
    const [timerRunning, setTimerRunning] = useState(false)
    const [attemptsLeft, setAttemptsLeft] = useState(5)

    const disabledButtonAttributes = `disabled:bg-gray-300 disabled:border-gray-300 disabled:text-gray-500
    disabled:cursor-not-allowed disabled:hover:bg-gray-300 disabled:hover:border-gray-300
    disabled:hover:text-gray-500 disabled:hover:shadow-none`

    // Тиканье таймера
    useEffect(() => {
        if (!timerRunning || timeWait <= 0) return;

        const interval = setInterval(() => {
            setTimeWait(prev => {
                if (prev <= 1) {
                    clearInterval(interval)
                    setTimerRunning(false)
                    return 0
                }
                return prev - 1
            });
        }, 1000)

        return () => clearInterval(interval)
    }, [timerRunning, timeWait])

    useEffect(() => {
        if (timeWait == 0) {
            setCouldResend(true)
        }
    }, [timeWait])

    // Глобальный слушатель клавиатуры
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === "Enter") {
                if (step === 1) {
                    if (ValidateEmail(email)) {
                        if (email == lastEmail && timeWait > 0) {
                            goToStep(2)
                            return
                        }
                        void SendResetPasswordEmail()
                    }
                } else if (step === 2) {
                    void CheckResetPasswordCode()
                }
            }

            if (step === 2) {
                if (/^\d$/.test(e.key) && code.length < 6) {
                    setCode(prev => prev + e.key)
                } else if (e.key === "Backspace" || e.key === "Delete") {
                    setCode(prev => prev.slice(0, -1))
                }
            }
        }

        const handlePaste = (e: ClipboardEvent) => {
            if (step !== 2) return
            e.preventDefault()
            const digits = (e.clipboardData?.getData("text") ?? "")
                .replace(/\D/g, "")
                .slice(0, 6)
            setCode(digits)
        }


        window.addEventListener("keydown", handleKeyDown)
        window.addEventListener("paste", handlePaste)

        return () => {
            window.removeEventListener("keydown", handleKeyDown)
            window.removeEventListener("paste", handlePaste)
        }
    }, [step, email, code, isLoading, lastEmail, timeWait])

    const goToStep = (newStep: number) => {
        setVisible(false)
        setTimeout(() => {
            setStep(newStep)
            setVisible(true)
        }, 200)
    }

    const declinationWord = (n: number, one: string, two: string, many: string) => {
        const lastTwoDigits = n % 100
        if (lastTwoDigits >= 11 && lastTwoDigits <= 14) {
            return many;
        }
        switch (n % 10) {
            case 1:
                return one
            case 2:
            case 3:
            case 4:
                return two
            default:
                return many
        }
    };

    const isValidEmail = (email: string) => {
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
    }

    const ValidateEmail = (email: string) => {
        if (email == "") {
            toast.error("Сначала введите почту")
            return false
        }

        if (!isValidEmail(email)) {
            toast.error("Введенная почта некорректна")
            return false
        }

        return true
    }

    const SendResetPasswordEmail = async() => {
        try {
            setIsLoading(true)
            const request = await fetch("http://localhost:8080/api/reset-password/send-code", {
                method: "POST",
                headers: {"Content-Type": "application/json"},
                body: JSON.stringify({
                    email: email
                })
            })

            const data = await request.json()

            if (request.ok) {
                if (step == 1) {
                    goToStep(2)
                } else {
                    toast.success("Письмо успешно переотправлено")
                }
                setTimeWait(60)
                setCouldResend(false)
                setTimerRunning(true)
                setAttemptsLeft(5)
            } else {
                if (step == 1) {
                    toast.error(data["error"])
                }
            }
        } catch (err) {
            toast.error("Не удалось связаться с сервером")
        } finally {
            setIsLoading(false)
        }
    }

    const CheckResetPasswordCode = async() => {
        if (code.length != 6) {
            toast.error("Введите код полностью")
            return
        }
        try {
            setIsLoading(true)
            const response = await fetch("http://localhost:8080/api/reset-password/check-code", {
                method: "POST",
                headers: {"Content-Type": "application/json"},
                body: JSON.stringify({
                    email: email,
                    code: code
                })
            })

            const data = await response.json()

            if (response.ok) {
                goToStep(3)
            } else {
                if (data["attempts"] != undefined) { // если ошибка вызвана неправильным кодом
                    setAttemptsLeft(data["attempts"])
                    if (data["attempts"] != 0) { // если остались попытки ввода
                        toast.error(`${data["error"]} ${data["attempts"]}`) // показываем их количество
                    } else {
                        toast.error(data["error"]) // иначе только ошибку без количества
                    }
                } else {
                    toast.error(data["error"]) // если ошибка не вызвана неправильным кодом (ошибка бд и тд)
                }
            }
        } catch (err) {
            toast.error("Не удалось связаться с сервером")
        } finally {
            setIsLoading(false)
        }

    }

    return (
        <div className="flex justify-center items-center min-h-screen">

            <Link
                to={step == 1 ? "/" : "#"}
                className="fixed top-4 left-4 flex flex-row items-center gap-2
                border-2 rounded-xl px-4 py-2 font-bold
                hover:bg-blue-600 hover:border-blue-600 hover:text-white hover:scale-105
                active:bg-blue-500 active:border-blue-500 active:text-white
                hover:shadow-lg hover:shadow-blue-500/50
                transition-all duration-200 cursor-pointer group"
                onClick={() => {
                    if (isLoading) return;
                    if (step == 2) setLastEmail(email)
                    if (step !== 1) goToStep(step - 1);
                }}
            >
                <img
                    src="/arrow.svg"
                    alt="Назад"
                    className="w-5 h-5 rotate-180 invert group-hover:filter-none transition-all duration-200"
                />
                <span key={step} className="text-base animate-fade-in translate-y-px">{step == 1 ? "На главную" : "Назад"}</span>
            </Link>

            {step == 1 && (
                <div className={`flex flex-col gap-4 w-96 transition-opacity duration-200 ${visible ? "opacity-100" : "opacity-0"}`}>
                    <p className="text-base text-center">Введите почту от Вашего аккаунта для смены пароля</p>
                    <input
                        type="email"
                        placeholder="example@gmail.com"
                        value={email}
                        className="border-2 rounded-xl px-4 py-2 focus:outline-none focus:border-black transition-all duration-200"
                        onChange={e => setEmail((e.target.value).trim())}
                    />
                    <button
                        className={`relative border-2 rounded-xl px-4 py-2 font-bold overflow-hidden
                            hover:bg-blue-600 hover:border-blue-600 hover:text-white
                            active:bg-blue-500 active:border-blue-500 active:text-white
                            hover:shadow-lg hover:shadow-blue-500/50
                            transition-all duration-200 cursor-pointer group
                            ${disabledButtonAttributes}`}

                        disabled={isLoading}
                        onClick={() => {
                            if (ValidateEmail(email)) {
                                if (email == lastEmail && timeWait > 0) {
                                    goToStep(2)
                                    return
                                }
                                void SendResetPasswordEmail()
                            }
                        }}
                    >
                    <span className={`transition-all duration-200
                     ${isLoading ? "group-hover:opacity-100" : "group-hover:opacity-0"}`}>
                        {isLoading ? "Подождите..." : "Далее"}
                    </span>
                        <img
                            src="/arrow.svg"
                            alt={isLoading ? "Подождите..." : "Далее"}
                            className={`absolute inset-0 m-auto w-5 h-5 opacity-0 -translate-x-4
                            ${isLoading ? "group-hover:opacity-0" : "group-hover:opacity-100"} group-hover:translate-x-0
                            transition-all duration-200`}
                        />
                    </button>
                </div>
            )}

            {step == 2 && (
                <div className={`flex flex-col gap-4 w-96 transition-opacity duration-200 ${visible ? "opacity-100" : "opacity-0"}`}>
                    <p className="text-xl text-center">
                        Проверьте почтовый ящик и введите<br/>
                        код для сброса пароля из письма
                    </p>

                    <input
                        type="text"
                        inputMode="numeric"
                        maxLength={6}
                        value={code}
                        readOnly
                        className="absolute opacity-0 w-0 h-0"
                        id="code-input"
                    />

                    <div
                        className="flex flex-row gap-3 cursor-text"
                        onClick={() => document.getElementById("code-input")?.focus()}
                    >
                        {Array.from({length: 6}).map((_, i) => (
                            <div
                                key={i}
                                className={`w-14 h-18 border-2 rounded-xl flex items-center justify-center text-2xl font-bold
                                transition-all duration-200
                                ${i === code.length ? "border-black scale-105" : "border-gray-300"}`}
                            >
                                {code[i] || ""}
                            </div>
                        ))}
                    </div>


                    <button className={`relative w-full border-2 rounded-xl px-4 py-2 font-bold overflow-hidden
                        hover:bg-green-600 hover:border-green-600 hover:text-white
                        active:bg-green-500 active:border-green-500 active:text-white
                        hover:shadow-lg hover:shadow-green-500/50
                        transition-all duration-200 cursor-pointer group
                        ${disabledButtonAttributes}`}
                        disabled={isLoading || attemptsLeft == 0}
                        onClick={() => {
                            void CheckResetPasswordCode()
                        }}>
            <span className={`transition-all duration-200 
            ${(isLoading || attemptsLeft == 0) ? "group-hover:opacity-100" : "group-hover:opacity-0"}`}>
                {isLoading ? "Подождите..." : "Подтвердить"}
            </span>
                        <img
                            src="/apply.svg"
                            className={`absolute inset-0 m-auto w-5 h-5 opacity-0 scale-75
                ${(isLoading || attemptsLeft == 0) ? "group-hover:opacity-0" : "group-hover:opacity-100"} group-hover:scale-100
                transition-all duration-200`}
                         alt={isLoading ? "Подождите..." : "Подтвердить"}/>
                    </button>


                    <button className={`relative w-full border-2 rounded-xl px-4 py-2 font-bold overflow-hidden
                        hover:bg-blue-600 hover:border-blue-600 hover:text-white
                        active:bg-blue-500 active:border-blue-500 active:text-white
                        hover:shadow-lg hover:shadow-blue-500/50
                        transition-all duration-200 cursor-pointer group
                        ${disabledButtonAttributes}`}
                        disabled={isLoading || !couldResend}
                        onClick={() => {
                            void SendResetPasswordEmail()
                        }}>
            <span className={`transition-all duration-200 group-hover:opacity-0
            ${(isLoading || !couldResend) ? "group-hover:opacity-100" : "group-hover:opacity-0"}`}>
                {(isLoading && couldResend) ? "Подождите..." : "Переотправить"}
            </span>
                        <img
                            src="/send.svg"
                            alt={isLoading ? "Подождите..." : "Переотправить"}
                            className={`absolute inset-0 m-auto w-5 h-5 opacity-0 scale-75
                ${(isLoading || !couldResend) ? "group-hover:opacity-0" : "group-hover:opacity-100"} group-hover:scale-100
                transition-all duration-200`}
                        />
                    </button>
                <div className="border-2 rounded-2xl border-gray-300 bg-gray-100">
                    <p className="text-sm py-3 text-center text-gray-400 whitespace-pre-line">
                        {timeWait > 0 ? (
                            <>
                                Если по каким-то причинам письмо не дошло{"\n"}
                                до Вас, то Вы можете запросить его ещё раз,{"\n"}
                                нажав на кнопку "Переотправить" через {timeWait}{" "}
                                {declinationWord(timeWait, "секунду", "секунды", "секунд")}
                            </>
                        ) : (
                            <>
                                Вы можете запросить новое письмо для{"\n"} подтверждения,
                                если Вам это необходимо
                            </>
                        )}
                    </p>
                </div>

                <div className="border-2 rounded-2xl border-gray-300 bg-gray-100">
                    <p className="text-sm py-3 text-center text-gray-400 whitespace-pre-line">
                        Если Вы ввели неправильный адрес электронной почты,{"\n"}
                        то вернитесь на прошлую страницу и введите почту снова
                    </p>
                </div>

                </div>
            )}
        </div>
    )
}

export default ResetPassword
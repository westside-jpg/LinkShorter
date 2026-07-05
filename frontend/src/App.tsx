import {AuthProvider, useAuth} from "./context/AuthContext.tsx";
import {BrowserRouter, Route, Routes, useLocation} from "react-router-dom";
import Home from "./pages/Home.tsx";
import LinksList from "./pages/LinksList.tsx";
import Sidebar from "./components/Sidebar.tsx";
import {useEffect, useState} from "react";
import ResetPassword from "./pages/ResetPassword.tsx";
import { Toaster } from "sonner";

function App() {
    const [sidebarOpen, setSidebarOpen] = useState(false)

    return (
        <AuthProvider>
            <BrowserRouter>
                <Toaster
                    position="top-center"
                    expand={false} visibleToasts={1}
                    style={{ fontFamily: "MyFont, sans-serif"}}
                    toastOptions={{
                        unstyled: true,
                        classNames: {
                            toast: "flex items-center border-2 rounded-xl text-center px-10 py-3 font-bold shadow-lg shadow-gray-300",
                            title: "text-base",
                            success: "bg-green-500 border-green-500 text-white shadow-green-300",
                            error: "bg-red-500 border-red-500 text-white shadow-red-300",
                            info: "bg-blue-500 border-blue-500 text-white shadow-blue-300",
                            icon: "hidden",
                        }
                    }}
                />
                <AppContent sidebarOpen={sidebarOpen} setSidebarOpen={setSidebarOpen} />
            </BrowserRouter>
        </AuthProvider>
    )
}

function AppContent({ sidebarOpen, setSidebarOpen }: {
    sidebarOpen: boolean
    setSidebarOpen: (v: boolean) => void
}) {
    const { user } = useAuth()
    const location = useLocation()
    const hideProfileButton = location.pathname === "/reset-password"

    // Закрытие сайдбара при переходе между страницами
    useEffect(() => {
        setSidebarOpen(false)
    }, [location.pathname])

    return (
        <>
            <Routes>
                <Route path="/" element={<Home />} />
                <Route path="/my-links" element={<LinksList />} />
                <Route path="/reset-password" element={<ResetPassword />}/>
            </Routes>
            {!hideProfileButton && <button className="fixed top-4 right-4" onClick={() => setSidebarOpen(true)}>
                <img src="/avatar.svg" alt="Профиль (меню)" className="w-15 h-15 rounded-sm cursor-pointer"/>
            </button>}
            {user?.username && <p className="text-xl fixed top-4 right-22 px-2 border-2 rounded-xl font-extrabold">{user.username}</p>}
            {user?.email && <p className="text-sm fixed top-13 right-22 px-2 border-2 rounded-xl">{user.email}</p>}
            {sidebarOpen && (
                <div className="fixed inset-0 bg-black/50 z-40 cursor-alias" onClick={() => setSidebarOpen(false)} />
            )}
            <Sidebar isOpen={sidebarOpen} />
        </>
    )
}

export default App
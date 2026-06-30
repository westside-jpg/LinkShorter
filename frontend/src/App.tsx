import { useState } from "react"
import Home from "./pages/Home"
import Sidebar from "./components/Sidebar"
import {AuthProvider} from "./context/AuthContext.tsx";

function App() {
    const [sidebarOpen, setSidebarOpen] = useState(false)

    return (
        <>
            <AuthProvider>
                <Home />
                <button
                    className="fixed top-4 right-4"
                    onClick={() => setSidebarOpen(true)}
                >
                    <img src="/avatar.svg" className="w-14 h-14 rounded-sm cursor-pointer" />
                </button>
                {sidebarOpen && (
                    <div
                        className="fixed inset-0 bg-black/50 z-40 cursor-alias"
                        onClick={() => setSidebarOpen(false)}
                    />
                )}
                <Sidebar isOpen={sidebarOpen} />
            </AuthProvider>
        </>

    )
}

export default App
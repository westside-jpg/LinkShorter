import {AuthProvider, useAuth} from "./context/AuthContext.tsx";
import {BrowserRouter, Route, Routes} from "react-router-dom";
import Home from "./pages/Home.tsx";
import LinksList from "./pages/LinksList.tsx";
import Sidebar from "./components/Sidebar.tsx";
import {useState} from "react";

function App() {
    const [sidebarOpen, setSidebarOpen] = useState(false)

    return (
        <AuthProvider>
            <BrowserRouter>
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

    return (
        <>
            <Routes>
                <Route path="/" element={<Home />} />
                <Route path="/my-links" element={<LinksList />} />
            </Routes>
            <button className="fixed top-4 right-4" onClick={() => setSidebarOpen(true)}>
                <img src="/avatar.svg" className="w-15 h-15 rounded-sm cursor-pointer" />
            </button>
            <p className="text-xl fixed top-4 right-22 px-2 border-2 rounded-xl font-extrabold">{user?.username}</p>
            <p className="text-sm fixed top-13 right-22 px-2 border-2 rounded-xl">{user?.email}</p>
            {sidebarOpen && (
                <div className="fixed inset-0 bg-black/50 z-40 cursor-alias" onClick={() => setSidebarOpen(false)} />
            )}
            <Sidebar isOpen={sidebarOpen} />
        </>
    )
}

export default App
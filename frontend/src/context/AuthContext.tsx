import { createContext, useContext, useEffect, useState, type ReactNode } from "react"

type User = {
    user_id: number
    email: string
    is_verified: boolean
} | null

type AuthContextType = {
    user: User
    loading: boolean
    refreshUser: () => void
}

const AuthContext = createContext<AuthContextType>({ user: null, loading: true, refreshUser: () => {} })

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User>(null)
    const [loading, setLoading] = useState(true)

    const fetchUser = async () => {
        try {
            const res = await fetch("http://localhost:8080/api/me", {
                credentials: "include"
            })
            if (res.ok) {
                const data = await res.json()
                setUser({ user_id: data.user_id, email: data.email, is_verified: data.is_verified })
            } else {
                setUser(null)
            }
        } catch {
            setUser(null)
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        fetchUser()
    }, [])

    return (
        <AuthContext.Provider value={{ user, loading, refreshUser: fetchUser }}>
            {children}
        </AuthContext.Provider>
    )
}

export function useAuth() {
    return useContext(AuthContext)
}



import { useState, useEffect } from 'react'
import { Login } from './components/Login'
import { Desktop } from './components/Desktop'

function App() {
  const [token, setToken] = useState<string | null>(null)
  const [user, setUser] = useState<any>(null)

  useEffect(() => {
    const savedToken = localStorage.getItem('token')
    if (savedToken) {
      setToken(savedToken)
      // Verify token
      fetch('/api/auth/me', {
        headers: { 'Authorization': `Bearer ${savedToken}` },
      }).then(res => {
        if (res.ok) {
          res.json().then(data => setUser(data))
        } else {
          localStorage.removeItem('token')
          setToken(null)
        }
      })
    }
  }, [])

  const handleLogin = (newToken: string, newUser: any) => {
    setToken(newToken)
    setUser(newUser)
  }

  const handleLogout = () => {
    localStorage.removeItem('token')
    setToken(null)
    setUser(null)
  }

  if (!token) {
    return <Login onLogin={handleLogin} />
  }

  return <Desktop user={user} onLogout={handleLogout} />
}

export default App

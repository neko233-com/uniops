import { useState } from 'react'

function App() {
  const [count, setCount] = useState(0)

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <h1 className="text-2xl p-4">UniOps</h1>
      <button
        className="bg-blue-500 px-4 py-2 rounded"
        onClick={() => setCount(c => c + 1)}
      >
        Count: {count}
      </button>
    </div>
  )
}

export default App

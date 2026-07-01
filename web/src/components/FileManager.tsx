import { useState, useEffect } from 'react'

interface FileItem {
  name: string
  size: number
  mode: string
  mod_time: string
  is_dir: boolean
}

interface FileManagerProps {
  serverId: number
}

export function FileManager({ serverId }: FileManagerProps) {
  const [files, setFiles] = useState<FileItem[]>([])
  const [currentPath, setCurrentPath] = useState('/')
  const [loading, setLoading] = useState(false)

  const fetchFiles = async (path: string) => {
    setLoading(true)
    try {
      const res = await fetch(`/api/files/${serverId}/list`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path }),
      })
      if (res.ok) {
        const data = await res.json()
        setFiles(data || [])
      }
    } catch (err) {
      console.error('Failed to fetch files:', err)
    }
    setLoading(false)
  }

  useEffect(() => {
    fetchFiles(currentPath)
  }, [currentPath])

  const handleDoubleClick = (file: FileItem) => {
    if (file.is_dir) {
      setCurrentPath(`${currentPath === '/' ? '' : currentPath}/${file.name}`)
    }
  }

  const handleBack = () => {
    if (currentPath === '/') return
    const parts = currentPath.split('/').filter(Boolean)
    parts.pop()
    setCurrentPath(parts.length === 0 ? '/' : '/' + parts.join('/'))
  }

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
  }

  return (
    <div className="flex flex-col gap-2" style={{ height: 'calc(100vh - 200px)' }}>
      {/* Path bar */}
      <div className="flex items-center gap-1 bg-gray-700 rounded px-2 py-1 text-sm">
        <button
          className="text-gray-400 hover:text-white px-1"
          onClick={() => setCurrentPath('/')}
        >
          /
        </button>
        {currentPath.split('/').filter(Boolean).map((part, i, arr) => (
          <span key={i} className="flex items-center">
            <span className="text-gray-500">/</span>
            <button
              className="text-gray-300 hover:text-white px-1"
              onClick={() => setCurrentPath('/' + arr.slice(0, i + 1).join('/'))}
            >
              {part}
            </button>
          </span>
        ))}
        <div className="flex-1" />
        <button
          className="text-gray-400 hover:text-white px-2 py-0.5 rounded bg-gray-600 hover:bg-gray-500"
          onClick={handleBack}
          disabled={currentPath === '/'}
        >
          ← Back
        </button>
      </div>

      {/* File list */}
      <div className="flex-1 overflow-auto bg-gray-900 rounded">
        {loading ? (
          <div className="p-4 text-center text-gray-400">Loading...</div>
        ) : files.length === 0 ? (
          <div className="p-4 text-center text-gray-500">Empty directory</div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-700 text-gray-400">
                <th className="text-left p-2">Name</th>
                <th className="text-right p-2 w-24">Size</th>
                <th className="text-right p-2 w-40">Modified</th>
                <th className="text-right p-2 w-24">Mode</th>
              </tr>
            </thead>
            <tbody>
              {files.map((file) => (
                <tr
                  key={file.name}
                  className="border-b border-gray-800 hover:bg-gray-800 cursor-pointer"
                  onDoubleClick={() => handleDoubleClick(file)}
                >
                  <td className="p-2">
                    <span className="mr-2">{file.is_dir ? '📁' : '📄'}</span>
                    {file.name}
                  </td>
                  <td className="p-2 text-right text-gray-400">
                    {file.is_dir ? '-' : formatSize(file.size)}
                  </td>
                  <td className="p-2 text-right text-gray-400">
                    {file.mod_time}
                  </td>
                  <td className="p-2 text-right text-gray-400 font-mono text-xs">
                    {file.mode}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}

interface WindowProps {
  title: string
  children: React.ReactNode
}

export function Window({ title, children }: WindowProps) {
  return (
    <div className="bg-gray-800 rounded-lg overflow-hidden h-full">
      <div className="bg-gray-700 px-4 py-2 flex items-center justify-between">
        <span className="text-white font-medium">{title}</span>
        <div className="flex gap-2">
          <button className="w-3 h-3 rounded-full bg-yellow-500" />
          <button className="w-3 h-3 rounded-full bg-green-500" />
          <button className="w-3 h-3 rounded-full bg-red-500" />
        </div>
      </div>
      <div className="p-4">
        {children}
      </div>
    </div>
  )
}

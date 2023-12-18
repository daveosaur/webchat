import { useEffect, createContext, useRef } from 'react'

const wsContext = createContext()

export default function WebSocketProvider({children}) {
	const ws = useRef(null)
	const channel = useRef([])

	useEffect(() => {
		ws.current = new WebSocket("ws://192.168.1.2:8080")
		ws.current.onopen = () => {
			channel.current.push("connected")
		}
		ws.current.onmessage = (msg) => {
			channel.current.push(msg)
		}

	})

}

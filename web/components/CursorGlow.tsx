'use client'

import { useEffect, useRef } from 'react'

export default function CursorGlow() {
  const glowRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    let raf: number
    let tx = -999, ty = -999
    let cx = -999, cy = -999

    const onMove = (e: MouseEvent) => { tx = e.clientX; ty = e.clientY }
    window.addEventListener('mousemove', onMove)

    // Smooth lerp toward cursor
    const tick = () => {
      cx += (tx - cx) * 0.08
      cy += (ty - cy) * 0.08
      if (glowRef.current) {
        glowRef.current.style.transform = `translate(${cx}px, ${cy}px)`
      }
      raf = requestAnimationFrame(tick)
    }
    raf = requestAnimationFrame(tick)

    return () => {
      window.removeEventListener('mousemove', onMove)
      cancelAnimationFrame(raf)
    }
  }, [])

  return (
    <div
      ref={glowRef}
      className="fixed pointer-events-none z-0 top-0 left-0"
      style={{
        width: 600,
        height: 600,
        marginLeft: -300,
        marginTop: -300,
        borderRadius: '50%',
        background: 'radial-gradient(circle, rgba(124,58,237,0.07) 0%, rgba(6,182,212,0.03) 40%, transparent 70%)',
        willChange: 'transform',
      }}
    />
  )
}

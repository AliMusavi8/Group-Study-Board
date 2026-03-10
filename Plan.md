# Group Study Board

A real-time collaborative whiteboard for studying with friends. Quick access with minimal barriers, supporting live drawing for all group members.

---

## Features

- **Quick Access:** Guest mode with auto-generated room IDs; optional login/signup for persistent boards  
- **Real-Time Drawing:** All users see updates immediately via WebSockets  
- **Board Management:** Create/join boards, optional naming or private links  
- **Optional Extras:** Chat, multiple boards, undo/redo, simple shapes/colors  

---

## Architecture Overview

- **MongoDB:** Persist boards, authentication, history, analytics  
- **Backend (Go):** Gin, handle WebSocket rooms, broadcast strokes, save board state  
- **Frontend (Angular):** HTML5 Canvas for drawing, WebSocket service, minimal UI with Tailwind CSS  

---

## MVP Strategy

1. Guest access with random room ID  
2. Real-time drawing (pen, erase, colors)  
3. Persistence and chat added later  

---

## Real-Time Communication

- **WebSockets:** Central server broadcasts events (recommended for MVP)  
- **WebRTC:** Optional for peer-to-peer, more complex  

---

## Conflict Handling

- **Authoritative server ordering:** Server timestamps and orders incoming stroke events, then broadcasts in that order  
- **Last-write-wins for pixels:** Clients render in received order; no merge conflicts at the stroke level  

---

## Board State Strategy

- **Event log + snapshots:** Store stroke events and periodic snapshots in MongoDB  
- **Join flow:** On room join, send latest snapshot, then stream remaining events and live updates  

---

## Room Access And Abuse Control

- **Room IDs:** Random, high-entropy IDs and share links for guest access  
- **Limits:** Max participants per room, basic rate limits on stroke events  
- **Cleanup:** Inactive room TTL and server-side eviction  

---

## Drawing Event Model

- **Stroke definition:** A stroke starts when mouse or touch input is pressed, continues with move events, ends on release  
- **Events:** `strokeStart`, `strokeMove`, `strokeEnd` with points, color, thickness, tool, timestamp  

---

## Recommendations

- Event model: `strokeStart`, `strokeMove`, `strokeEnd` with payload structure  
- Join flow: client requests room, server sends current snapshot + later events  
- Room lifecycle: inactive room TTL, max participants  
- Minimal auth: guest room ID entropy (e.g., 6–8 words or 10+ chars), optional share link  
- MVP non-goals: chat, shapes, undo/redo out of scope until v2  

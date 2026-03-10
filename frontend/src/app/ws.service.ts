export type Point = { x: number; y: number };

export type ClientStrokeEvent = {
  type: 'strokeStart' | 'strokeMove' | 'strokeEnd';
  point?: Point;
  color?: string;
  thickness?: number;
  tool?: 'pen' | 'eraser';
};

export type StrokeEvent = ClientStrokeEvent & {
  roomId: string;
  clientId: string;
  seq: number;
  serverTs: number;
};

export type SnapshotMessage = {
  type: 'snapshot';
  roomId: string;
  snapshot?: { events: StrokeEvent[]; createdAt: number };
  events: StrokeEvent[];
};

export type EventMessage = {
  type: 'event';
  event: StrokeEvent;
};

export type ErrorMessage = {
  type: 'error';
  message: string;
};

export type ServerMessage = SnapshotMessage | EventMessage | ErrorMessage;

const API_BASE = (window as any).GSB_API_BASE || 'http://localhost:8080';
const WS_BASE = API_BASE.replace(/^http/, 'ws');

import { Injectable } from '@angular/core';

@Injectable({ providedIn: 'root' })
export class WsService {
  private ws?: WebSocket;
  private listeners: Array<(msg: ServerMessage) => void> = [];

  async createRoom(): Promise<string> {
    const res = await fetch(`${API_BASE}/rooms`, { method: 'POST' });
    if (!res.ok) {
      throw new Error('Failed to create room');
    }
    const data = await res.json();
    return data.roomId as string;
  }

  connect(roomId: string, clientId: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws) {
        this.ws.close();
      }
      const url = `${WS_BASE}/rooms/${roomId}/ws?clientId=${encodeURIComponent(clientId)}`;
      const ws = new WebSocket(url);
      this.ws = ws;

      ws.onopen = () => resolve();
      ws.onerror = () => reject(new Error('WebSocket connection failed'));
      ws.onclose = () => {
        this.ws = undefined;
      };
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as ServerMessage;
          this.listeners.forEach((cb) => cb(data));
        } catch {
          // Ignore malformed messages
        }
      };
    });
  }

  send(event: ClientStrokeEvent): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      return;
    }
    this.ws.send(JSON.stringify(event));
  }

  onMessage(cb: (msg: ServerMessage) => void): void {
    this.listeners.push(cb);
  }
}

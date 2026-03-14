import { AfterViewInit, Component, ElementRef, HostListener, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { WsService, ClientStrokeEvent, Point, ServerMessage, StrokeEvent } from './ws.service';

const CANVAS_BG = '#f5f1e8';
const ERASER_SCALE = 3.5;

type Tool = 'pen' | 'eraser';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css'
})
export class AppComponent implements OnInit, AfterViewInit, OnDestroy {
  @ViewChild('canvas', { static: true }) canvasRef!: ElementRef<HTMLCanvasElement>;

  roomInput = '';
  roomId = '';
  status = 'disconnected';

  color = '#1f2937';
  thickness = 3;
  tool: Tool = 'pen';

  private ctrlDrawing = false;  // true while Left Ctrl is held and used as a draw trigger

  private ctx?: CanvasRenderingContext2D;
  private drawing = false;
  private lastPoint?: Point;
  private lastPointByClient = new Map<string, Point>();
  private clientId = this.createClientId();
  private deviceScale = 1;
  private resizeObserver?: ResizeObserver;

  constructor(private ws: WsService) {
    this.ws.onMessage((msg) => this.handleServerMessage(msg));
  }

  ngOnInit(): void {
    const urlRoom = this.getRoomFromUrl();
    if (urlRoom) {
      this.roomInput = urlRoom;
    }
  }

  ngAfterViewInit(): void {
    const canvas = this.canvasRef.nativeElement;
    const ctx = canvas.getContext('2d');
    if (!ctx) {
      return;
    }
    this.ctx = ctx;
    this.resizeCanvas();
    requestAnimationFrame(() => this.resizeCanvas());
    if (typeof ResizeObserver !== 'undefined') {
      this.resizeObserver = new ResizeObserver(() => this.resizeCanvas());
      this.resizeObserver.observe(canvas);
    }
  }

  ngOnDestroy(): void {
    this.resizeObserver?.disconnect();
  }

  @HostListener('window:resize')
  onResize(): void {
    this.resizeCanvas();
  }

  async createRoom(): Promise<void> {
    try {
      this.status = 'creating';
      const roomId = await this.ws.createRoom();
      this.roomInput = roomId;
      await this.joinRoom();
    } catch {
      this.status = 'error';
    }
  }

  async joinRoom(): Promise<void> {
    const roomId = this.roomInput.trim();
    if (!roomId) {
      return;
    }
    try {
      this.status = 'connecting';
      await this.ws.connect(roomId, this.clientId);
      this.roomId = roomId;
      this.status = 'connected';
      this.setRoomInUrl(roomId);
    } catch {
      this.status = 'error';
    }
  }

  async copyShareLink(): Promise<void> {
    if (!this.roomId) {
      return;
    }
    const link = `${window.location.origin}?room=${this.roomId}`;
    try {
      await navigator.clipboard.writeText(link);
    } catch {
      // Ignore clipboard failures
    }
  }

  onPointerDown(event: PointerEvent): void {
    if (!this.ctx) {
      return;
    }
    this.canvasRef.nativeElement.setPointerCapture?.(event.pointerId);
    const point = this.getPoint(event);
    this.startStroke(point);
  }

  onPointerMove(event: PointerEvent): void {
    if (!this.ctx) {
      return;
    }
    if (!this.drawing || !this.lastPoint) {
      return;
    }
    const point = this.getPoint(event);
    const thickness = this.getEffectiveThickness();
    this.drawLine(this.lastPoint, point, this.getStrokeColor(), thickness);
    this.lastPoint = point;
    this.sendEvent({ type: 'strokeMove', point });
  }

  onPointerUp(event: PointerEvent): void {
    this.canvasRef.nativeElement.releasePointerCapture?.(event.pointerId);
    this.endStroke();
  }

  private handleServerMessage(message: ServerMessage): void {
    if (!this.ctx) {
      return;
    }
    if (message.type === 'snapshot') {
      this.clearCanvas();
      if (message.snapshot?.events?.length) {
        message.snapshot.events.forEach((ev) => this.replayEvent(ev));
      }
      if (message.events?.length) {
        message.events.forEach((ev) => this.replayEvent(ev));
      }
      return;
    }

    if (message.type === 'event') {
      this.replayEvent(message.event);
    }
  }

  private replayEvent(event: StrokeEvent): void {
    if (!this.ctx) {
      return;
    }
    if (event.clientId === this.clientId) {
      return;
    }

    const color = event.tool === 'eraser' ? CANVAS_BG : (event.color || this.color);
    const tool: Tool = event.tool === 'eraser' ? 'eraser' : 'pen';
    const thickness = event.thickness ?? this.getEffectiveThickness(this.thickness, tool);

    if (event.type === 'clear') {
      this.clearCanvas();
      this.lastPointByClient.clear();
      return;
    }

    if (event.type === 'strokeStart' && event.point) {
      this.lastPointByClient.set(event.clientId, event.point);
      this.drawPoint(event.point, color, thickness);
      return;
    }

    if (event.type === 'strokeMove' && event.point) {
      const last = this.lastPointByClient.get(event.clientId);
      if (last) {
        this.drawLine(last, event.point, color, thickness);
      }
      this.lastPointByClient.set(event.clientId, event.point);
      return;
    }

    if (event.type === 'strokeEnd') {
      this.lastPointByClient.delete(event.clientId);
    }
  }

  private sendEvent(event: ClientStrokeEvent): void {
    if (!this.roomId) {
      return;
    }
    const tool: Tool = event.tool ?? this.tool;
    const thickness = this.getEffectiveThickness(event.thickness ?? this.thickness, tool);
    const payload: ClientStrokeEvent = {
      ...event,
      color: this.color,
      thickness,
      tool
    };
    this.ws.send(payload);
  }

  private drawLine(from: Point, to: Point, color: string, thickness: number): void {
    if (!this.ctx) {
      return;
    }
    const scale = this.deviceScale;
    this.ctx.strokeStyle = color;
    this.ctx.lineWidth = thickness * scale;
    this.ctx.lineCap = 'round';
    this.ctx.lineJoin = 'round';
    this.ctx.beginPath();
    this.ctx.moveTo(from.x * scale, from.y * scale);
    this.ctx.lineTo(to.x * scale, to.y * scale);
    this.ctx.stroke();
  }

  private drawPoint(point: Point, color: string, thickness = this.thickness): void {
    if (!this.ctx) {
      return;
    }
    const scale = this.deviceScale;
    this.ctx.fillStyle = color;
    this.ctx.beginPath();
    this.ctx.arc(point.x * scale, point.y * scale, (thickness / 2) * scale, 0, Math.PI * 2);
    this.ctx.fill();
  }

  private getPoint(event: PointerEvent): Point {
    if (Number.isFinite(event.offsetX) && Number.isFinite(event.offsetY)) {
      return { x: event.offsetX, y: event.offsetY };
    }
    const rect = this.canvasRef.nativeElement.getBoundingClientRect();
    return {
      x: event.clientX - rect.left,
      y: event.clientY - rect.top
    };
  }

  private resizeCanvas(): void {
    const canvas = this.canvasRef.nativeElement;
    const rect = canvas.getBoundingClientRect();
    this.deviceScale = window.devicePixelRatio || 1;
    canvas.width = rect.width * this.deviceScale;
    canvas.height = rect.height * this.deviceScale;
    if (!this.ctx) {
      return;
    }
    this.ctx.setTransform(1, 0, 0, 1, 0, 0);
    this.clearCanvas();
  }

  private clearCanvas(): void {
    if (!this.ctx) {
      return;
    }
    const canvas = this.canvasRef.nativeElement;
    this.ctx.clearRect(0, 0, canvas.width, canvas.height);
    this.ctx.fillStyle = CANVAS_BG;
    this.ctx.fillRect(0, 0, canvas.width, canvas.height);
  }

  private getStrokeColor(): string {
    return this.tool === 'eraser' ? CANVAS_BG : this.color;
  }

  private getEffectiveThickness(thickness = this.thickness, tool: Tool = this.tool): number {
    return tool === 'eraser' ? thickness * ERASER_SCALE : thickness;
  }

  private startStroke(point: Point): void {
    this.drawing = true;
    this.lastPoint = point;
    this.drawPoint(point, this.getStrokeColor(), this.getEffectiveThickness());
    this.sendEvent({ type: 'strokeStart', point });
  }

  private endStroke(): void {
    if (!this.drawing) {
      return;
    }
    this.drawing = false;
    this.lastPoint = undefined;
    this.sendEvent({ type: 'strokeEnd' });
  }

  @HostListener('window:keydown', ['$event'])
  onKeyDown(event: KeyboardEvent): void {
    // Left Ctrl acts as a virtual "mouse button" for touchpad drawing.
    // Holding it starts a stroke at the current cursor position on the canvas.
    if (event.code !== 'ControlLeft' || this.ctrlDrawing) {
      return;
    }
    const tag = (event.target as HTMLElement)?.tagName?.toLowerCase();
    if (tag === 'input' || tag === 'textarea' || tag === 'select') {
      return;
    }
    this.ctrlDrawing = true;
    // Synthesise a strokeStart at wherever the pointer currently is.
    // We rely on the next pointermove to get the real position;
    // store the canvas centre as a safe fallback start point.
    if (!this.drawing && this.ctx) {
      const canvas = this.canvasRef.nativeElement;
      const rect = canvas.getBoundingClientRect();
      // Use the last known mouse position if available, else canvas centre.
      const x = this.lastCtrlPoint?.x ?? rect.width / 2;
      const y = this.lastCtrlPoint?.y ?? rect.height / 2;
      this.startStroke({ x, y });
    }
  }

  @HostListener('window:keyup', ['$event'])
  onKeyUp(event: KeyboardEvent): void {
    if (event.code !== 'ControlLeft') {
      return;
    }
    this.ctrlDrawing = false;
    this.endStroke();
  }

  // Tracks the last cursor position on the canvas so Ctrl+hold can start
  // a stroke there rather than at the canvas centre.
  private lastCtrlPoint?: Point;

  onCanvasMouseMove(event: MouseEvent): void {
    const rect = this.canvasRef.nativeElement.getBoundingClientRect();
    this.lastCtrlPoint = {
      x: event.clientX - rect.left,
      y: event.clientY - rect.top
    };
    if (this.ctrlDrawing && this.drawing && this.lastPoint) {
      const point = this.lastCtrlPoint;
      const thickness = this.getEffectiveThickness();
      this.drawLine(this.lastPoint, point, this.getStrokeColor(), thickness);
      this.lastPoint = point;
      this.sendEvent({ type: 'strokeMove', point });
    }
  }

  private createClientId(): string {
    if (crypto && 'randomUUID' in crypto) {
      return crypto.randomUUID();
    }
    return `guest-${Math.random().toString(36).slice(2, 10)}`;
  }

  private getRoomFromUrl(): string | null {
    const params = new URLSearchParams(window.location.search);
    return params.get('room');
  }

  private setRoomInUrl(roomId: string): void {
    const params = new URLSearchParams(window.location.search);
    params.set('room', roomId);
    const newUrl = `${window.location.pathname}?${params.toString()}`;
    window.history.replaceState({}, '', newUrl);
  }

  clearBoard(): void {
    this.clearCanvas();
    this.lastPoint = undefined;
    this.lastPointByClient.clear();
    this.sendEvent({ type: 'clear' });
  }
}

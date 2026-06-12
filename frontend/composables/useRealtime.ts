/**
 * useRealtime — single-shared WebSocket connection to /api/ws.
 *
 * The hub authenticates by the same JWT used for REST (passed as a `token`
 * query param because browsers cannot set custom headers on WS handshakes).
 *
 * One global socket per app: all consumers share `state` (online users +
 * latest event); subscribers receive every event matching a given topic
 * prefix (e.g. "notif", "tasks", "kanban").
 */
import type { Ref } from "vue";

interface RealtimeEvent {
  topic: string;
  action?: string;
  from?: number;
  to?: number;
  payload?: any;
  at?: number;
}

type Listener = (ev: RealtimeEvent) => void;

interface Realtime {
  connected: Ref<boolean>;
  online: Ref<number[]>;
  lastEvent: Ref<RealtimeEvent | null>;
  on: (topicPrefix: string, fn: Listener) => () => void;
  send: (ev: RealtimeEvent) => void;
}

let _instance: Realtime | null = null;

export const useRealtime = (): Realtime => {
  if (_instance) return _instance;

  const auth = useAuthStore();
  const baseURL = useRuntimeConfig().public.apiBase || "";
  const connected = ref(false);
  const online = ref<number[]>([]);
  const lastEvent = ref<RealtimeEvent | null>(null);
  const listeners = new Map<string, Set<Listener>>();

  let ws: WebSocket | null = null;
  let retry = 0;
  let stopped = false;

  // Build the WS URL. Convert http→ws / https→wss based on the base URL or
  // the current page (fallback for "/api" relative paths in dev).
  const buildURL = (): string => {
    const token = auth.token || "";
    const url = (baseURL || window.location.origin).replace(/^http/, "ws");
    return url + "/api/ws?token=" + encodeURIComponent(token);
  };

  const open = () => {
    if (stopped || !auth.token) return;
    try {
      ws = new WebSocket(buildURL());
    } catch {
      return scheduleReconnect();
    }
    ws.addEventListener("open", () => {
      connected.value = true;
      retry = 0;
    });
    ws.addEventListener("message", (e) => {
      try {
        const ev = JSON.parse(e.data) as RealtimeEvent;
        lastEvent.value = ev;
        if (ev.topic === "presence") {
          // The hub pushes per-user presence; we keep a simple online set.
          const uid = ev.payload?.user_id as number | undefined;
          if (typeof uid === "number") {
            const isOnline = !!ev.payload?.online;
            const set = new Set(online.value);
            isOnline ? set.add(uid) : set.delete(uid);
            online.value = [...set];
          }
        }
        for (const [prefix, fns] of listeners) {
          if (ev.topic === prefix || ev.topic.startsWith(prefix + ":")) {
            fns.forEach((f) => {
              try { f(ev); } catch { /* */ }
            });
          }
        }
      } catch { /* */ }
    });
    ws.addEventListener("close", () => {
      connected.value = false;
      scheduleReconnect();
    });
    ws.addEventListener("error", () => {
      connected.value = false;
    });
  };

  const scheduleReconnect = () => {
    if (stopped) return;
    retry++;
    const delay = Math.min(30000, 500 * Math.pow(2, Math.min(retry, 6)));
    setTimeout(open, delay);
  };

  const send = (ev: RealtimeEvent) => {
    try {
      ws?.send(JSON.stringify(ev));
    } catch { /* */ }
  };

  const on = (topicPrefix: string, fn: Listener) => {
    if (!listeners.has(topicPrefix)) listeners.set(topicPrefix, new Set());
    listeners.get(topicPrefix)!.add(fn);
    return () => listeners.get(topicPrefix)?.delete(fn);
  };

  // Auto-reconnect when the user logs in (token appears).
  if (typeof window !== "undefined") {
    if (auth.token) open();
    watch(
      () => auth.token,
      (t) => {
        try { ws?.close(); } catch { /* */ }
        ws = null;
        connected.value = false;
        if (t) open();
      }
    );
    window.addEventListener("beforeunload", () => {
      stopped = true;
      try { ws?.close(); } catch { /* */ }
    });
  }

  _instance = { connected, online, lastEvent, on, send };
  return _instance;
};

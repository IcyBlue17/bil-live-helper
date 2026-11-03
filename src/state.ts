// App state and the actions that call Go. Views read the signals here and
// call the actions; nothing else talks to the backend.
import { Channel } from "mygo-runtime";
import {
  Auth,
  Danmaku,
  Live,
  System,
  events,
  type AppInfo,
  type AreaGroup,
  type ConnState,
  type FeedEvent,
  type LogEntry,
  type LogLevel,
  type Message,
  type QRUpdate,
  type RoomInfo,
  type RoomSettings,
  type StartUpdate,
  type StreamInfo,
  type User,
} from "./mygo";
import { batch, computed, effect, signal } from "./lib/signal";
import { errorMessage } from "./lib/format";
import { toast } from "./ui/toast";

export type Phase = "boot" | "unauth" | "app" | "error";

export const phase = signal<Phase>("boot");
export const bootError = signal("");
export const info = signal<AppInfo | null>(null);
export const user = signal<User | null>(null);
export const room = signal<RoomInfo | null>(null);
export const stream = signal<StreamInfo | null>(null);
export const areas = signal<AreaGroup[]>([]);
export const settings = signal<RoomSettings | null>(null);
export const refreshing = signal(false);

/** Ticks every second, for durations. */
export const now = signal(Date.now());
setInterval(() => now.set(Date.now()), 1000);

// ---- Activity log ----

export const logs = signal<LogEntry[]>([]);
export const unreadLogs = signal(0);
export const logOpen = signal(false);

function pushLog(entry: LogEntry) {
  logs.update((l) => [...l.slice(-299), entry]);
  if (!logOpen.peek()) unreadLogs.update((n) => n + 1);
}

/** Adds a line for an action the backend does not log itself. */
export function log(level: LogLevel, message: string) {
  pushLog({ time: new Date().toISOString(), level, message });
}

events.log.on(pushLog);

function fail(prefix: string, err: unknown): string {
  const msg = `${prefix}：${errorMessage(err)}`;
  toast(msg, "error");
  log("error", msg);
  return msg;
}

// ---- Boot and auth ----

export async function boot() {
  phase.set("boot");
  try {
    const [appInfo, auth] = await Promise.all([System.info(), Auth.state()]);
    info.set(appInfo);
    if (auth.loggedIn) await enterApp(auth.user);
    else phase.set("unauth");
  } catch (err) {
    bootError.set(errorMessage(err));
    phase.set("error");
  }
}

async function enterApp(u: User | null) {
  user.set(u);
  try {
    const [r, s, a, st] = await Promise.all([Live.room(), Live.stream(), Live.areas(), Live.settings()]);
    batch(() => {
      room.set(r);
      stream.set(s);
      areas.set(a);
      settings.set(st);
      phase.set("app");
    });
  } catch (err) {
    phase.set("app");
    fail("加载直播间失败", err);
  }
}

events.authChanged.on((s) => {
  if (s.loggedIn && phase.peek() !== "app") void enterApp(s.user);
  else if (!s.loggedIn && phase.peek() === "app") leaveApp();
});

function leaveApp() {
  batch(() => {
    user.set(null);
    room.set(null);
    stream.set(null);
    phase.set("unauth");
  });
}

export type LoginView = { stage: QRUpdate["stage"] | "loading" | "error"; svg: string; message: string };
export const login = signal<LoginView>({ stage: "loading", svg: "", message: "" });

/** Starts a QR login; returns a canceller. */
export function startLogin(): () => void {
  login.set({ stage: "loading", svg: "", message: "正在获取二维码…" });
  const ch = new Channel<QRUpdate>((u) => {
    login.update((v) => ({ stage: u.stage, svg: u.code?.svg ?? v.svg, message: u.message }));
  });
  Auth.login(ch)
    .then(() => {
      if (login.peek().stage === "done" && phase.peek() !== "app") void enterApp(null).then(refreshUser);
    })
    .catch((err: unknown) => {
      if (ch.closed && login.peek().stage !== "error") return; // cancelled
      login.update((v) => ({ ...v, stage: "error", message: errorMessage(err) }));
    });
  return () => {
    if (!ch.closed) ch.close();
  };
}

async function refreshUser() {
  try {
    const s = await Auth.state();
    if (s.user) user.set(s.user);
  } catch {
    /* the authChanged event carries it as well */
  }
}

export async function logout() {
  try {
    await Auth.logout();
    leaveApp();
  } catch (err) {
    fail("退出登录失败", err);
  }
}

// ---- Room ----

events.roomChanged.on((r) => room.set(r));

export async function refreshRoom() {
  if (refreshing.peek()) return;
  refreshing.set(true);
  try {
    const [r, s] = await Promise.all([Live.room(), Live.stream()]);
    batch(() => {
      room.set(r);
      if (s) stream.set(s);
    });
    toast("直播间信息已刷新", "ok");
  } catch (err) {
    fail("刷新失败", err);
  } finally {
    refreshing.set(false);
  }
}

export async function saveSettings(s: RoomSettings): Promise<boolean> {
  try {
    const r = await Live.saveSettings(s);
    batch(() => {
      room.set(r);
      settings.set(s);
    });
    toast("直播间设置已保存", "ok");
    return true;
  } catch (err) {
    fail("保存失败", err);
    return false;
  }
}

// ---- Broadcast ----

export type LiveAction = "idle" | "starting" | "faceAuth" | "stopping";
export const liveAction = signal<LiveAction>("idle");
export const face = signal<{ stage: QRUpdate["stage"]; svg: string; message: string } | null>(null);
let startChannel: Channel<StartUpdate> | null = null;

export const isLive = () => room()?.status === "live";

export function startLive() {
  if (liveAction.peek() !== "idle") return;
  liveAction.set("starting");
  const ch = new Channel<StartUpdate>((u) => {
    if (u.stage === "faceAuth" && u.face) {
      const f = u.face;
      face.update((v) => ({ stage: f.stage, svg: f.code?.svg ?? v?.svg ?? "", message: f.message || u.message }));
      liveAction.set("faceAuth");
    } else if (u.stage === "live") {
      face.set(null);
      if (u.stream) stream.set(u.stream);
    } else if (u.stage === "requesting") {
      liveAction.set("starting");
    }
  });
  startChannel = ch;
  Live.start(ch)
    .then((s) => {
      stream.set(s);
      toast(s.changed ? "开播成功 · 推流码已变更" : "开播成功", s.changed ? "warn" : "ok");
    })
    .catch((err: unknown) => {
      if (ch.closed && startChannel !== ch) return; // cancelled by the user
      fail("开播失败", err);
    })
    .finally(() => {
      if (startChannel === ch) startChannel = null;
      face.set(null);
      liveAction.set("idle");
    });
}

/** Cancels a running start, e.g. from the face-auth dialog. */
export function cancelStart() {
  const ch = startChannel;
  if (!ch) return;
  startChannel = null;
  ch.close();
  face.set(null);
  liveAction.set("idle");
  log("info", "已取消开播");
}

export async function stopLive() {
  if (liveAction.peek() !== "idle") return;
  liveAction.set("stopping");
  try {
    await Live.stop();
    toast("已结束直播", "ok");
  } catch (err) {
    fail("下播失败", err);
  } finally {
    liveAction.set("idle");
  }
}

// ---- Clipboard ----

export async function copy(text: string, label: string): Promise<boolean> {
  try {
    await System.copy(text);
    toast(`${label}已复制`, "ok", 1800);
    log("info", `${label}已复制到剪贴板`);
    return true;
  } catch (err) {
    fail("复制失败", err);
    return false;
  }
}

export function openURL(url: string) {
  System.openURL(url).catch((err: unknown) => fail("无法打开链接", err));
}

// ---- Danmaku ----

export const MAX_FEED = 500;
export type FeedItem = Message & { fresh: boolean };
export const feed = signal<FeedItem[]>([]);
export const conn = signal<ConnState>("closed");
export const popularity = signal(0);
/** Messages appended since the last render, consumed by the feed view. */
export const feedListeners = new Set<(items: FeedItem[], reset: boolean) => void>();

function emitFeed(items: FeedItem[], reset: boolean) {
  for (const l of feedListeners) l(items, reset);
}

let pendingFeed: FeedItem[] = [];
let feedFrame = 0;
const LOCAL_ECHO_MS = 15000;
function queueMessage(m: Message) {
  // A comment we inserted locally when sending: drop the server's echo.
  if (!m.id.startsWith("local-") && (m.kind === "self" || m.uid === user.peek()?.uid)) {
    const t = Date.now();
    const echo = feed.peek().find((f) => f.id.startsWith("local-") && f.text === m.text && t - Date.parse(f.time) < LOCAL_ECHO_MS);
    if (echo && !m.historical) return;
  }
  pendingFeed.push({ ...m, fresh: !m.historical });
  if (!feedFrame)
    feedFrame = requestAnimationFrame(() => {
      feedFrame = 0;
      const items = pendingFeed;
      pendingFeed = [];
      feed.update((f) => [...f, ...items].slice(-MAX_FEED));
      emitFeed(items, false);
    });
}

// One watch per logged-in room: restarts when the room changes, ends on logout.
const watchedRoom = computed(() => (phase() === "app" ? room()?.roomId : undefined));
effect(() => {
  const id = watchedRoom();
  if (!id) return;
  return watchRoom(id);
});

function watchRoom(roomId: number): () => void {
  let stopped = false;
  let ch: Channel<FeedEvent> | null = null;
  let retry: ReturnType<typeof setTimeout> | undefined;
  batch(() => {
    feed.set([]);
    popularity.set(0);
  });
  emitFeed([], true);
  const run = () => {
    const c = new Channel<FeedEvent>((e) => {
      if (e.state) conn.set(e.state);
      if (e.popularity) popularity.set(e.popularity);
      if (e.message) queueMessage(e.message);
    });
    ch = c;
    conn.set("connecting");
    Danmaku.watch(c)
      .catch((err: unknown) => {
        if (!stopped) log("warn", `弹幕连接中断：${errorMessage(err)}`);
      })
      .finally(() => {
        if (stopped) return;
        // The watch ended on its own: reconnect after a pause.
        conn.set("reconnecting");
        retry = setTimeout(run, 3000);
      });
  };
  run();
  log("info", `开始接收直播间 ${roomId} 的弹幕`);
  return () => {
    stopped = true;
    clearTimeout(retry);
    if (ch && !ch.closed) ch.close();
    conn.set("closed");
  };
}

let localSeq = 0;
/** Sends a comment and shows it in the feed right away. */
export async function sendDanmaku(text: string) {
  await Danmaku.send(text);
  const u = user.peek();
  queueMessage({
    id: `local-${++localSeq}`, kind: "self", time: new Date().toISOString(), uid: u?.uid ?? 0, name: u?.name ?? "我",
    text, admin: false, guard: 0, userLevel: 0, medal: null, gift: null, historical: false,
    face: u?.face ?? "", emoticon: null, emots: {}, superChat: null,
  });
}

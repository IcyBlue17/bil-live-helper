// App state (Svelte runes) and the actions that call Go. Components read the
// stores here and call the actions; nothing else talks to the backend.
import { Channel } from "mygo-runtime";
import { untrack } from "svelte";
import { toast } from "svelte-sonner";
import {
  Auth,
  Danmaku,
  Live,
  System,
  events,
  type AppInfo,
  type AreaGroup,
  type ConnState,
  type Emoticon,
  type FeedEvent,
  type LogEntry,
  type LogLevel,
  type Message,
  type QRUpdate,
  type RoomInfo,
  type RoomSettings,
  type SessionStats,
  type StartUpdate,
  type StreamInfo,
  type User,
} from "../mygo";
import { errorMessage } from "./format";

export type Phase = "boot" | "unauth" | "app" | "error";

class AppStore {
  phase = $state<Phase>("boot");
  bootError = $state("");
  info = $state<AppInfo | null>(null);
  user = $state<User | null>(null);
  room = $state<RoomInfo | null>(null);
  stream = $state<StreamInfo | null>(null);
  areas = $state.raw<AreaGroup[]>([]);
  settings = $state<RoomSettings | null>(null);
  refreshing = $state(false);
  /** Ticks every second, for durations. */
  now = $state(Date.now());
  isLive = $derived(this.room?.status === "live");
}
export const app = new AppStore();
setInterval(() => (app.now = Date.now()), 1000);

// ---- Activity log ----

class LogStore {
  entries = $state.raw<LogEntry[]>([]);
  unread = $state(0);
  open = $state(false);
}
export const logs = new LogStore();

function pushLog(entry: LogEntry) {
  logs.entries = [...logs.entries.slice(-299), entry];
  if (!logs.open) logs.unread++;
}

/** Adds a line for an action the backend does not log itself. */
export function log(level: LogLevel, message: string) {
  pushLog({ time: new Date().toISOString(), level, message });
}
events.log.on(pushLog);

/** Reports a failed action as a toast and a log line. */
export function fail(prefix: string, err: unknown): string {
  const msg = `${prefix}：${errorMessage(err)}`;
  toast.error(msg);
  log("error", msg);
  return msg;
}

// ---- Boot and auth ----

export async function boot() {
  app.phase = "boot";
  try {
    const [info, auth] = await Promise.all([System.info(), Auth.state()]);
    app.info = info;
    if (auth.loggedIn) await enterApp(auth.user);
    else app.phase = "unauth";
  } catch (err) {
    app.bootError = errorMessage(err);
    app.phase = "error";
  }
}

async function enterApp(u: User | null) {
  if (u) app.user = u;
  try {
    const [r, s, a, st] = await Promise.all([Live.room(), Live.stream(), Live.areas(), Live.settings()]);
    app.room = r;
    app.stream = s;
    app.areas = a;
    app.settings = st;
  } catch (err) {
    fail("加载直播间失败", err);
  }
  app.phase = "app";
}

events.authChanged.on((s) => {
  if (s.loggedIn && app.phase !== "app") void enterApp(s.user);
  else if (s.loggedIn && s.user) app.user = s.user;
  else if (!s.loggedIn && app.phase === "app") leaveApp();
});

function leaveApp() {
  app.user = null;
  app.room = null;
  app.stream = null;
  app.phase = "unauth";
}

export type LoginStage = QRUpdate["stage"] | "loading" | "error";
export const login = $state({ stage: "loading" as LoginStage, svg: "", message: "" });

/** Starts a QR login; returns a canceller. */
export function startLogin(): () => void {
  Object.assign(login, { stage: "loading", svg: "", message: "正在获取二维码…" });
  const ch = new Channel<QRUpdate>((u) => {
    login.stage = u.stage;
    if (u.code) login.svg = u.code.svg;
    login.message = u.message;
  });
  Auth.login(ch)
    .then(async () => {
      if (login.stage !== "done" || app.phase === "app") return;
      const s = await Auth.state().catch(() => null);
      await enterApp(s?.user ?? null);
    })
    .catch((err: unknown) => {
      if (ch.closed && login.stage !== "error") return; // cancelled
      login.stage = "error";
      login.message = errorMessage(err);
    });
  return () => {
    if (!ch.closed) ch.close();
  };
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

events.roomChanged.on((r) => (app.room = r));

export async function refreshRoom() {
  if (app.refreshing) return;
  app.refreshing = true;
  try {
    const [r, s] = await Promise.all([Live.room(), Live.stream()]);
    app.room = r;
    if (s) app.stream = s;
    toast.success("直播间信息已刷新");
  } catch (err) {
    fail("刷新失败", err);
  } finally {
    app.refreshing = false;
  }
}

export async function saveSettings(s: RoomSettings): Promise<boolean> {
  try {
    app.room = await Live.saveSettings(s);
    app.settings = s;
    toast.success("直播间设置已保存");
    return true;
  } catch (err) {
    fail("保存失败", err);
    return false;
  }
}

// ---- Broadcast ----

export type LiveAction = "idle" | "starting" | "faceAuth" | "stopping";
class LiveStore {
  action = $state<LiveAction>("idle");
  face = $state<{ stage: QRUpdate["stage"]; svg: string; message: string } | null>(null);
  confirmingStop = $state(false);
}
export const live = new LiveStore();
let startChannel: Channel<StartUpdate> | null = null;

export function startLive() {
  if (live.action !== "idle") return;
  live.action = "starting";
  const ch = new Channel<StartUpdate>((u) => {
    if (u.stage === "faceAuth" && u.face) {
      const f = u.face;
      live.face = { stage: f.stage, svg: f.code?.svg ?? live.face?.svg ?? "", message: f.message || u.message };
      live.action = "faceAuth";
    } else if (u.stage === "live") {
      live.face = null;
      if (u.stream) app.stream = u.stream;
    } else if (u.stage === "requesting") {
      live.action = "starting";
    }
  });
  startChannel = ch;
  Live.start(ch)
    .then((s) => {
      app.stream = s;
      if (s.changed) toast.warning("开播成功 · 推流码已变更，请更新 OBS");
      else toast.success("开播成功");
    })
    .catch((err: unknown) => {
      if (ch.closed && startChannel !== ch) return; // cancelled by the user
      fail("开播失败", err);
    })
    .finally(() => {
      if (startChannel === ch) startChannel = null;
      live.face = null;
      live.action = "idle";
    });
}

/** Cancels a running start, e.g. from the face-auth dialog. */
export function cancelStart() {
  const ch = startChannel;
  if (!ch) return;
  startChannel = null;
  ch.close();
  live.face = null;
  live.action = "idle";
  log("info", "已取消开播");
}

export async function stopLive() {
  if (live.action !== "idle") return;
  live.confirmingStop = false;
  live.action = "stopping";
  try {
    await Live.stop();
    toast.success("已结束直播");
  } catch (err) {
    fail("下播失败", err);
  } finally {
    live.action = "idle";
  }
}

/** What Space does: start, ask to stop, or confirm the stop. */
export function toggleLive() {
  if (live.action !== "idle" || !app.room) return;
  if (!app.isLive) startLive();
  else if (!live.confirmingStop) live.confirmingStop = true;
  else void stopLive();
}

// ---- Clipboard and links ----

export async function copy(text: string, label: string): Promise<boolean> {
  try {
    await System.copy(text);
    toast.success(`${label}已复制`, { duration: 1800 });
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

export const streamText = (s: StreamInfo) => `服务器：${s.addr}\n推流码：${s.code}`;

// ---- Danmaku ----

export const MAX_FEED = 500;
/** A feed row: a message, or the divider after the history. */
export type FeedItem = (Message & { arrived: number }) | { kind: "divider"; id: string; arrived: number };

class FeedStore {
  items = $state.raw<FeedItem[]>([]);
  conn = $state<ConnState>("closed");
  popularity = $state(0);
  watched = $state(0);
  online = $state(0);
  likes = $state(0);
  stats = $state.raw<SessionStats | null>(null);
}
export const feed = new FeedStore();
/** Called with each batch of new rows (and with reset=true when the feed restarts). */
export const feedListeners = new Set<(items: FeedItem[], reset: boolean) => void>();

let pending: FeedItem[] = [];
let frame = 0;
let hasDivider = false;
let sawHistory = false;
const LOCAL_ECHO_MS = 15000;

function queueMessage(m: Message) {
  // A comment we inserted locally when sending: drop the server's echo.
  if (!m.id.startsWith("local-") && !m.historical && (m.kind === "self" || m.uid === app.user?.uid)) {
    const t = Date.now();
    const echo = [...feed.items, ...pending].some(
      (f) => f.kind !== "divider" && f.id.startsWith("local-") && f.text === m.text && t - f.arrived < LOCAL_ECHO_MS,
    );
    if (echo) return;
  }
  if (m.historical) sawHistory = true;
  else if (sawHistory && !hasDivider) {
    hasDivider = true;
    pending.push({ kind: "divider", id: "divider", arrived: Date.now() });
  }
  pending.push({ ...m, arrived: m.historical ? 0 : Date.now() });
  if (!frame)
    frame = requestAnimationFrame(() => {
      frame = 0;
      const batch = pending;
      pending = [];
      feed.items = [...feed.items, ...batch].slice(-MAX_FEED);
      for (const l of feedListeners) l(batch, false);
    });
}

function resetFeed() {
  cancelAnimationFrame(frame);
  frame = 0;
  pending = [];
  hasDivider = sawHistory = false;
  feed.items = [];
  feed.popularity = feed.watched = feed.online = feed.likes = 0;
  feed.stats = null;
  for (const l of feedListeners) l([], true);
}

function onFeedEvent(e: FeedEvent) {
  if (e.state) feed.conn = e.state;
  if (e.popularity) feed.popularity = e.popularity;
  if (e.watched) feed.watched = e.watched;
  if (e.online) feed.online = e.online;
  if (e.likes) feed.likes = e.likes;
  if (e.stats) feed.stats = e.stats;
  if (e.message) queueMessage(e.message);
}

function watchRoom(roomId: number): () => void {
  let stopped = false;
  let ch: Channel<FeedEvent> | null = null;
  let retry: ReturnType<typeof setTimeout> | undefined;
  resetFeed();
  const run = () => {
    const c = new Channel<FeedEvent>(onFeedEvent);
    ch = c;
    feed.conn = "connecting";
    Danmaku.watch(c)
      .catch((err: unknown) => {
        if (!stopped) log("warn", `弹幕连接中断：${errorMessage(err)}`);
      })
      .finally(() => {
        if (stopped) return;
        // The watch ended on its own: reconnect after a pause.
        feed.conn = "reconnecting";
        retry = setTimeout(run, 3000);
      });
  };
  run();
  log("info", `开始接收直播间 ${roomId} 的弹幕`);
  return () => {
    stopped = true;
    clearTimeout(retry);
    if (ch && !ch.closed) ch.close();
    feed.conn = "closed";
  };
}

// One watch per logged-in room: restarts when the room changes, ends on logout.
$effect.root(() => {
  const roomId = $derived(app.phase === "app" ? app.room?.roomId : undefined);
  $effect(() => {
    const id = roomId;
    // Untracked: watchRoom reads the feed and log state it writes.
    if (id) return untrack(() => watchRoom(id));
  });
});

let localSeq = 0;
function localMessage(text: string, emoticon: Emoticon | null): Message {
  const u = app.user;
  return {
    id: `local-${++localSeq}`, kind: "self", time: new Date().toISOString(), uid: u?.uid ?? 0, name: u?.name ?? "我",
    text, admin: false, guard: 0, userLevel: 0, medal: null, gift: null, historical: false,
    face: u?.face ?? "", emoticon, emots: {}, superChat: null,
  };
}

/** Sends a comment and shows it in the feed right away. */
export async function sendDanmaku(text: string) {
  await Danmaku.send(text);
  queueMessage(localMessage(text, null));
}

/** Sends a sticker (表情包弹幕) and shows it in the feed right away. */
export async function sendSticker(e: Emoticon) {
  await Danmaku.sendEmoticon(e.token);
  queueMessage(localMessage(e.name, e));
}

// ---- Dialog and popover state shared across views ----

class UiStore {
  aboutOpen = $state(false);
  prefsOpen = $state(false);
  areaOpen = $state(false);
  /** The page tab under the top bar (概览, 本场数据, 榜单, …). */
  tab = $state<"overview" | "stats" | "rank" | "announce" | "muted">("overview");
  /** Which column shows in the narrow (tabbed) layout. */
  narrowTab = $state<"control" | "danmaku">("control");
  /** The viewer whose card is open, and the element it points at. */
  viewer = $state<{ uid: number; name: string; anchor: HTMLElement } | null>(null);
  /** Bumped to ask the title input to take focus. */
  focusTitle = $state(0);
  /** Bumped when the mute list changed (mute or unmute anywhere). */
  mutedVersion = $state(0);
}
export const ui = new UiStore();

// Placeholder page proving the Go ↔ TypeScript contract works end to end.
// The real UI replaces this file.
import { Auth, Live, System } from "./mygo";

const root = document.querySelector<HTMLElement>("#app")!;
const [info, auth, room, areas] = await Promise.all([
  System.info(),
  Auth.state(),
  Live.room(),
  Live.areas(),
]);
root.innerHTML = `
  <h1>Bil Live Helper <small>v${info.version}${info.demo ? " · demo" : ""}</small></h1>
  <p>Contract check — the real interface lands in the frontend PR.</p>
  <dl>
    <dt>Auth</dt><dd>${auth.loggedIn ? auth.user?.name : "未登录"}</dd>
    <dt>Room</dt><dd>${room.roomId} · ${room.title} · ${room.parentAreaName}/${room.areaName} · ${room.status}</dd>
    <dt>Areas</dt><dd>${areas.map((g) => g.name).join(" / ")}</dd>
  </dl>`;

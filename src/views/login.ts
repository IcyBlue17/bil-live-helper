import { h, html, text } from "../lib/dom";
import { effect, onCleanup, signal } from "../lib/signal";
import { info, login, startLogin } from "../state";
import { icons, mark } from "../ui/icons";
import { spinner } from "../ui/widgets";

const stageLabel: Record<string, string> = {
  loading: "正在获取二维码…",
  waiting: "请使用哔哩哔哩 App 扫码登录",
  scanned: "已扫码，请在手机上确认",
  done: "登录成功，正在进入…",
  expired: "二维码已过期",
  error: "获取二维码失败",
};

export function LoginView() {
  const attempt = signal(0);
  effect(() => {
    attempt();
    return startLogin();
  });
  const retry = () => attempt.update((n) => n + 1);
  const stage = () => login().stage;

  const qrSlot = h("div", { class: "qr-code" });
  effect(() => {
    const svg = login().svg;
    qrSlot.replaceChildren(svg ? html(svg) : spinner());
  });

  const onKey = (e: KeyboardEvent) => {
    if ((e.key === "r" || e.key === "R") && (stage() === "expired" || stage() === "error")) retry();
  };
  window.addEventListener("keydown", onKey);
  onCleanup(() => window.removeEventListener("keydown", onKey));

  return h(
    "div",
    { class: "login-screen" },
    h(
      "section",
      { class: "login-card" },
      h(
        "div",
        { class: "login-brand" },
        html(mark),
        h("div", null, h("h1", null, "Bil Live Helper"), h("p", { class: "muted" }, "B 站直播推流助手")),
      ),
      h(
        "p",
        { class: "login-intro" },
        "扫码登录后，即可一键开播、获取 RTMP 推流地址与推流码，修改标题与分区，并实时查看弹幕。",
      ),
      h(
        "div",
        { class: () => `qr-tile stage-${stage()}` },
        qrSlot,
        h("div", { class: "qr-scanned" }, html(icons.phone), h("span", null, "请在手机上确认")),
        h(
          "div",
          { class: "qr-expired" },
          h("p", null, text(() => (stage() === "error" ? "获取失败" : "二维码已过期"))),
          h("button", { class: "btn btn-primary btn-sm", onclick: retry }, html(icons.refresh), "刷新二维码"),
        ),
      ),
      h(
        "div",
        { class: () => `login-stage stage-${stage()}` },
        h("span", { class: "stage-dot" }),
        h("span", { class: "stage-text" }, text(() => {
          const v = login();
          return v.stage === "error" ? v.message || stageLabel.error! : stageLabel[v.stage] ?? v.message;
        })),
      ),
      h(
        "ol",
        { class: "login-steps" },
        h("li", { class: () => (stage() === "loading" ? "active" : "done") }, h("b", null, "1"), "打开哔哩哔哩 App"),
        h("li", { class: () => (stage() === "waiting" ? "active" : stage() === "scanned" || stage() === "done" ? "done" : "") }, h("b", null, "2"), "扫描二维码"),
        h("li", { class: () => (stage() === "scanned" ? "active" : stage() === "done" ? "done" : "") }, h("b", null, "3"), "在手机上确认登录"),
      ),
    ),
    h(
      "footer",
      { class: "login-foot muted" },
      text(() => `v${info()?.version ?? ""}`),
      info()?.demo ? h("span", { class: "badge badge-blue" }, "演示模式") : null,
    ),
  );
}

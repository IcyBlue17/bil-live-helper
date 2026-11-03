// The 直播控制 and 推流信息 cards, and the face-auth dialog.
import { h, html, listen, text } from "../lib/dom";
import { effect, signal } from "../lib/signal";
import { duration, maskKey } from "../lib/format";
import { cancelStart, copy, face, isLive, liveAction, now, room, startLive, stopLive, stream } from "../state";
import { icons } from "../ui/icons";
import { openModal } from "../ui/modal";
import { copyButton, flashCopied, spinner } from "../ui/widgets";

/** Set while the stop confirmation is showing. */
export const confirmingStop = signal(false);

/** What Space does: start, ask to stop, or confirm the stop. */
export function toggleLive() {
  if (liveAction.peek() !== "idle") return;
  if (!isLive()) startLive();
  else if (!confirmingStop.peek()) confirmingStop.set(true);
  else {
    confirmingStop.set(false);
    void stopLive();
  }
}

export function LiveCard() {
  // Drop a pending confirmation when the room stops on its own.
  effect(() => {
    if (!isLive()) confirmingStop.set(false);
  });
  let confirmTimer: ReturnType<typeof setTimeout> | undefined;
  effect(() => {
    clearTimeout(confirmTimer);
    if (confirmingStop()) confirmTimer = setTimeout(() => confirmingStop.set(false), 6000);
  });

  const label = () => {
    const a = liveAction();
    if (a === "starting") return "正在开播…";
    if (a === "faceAuth") return "等待人脸认证…";
    if (a === "stopping") return "正在下播…";
    return isLive() ? "结束直播" : "开始直播";
  };

  const mainBtn = h(
    "button",
    {
      class: () => `live-btn ${isLive() ? "is-live" : ""} ${liveAction() !== "idle" ? "busy" : ""}`,
      disabled: () => liveAction() !== "idle" || !room(),
      onclick: toggleLive,
      title: "快捷键：Space",
    },
    h("span", { class: "live-btn-icon" }, spinner(), html(icons.play), html(icons.stop)),
    h("span", { class: "live-btn-label" }, text(label)),
    h("kbd", null, "Space"),
  );

  return h(
    "section",
    { class: () => `card hero ${isLive() ? "hero-live" : ""}` },
    h(
      "div",
      { class: "hero-main" },
      h(
        "div",
        { class: "hero-text" },
        h("span", { class: "eyebrow" }, "直播控制"),
        h("h2", { class: "hero-status" }, text(() => {
          const r = room();
          if (r?.status === "live") return "正在直播";
          if (r?.status === "round") return "轮播中";
          return "准备好了吗？";
        })),
        h("p", { class: "hero-sub muted" }, text(() => {
          const r = room();
          if (!r) return "正在加载直播间…";
          if (r.status === "live") return `已直播 ${duration(r.liveSince, now())} · 记得在 OBS 中开始推流`;
          return "开播后会生成推流地址，粘贴到 OBS 即可推流。";
        })),
      ),
      h(
        "div",
        { class: () => `hero-actions ${confirmingStop() ? "confirming" : ""}` },
        mainBtn,
        h(
          "div",
          { class: "confirm-stop", role: "alertdialog", "aria-label": "确认结束直播" },
          h("span", null, "确定结束直播？"),
          h("button", { class: "btn btn-ghost btn-sm", onclick: () => confirmingStop.set(false) }, "取消"),
          h("button", { class: "btn btn-danger btn-sm", onclick: () => { confirmingStop.set(false); void stopLive(); } }, "结束直播"),
        ),
      ),
    ),
    h(
      "dl",
      { class: "hero-summary" },
      summaryItem("标题", () => room()?.title || "未设置", "title-ellipsis"),
      summaryItem("分区", () => {
        const r = room();
        return r && r.areaName ? `${r.parentAreaName} / ${r.areaName}` : "未设置";
      }),
    ),
  );
}

function summaryItem(label: string, value: () => string, cls = "") {
  return h("div", null, h("dt", null, label), h("dd", { class: cls, title: value }, text(value)));
}

export function StreamCard() {
  const reveal = signal(false);
  const addr = () => stream()?.addr ?? "";
  const code = () => stream()?.code ?? "";
  const copyAllBtn = copyButton(() => (stream() ? `服务器：${addr()}\n推流码：${code()}` : ""), "推流信息", { text: "复制全部", class: "copy-all" });

  const body = h("div", { class: "stream-body" });
  effect(() => {
    if (!stream()) {
      body.replaceChildren(
        h(
          "div",
          { class: "empty" },
          h("span", { class: "empty-icon" }, html(icons.server)),
          h("p", null, "还没有推流信息"),
          h("p", { class: "muted" }, "点击「开始直播」后，这里会显示服务器地址和推流码，复制到 OBS 的「推流」设置中即可。"),
        ),
      );
      return;
    }
    body.replaceChildren(
      h(
        "div",
        { class: () => `banner banner-warn ${stream()?.changed ? "" : "hidden"}` },
        html(icons.alert),
        h("span", null, "推流码已变更，请更新 OBS 中的推流码。"),
      ),
      field("服务器地址", icons.server, () => addr(), copyButton(addr, "服务器地址", { class: "copy-addr" })),
      field(
        "推流码",
        icons.key,
        () => (reveal() ? code() : maskKey(code())),
        h(
          "button",
          {
            class: "icon-btn",
            title: () => (reveal() ? "隐藏推流码" : "显示推流码"),
            "aria-label": () => (reveal() ? "隐藏推流码" : "显示推流码"),
            "aria-pressed": () => String(reveal()),
            onclick: () => reveal.update((v) => !v),
          },
          h("span", { class: () => `swap-icons ${reveal() ? "on" : ""}` }, html(icons.eye), html(icons.eyeOff)),
        ),
        copyButton(code, "推流码", { class: "copy-code" }),
      ),
      h("p", { class: "hint muted" }, "推流码仅本次开播有效，请勿泄露给他人。"),
    );
  });

  return h(
    "section",
    { class: "card stream-card" },
    h(
      "header",
      { class: "card-head" },
      h("h3", null, html(icons.server), "推流信息"),
      h("div", { class: () => `card-tools ${stream() ? "" : "hidden"}` }, copyAllBtn),
    ),
    body,
  );
}

function field(label: string, icon: string, value: () => string, ...tools: Node[]) {
  return h(
    "div",
    { class: "field" },
    h("label", null, html(icon), label),
    h("div", { class: "field-box mono" }, h("span", { class: "field-value", title: value }, text(value)), h("div", { class: "field-tools" }, tools)),
  );
}

/** Opens the face-auth dialog while a start waits for it. */
export function FaceAuthWatcher() {
  let close: (() => void) | null = null;
  effect(() => {
    const open = face() !== null;
    if (open && !close) {
      close = openModal(
        { title: "需要人脸认证", subtitle: "B 站要求本次开播前完成实名人脸认证", width: 440, onDismiss: cancelStart },
        () => FaceAuthBody(),
      );
    } else if (!open && close) {
      close();
      close = null;
    }
  });
  listen(window, "beforeunload", () => close?.());
}

function FaceAuthBody() {
  const qr = h("div", { class: "qr-code" });
  effect(() => {
    const svg = face()?.svg;
    qr.replaceChildren(svg ? html(svg) : spinner());
  });
  const stage = () => face()?.stage ?? "waiting";
  return h(
    "div",
    { class: "modal-body face" },
    h(
      "div",
      { class: () => `qr-tile qr-tile-sm stage-${stage()}` },
      qr,
      h("div", { class: "qr-scanned" }, html(icons.phone), h("span", null, "请在手机上完成认证")),
      h("div", { class: "qr-expired" }, h("p", null, "二维码已过期"), h("p", { class: "muted" }, "请取消后重新开播")),
    ),
    h(
      "ol",
      { class: "face-steps" },
      h("li", null, "使用", h("b", null, "哔哩哔哩 App"), "扫描上方二维码"),
      h("li", null, "按提示完成人脸识别"),
      h("li", null, "认证通过后将自动开播，无需其他操作"),
    ),
    h("div", { class: () => `login-stage stage-${stage()}` }, h("span", { class: "stage-dot" }), h("span", null, text(() => face()?.message || "等待扫码…"))),
    h("div", { class: "modal-actions" }, h("button", { class: "btn btn-ghost", onclick: cancelStart }, "取消开播")),
  );
}

/** Copies the stream key (C) or address + key (Shift+C) from the keyboard. */
export async function copyStream(all: boolean) {
  const s = stream.peek();
  if (!s) return;
  const ok = all ? await copy(`服务器：${s.addr}\n推流码：${s.code}`, "推流信息") : await copy(s.code, "推流码");
  if (ok) flashCopied(document.querySelector<HTMLElement>(all ? ".copy-all" : ".copy-code"));
}

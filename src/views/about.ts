import { h, html } from "../lib/dom";
import { info, openURL } from "../state";
import { icons, mark } from "../ui/icons";
import { openModal } from "../ui/modal";
import { kbd } from "../ui/widgets";

export const REPO_URL = "https://github.com/IcyBlue17/bil-live-helper";
export const UPSTREAM_URL = "https://github.com/chenxi-Eumenides/bilibili_live_tool";

export const shortcuts: [string[], string][] = [
  [["Space"], "开始 / 结束直播"],
  [["T"], "编辑直播标题"],
  [["A"], "选择分区"],
  [["R"], "刷新直播间信息"],
  [["C"], "复制推流码"],
  [["Shift", "C"], "复制服务器地址与推流码"],
  [["L"], "展开 / 收起活动日志"],
  [["?"], "关于与快捷键"],
  [["Esc"], "关闭弹窗 / 取消"],
];

let close: (() => void) | null = null;

export function openAbout() {
  if (close) return;
  const i = info.peek();
  close = openModal({ title: "关于与快捷键", width: 560, class: "about-modal", onDismiss: () => close?.() }, () =>
    h(
      "div",
      { class: "modal-body about" },
      h(
        "div",
        { class: "about-brand" },
        html(mark),
        h(
          "div",
          null,
          h("h3", null, "Bil Live Helper", h("span", { class: "badge" }, `v${i?.version ?? "?"}`), i?.demo ? h("span", { class: "badge badge-blue" }, "演示模式") : null),
          h("p", { class: "muted" }, "绕过直播姬，直接获取推流码的 B 站直播助手。基于 ", linkTo(UPSTREAM_URL, "bilibili_live_tool"), " 重写。"),
        ),
      ),
      h("h4", null, "快捷键"),
      h(
        "div",
        { class: "shortcut-grid" },
        shortcuts.map(([keys, label]) => h("div", { class: "shortcut" }, h("span", null, label), kbd(...keys))),
      ),
      h("p", { class: "hint muted" }, "在输入框中输入时，除 Esc 外的快捷键不会触发。"),
      h("h4", null, "配置文件"),
      h("code", { class: "config-path" }, i?.configPath || "—"),
      h("div", { class: "disclaimer" }, html(icons.info), h("p", null, "本程序仅用于声明的功能，用户使用本程序产生的账号问题，与本程序无关。请勿用于不当用途。")),
      h(
        "div",
        { class: "modal-actions" },
        h("button", { class: "btn btn-ghost", onclick: () => openURL(REPO_URL) }, html(icons.external), "GitHub 仓库"),
        h("button", { class: "btn btn-secondary", autofocus: true, onclick: () => close?.() }, "好的"),
      ),
    ),
  );
  const c = close;
  close = () => {
    c();
    close = null;
  };
}

function linkTo(url: string, label: string) {
  return h("a", { href: url, onclick: (e: Event) => { e.preventDefault(); openURL(url); } }, label);
}

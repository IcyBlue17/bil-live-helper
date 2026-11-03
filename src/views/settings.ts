// The 直播间设置 card: title and area, saved together.
import { h, html, text } from "../lib/dom";
import { computed, effect, signal, untrack } from "../lib/signal";
import { charCount } from "../lib/format";
import { areas, room, saveSettings, settings } from "../state";
import type { Area, AreaGroup } from "../mygo";
import { icons } from "../ui/icons";
import { openModal } from "../ui/modal";
import { spinner } from "../ui/widgets";

export const TITLE_MAX = 40;

let titleInput: HTMLInputElement | null = null;
let openPicker: (() => void) | null = null;

export function focusTitle() {
  titleInput?.focus();
  titleInput?.select();
}
export function openAreaPicker() {
  openPicker?.();
}

function findArea(groups: AreaGroup[], id: number): { group: AreaGroup; area: Area } | null {
  for (const group of groups) for (const area of group.areas) if (area.id === id) return { group, area };
  return null;
}

export function SettingsCard() {
  const title = signal("");
  const areaId = signal(0);
  const saving = signal(false);
  const touched = signal(false);

  // The saved values, from settings (preferred) or the room.
  const savedTitle = () => settings()?.title ?? room()?.title ?? "";
  const savedArea = () => settings()?.areaId || room()?.areaId || 0;
  // Reset the form whenever the saved values change.
  effect(() => {
    const t = savedTitle();
    const a = savedArea();
    untrack(() => {
      title.set(t);
      areaId.set(a);
      touched.set(false);
    });
  });

  const count = computed(() => charCount(title()));
  const titleError = computed(() => {
    if (!touched()) return "";
    if (title().trim() === "") return "标题不能为空";
    if (count() > TITLE_MAX) return `标题最多 ${TITLE_MAX} 个字`;
    return "";
  });
  const dirty = computed(() => title() !== savedTitle() || areaId() !== savedArea());
  const valid = computed(() => title().trim() !== "" && count() <= TITLE_MAX && areaId() > 0);
  const area = computed(() => findArea(areas(), areaId()));

  const save = async () => {
    touched.set(true);
    if (!valid() || !dirty() || saving()) return;
    saving.set(true);
    await saveSettings({ title: title().trim(), areaId: areaId() });
    saving.set(false);
  };
  const reset = () => {
    title.set(savedTitle());
    areaId.set(savedArea());
    touched.set(false);
  };

  const input = h("input", {
    id: "title-input",
    class: () => `input ${titleError() ? "invalid" : ""}`,
    type: "text",
    value: () => title(),
    placeholder: "给直播起个标题",
    autocomplete: "off",
    spellcheck: "false",
    "aria-invalid": () => String(!!titleError()),
    "aria-describedby": "title-help",
    oninput: (e: Event) => {
      title.set((e.target as HTMLInputElement).value);
      touched.set(true);
    },
    onkeydown: (e: KeyboardEvent) => {
      if (e.key === "Enter" && !e.isComposing) void save();
    },
  });
  titleInput = input;

  const pick = () =>
    AreaPicker(areas(), areaId(), (id) => {
      areaId.set(id);
      touched.set(true);
    });
  openPicker = pick;

  const areaBtn = h(
    "button",
    { class: "input area-btn", type: "button", onclick: pick, title: "选择分区 (A)" },
    h("span", { class: "area-value" }),
    html(icons.chevronRight),
  );
  const areaValue = areaBtn.querySelector(".area-value")!;
  effect(() => {
    const a = area();
    areaValue.replaceChildren(
      a ? h("span", null, h("span", { class: "muted" }, a.group.name), h("span", { class: "crumb" }, "/"), a.area.name) : h("span", { class: "muted" }, areaId() ? `分区 #${areaId()}` : "选择分区"),
    );
  });

  return h(
    "section",
    { class: "card settings-card" },
    h(
      "header",
      { class: "card-head" },
      h("h3", null, html(icons.edit), "直播间设置"),
      h("span", { class: () => `dirty-badge ${dirty() ? "show" : ""}` }, h("span", { class: "dirty-dot" }), "未保存"),
    ),
    h(
      "div",
      { class: "form" },
      h(
        "div",
        { class: "form-row" },
        h("label", { for: "title-input" }, "直播标题", h("kbd", null, "T")),
        input,
        h(
          "div",
          { class: "form-help", id: "title-help" },
          h("span", { class: "error-text" }, text(() => titleError())),
          h("span", { class: () => `counter num ${count() > TITLE_MAX ? "over" : count() > TITLE_MAX - 5 ? "near" : ""}` }, text(() => `${count()}/${TITLE_MAX}`)),
        ),
      ),
      h("div", { class: "form-row" }, h("label", null, "直播分区", h("kbd", null, "A")), areaBtn),
    ),
    h(
      "footer",
      { class: "card-foot" },
      h("button", { class: () => `btn btn-ghost ${dirty() ? "" : "invisible"}`, onclick: reset, disabled: () => !dirty() || saving() }, "撤销修改"),
      h(
        "button",
        { class: "btn btn-secondary", onclick: () => void save(), disabled: () => !dirty() || !valid() || saving() },
        h("span", { class: () => (saving() ? "" : "hidden") }, spinner()),
        text(() => (saving() ? "保存中…" : "保存设置")),
      ),
    ),
  );
}

/** The two-level area picker: parent list + child grid, searchable across all. */
function AreaPicker(groups: AreaGroup[], current: number, onPick: (id: number) => void) {
  const initial = findArea(groups, current)?.group.id ?? groups[0]?.id ?? 0;
  const query = signal("");
  const groupId = signal(initial);
  const close = openModal({ title: "选择分区", subtitle: "选择直播内容所属的分区", width: 680, class: "picker-modal" }, (close) => {
    const search = h("input", {
      class: "input search-input",
      type: "search",
      placeholder: "搜索分区，如「单机」「唱见」",
      autofocus: true,
      value: () => query(),
      oninput: (e: Event) => query.set((e.target as HTMLInputElement).value),
      onkeydown: (e: KeyboardEvent) => {
        if (e.key === "Enter" && !e.isComposing) {
          const first = grid.querySelector<HTMLButtonElement>(".area-chip");
          first?.click();
        }
      },
    });
    const choose = (id: number) => {
      onPick(id);
      close();
    };
    const grid = h("div", { class: "area-grid" });
    const groupList = h("nav", { class: "group-list" });

    effect(() => {
      const q = query().trim().toLowerCase();
      groupList.classList.toggle("dimmed", q !== "");
      groupList.replaceChildren(
        ...groups.map((g) =>
          h(
            "button",
            {
              class: () => `group-item ${!query().trim() && groupId() === g.id ? "active" : ""}`,
              onclick: () => {
                query.set("");
                groupId.set(g.id);
              },
            },
            h("span", null, g.name),
            h("span", { class: "muted num" }, String(g.areas.length)),
          ),
        ),
      );
    });
    effect(() => {
      const q = query().trim().toLowerCase();
      let items: { group: AreaGroup; area: Area }[];
      if (q) {
        items = groups.flatMap((group) =>
          group.areas.filter((a) => a.name.toLowerCase().includes(q) || group.name.toLowerCase().includes(q)).map((area) => ({ group, area })),
        );
      } else {
        const g = groups.find((x) => x.id === groupId());
        items = g ? g.areas.map((area) => ({ group: g, area })) : [];
      }
      grid.replaceChildren(
        ...(items.length
          ? items.map(({ group, area }) =>
              h(
                "button",
                { class: `area-chip ${area.id === current ? "selected" : ""}`, onclick: () => choose(area.id), title: `${group.name} / ${area.name}` },
                h("span", { class: "chip-name" }, area.name),
                q ? h("span", { class: "chip-group muted" }, group.name) : null,
                area.id === current ? html(icons.check) : null,
              ),
            )
          : [h("div", { class: "empty empty-sm" }, h("p", null, groups.length ? "没有匹配的分区" : "分区列表加载中…"), q ? h("p", { class: "muted" }, "换个关键词试试") : null)]),
      );
      grid.classList.remove("fade-in");
      void grid.offsetWidth;
      grid.classList.add("fade-in");
    });
    return h("div", { class: "modal-body picker" }, h("div", { class: "search-wrap" }, html(icons.search), search), h("div", { class: "picker-cols" }, groupList, grid));
  });
  return close;
}

<script lang="ts">
  // Preferences: appearance (local only), desktop notifications and blocked words.
  import { Monitor, Moon, Sun, X } from "@lucide/svelte";
  import { Switch } from "bits-ui";
  import { toast } from "svelte-sonner";
  import Modal from "../components/Modal.svelte";
  import { System } from "../mygo";
  import { errorMessage } from "../lib/format";
  import { fail, log, ui } from "../lib/state.svelte";
  import { setTheme, theme, type ThemePref } from "../lib/theme.svelte";

  const themes: { value: ThemePref; label: string; icon: typeof Sun }[] = [
    { value: "system", label: "跟随系统", icon: Monitor },
    { value: "dark", label: "暗色", icon: Moon },
    { value: "light", label: "亮色", icon: Sun },
  ];

  let loading = $state(true);
  let loadError = $state("");
  let saving = $state(false);
  let notify = $state(false);
  /** Gift threshold in yuan as typed; stored ×1000 as gold coins. */
  let giftYuan = $state("");
  let blocked = $state<string[]>([]);
  let draft = $state("");

  $effect(() => {
    if (!ui.prefsOpen) return;
    loading = true;
    loadError = "";
    draft = "";
    System.preferences()
      .then((p) => {
        notify = p.notify;
        giftYuan = p.notifyGiftMin ? String(p.notifyGiftMin / 1000) : "";
        blocked = [...p.blocked];
      })
      .catch((err: unknown) => (loadError = errorMessage(err)))
      .finally(() => (loading = false));
  });

  const giftValue = $derived(giftYuan.trim() === "" ? 0 : Number(giftYuan));
  const giftError = $derived(Number.isFinite(giftValue) && giftValue >= 0 ? "" : "请输入不小于 0 的金额");

  function addWord() {
    const words = draft.split(/[,，\s]+/).map((w) => w.trim()).filter(Boolean);
    const seen = new Set(blocked.map((w) => w.toLowerCase()));
    for (const w of words) if (!seen.has(w.toLowerCase())) {
      seen.add(w.toLowerCase());
      blocked.push(w);
    }
    draft = "";
  }
  function onWordKey(e: KeyboardEvent) {
    if (e.isComposing) return;
    if (e.key === "Enter" || e.key === "," || e.key === "，") {
      e.preventDefault();
      addWord();
    } else if (e.key === "Backspace" && draft === "" && blocked.length) {
      blocked.pop();
    }
  }

  async function save() {
    if (draft.trim()) addWord();
    if (giftError || saving) return;
    saving = true;
    try {
      await System.savePreferences({ notify, notifyGiftMin: Math.round(giftValue * 1000), blocked: [...blocked] });
      toast.success("偏好设置已保存");
      log("ok", "偏好设置已保存");
      ui.prefsOpen = false;
    } catch (err) {
      fail("保存偏好设置失败", err);
    } finally {
      saving = false;
    }
  }
</script>

<Modal bind:open={ui.prefsOpen} title="偏好设置" subtitle="外观立即生效；提醒与过滤在保存后生效。" width={540}>
  <div class="modal-body prefs">
    <div class="pref-group">
      <h4>外观</h4>
      <div class="pref-list">
        <div class="pref-row">
          <div class="pref-text">
            <b>主题</b>
            <span>只保存在本机，默认跟随系统</span>
          </div>
          <div class="seg theme-seg" role="radiogroup" aria-label="主题">
            {#each themes as t (t.value)}
              <button class="seg-item" class:active={theme.pref === t.value} role="radio" aria-checked={theme.pref === t.value} onclick={() => setTheme(t.value)}><t.icon class="icon" />{t.label}</button>
            {/each}
          </div>
        </div>
      </div>
    </div>
    {#if loading}
      <div class="prefs-loading"><span class="skel" style="height:56px"></span><span class="skel" style="height:56px"></span><span class="skel" style="height:96px"></span></div>
    {:else if loadError}
      <div class="empty"><p>读取偏好设置失败</p><p>{loadError}</p></div>
    {:else}
      <div class="pref-group">
        <h4>通知</h4>
        <div class="pref-list">
          <div class="pref-row">
            <label class="pref-text" for="pref-notify">
              <b>桌面通知</b>
              <span>收到醒目留言（SC）和上舰时弹出系统通知</span>
            </label>
            <Switch.Root id="pref-notify" class="switch" bind:checked={notify}><Switch.Thumb class="switch-thumb" /></Switch.Root>
          </div>
          <div class="pref-row" class:disabled={!notify}>
            <label class="pref-text" for="pref-gift">
              <b>礼物通知</b>
              <span class:error-text={!!giftError}>{giftError || "单次礼物达到该金额时也通知，留空或 0 为关闭"}</span>
            </label>
            <div class="input-affix" class:invalid={!!giftError}>
              <span>¥</span>
              <input id="pref-gift" class="num" type="text" inputmode="decimal" placeholder="0" disabled={!notify} bind:value={giftYuan} />
            </div>
          </div>
        </div>
      </div>
      <div class="pref-group">
        <h4>弹幕过滤</h4>
        <div class="pref-list">
          <div class="pref-row bare">
            <label class="pref-text" for="pref-word">
              <b>屏蔽词</b>
              <span>包含这些词的弹幕不显示，不区分大小写</span>
            </label>
          </div>
          <div class="pref-tags">
            <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
            <div class="tag-input" onclick={(e) => (e.currentTarget.querySelector("input") as HTMLInputElement).focus()}>
              {#each blocked as w, i (w)}
                <span class="tag-chip">{w}<button aria-label="移除 {w}" onclick={(e) => { e.stopPropagation(); blocked.splice(i, 1); }}><X class="icon" /></button></span>
              {/each}
              <input id="pref-word" type="text" autocomplete="off" placeholder={blocked.length ? "" : "输入后按 Enter 添加"} bind:value={draft} onkeydown={onWordKey} onblur={() => draft.trim() && addWord()} />
            </div>
            <p class="hint"><span class="num">{blocked.length}</span> 个屏蔽词 · 用 Enter 或逗号分隔，退格删除最后一个</p>
          </div>
        </div>
      </div>
    {/if}
  </div>
  <footer class="modal-actions">
    <button class="btn btn-secondary" onclick={() => (ui.prefsOpen = false)}>取消</button>
    <button class="btn btn-primary" disabled={loading || !!loadError || !!giftError || saving} onclick={() => void save()}>
      {#if saving}<span class="spinner"></span>保存中…{:else}保存{/if}
    </button>
  </footer>
</Modal>

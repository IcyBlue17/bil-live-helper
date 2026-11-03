<script lang="ts">
  // Preferences: desktop notifications and blocked words.
  import { Bell, Gift, ShieldOff, X } from "@lucide/svelte";
  import { Switch } from "bits-ui";
  import { toast } from "svelte-sonner";
  import Modal from "../components/Modal.svelte";
  import { System } from "../mygo";
  import { errorMessage } from "../lib/format";
  import { fail, log, ui } from "../lib/state.svelte";

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

<Modal bind:open={ui.prefsOpen} title="偏好设置" subtitle="提醒与弹幕过滤，保存后立即生效" width={500}>
  <div class="modal-body prefs">
    {#if loading}
      <div class="prefs-loading"><span class="skel" style="height:56px"></span><span class="skel" style="height:56px"></span><span class="skel" style="height:96px"></span></div>
    {:else if loadError}
      <div class="empty"><p>读取偏好设置失败</p><p class="muted">{loadError}</p></div>
    {:else}
      <div class="pref-row">
        <span class="pref-icon"><Bell class="icon" /></span>
        <label class="pref-text" for="pref-notify">
          <b>桌面通知</b>
          <span class="muted">收到醒目留言（SC）和上舰时弹出系统通知</span>
        </label>
        <Switch.Root id="pref-notify" class="switch" bind:checked={notify}><Switch.Thumb class="switch-thumb" /></Switch.Root>
      </div>
      <div class="pref-row" class:disabled={!notify}>
        <span class="pref-icon"><Gift class="icon" /></span>
        <label class="pref-text" for="pref-gift">
          <b>礼物通知</b>
          <span class="muted">{giftError || "单次礼物价值达到该金额时也通知，留空或 0 为关闭"}</span>
        </label>
        <div class="input-affix" class:invalid={!!giftError}>
          <span>¥</span>
          <input id="pref-gift" class="num" type="text" inputmode="decimal" placeholder="0" disabled={!notify} bind:value={giftYuan} />
        </div>
      </div>
      <div class="pref-block">
        <div class="pref-row bare">
          <span class="pref-icon"><ShieldOff class="icon" /></span>
          <label class="pref-text" for="pref-word">
            <b>屏蔽词</b>
            <span class="muted">包含这些词的弹幕不显示，不区分大小写</span>
          </label>
        </div>
        <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
        <div class="tag-input" onclick={(e) => (e.currentTarget.querySelector("input") as HTMLInputElement).focus()}>
          {#each blocked as w, i (w)}
            <span class="tag-chip">{w}<button aria-label="移除 {w}" onclick={(e) => { e.stopPropagation(); blocked.splice(i, 1); }}><X class="icon" /></button></span>
          {/each}
          <input id="pref-word" type="text" autocomplete="off" placeholder={blocked.length ? "" : "输入后按 Enter 添加"} bind:value={draft} onkeydown={onWordKey} onblur={() => draft.trim() && addWord()} />
        </div>
        <p class="hint muted">{blocked.length} 个屏蔽词 · 用 Enter 或逗号分隔，退格删除最后一个</p>
      </div>
    {/if}
  </div>
  <footer class="modal-actions">
    <button class="btn btn-ghost" onclick={() => (ui.prefsOpen = false)}>取消</button>
    <button class="btn btn-primary" disabled={loading || !!loadError || !!giftError || saving} onclick={() => void save()}>
      {#if saving}<span class="spinner"></span>保存中…{:else}保存{/if}
    </button>
  </footer>
</Modal>

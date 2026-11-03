<script lang="ts">
  import { ChevronRight, PencilLine } from "@lucide/svelte";
  import { tick } from "svelte";
  import { findArea } from "../lib/areaSearch";
  import { charCount } from "../lib/format";
  import { app, saveSettings, ui } from "../lib/state.svelte";
  import AreaPicker from "./AreaPicker.svelte";

  const TITLE_MAX = 40;
  const savedTitle = $derived(app.settings?.title ?? app.room?.title ?? "");
  const savedArea = $derived(app.settings?.areaId || app.room?.areaId || 0);

  let title = $state("");
  let areaId = $state(0);
  let touched = $state(false);
  let saving = $state(false);
  let input: HTMLInputElement | undefined = $state();

  // Reset the form whenever the saved values change.
  $effect.pre(() => {
    title = savedTitle;
    areaId = savedArea;
    touched = false;
  });
  // T focuses the title.
  $effect(() => {
    if (!ui.focusTitle) return;
    void tick().then(() => {
      input?.focus();
      input?.select();
    });
  });

  const count = $derived(charCount(title));
  const error = $derived(!touched ? "" : title.trim() === "" ? "标题不能为空" : count > TITLE_MAX ? `标题最多 ${TITLE_MAX} 个字` : "");
  const dirty = $derived(title !== savedTitle || areaId !== savedArea);
  const valid = $derived(title.trim() !== "" && count <= TITLE_MAX && areaId > 0);
  const area = $derived(findArea(app.areas, areaId));

  async function save() {
    touched = true;
    if (!valid || !dirty || saving) return;
    saving = true;
    await saveSettings({ title: title.trim(), areaId });
    saving = false;
  }
  function reset() {
    title = savedTitle;
    areaId = savedArea;
    touched = false;
  }
</script>

<section class="card settings-card">
  <header class="card-head">
    <h3><PencilLine class="icon" />直播间设置</h3>
    <span class="dirty-badge" class:show={dirty}><span class="dirty-dot"></span>未保存</span>
  </header>
  <div class="form">
    <div class="form-row">
      <label for="title-input">直播标题<kbd>T</kbd></label>
      <input
        id="title-input"
        bind:this={input}
        class="input"
        class:invalid={!!error}
        type="text"
        placeholder="给直播起个标题"
        autocomplete="off"
        spellcheck="false"
        aria-invalid={!!error}
        aria-describedby="title-help"
        bind:value={title}
        oninput={() => (touched = true)}
        onkeydown={(e) => e.key === "Enter" && !e.isComposing && void save()}
      />
      <div class="form-help" id="title-help">
        <span class="error-text">{error}</span>
        <span class="counter num" class:near={count > TITLE_MAX - 5 && count <= TITLE_MAX} class:over={count > TITLE_MAX}>{count}/{TITLE_MAX}</span>
      </div>
    </div>
    <div class="form-row">
      <span class="label">直播分区<kbd>A</kbd></span>
      <button class="input area-btn" type="button" onclick={() => (ui.areaOpen = true)} title="选择分区 (A)">
        <span class="area-value">
          {#if area}<span class="muted">{area.group.name}</span><span class="crumb">/</span>{area.area.name}
          {:else}<span class="muted">{areaId ? `分区 #${areaId}` : "选择分区"}</span>{/if}
        </span>
        <ChevronRight class="icon" />
      </button>
    </div>
  </div>
  <footer class="card-foot">
    <button class="btn btn-ghost" class:invisible={!dirty} onclick={reset} disabled={!dirty || saving}>撤销修改</button>
    <button class="btn btn-secondary" onclick={() => void save()} disabled={!dirty || !valid || saving}>
      {#if saving}<span class="spinner"></span>保存中…{:else}保存设置{/if}
    </button>
  </footer>
</section>

<AreaPicker bind:open={ui.areaOpen} current={areaId} onPick={(id) => {
  areaId = id;
  touched = true;
}} />

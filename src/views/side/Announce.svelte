<script lang="ts">
  import { toast } from "svelte-sonner";
  import { Live } from "../../mygo";
  import { charCount, errorMessage } from "../../lib/format";
  import { fail, log } from "../../lib/state.svelte";

  const MAX = 60;
  let saved = $state<string | null>(null);
  let text = $state("");
  let loadError = $state("");
  let saving = $state(false);

  async function load() {
    loadError = "";
    try {
      saved = await Live.announcement();
      text = saved;
    } catch (err) {
      loadError = errorMessage(err);
    }
  }
  void load();

  const count = $derived(charCount(text));
  const dirty = $derived(saved !== null && text !== saved);
  async function save() {
    if (!dirty || count > MAX || saving) return;
    saving = true;
    try {
      await Live.setAnnouncement(text.trim());
      saved = text = text.trim();
      toast.success("公告已更新");
      log("ok", "直播间公告已更新");
    } catch (err) {
      fail("更新公告失败", err);
    } finally {
      saving = false;
    }
  }
</script>

<section class="card">
  {#if loadError}
    <div class="empty"><p>公告加载失败</p><p>{loadError}</p><button class="btn btn-secondary btn-sm" onclick={() => void load()}>重试</button></div>
  {:else}
    <div class="section-body announce">
      <label class="section-title" for="announce-input">直播间公告</label>
      <p class="section-desc">显示在直播间简介区，所有观众可见。</p>
      <textarea
        id="announce-input"
        class="input textarea"
        class:invalid={count > MAX}
        rows="4"
        placeholder={saved === null ? "加载中…" : "写点什么告诉观众吧"}
        disabled={saved === null}
        bind:value={text}
        onkeydown={(e) => (e.ctrlKey || e.metaKey) && e.key === "Enter" && void save()}
      ></textarea>
      <div class="form-help">
        <span class="error-text">{count > MAX ? `公告最多 ${MAX} 个字` : ""}</span>
        <span class="counter num" class:near={count > MAX - 6 && count <= MAX} class:over={count > MAX}>{count}/{MAX}</span>
      </div>
    </div>
    <footer class="section-foot">
      <span>最多 {MAX} 个字，<kbd>Ctrl</kbd> <kbd>Enter</kbd> 保存。</span>
      <div class="actions">
        <button class="btn btn-ghost btn-sm" class:invisible={!dirty} disabled={!dirty || saving} onclick={() => (text = saved ?? "")}>撤销</button>
        <button class="btn btn-primary btn-sm" disabled={!dirty || count > MAX || saving} onclick={() => void save()}>
          {#if saving}<span class="spinner"></span>保存中…{:else}保存{/if}
        </button>
      </div>
    </footer>
  {/if}
</section>

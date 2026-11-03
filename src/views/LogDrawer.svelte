<script lang="ts">
  import { ChevronDown, ScrollText } from "@lucide/svelte";
  import { tick } from "svelte";
  import Kbd from "../components/Kbd.svelte";
  import { clock } from "../lib/format";
  import { logs } from "../lib/state.svelte";

  const LEVEL = { info: "信息", ok: "成功", warn: "警告", error: "错误" } as const;
  const last = $derived(logs.entries.at(-1));
  let list: HTMLDivElement | undefined = $state();

  $effect(() => {
    if (logs.open) logs.unread = 0;
  });
  // Keep the newest line in view while open.
  $effect(() => {
    void logs.entries.length;
    if (!logs.open) return;
    void tick().then(() => list && (list.scrollTop = list.scrollHeight));
  });
</script>

<section class="log-drawer" class:open={logs.open}>
  <button class="log-bar" aria-expanded={logs.open} aria-controls="log-body" onclick={() => (logs.open = !logs.open)}>
    <span class="log-title"><ScrollText class="icon" />活动日志</span>
    <span class="log-unread num" class:show={logs.unread > 0}>{logs.unread}</span>
    <span class="log-last log-{last?.level ?? 'info'}">{logs.open ? "" : (last?.message ?? "")}</span>
    <Kbd keys={["L"]} />
    <span class="log-chevron"><ChevronDown class="icon" /></span>
  </button>
  <div class="log-body" id="log-body">
    <div class="log-list" bind:this={list}>
      {#each logs.entries as e, i (i)}
        <div class="log-line log-{e.level}">
          <time class="num">{clock(e.time)}</time>
          <span class="log-level">{LEVEL[e.level]}</span>
          <span class="log-msg">{e.message}</span>
        </div>
      {:else}
        <div class="log-line"><span class="log-msg">暂无日志</span></div>
      {/each}
    </div>
  </div>
</section>

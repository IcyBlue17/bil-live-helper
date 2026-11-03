<script lang="ts">
  import { Play, Square } from "@lucide/svelte";
  import { elapsed } from "../lib/format";
  import { app, live, stopLive, toggleLive } from "../lib/state.svelte";

  // Drop a pending confirmation after a while, or when the room stops on its own.
  $effect(() => {
    if (!live.confirmingStop) return;
    if (!app.isLive) {
      live.confirmingStop = false;
      return;
    }
    const t = setTimeout(() => (live.confirmingStop = false), 6000);
    return () => clearTimeout(t);
  });

  const label = $derived.by(() => {
    if (live.action === "starting") return "正在开播…";
    if (live.action === "faceAuth") return "等待人脸认证…";
    if (live.action === "stopping") return "正在下播…";
    return app.isLive ? "结束直播" : "开始直播";
  });
  const status = $derived(app.room?.status);
</script>

<section class="card hero" class:hero-live={app.isLive}>
  <div class="hero-main">
    <div class="hero-text">
      <span class="eyebrow">直播控制</span>
      <h2 class="hero-status">{status === "live" ? "正在直播" : status === "round" ? "轮播中" : "准备好了吗？"}</h2>
      <p class="hero-sub muted">
        {#if !app.room}正在加载直播间…
        {:else if status === "live"}已直播 <span class="num">{elapsed(app.room.liveSince, app.now)}</span> · 记得在 OBS 中开始推流
        {:else}开播后会生成推流地址，粘贴到 OBS 即可推流。{/if}
      </p>
    </div>
    <div class="hero-actions" class:confirming={live.confirmingStop}>
      <button
        class="live-btn"
        class:is-live={app.isLive}
        class:busy={live.action !== "idle"}
        disabled={live.action !== "idle" || !app.room}
        onclick={toggleLive}
        title="快捷键：Space"
      >
        <span class="live-btn-icon"><span class="spinner"></span><Play class="icon" fill="currentColor" stroke="none" /><Square class="icon" fill="currentColor" stroke="none" /></span>
        <span class="live-btn-label">{label}</span>
        <kbd>Space</kbd>
      </button>
      <div class="confirm-stop" role="alertdialog" aria-label="确认结束直播" inert={!live.confirmingStop}>
        <span>确定结束直播？</span>
        <button class="btn btn-ghost btn-sm" onclick={() => (live.confirmingStop = false)}>取消</button>
        <button class="btn btn-danger btn-sm" onclick={() => void stopLive()}>结束直播</button>
      </div>
    </div>
  </div>
</section>

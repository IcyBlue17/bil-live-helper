<script lang="ts">
  // The broadcast card, laid out like a production deployment: status,
  // duration and room details on the left, the start/stop action on the right.
  import { Check, Copy, Play, Square } from "@lucide/svelte";
  import { elapsed } from "../lib/format";
  import { app, copy, live, stopLive, toggleLive, ui } from "../lib/state.svelte";

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

  let copied = $state(false);
  async function copyRoom() {
    const id = app.room?.roomId;
    if (id && (await copy(String(id), "房间号"))) {
      copied = true;
      setTimeout(() => (copied = false), 1400);
    }
  }
</script>

<section class="card hero" class:hero-live={app.isLive}>
  <div class="hero-main">
    <div class="hero-text">
      <h2 class="hero-status">{status === "live" ? "正在直播" : status === "round" ? "轮播中" : "准备开播"}</h2>
      <p class="hero-sub">
        {#if !app.room}正在加载直播间…
        {:else if status === "live"}记得在 OBS 中开始推流，观众才能看到画面。
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
  <dl class="hero-meta">
    <div>
      <dt>状态</dt>
      <dd class="hero-state">
        <span class="dot" class:dot-live={status === "live"} class:dot-info={status === "round"}></span>
        {status === "live" ? "直播中" : status === "round" ? "轮播中" : "未开播"}
      </dd>
    </div>
    <div>
      <dt>时长</dt>
      <dd class="mono num" class:dim={status !== "live"}>{status === "live" ? elapsed(app.room?.liveSince ?? null, app.now) : "--:--:--"}</dd>
    </div>
    <div>
      <dt>房间号</dt>
      <dd>
        <button class="room-id mono num" class:copied title="点击复制房间号" onclick={copyRoom}>
          {app.room?.roomId ?? "—"}
          <span class="copy-icons"><Copy class="icon" /><Check class="icon" /></span>
        </button>
      </dd>
    </div>
    <div>
      <dt>分区</dt>
      <dd>
        {#if app.room}
          <button class="area-link" title="选择分区 (A)" onclick={() => (ui.areaOpen = true)}>{app.room.parentAreaName}<span class="crumb">/</span>{app.room.areaName}</button>
        {:else}—{/if}
      </dd>
    </div>
  </dl>
</section>

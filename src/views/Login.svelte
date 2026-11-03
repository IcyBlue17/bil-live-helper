<script lang="ts">
  import { RefreshCw } from "@lucide/svelte";
  import Mark from "../components/Mark.svelte";
  import QrTile from "../components/QrTile.svelte";
  import { app, login, startLogin } from "../lib/state.svelte";

  let attempt = $state(0);
  $effect(() => {
    void attempt;
    return startLogin();
  });
  const retry = () => attempt++;

  const labels: Record<string, string> = {
    loading: "正在获取二维码…",
    waiting: "请使用哔哩哔哩 App 扫码登录",
    scanned: "已扫码，请在手机上确认",
    done: "登录成功，正在进入…",
    expired: "二维码已过期",
    error: "获取二维码失败",
  };
  const stage = $derived(login.stage);
  const text = $derived(stage === "error" ? login.message || labels.error : (labels[stage] ?? login.message));
  const step = (n: 1 | 2 | 3) => {
    if (n === 1) return stage === "loading" ? "active" : "done";
    if (n === 2) return stage === "waiting" ? "active" : stage === "scanned" || stage === "done" ? "done" : "";
    return stage === "scanned" ? "active" : stage === "done" ? "done" : "";
  };
  function onkeydown(e: KeyboardEvent) {
    if (e.code === "KeyR" && (stage === "expired" || stage === "error")) retry();
  }
</script>

<svelte:window {onkeydown} />

<div class="login-screen">
  <section class="login-card">
    <div class="login-brand">
      <Mark size={40} />
      <div>
        <h1>Bil Live Helper</h1>
        <p class="muted">B 站直播推流助手</p>
      </div>
    </div>
    <p class="login-intro">扫码登录后，即可一键开播、获取 RTMP 推流地址与推流码，修改标题与分区，并实时查看弹幕。</p>
    <QrTile svg={login.svg} {stage} scannedText="请在手机上确认">
      {#snippet expired()}
        <p>{stage === "error" ? "获取失败" : "二维码已过期"}</p>
        <button class="btn btn-primary btn-sm" onclick={retry}><RefreshCw class="icon" />刷新二维码</button>
      {/snippet}
    </QrTile>
    <div class="login-stage stage-{stage}" aria-live="polite"><span class="stage-dot"></span><span>{text}</span></div>
    <ol class="login-steps">
      <li class={step(1)}><b>1</b>打开哔哩哔哩 App</li>
      <li class={step(2)}><b>2</b>扫描二维码</li>
      <li class={step(3)}><b>3</b>在手机上确认登录</li>
    </ol>
  </section>
  <footer class="login-foot muted">
    v{app.info?.version ?? ""}
    {#if app.info?.demo}<span class="badge badge-blue">演示模式</span>{/if}
  </footer>
</div>

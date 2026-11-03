<script lang="ts">
  import Modal from "../components/Modal.svelte";
  import QrTile from "../components/QrTile.svelte";
  import { cancelStart, live } from "../lib/state.svelte";

  const open = $derived(live.face !== null);
  // Keep the last content while the dialog animates out.
  let last = $state({ stage: "waiting", svg: "", message: "" });
  $effect(() => {
    if (live.face) last = { ...live.face };
  });
</script>

<Modal open={open} title="需要人脸认证" subtitle="B 站要求本次开播前完成实名人脸认证" width={440} onDismiss={cancelStart}>
  <div class="modal-body face">
    <QrTile svg={last.svg} stage={last.stage} small scannedText="请在手机上完成认证">
      {#snippet expired()}
        <p>二维码已过期</p>
        <p class="muted">请取消后重新开播</p>
      {/snippet}
    </QrTile>
    <ol class="face-steps">
      <li>使用<b>哔哩哔哩 App</b>扫描上方二维码</li>
      <li>按提示完成人脸识别</li>
      <li>认证通过后将自动开播，无需其他操作</li>
    </ol>
    <div class="login-stage stage-{last.stage}"><span class="stage-dot"></span><span>{last.message || "等待扫码…"}</span></div>
    <div class="modal-actions"><button class="btn btn-ghost" onclick={cancelStart}>取消开播</button></div>
  </div>
</Modal>

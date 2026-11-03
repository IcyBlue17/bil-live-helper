<script lang="ts">
  import { ExternalLink, Info } from "@lucide/svelte";
  import Kbd from "../components/Kbd.svelte";
  import Mark from "../components/Mark.svelte";
  import Modal from "../components/Modal.svelte";
  import { app, openURL, ui } from "../lib/state.svelte";

  const REPO_URL = "https://github.com/IcyBlue17/bil-live-helper";
  const UPSTREAM_URL = "https://github.com/chenxi-Eumenides/bilibili_live_tool";
  const shortcuts: [string[], string][] = [
    [["Space"], "开播 / 下播"],
    [["T"], "编辑标题"],
    [["A"], "选择分区"],
    [["R"], "刷新直播间"],
    [["C"], "复制推流码"],
    [["Shift", "C"], "复制服务器 + 推流码"],
    [["L"], "活动日志"],
    [["P"], "偏好设置"],
    [["Enter"], "发送弹幕"],
    [["?"], "关于与快捷键"],
  ];
</script>

<Modal bind:open={ui.aboutOpen} title="关于" width={560} class="about">
  <div class="modal-body">
    <div class="about-brand">
      <Mark size={44} />
      <div>
        <h3>Bil Live Helper <span class="badge num">v{app.info?.version ?? "?"}</span>{#if app.info?.demo}<span class="badge badge-amber">演示模式</span>{/if}</h3>
        <p class="muted">B 站直播助手：扫码登录、一键开播、获取推流码、收发弹幕。</p>
      </div>
    </div>
    <h4>快捷键</h4>
    <div class="shortcut-grid">
      {#each shortcuts as [keys, what] (what)}
        <div class="shortcut"><span>{what}</span><Kbd {keys} /></div>
      {/each}
    </div>
    <h4>配置文件</h4>
    <code class="config-path">{app.info?.configPath ?? "—"}</code>
    <p class="hint muted">登录凭据与直播间设置保存在这里，删除即退出登录。</p>
    <div class="disclaimer">
      <Info class="icon" />
      <span>本程序仅用于声明的功能，用户使用本程序产生的账号问题，与本程序无关。请勿用于不当用途。</span>
    </div>
  </div>
  <footer class="modal-actions">
    <button class="btn btn-ghost" onclick={() => openURL(UPSTREAM_URL)}>原版 bilibili_live_tool<ExternalLink class="icon" /></button>
    <button class="btn btn-secondary" onclick={() => openURL(REPO_URL)}>项目主页<ExternalLink class="icon" /></button>
  </footer>
</Modal>

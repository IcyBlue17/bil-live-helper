<script lang="ts">
  // RTMP server and stream key, as key/value rows like environment variables.
  import { EyeOff, Eye, Server, TriangleAlert } from "@lucide/svelte";
  import CopyButton from "../components/CopyButton.svelte";
  import { maskKey } from "../lib/format";
  import { rise } from "../lib/motion";
  import { app, streamText } from "../lib/state.svelte";

  let reveal = $state(false);
  const s = $derived(app.stream);
</script>

<section class="card stream-card">
  <header class="card-head">
    <h3>推流信息</h3>
    {#if s}<div class="card-tools"><CopyButton get={() => (s ? streamText(s) : "")} label="推流信息" text="复制全部" /></div>{/if}
  </header>
  {#if s}
    <div class="stream-body" in:rise>
      {#if s.changed}
        <div class="banner banner-warn" role="status"><TriangleAlert class="icon" /><span>推流码已变更，请更新 OBS 中的推流码。</span></div>
      {/if}
      <div class="kv">
        <div class="kv-row">
          <span class="kv-key">服务器</span>
          <span class="kv-value mono" title={s.addr}>{s.addr}</span>
          <div class="kv-tools"><CopyButton get={() => s.addr} label="服务器地址" /></div>
        </div>
        <div class="kv-row">
          <span class="kv-key">推流码</span>
          <span class="kv-value mono" class:masked={!reveal} title={reveal ? s.code : ""}>{reveal ? s.code : maskKey(s.code)}</span>
          <div class="kv-tools">
            <button class="icon-btn" title={reveal ? "隐藏推流码" : "显示推流码"} aria-label={reveal ? "隐藏推流码" : "显示推流码"} aria-pressed={reveal} onclick={() => (reveal = !reveal)}>
              <span class="swap-icons" class:on={reveal}><Eye class="icon" /><EyeOff class="icon" /></span>
            </button>
            <CopyButton get={() => s.code} label="推流码" />
          </div>
        </div>
      </div>
    </div>
    <footer class="section-foot">
      <span>推流码仅本次开播有效，请勿泄露。</span>
      <span class="foot-keys"><span><kbd>C</kbd>复制推流码</span><span><kbd>Shift</kbd><kbd>C</kbd>复制全部</span></span>
    </footer>
  {:else}
    <div class="empty">
      <span class="empty-icon"><Server /></span>
      <p>还没有推流信息</p>
      <p>点击「开始直播」后，这里会显示服务器地址和推流码，复制到 OBS 的「推流」设置中即可。</p>
    </div>
  {/if}
</section>

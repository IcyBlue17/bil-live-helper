<script lang="ts">
  import NumberFlow from "@number-flow/svelte";
  import { compactFormat, dayjs, elapsed } from "../../lib/format";
  import { app, feed } from "../../lib/state.svelte";
  import OnlineChart from "./OnlineChart.svelte";

  const s = $derived(feed.stats);
  const tiles = $derived(
    s
      ? [
          { label: "弹幕", value: s.danmaku },
          { label: "发言人数", value: s.senders },
          { label: "进场", value: s.enters },
          { label: "新增关注", value: s.follows },
          { label: "点赞", value: s.likes },
          { label: "礼物", value: s.gifts },
          { label: "醒目留言", value: s.superChats },
          { label: "上舰", value: s.guards },
        ]
      : [],
  );
</script>

{#if !s}
  <div class="empty empty-sm">
    {#if feed.conn === "connected"}
      <p>正在统计本场数据…</p>
      <p class="muted">弹幕、礼物和在线人数会在这里汇总</p>
    {:else}
      <p>正在连接弹幕服务器…</p>
      <p class="muted">连接后开始统计本场数据</p>
    {/if}
  </div>
{:else}
  <div class="stats">
    <div class="stats-hero">
      <div>
        <span class="eyebrow">礼物流水</span>
        <b class="stats-big num">¥<NumberFlow value={s.giftValue / 1000} format={{ maximumFractionDigits: 1 }} locales="zh-CN" /></b>
      </div>
      <div>
        <span class="eyebrow">峰值在线</span>
        <b class="stats-big num"><NumberFlow value={s.peakOnline} format={compactFormat} locales="zh-CN" /></b>
      </div>
      <div class="stats-since muted">
        统计自 {dayjs(s.since).format("HH:mm")}<br /><span class="num">{elapsed(s.since, app.now)}</span>
      </div>
    </div>
    <dl class="stats-grid">
      {#each tiles as t (t.label)}
        <div class="stats-tile">
          <dt>{t.label}</dt>
          <dd class="num"><NumberFlow value={t.value} format={compactFormat} locales="zh-CN" /></dd>
        </div>
      {/each}
    </dl>
    <div class="chart-box">
      <div class="chart-head"><span class="eyebrow">在线人数</span><span class="muted">每分钟采样</span></div>
      <OnlineChart samples={s.samples} />
    </div>
  </div>
{/if}

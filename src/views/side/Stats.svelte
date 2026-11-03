<script lang="ts">
  // 本场数据, laid out like Vercel Analytics: headline metrics over the chart,
  // then a grid of counters.
  import { ChartLine } from "@lucide/svelte";
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
  <section class="card">
    <div class="empty">
      <span class="empty-icon"><ChartLine /></span>
      {#if feed.conn === "connected"}
        <p>正在统计本场数据…</p>
        <p>弹幕、礼物和在线人数会在这里汇总</p>
      {:else}
        <p>正在连接弹幕服务器…</p>
        <p>连接后开始统计本场数据</p>
      {/if}
    </div>
  </section>
{:else}
  <section class="card analytics">
    <div class="metrics">
      <div class="metric active">
        <span class="metric-label">在线人数</span>
        <b class="metric-value num"><NumberFlow value={feed.online || app.room?.online || 0} format={compactFormat} locales="zh-CN" /></b>
      </div>
      <div class="metric">
        <span class="metric-label">峰值在线</span>
        <b class="metric-value num"><NumberFlow value={s.peakOnline} format={compactFormat} locales="zh-CN" /></b>
      </div>
      <div class="metric">
        <span class="metric-label">礼物流水</span>
        <b class="metric-value num">¥<NumberFlow value={s.giftValue / 1000} format={{ maximumFractionDigits: 1 }} locales="zh-CN" /></b>
      </div>
      <div class="metric">
        <span class="metric-label">统计时长</span>
        <b class="metric-value mono num">{elapsed(s.since, app.now)}</b>
      </div>
    </div>
    <div class="chart-box">
      <OnlineChart samples={s.samples} />
    </div>
    <footer class="section-foot">
      <span>统计自 <span class="mono">{dayjs(s.since).format("HH:mm")}</span>，在线人数每分钟采样一次。</span>
    </footer>
  </section>
  <section class="card">
    <header class="card-head"><h3>互动</h3></header>
    <dl class="stats-grid">
      {#each tiles as t (t.label)}
        <div class="stats-tile">
          <dt>{t.label}</dt>
          <dd class="num"><NumberFlow value={t.value} format={compactFormat} locales="zh-CN" /></dd>
        </div>
      {/each}
    </dl>
  </section>
{/if}

<script lang="ts">
  // The room's live counters as a row of metric cells, like Vercel Analytics.
  import NumberFlow from "@number-flow/svelte";
  import { compactFormat } from "../lib/format";
  import { app, feed } from "../lib/state.svelte";

  // Online: the high-energy count from the feed when reported, else the room's.
  const online = $derived(feed.online || app.room?.online || 0);
  const stats = $derived([
    { label: "在线", hint: "高能在线人数", value: online },
    { label: "看过", hint: "本场看过人数", value: feed.watched },
    { label: "点赞", hint: "直播间点赞数", value: feed.likes },
    { label: "粉丝", hint: "粉丝数", value: app.room?.attention ?? 0 },
  ]);
</script>

<section class="card kpis" aria-label="直播间数据">
  {#each stats as s (s.label)}
    <div class="kpi" title={s.hint}>
      <span class="kpi-label">{s.label}</span>
      <span class="kpi-value num">
        {#if s.value}<NumberFlow value={s.value} format={compactFormat} locales="zh-CN" />{:else}<span class="dim">—</span>{/if}
      </span>
    </div>
  {/each}
</section>

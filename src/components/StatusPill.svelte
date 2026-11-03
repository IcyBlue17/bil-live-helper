<script lang="ts">
  import { elapsed } from "../lib/format";
  import { app } from "../lib/state.svelte";
  const status = $derived(app.room?.status ?? "offline");
</script>

<span class="status-pill status-{status}">
  <span class="status-dot"></span>
  <span>{status === "live" ? "直播中" : status === "round" ? "轮播中" : "未开播"}</span>
  {#if status === "live"}<span class="status-time num">{elapsed(app.room?.liveSince ?? null, app.now)}</span>{/if}
</span>

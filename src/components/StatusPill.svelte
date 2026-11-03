<script lang="ts">
  // The room's broadcast state as a dot and a label, like a deployment status.
  import { app } from "../lib/state.svelte";
  const status = $derived(app.room?.status ?? "offline");
</script>

<span class="status-pill status-{status}">
  <span class="dot" class:dot-live={status === "live"} class:dot-info={status === "round"}></span>
  <span>{status === "live" ? "直播中" : status === "round" ? "轮播中" : "未开播"}</span>
</span>

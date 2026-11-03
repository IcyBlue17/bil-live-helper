<script lang="ts">
  // A round avatar; falls back to the first character of the name.
  let { src = "", name = "", size = 28 }: { src?: string; name?: string; size?: number } = $props();
  let failed = $state(false);
  $effect(() => {
    void src;
    failed = false;
  });
</script>

<span class="avatar" style="width:{size}px;height:{size}px;font-size:{Math.round(size * 0.46)}px">
  {#if src && !failed}
    <img {src} alt="" referrerpolicy="no-referrer" onerror={() => (failed = true)} />
  {:else}
    <span>{[...name][0] ?? "?"}</span>
  {/if}
</span>

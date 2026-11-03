<script lang="ts">
  let { keys }: { keys: string[] } = $props();
</script>

<span class="kbds">{#each keys as k (k)}<kbd>{k}</kbd>{/each}</span>

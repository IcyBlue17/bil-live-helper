<script lang="ts">
  // Copies text and shows a check mark for a moment.
  import { Check, Copy } from "@lucide/svelte";
  import { copy } from "../lib/state.svelte";

  let { get, label, text = "", class: cls = "", flash = $bindable(0) }: { get: () => string; label: string; text?: string; class?: string; flash?: number } = $props();
  let copied = $state(false);
  let timer: ReturnType<typeof setTimeout> | undefined;

  function show() {
    copied = true;
    clearTimeout(timer);
    timer = setTimeout(() => (copied = false), 1400);
  }
  // Lets keyboard shortcuts show the same confirmation.
  $effect(() => {
    if (flash) show();
  });
  async function onclick() {
    const v = get();
    if (v && (await copy(v, label))) show();
  }
</script>

<button class="copy-btn {text ? 'btn btn-ghost btn-sm' : 'icon-btn'} {cls}" class:copied title="复制{label}" aria-label="复制{label}" {onclick}>
  <span class="copy-icons"><Copy class="icon" /><Check class="icon" /></span>
  {#if text}<span>{text}</span>{/if}
</button>

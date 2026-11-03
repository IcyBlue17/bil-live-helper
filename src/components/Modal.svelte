<script lang="ts">
  // A dialog on bits-ui (focus trap, Esc, scroll lock, a11y) with our styling.
  import { Dialog } from "bits-ui";
  import { X } from "@lucide/svelte";
  import type { Snippet } from "svelte";
  import { fadeIn, pop } from "../lib/motion";

  let {
    open = $bindable(false),
    title,
    subtitle = "",
    width = 520,
    class: cls = "",
    onDismiss,
    children,
  }: {
    open?: boolean;
    title: string;
    subtitle?: string;
    width?: number;
    class?: string;
    /** Called instead of closing on Esc, outside click or ×. */
    onDismiss?: () => void;
    children: Snippet;
  } = $props();

  /** Focus the first field rather than the close button. */
  function onOpenAutoFocus(e: Event) {
    const root = e.target as HTMLElement;
    const field = root.querySelector<HTMLElement>("input:not([type=hidden]):not(:disabled), textarea:not(:disabled)");
    e.preventDefault();
    (field ?? root).focus({ preventScroll: true });
  }

  function onOpenChange(v: boolean) {
    if (!v && onDismiss) onDismiss();
    else open = v;
  }
</script>

<Dialog.Root bind:open={() => open, onOpenChange}>
  <Dialog.Portal>
    <Dialog.Overlay forceMount>
      {#snippet child({ props, open: o })}
        {#if o}<div {...props} class="backdrop" transition:fadeIn></div>{/if}
      {/snippet}
    </Dialog.Overlay>
    <Dialog.Content forceMount {onOpenAutoFocus}>
      {#snippet child({ props, open: o })}
        {#if o}
          <div class="modal-wrap">
            <div {...props} class="modal {cls}" style="width:min({width}px, calc(100vw - 48px))" transition:pop>
              <header class="modal-head">
                <div>
                  <Dialog.Title level={2}>{title}</Dialog.Title>
                  {#if subtitle}<Dialog.Description class="muted">{subtitle}</Dialog.Description>{/if}
                </div>
                <Dialog.Close class="icon-btn" aria-label="关闭" title="关闭 (Esc)"><X class="icon" /></Dialog.Close>
              </header>
              {@render children()}
            </div>
          </div>
        {/if}
      {/snippet}
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>

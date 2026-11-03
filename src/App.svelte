<script lang="ts">
  import { RefreshCw } from "@lucide/svelte";
  import { Toaster } from "svelte-sonner";
  import Mark from "./components/Mark.svelte";
  import { app, boot, copy, logs, refreshRoom, streamText, toggleLive, ui } from "./lib/state.svelte";
  import { rise } from "./lib/motion";
  import Dashboard from "./views/Dashboard.svelte";
  import Login from "./views/Login.svelte";
  import About from "./views/About.svelte";

  void boot();

  function isTyping(t: EventTarget | null): boolean {
    if (!(t instanceof HTMLElement)) return false;
    if (t instanceof HTMLInputElement) return !["button", "checkbox", "radio"].includes(t.type);
    return t instanceof HTMLTextAreaElement || t.isContentEditable;
  }
  /** A bits-ui dialog or popover is open: it owns the keyboard. */
  const layerOpen = () => !!document.querySelector("[role=dialog][data-state=open], [data-popover-content][data-state=open], .menu-content");

  function onkeydown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      if (!layerOpen() && isTyping(e.target)) (e.target as HTMLElement).blur();
      return;
    }
    if (e.ctrlKey || e.metaKey || e.altKey || e.isComposing || isTyping(e.target)) return;
    if (e.key === "?") {
      e.preventDefault();
      ui.aboutOpen = true;
      return;
    }
    if (layerOpen() || app.phase !== "app") return;
    const run = (fn: () => void) => {
      e.preventDefault();
      fn();
    };
    switch (e.code) {
      case "Space":
        return run(toggleLive);
      case "KeyT":
        return run(() => {
          ui.narrowTab = "control";
          ui.focusTitle++;
        });
      case "KeyA":
        return run(() => (ui.areaOpen = true));
      case "KeyR":
        return run(() => void refreshRoom());
      case "KeyC":
        return run(() => {
          const s = app.stream;
          if (s) void (e.shiftKey ? copy(streamText(s), "推流信息") : copy(s.code, "推流码"));
        });
      case "KeyL":
        return run(() => (logs.open = !logs.open));
      case "KeyP":
        return run(() => (ui.prefsOpen = true));
    }
  }
</script>

<svelte:window {onkeydown} />

{#key app.phase}
  <div class="view" in:rise>
    {#if app.phase === "boot"}
      <div class="boot">
        <div class="skel-topbar"><span class="skel skel-mark"></span><span class="skel skel-line w-40"></span><span class="skel skel-pill"></span></div>
        <div class="skel-columns">
          <div class="skel-col">
            <div class="skel skel-card tall"></div>
            <div class="skel-row"><div class="skel skel-card"></div><div class="skel skel-card"></div></div>
          </div>
          <div class="skel skel-card side"></div>
        </div>
      </div>
    {:else if app.phase === "unauth"}
      <Login />
    {:else if app.phase === "error"}
      <div class="login-screen">
        <section class="login-card error-card">
          <Mark size={40} />
          <h1>无法连接后端</h1>
          <p class="muted">{app.bootError}</p>
          <button class="btn btn-primary" onclick={() => void boot()}><RefreshCw class="icon" />重试</button>
        </section>
      </div>
    {:else}
      <Dashboard />
    {/if}
  </div>
{/key}

<About />
<Toaster theme="dark" position="top-center" offset={72} duration={2600} closeButton={false} toastOptions={{ class: "toast" }} />

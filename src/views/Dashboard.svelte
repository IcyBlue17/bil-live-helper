<script lang="ts">
  import { MessageSquareText, SlidersHorizontal } from "@lucide/svelte";
  import { onDestroy } from "svelte";
  import { feedListeners, ui } from "../lib/state.svelte";
  import DanmakuPanel from "./DanmakuPanel.svelte";
  import FaceAuth from "./FaceAuth.svelte";
  import LiveCard from "./LiveCard.svelte";
  import LogDrawer from "./LogDrawer.svelte";
  import Prefs from "./Prefs.svelte";
  import SettingsCard from "./SettingsCard.svelte";
  import SideCard from "./side/SideCard.svelte";
  import StreamCard from "./StreamCard.svelte";
  import Topbar from "./Topbar.svelte";
  import ViewerCard from "./ViewerCard.svelte";

  // Unread danmaku while the narrow layout shows the control column.
  let unread = $state(0);
  const onFeed = (items: unknown[], reset: boolean) => {
    if (reset) unread = 0;
    else if (ui.narrowTab === "control") unread += items.length;
  };
  feedListeners.add(onFeed);
  onDestroy(() => feedListeners.delete(onFeed));
  $effect(() => {
    if (ui.narrowTab === "danmaku") unread = 0;
  });
</script>

<div class="dashboard tab-{ui.narrowTab}">
  <Topbar />
  <nav class="tabs" aria-label="视图">
    <button class="tab" class:active={ui.narrowTab === "control"} onclick={() => (ui.narrowTab = "control")}>
      <SlidersHorizontal class="icon" />直播控制
    </button>
    <button class="tab" class:active={ui.narrowTab === "danmaku"} onclick={() => (ui.narrowTab = "danmaku")}>
      <MessageSquareText class="icon" />弹幕
      <span class="tab-badge num" class:show={unread > 0}>{unread > 99 ? "99+" : unread}</span>
    </button>
  </nav>
  <main class="columns">
    <div class="col col-control">
      <LiveCard />
      <div class="row-2">
        <StreamCard />
        <SettingsCard />
      </div>
      <SideCard />
    </div>
    <div class="col col-danmaku">
      <DanmakuPanel />
    </div>
  </main>
  <LogDrawer />
</div>

<FaceAuth />
<ViewerCard />
<Prefs />

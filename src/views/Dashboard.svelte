<script lang="ts">
  import { ChartLine, Crown, LayoutGrid, Megaphone, MessageSquareText, ShieldBan } from "@lucide/svelte";
  import { Tabs } from "bits-ui";
  import { onDestroy } from "svelte";
  import { feedListeners, ui } from "../lib/state.svelte";
  import DanmakuPanel from "./DanmakuPanel.svelte";
  import FaceAuth from "./FaceAuth.svelte";
  import Kpis from "./Kpis.svelte";
  import LiveCard from "./LiveCard.svelte";
  import LogDrawer from "./LogDrawer.svelte";
  import Prefs from "./Prefs.svelte";
  import SettingsCard from "./SettingsCard.svelte";
  import Announce from "./side/Announce.svelte";
  import Muted from "./side/Muted.svelte";
  import Rank from "./side/Rank.svelte";
  import Stats from "./side/Stats.svelte";
  import StreamCard from "./StreamCard.svelte";
  import Topbar from "./Topbar.svelte";
  import ViewerCard from "./ViewerCard.svelte";

  const tabs = [
    { value: "overview", label: "概览", icon: LayoutGrid },
    { value: "stats", label: "本场数据", icon: ChartLine },
    { value: "rank", label: "榜单", icon: Crown },
    { value: "announce", label: "公告", icon: Megaphone },
    { value: "muted", label: "禁言", icon: ShieldBan },
  ] as const;

  // In the narrow layout the danmaku column becomes one more tab.
  const mq = matchMedia("(max-width: 1000px)");
  let narrow = $state(mq.matches);
  const onMq = (e: MediaQueryListEvent) => (narrow = e.matches);
  mq.addEventListener("change", onMq);
  onDestroy(() => mq.removeEventListener("change", onMq));

  const current = () => (narrow && ui.narrowTab === "danmaku" ? "danmaku" : ui.tab);
  function select(v: string) {
    if (v === "danmaku") ui.narrowTab = "danmaku";
    else {
      ui.narrowTab = "control";
      ui.tab = v as typeof ui.tab;
    }
  }

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
  <Tabs.Root class="dash-body" bind:value={current, select}>
    <Tabs.List class="page-tabs" aria-label="页面">
      {#each tabs as t (t.value)}
        <Tabs.Trigger value={t.value} class="page-tab"><t.icon class="icon" />{t.label}</Tabs.Trigger>
      {/each}
      <Tabs.Trigger value="danmaku" class="page-tab narrow-only">
        <MessageSquareText class="icon" />弹幕
        <span class="tab-badge num" class:show={unread > 0}>{unread > 99 ? "99+" : unread}</span>
      </Tabs.Trigger>
    </Tabs.List>
    <main class="columns">
      <div class="col col-control">
        <Tabs.Content value="overview" class="page">
          <LiveCard />
          <StreamCard />
          <Kpis />
          <SettingsCard />
        </Tabs.Content>
        <Tabs.Content value="stats" class="page"><Stats /></Tabs.Content>
        <Tabs.Content value="rank" class="page"><Rank /></Tabs.Content>
        <Tabs.Content value="announce" class="page"><Announce /></Tabs.Content>
        <Tabs.Content value="muted" class="page"><Muted /></Tabs.Content>
      </div>
      <div class="col col-danmaku">
        <DanmakuPanel />
      </div>
    </main>
  </Tabs.Root>
  <LogDrawer />
</div>

<FaceAuth />
<ViewerCard />
<Prefs />

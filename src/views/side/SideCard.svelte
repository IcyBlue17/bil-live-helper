<script lang="ts">
  // The secondary card under the stream and settings cards: stats, ranks,
  // announcement and the mute list, one tab at a time.
  import { ChartLine, Crown, Megaphone, ShieldBan } from "@lucide/svelte";
  import { Tabs } from "bits-ui";
  import { ui } from "../../lib/state.svelte";
  import Announce from "./Announce.svelte";
  import Muted from "./Muted.svelte";
  import Rank from "./Rank.svelte";
  import Stats from "./Stats.svelte";

  const tabs = [
    { value: "stats", label: "本场数据", icon: ChartLine },
    { value: "rank", label: "榜单", icon: Crown },
    { value: "announce", label: "公告", icon: Megaphone },
    { value: "muted", label: "禁言", icon: ShieldBan },
  ] as const;
</script>

<section class="card side-card">
  <Tabs.Root bind:value={ui.sideTab}>
    <Tabs.List class="side-tabs" aria-label="更多">
      {#each tabs as t (t.value)}
        <Tabs.Trigger value={t.value} class="side-tab"><t.icon class="icon" />{t.label}</Tabs.Trigger>
      {/each}
    </Tabs.List>
    <Tabs.Content value="stats" class="side-panel"><Stats /></Tabs.Content>
    <Tabs.Content value="rank" class="side-panel"><Rank /></Tabs.Content>
    <Tabs.Content value="announce" class="side-panel"><Announce /></Tabs.Content>
    <Tabs.Content value="muted" class="side-panel"><Muted /></Tabs.Content>
  </Tabs.Root>
</section>

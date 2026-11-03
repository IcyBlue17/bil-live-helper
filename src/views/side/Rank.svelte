<script lang="ts">
  import { RefreshCw } from "@lucide/svelte";
  import { Tabs } from "bits-ui";
  import Avatar from "../../components/Avatar.svelte";
  import { Live, type RankKind, type RankUser } from "../../mygo";
  import { animateList } from "../../lib/actions";
  import { compact, errorMessage, guardName } from "../../lib/format";
  import { ui } from "../../lib/state.svelte";

  let kind = $state<RankKind>("online");
  let rows = $state.raw<RankUser[] | null>(null);
  let error = $state("");
  let loading = $state(false);
  let seq = 0;

  async function load() {
    const id = ++seq;
    const k = kind;
    loading = true;
    error = "";
    try {
      const r = await Live.rank(k);
      if (id === seq) rows = r;
    } catch (err) {
      if (id === seq) error = errorMessage(err);
    } finally {
      if (id === seq) loading = false;
    }
  }
  $effect(() => {
    void kind;
    rows = null;
    void load();
  });
  function openViewer(r: RankUser, e: MouseEvent) {
    ui.viewer = { uid: r.uid, name: r.name, anchor: e.currentTarget as HTMLElement };
  }
</script>

<div class="side-tools">
  <Tabs.Root bind:value={() => kind, (v) => (kind = v as RankKind)}>
    <Tabs.List class="seg">
      <Tabs.Trigger value="online" class="seg-item">高能榜</Tabs.Trigger>
      <Tabs.Trigger value="guard" class="seg-item">大航海</Tabs.Trigger>
    </Tabs.List>
  </Tabs.Root>
  <button class="icon-btn" class:spinning={loading} title="刷新榜单" aria-label="刷新榜单" onclick={() => void load()}><RefreshCw class="icon" /></button>
</div>

{#if error}
  <div class="empty empty-sm"><p>榜单加载失败</p><p class="muted">{error}</p><button class="btn btn-secondary btn-sm" onclick={() => void load()}>重试</button></div>
{:else if !rows}
  <div class="rank-list">{#each { length: 4 } as _, i (i)}<span class="skel rank-skel"></span>{/each}</div>
{:else if rows.length === 0}
  <div class="empty empty-sm"><p>{kind === "online" ? "高能榜暂时为空" : "还没有舰长"}</p><p class="muted">{kind === "online" ? "本场送礼或发弹幕的观众会出现在这里" : "上舰的观众会出现在这里"}</p></div>
{:else}
  <ol class="rank-list" use:animateList>
    {#each rows as r (r.uid)}
      <li class="rank-row">
        <span class="rank-no num" class:top={r.rank <= 3}>{r.rank}</span>
        <Avatar src={r.face} name={r.name} size={28} />
        <button class="rank-name" onclick={(e) => openViewer(r, e)}>{r.name}</button>
        {#if r.guard}<span class="tag tag-guard-{r.guard}">{guardName[r.guard]}</span>{/if}
        {#if r.medal}<span class="medal"><span>{r.medal.name}</span><b class="num">{r.medal.level}</b></span>{/if}
        <span class="rank-score num">{kind === "online" ? compact(r.score) : ""}</span>
      </li>
    {/each}
  </ol>
{/if}

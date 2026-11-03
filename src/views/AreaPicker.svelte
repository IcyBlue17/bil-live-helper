<script lang="ts">
  // Two-level area picker: parent list + child grid, with fuzzy and pinyin
  // search across every area ("cj" or "changjian" finds 视频唱见).
  import { Check, Search } from "@lucide/svelte";
  import Modal from "../components/Modal.svelte";
  import { areaIndex, findArea, type AreaHit } from "../lib/areaSearch";
  import { app } from "../lib/state.svelte";

  let { open = $bindable(false), current, onPick }: { open?: boolean; current: number; onPick: (id: number) => void } = $props();

  let query = $state("");
  let groupId = $state(0);
  const search = $derived(areaIndex(app.areas));

  $effect.pre(() => {
    if (!open) return;
    query = "";
    groupId = findArea(app.areas, current)?.group.id ?? app.areas[0]?.id ?? 0;
  });

  const q = $derived(query.trim());
  const items: AreaHit[] = $derived.by(() => {
    if (q) return search(q);
    const g = app.areas.find((x) => x.id === groupId);
    return g ? g.areas.map((area) => ({ group: g, area })) : [];
  });
  function choose(id: number) {
    onPick(id);
    open = false;
  }
</script>

<Modal bind:open title="选择分区" subtitle="选择直播内容所属的分区，支持拼音与首字母搜索" width={680} class="picker-modal">
  <div class="modal-body picker">
    <div class="search-wrap">
      <Search class="icon" />
      <input
        class="input search-input"
        type="search"
        placeholder="搜索分区，如「唱见」「changjian」「cj」"
        bind:value={query}
        onkeydown={(e) => {
          if (e.key === "Enter" && !e.isComposing && items[0]) choose(items[0].area.id);
        }}
      />
    </div>
    <div class="picker-cols">
      <nav class="group-list" class:dimmed={!!q}>
        {#each app.areas as g (g.id)}
          <button class="group-item" class:active={!q && groupId === g.id} onclick={() => { query = ""; groupId = g.id; }}>
            <span>{g.name}</span><span class="muted num">{g.areas.length}</span>
          </button>
        {/each}
      </nav>
      {#key q ? "q" : groupId}
        <div class="area-grid fade-in">
          {#each items as { group, area } (area.id)}
            <button class="area-chip" class:selected={area.id === current} onclick={() => choose(area.id)} title="{group.name} / {area.name}">
              <span class="chip-name">{area.name}</span>
              {#if q}<span class="chip-group muted">{group.name}</span>{/if}
              {#if area.id === current}<Check class="icon" />{/if}
            </button>
          {:else}
            <div class="empty empty-sm">
              <p>{app.areas.length ? "没有匹配的分区" : "分区列表加载中…"}</p>
              {#if q}<p class="muted">换个关键词试试</p>{/if}
            </div>
          {/each}
        </div>
      {/key}
    </div>
  </div>
</Modal>

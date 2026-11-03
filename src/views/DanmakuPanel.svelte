<script lang="ts">
  // The 弹幕 panel: connection state, a virtualized feed, and the send box.
  import { ArrowDown, Flame, MessageSquareText } from "@lucide/svelte";
  import NumberFlow from "@number-flow/svelte";
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { onMount } from "svelte";
  import { compactFormat } from "../lib/format";
  import { feed, feedListeners, type FeedItem } from "../lib/state.svelte";
  import MessageRow from "./MessageRow.svelte";
  import SendBox from "./SendBox.svelte";

  const connLabel = { connecting: "连接中", connected: "已连接", reconnecting: "重连中", closed: "未连接" } as const;
  const connDot = { connecting: "dot-busy", connected: "dot-ok", reconnecting: "dot-busy", closed: "" } as const;

  let scroller: HTMLDivElement | undefined = $state();
  let atBottom = true;
  let unseen = $state(0);
  /** Rows arrived after mount slide in; the initial batch and history do not. */
  let mountedAt = $state(Date.now());

  const virtualizer = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0,
    getScrollElement: () => scroller ?? null,
    estimateSize: () => 30,
    overscan: 12,
    getItemKey: (i) => feed.items[i]?.id ?? i,
  });

  $effect(() => {
    $virtualizer.setOptions({ count: feed.items.length, getItemKey: (i) => feed.items[i]?.id ?? i });
  });

  function toBottom() {
    const n = feed.items.length;
    if (n) $virtualizer.scrollToIndex(n - 1, { align: "end" });
    unseen = 0;
  }
  function onscroll() {
    if (!scroller) return;
    atBottom = scroller.scrollHeight - scroller.scrollTop - scroller.clientHeight < 40;
    if (atBottom) unseen = 0;
  }

  onMount(() => {
    const onItems = (items: FeedItem[], reset: boolean) => {
      if (reset) {
        unseen = 0;
        atBottom = true;
        mountedAt = Date.now();
        return;
      }
      if (atBottom) requestAnimationFrame(() => requestAnimationFrame(toBottom));
      else unseen += items.filter((m) => m.kind !== "divider" && !m.historical).length;
    };
    feedListeners.add(onItems);
    requestAnimationFrame(toBottom);
    return () => feedListeners.delete(onItems);
  });

  function measure(node: HTMLDivElement) {
    $virtualizer.measureElement(node);
  }
</script>

<section class="card danmaku-panel">
  <header class="card-head">
    <h3>弹幕</h3>
    <div class="card-tools">
      <span class="stat-sm num" title="人气值"><Flame class="icon" />{#if feed.popularity}<NumberFlow value={feed.popularity} format={compactFormat} locales="zh-CN" />{:else}—{/if}</span>
      <span class="conn conn-{feed.conn}"><span class="dot {connDot[feed.conn]}"></span>{connLabel[feed.conn]}</span>
    </div>
  </header>
  <div class="feed">
    <div class="feed-scroll" bind:this={scroller} {onscroll}>
      <div class="feed-list" role="log" aria-live="off" style="height:{$virtualizer.getTotalSize()}px">
        {#each $virtualizer.getVirtualItems() as row (row.key)}
          {@const item = feed.items[row.index]}
          {#if item}
            <div class="feed-row" data-index={row.index} use:measure style="transform:translateY({row.start}px)">
              <MessageRow {item} fresh={item.arrived > mountedAt} />
            </div>
          {/if}
        {/each}
      </div>
    </div>
    {#if feed.items.length === 0}
      <div class="empty feed-empty">
        <span class="empty-icon"><MessageSquareText /></span>
        <p>暂无弹幕</p>
        <p class="muted">观众发送的弹幕、礼物、醒目留言和进场消息会显示在这里。点击昵称可查看观众卡片。</p>
      </div>
    {/if}
    <button class="new-pill" class:show={unseen > 0} onclick={toBottom} tabindex={unseen > 0 ? 0 : -1}>
      {unseen > 99 ? "99+" : unseen} 条新弹幕<ArrowDown class="icon" />
    </button>
  </div>
  <SendBox />
</section>

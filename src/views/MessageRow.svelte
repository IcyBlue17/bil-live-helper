<script lang="ts">
  import { Gift, Heart, Plus } from "@lucide/svelte";
  import { clock, guardName, yuan } from "../lib/format";
  import { app, ui, type FeedItem } from "../lib/state.svelte";
  import type { Emoticon, Message } from "../mygo";

  let { item, fresh }: { item: FeedItem; fresh: boolean } = $props();

  /** Splits text into strings and inline emotes. */
  function parts(m: Message): (string | Emoticon)[] {
    const tokens = Object.keys(m.emots ?? {});
    if (!tokens.length) return [m.text];
    const re = new RegExp(tokens.map((t) => t.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")).join("|"), "g");
    const out: (string | Emoticon)[] = [];
    let last = 0;
    for (const match of m.text.matchAll(re)) {
      const e = m.emots[match[0]];
      if (!e) continue;
      out.push(m.text.slice(last, match.index), e);
      last = match.index + match[0].length;
    }
    out.push(m.text.slice(last));
    return out.filter((p) => p !== "");
  }

  function openViewer(e: MouseEvent, m: Message) {
    if (!m.uid || m.uid === app.user?.uid) return;
    ui.viewer = { uid: m.uid, name: m.name, anchor: e.currentTarget as HTMLElement };
  }
  const isSelf = (m: Message) => m.kind === "self" || (m.kind === "danmaku" && !!app.user && m.uid === app.user.uid);
</script>

{#snippet who(m: Message)}
  <button class="msg-name" onclick={(e) => openViewer(e, m)} disabled={isSelf(m)} title={isSelf(m) ? undefined : "查看观众卡片"}>
    {isSelf(m) ? `${m.name}（我）` : m.name}
  </button>
{/snippet}
{#snippet medal(m: Message)}
  {#if m.medal}<span class="medal"><span>{m.medal.name}</span><b class="num">{m.medal.level}</b></span>{/if}
{/snippet}

{#if item.kind === "divider"}
  <div class="feed-divider"><span>以上为历史弹幕</span></div>
{:else}
  {@const m = item}
  {@const base = `msg msg-${m.kind}${m.historical ? " historical" : ""}${fresh ? " fresh" : ""}`}
  {#if m.kind === "superchat"}
    <div class={base}>
      <div class="sc-card">
        <div class="sc-head">
          {@render medal(m)}
          {@render who(m)}
          <span class="sc-price num">{m.superChat ? `¥${m.superChat.price}` : "SC"}</span>
        </div>
        <div class="sc-text">{m.text}</div>
      </div>
    </div>
  {:else if m.kind === "gift" || m.kind === "guard"}
    {@const g = m.gift}
    <div class="{base}{m.guard ? ` guard-${m.guard}` : g && g.price >= 50_000 ? ' gift-big' : ''}">
      <time class="msg-time num">{clock(m.time)}</time>
      <div class="msg-body">
        <Gift class="icon" />
        {@render who(m)}
        <span class="muted">{m.kind === "guard" || m.guard ? "开通了" : "送出"}</span>
        <span class="gift-name">{g ? `${g.name}${g.num > 1 ? ` ×${g.num}` : ""}` : m.text}</span>
        {#if g && g.price > 0}<span class="gift-price num">{yuan(g.price)}</span>{/if}
      </div>
    </div>
  {:else if m.kind === "enter" || m.kind === "follow" || m.kind === "like"}
    <div class={base}>
      <time class="msg-time num">{clock(m.time)}</time>
      <div class="msg-body">
        {#if m.kind === "follow"}<Plus class="icon kind-icon follow" />{:else if m.kind === "like"}<Heart class="icon kind-icon like" />{/if}
        {@render who(m)}
        <span>{m.text || (m.kind === "enter" ? "进入直播间" : m.kind === "follow" ? "关注了直播间" : "点赞了直播间")}</span>
      </div>
    </div>
  {:else if m.kind === "notice" || m.kind === "system"}
    <div class={base}><div class="msg-body">{m.text}</div></div>
  {:else}
    {@const cls = [isSelf(m) && "msg-self", m.guard ? `guard-${m.guard}` : m.medal && "fan", m.admin && "admin"].filter(Boolean).join(" ")}
    <div class="{base} {cls}">
      <time class="msg-time num">{clock(m.time)}</time>
      <div class="msg-body">
        <span class="msg-who">
          {#if m.admin}<span class="tag tag-admin" title="房管">房</span>{/if}
          {#if m.guard}<span class="tag tag-guard-{m.guard}">{guardName[m.guard]}</span>{/if}
          {@render medal(m)}
          {#if m.userLevel}<span class="ul num" title="用户等级">UL{m.userLevel}</span>{/if}
          {@render who(m)}<span class="colon">:</span>
        </span>
        {#if m.emoticon}
          <img class="emote-sticker" src={m.emoticon.url} alt={m.emoticon.name} title={m.emoticon.name} referrerpolicy="no-referrer" />
        {:else}
          <span class="msg-text">{#each parts(m) as p, i (i)}{#if typeof p === "string"}{p}{:else}<img class="emote" src={p.url} alt={p.token} title={p.name} referrerpolicy="no-referrer" />{/if}{/each}</span>
        {/if}
      </div>
    </div>
  {/if}
{/if}

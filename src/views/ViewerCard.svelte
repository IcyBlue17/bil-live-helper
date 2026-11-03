<script lang="ts">
  // The card shown when clicking a name in the feed: profile, mute, homepage.
  import { Ban, ExternalLink, ShieldCheck, Users } from "@lucide/svelte";
  import { Popover } from "bits-ui";
  import { toast } from "svelte-sonner";
  import Avatar from "../components/Avatar.svelte";
  import { Danmaku, type ViewerCard } from "../mygo";
  import { compact, errorMessage, guardName, spaceURL } from "../lib/format";
  import { fail, log, openURL, ui } from "../lib/state.svelte";

  let card = $state<ViewerCard | null>(null);
  let error = $state("");
  let busy = $state(false);

  const open = $derived(ui.viewer !== null);
  $effect(() => {
    const v = ui.viewer;
    if (!v) return;
    card = null;
    error = "";
    let live = true;
    Danmaku.viewer(v.uid)
      .then((c) => live && (card = c))
      .catch((err: unknown) => live && (error = errorMessage(err)));
    return () => (live = false);
  });

  async function mute(hours: number) {
    const c = card;
    if (!c || busy) return;
    busy = true;
    try {
      await Danmaku.mute(c.uid, hours);
      card = { ...c, muted: true };
      ui.mutedVersion++;
      const what = hours ? `${hours} 小时` : "至本场结束";
      toast.success(`已禁言 ${c.name}（${what}）`);
      log("ok", `已禁言 ${c.name}（${what}）`);
    } catch (err) {
      fail("禁言失败", err);
    } finally {
      busy = false;
    }
  }
</script>

<Popover.Root open={open} onOpenChange={(v) => !v && (ui.viewer = null)}>
  <Popover.Portal>
    <Popover.Content class="popover viewer-card" customAnchor={ui.viewer?.anchor ?? null} side="left" align="start" sideOffset={12} collisionPadding={12} onOpenAutoFocus={(e) => e.preventDefault()}>
      {@const name = card?.name ?? ui.viewer?.name ?? ""}
      <div class="vc-main">
        <div class="vc-head">
          <Avatar src={card?.face} {name} size={40} />
          <div class="vc-id">
            <div class="vc-name">
              <b>{name}</b>
              {#if card?.level}<span class="ul num">UL{card.level}</span>{/if}
            </div>
            <span class="mono num">UID {ui.viewer?.uid ?? ""}</span>
          </div>
        </div>
        {#if error}
          <p class="vc-error">{error}</p>
        {:else if !card}
          <div class="vc-loading"><span class="skel" style="height:12px;width:70%"></span><span class="skel" style="height:12px;width:45%"></span></div>
        {:else}
          {#if card.sign}<p class="vc-sign">{card.sign}</p>{/if}
          <div class="vc-tags">
            {#if card.guard}<span class="tag tag-guard-{card.guard}">{guardName[card.guard]}</span>{/if}
            {#if card.medal}<span class="medal"><span>{card.medal.name}</span><b class="num">{card.medal.level}</b></span>{/if}
            <span class="vc-fans"><Users class="icon" /><span class="num">{compact(card.followers)}</span> 粉丝</span>
            {#if card.muted}<span class="badge badge-red"><Ban class="icon" />禁言中</span>{/if}
          </div>
        {/if}
      </div>
      <div class="vc-actions">
        <button class="btn btn-ghost btn-sm" onclick={() => ui.viewer && openURL(spaceURL(ui.viewer.uid))}>主页<ExternalLink class="icon" /></button>
        <span class="vc-spacer"></span>
        {#if card?.muted}
          <span class="vc-muted"><ShieldCheck class="icon" />已禁言</span>
        {:else}
          <button class="btn btn-secondary btn-sm" disabled={!card || busy} onclick={() => void mute(1)}>禁言 1 小时</button>
          <button class="btn btn-danger-soft btn-sm" disabled={!card || busy} onclick={() => void mute(0)}>禁言本场</button>
        {/if}
      </div>
    </Popover.Content>
  </Popover.Portal>
</Popover.Root>

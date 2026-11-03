<script lang="ts">
  import { RefreshCw, ShieldBan } from "@lucide/svelte";
  import Avatar from "../../components/Avatar.svelte";
  import { toast } from "svelte-sonner";
  import { Danmaku, type MutedUser } from "../../mygo";
  import { animateList } from "../../lib/actions";
  import { dayjs, errorMessage } from "../../lib/format";
  import { app, fail, log, ui } from "../../lib/state.svelte";

  let rows = $state.raw<MutedUser[] | null>(null);
  let error = $state("");
  let loading = $state(false);
  let busy = $state<number | null>(null);

  async function load() {
    loading = true;
    error = "";
    try {
      rows = await Danmaku.muted();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      loading = false;
    }
  }
  $effect(() => {
    void ui.mutedVersion;
    void load();
  });

  async function unmute(m: MutedUser) {
    busy = m.id;
    try {
      await Danmaku.unmute(m.id);
      rows = rows?.filter((r) => r.id !== m.id) ?? null;
      toast.success(`已解除 ${m.name} 的禁言`);
      log("ok", `已解除 ${m.name} 的禁言`);
    } catch (err) {
      fail("解除禁言失败", err);
    } finally {
      busy = null;
    }
  }
  function until(m: MutedUser): string {
    if (!m.until) return "永久";
    const t = dayjs(m.until);
    return t.isBefore(app.now + 86400_000) ? `${t.format("HH:mm")} 解除（${t.from(app.now)}）` : `${t.format("M月D日 HH:mm")} 解除`;
  }
</script>

<section class="card list-card">
  <header class="card-head">
    <h3>禁言列表</h3>
    <div class="card-tools">
      <button class="icon-btn" class:spinning={loading} title="刷新禁言列表" aria-label="刷新禁言列表" onclick={() => void load()}><RefreshCw class="icon" /></button>
    </div>
  </header>
  {#if error}
    <div class="empty"><p>禁言列表加载失败</p><p>{error}</p><button class="btn btn-secondary btn-sm" onclick={() => void load()}>重试</button></div>
  {:else if !rows}
    <div class="rank-list">{#each { length: 3 } as _, i (i)}<span class="skel rank-skel"></span>{/each}</div>
  {:else if rows.length === 0}
    <div class="empty"><span class="empty-icon"><ShieldBan /></span><p>没有被禁言的观众</p><p>在弹幕中点击昵称，可以从观众卡片禁言。</p></div>
  {:else}
    <ul class="rank-list" use:animateList>
      {#each rows as m (m.id)}
        <li class="rank-row muted-row">
          <Avatar name={m.name} size={28} />
          <div class="muted-who">
            <b>{m.name}</b>
            <span>{until(m)}{m.admin ? ` · 由 ${m.admin} 禁言` : ""}</span>
          </div>
          <button class="btn btn-secondary btn-sm" disabled={busy === m.id} onclick={() => void unmute(m)}>
            {#if busy === m.id}<span class="spinner"></span>{/if}解除
          </button>
        </li>
      {/each}
    </ul>
  {/if}
  <footer class="section-foot"><span>{rows ? `共 ${rows.length} 人 · ` : ""}在弹幕中点击昵称可禁言</span></footer>
</section>

<script lang="ts">
  import { Check, ChevronDown, Copy, Eye, Heart, Keyboard, LogOut, RefreshCw, Settings, ThumbsUp, Users } from "@lucide/svelte";
  import NumberFlow from "@number-flow/svelte";
  import { DropdownMenu, Tooltip } from "bits-ui";
  import Avatar from "../components/Avatar.svelte";
  import Kbd from "../components/Kbd.svelte";
  import Mark from "../components/Mark.svelte";
  import StatusPill from "../components/StatusPill.svelte";
  import { compactFormat } from "../lib/format";
  import { app, copy, feed, logout, refreshRoom, ui } from "../lib/state.svelte";

  let copied = $state(false);
  async function copyRoom() {
    const id = app.room?.roomId;
    if (id && (await copy(String(id), "房间号"))) {
      copied = true;
      setTimeout(() => (copied = false), 1400);
    }
  }
  // Online: the high-energy count from the feed when reported, else the room's.
  const online = $derived(feed.online || app.room?.online || 0);
  const stats = $derived([
    { label: "在线", hint: "高能在线人数", icon: Users, value: online },
    { label: "看过", hint: "本场看过人数", icon: Eye, value: feed.watched },
    { label: "点赞", hint: "直播间点赞数", icon: ThumbsUp, value: feed.likes },
    { label: "粉丝", hint: "粉丝数", icon: Heart, value: app.room?.attention ?? 0 },
  ]);
</script>

<header class="topbar">
  <div class="topbar-left">
    <Mark />
    <div class="topbar-room">
      <div class="room-title" title={app.room?.title ?? ""}>{app.room?.title || "我的直播间"}</div>
      <div class="room-meta">
        <button class="room-id" class:copied title="点击复制房间号" onclick={copyRoom}>
          <span class="muted">房间</span><span class="num">{app.room?.roomId ?? "—"}</span>
          <span class="copy-icons"><Copy class="icon" /><Check class="icon" /></span>
        </button>
        {#if app.room}
          <span class="sep" aria-hidden="true"></span>
          <span class="area-crumb">{app.room.parentAreaName}<span class="crumb">/</span>{app.room.areaName}</span>
        {/if}
      </div>
    </div>
  </div>

  <div class="topbar-right">
    <StatusPill />
    <Tooltip.Provider delayDuration={300}>
      <div class="stats-strip">
        {#each stats as s (s.label)}
          <Tooltip.Root>
            <Tooltip.Trigger class="stat">
              <s.icon class="icon" />
              {#if s.value}<NumberFlow value={s.value} format={compactFormat} locales="zh-CN" />{:else}<span class="num muted">—</span>{/if}
            </Tooltip.Trigger>
            <Tooltip.Portal>
              <Tooltip.Content class="tooltip" sideOffset={8}>{s.hint}</Tooltip.Content>
            </Tooltip.Portal>
          </Tooltip.Root>
        {/each}
      </div>
    </Tooltip.Provider>
    <button class="icon-btn" class:spinning={app.refreshing} title="刷新直播间信息 (R)" aria-label="刷新直播间信息" onclick={() => void refreshRoom()}>
      <RefreshCw class="icon" />
    </button>
    <DropdownMenu.Root>
      <DropdownMenu.Trigger class="user-btn">
        <Avatar src={app.user?.face} name={app.user?.name} />
        <span class="user-name">{app.user?.name ?? "已登录"}</span>
        <ChevronDown class="icon chev" />
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content class="menu-content" align="end" sideOffset={8}>
          <div class="menu-head">
            <b>{app.user?.name ?? ""}</b>
            <span class="muted num">{app.user ? `UID ${app.user.uid}` : ""}</span>
          </div>
          <DropdownMenu.Item class="menu-item" onSelect={() => (ui.prefsOpen = true)}><Settings class="icon" />偏好设置<Kbd keys={["P"]} /></DropdownMenu.Item>
          <DropdownMenu.Item class="menu-item" onSelect={() => (ui.aboutOpen = true)}><Keyboard class="icon" />关于与快捷键<Kbd keys={["?"]} /></DropdownMenu.Item>
          <DropdownMenu.Separator class="menu-sep" />
          <DropdownMenu.Item class="menu-item danger" onSelect={() => void logout()}><LogOut class="icon" />退出登录</DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  </div>
</header>

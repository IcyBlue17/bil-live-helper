<script lang="ts">
  import { ChevronsUpDown, CircleHelp, LogOut, RefreshCw, Settings } from "@lucide/svelte";
  import { DropdownMenu } from "bits-ui";
  import Avatar from "../components/Avatar.svelte";
  import Kbd from "../components/Kbd.svelte";
  import Mark from "../components/Mark.svelte";
  import StatusPill from "../components/StatusPill.svelte";
  import { app, logout, refreshRoom, ui } from "../lib/state.svelte";
</script>

<header class="topbar">
  <nav class="crumbs" aria-label="位置">
    <Mark size={24} />
    <span class="crumb-sep" aria-hidden="true"></span>
    <DropdownMenu.Root>
      <DropdownMenu.Trigger class="crumb-btn">
        <Avatar src={app.user?.face} name={app.user?.name} size={20} />
        <span class="crumb-text">{app.user?.name ?? "已登录"}</span>
        <ChevronsUpDown class="icon chev" />
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content class="menu-content" align="start" sideOffset={8}>
          <div class="menu-head">
            <b>{app.user?.name ?? ""}</b>
            <span>{app.user ? `UID ${app.user.uid}` : ""}</span>
          </div>
          <DropdownMenu.Item class="menu-item" onSelect={() => (ui.prefsOpen = true)}><Settings class="icon" />偏好设置<Kbd keys={["P"]} /></DropdownMenu.Item>
          <DropdownMenu.Item class="menu-item" onSelect={() => (ui.aboutOpen = true)}><CircleHelp class="icon" />关于与快捷键<Kbd keys={["?"]} /></DropdownMenu.Item>
          <DropdownMenu.Separator class="menu-sep" />
          <DropdownMenu.Item class="menu-item danger" onSelect={() => void logout()}><LogOut class="icon" />退出登录</DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
    <span class="crumb-sep" aria-hidden="true"></span>
    <span class="crumb-room" title={app.room?.title ?? ""}>{app.room?.title || "我的直播间"}</span>
    <StatusPill />
  </nav>

  <div class="topbar-right">
    <button class="btn btn-secondary btn-sm" class:spinning={app.refreshing} title="刷新直播间信息 (R)" onclick={() => void refreshRoom()}>
      <RefreshCw class="icon" />刷新
    </button>
    <button class="icon-btn" title="偏好设置 (P)" aria-label="偏好设置" onclick={() => (ui.prefsOpen = true)}><Settings class="icon" /></button>
    <button class="icon-btn" title="关于与快捷键 (?)" aria-label="关于与快捷键" onclick={() => (ui.aboutOpen = true)}><CircleHelp class="icon" /></button>
  </div>
</header>

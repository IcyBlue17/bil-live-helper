<script lang="ts">
  import { CircleAlert, SendHorizontal, Smile } from "@lucide/svelte";
  import { Popover, Tabs } from "bits-ui";
  import { Danmaku, type Emoticon, type EmoticonPackage } from "../mygo";
  import { charCount, errorMessage } from "../lib/format";
  import { app, sendDanmaku, sendSticker } from "../lib/state.svelte";

  const MAX = 40;
  let value = $state("");
  let sending = $state(false);
  let error = $state("");
  let input: HTMLInputElement | undefined = $state();
  const count = $derived(charCount(value));
  const canSend = $derived(value.trim() !== "" && count <= MAX && !sending && app.phase === "app");

  async function run(fn: () => Promise<void>) {
    sending = true;
    error = "";
    try {
      await fn();
      return true;
    } catch (err) {
      error = `发送失败：${errorMessage(err)}`;
      return false;
    } finally {
      sending = false;
      input?.focus();
    }
  }
  async function send() {
    if (!canSend) {
      if (value.trim() === "") error = "弹幕内容不能为空";
      return;
    }
    const text = value.trim();
    if (await run(() => sendDanmaku(text))) value = "";
  }

  // ---- Emote picker ----
  let pickerOpen = $state(false);
  let packages = $state.raw<EmoticonPackage[] | null>(null);
  let loadError = $state("");
  let tab = $state("");
  $effect(() => {
    if (!pickerOpen || packages) return;
    Danmaku.emoticons()
      .then((p) => {
        packages = p;
        tab = String(p[0]?.id ?? "");
      })
      .catch((err: unknown) => (loadError = errorMessage(err)));
  });
  /** Inline emotes ("[dog]") go into the text; stickers are sent at once. */
  const isInline = (e: Emoticon) => e.token.startsWith("[");
  async function pick(e: Emoticon) {
    if (isInline(e)) {
      const el = input;
      const at = el?.selectionStart ?? value.length;
      value = value.slice(0, at) + e.token + value.slice(el?.selectionEnd ?? at);
      requestAnimationFrame(() => {
        el?.focus();
        el?.setSelectionRange(at + e.token.length, at + e.token.length);
      });
    } else {
      pickerOpen = false;
      await run(() => sendSticker(e));
    }
  }
</script>

<footer class="send-box">
  <div class="send-row">
    <div class="send-field">
      <Popover.Root bind:open={pickerOpen}>
        <Popover.Trigger class="icon-btn emote-btn" aria-label="表情" title="表情"><Smile class="icon" /></Popover.Trigger>
        <Popover.Portal>
          <Popover.Content class="popover emote-picker" side="top" align="start" sideOffset={8} onOpenAutoFocus={(e) => e.preventDefault()}>
            {#if loadError}
              <div class="empty empty-sm"><p>表情加载失败</p><p>{loadError}</p></div>
            {:else if !packages}
              <div class="empty empty-sm"><span class="spinner"></span></div>
            {:else if packages.length === 0}
              <div class="empty empty-sm"><p>没有可用的表情</p></div>
            {:else}
              <Tabs.Root bind:value={tab}>
                <Tabs.List class="emote-tabs">
                  {#each packages as p (p.id)}<Tabs.Trigger value={String(p.id)} class="emote-tab">{p.name}</Tabs.Trigger>{/each}
                </Tabs.List>
                {#each packages as p (p.id)}
                  <Tabs.Content value={String(p.id)}>
                    <div class="emote-grid" class:stickers={!isInline(p.items[0] ?? { token: "[" } as Emoticon)}>
                      {#each p.items as e (e.token)}
                        <button class="emote-cell" title={isInline(e) ? `${e.name}（插入）` : `${e.name}（直接发送）`} onclick={() => void pick(e)}>
                          <img src={e.url} alt={e.name} referrerpolicy="no-referrer" />
                          {#if !isInline(e)}<span>{e.name}</span>{/if}
                        </button>
                      {/each}
                    </div>
                  </Tabs.Content>
                {/each}
              </Tabs.Root>
              <p class="picker-hint">小表情插入到输入框，大表情点击即发送</p>
            {/if}
          </Popover.Content>
        </Popover.Portal>
      </Popover.Root>
      <input
        bind:this={input}
        class="input send-input"
        class:invalid={count > MAX}
        id="danmaku-input"
        type="text"
        placeholder="发送弹幕…"
        autocomplete="off"
        aria-label="弹幕内容"
        bind:value
        oninput={() => (error = "")}
        onkeydown={(e) => {
          if (e.key === "Enter" && !e.isComposing) {
            e.preventDefault();
            void send();
          }
        }}
      />
      <span class="counter num" class:near={count > MAX - 5 && count <= MAX} class:over={count > MAX}>{count}/{MAX}</span>
    </div>
    <button class="btn btn-primary send-btn" disabled={!canSend} onclick={() => void send()} title="发送 (Enter)" aria-label="发送弹幕">
      {#if sending}<span class="spinner"></span>{:else}<SendHorizontal class="icon" />{/if}
    </button>
  </div>
  <p class="send-error" class:show={!!error} role="alert"><CircleAlert class="icon" />{error}</p>
</footer>

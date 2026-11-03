import { effect, onCleanup, root } from "./signal";

type Child = Node | string | number | null | undefined | false | Child[];
type Reactive<T> = T | (() => T);
type Props = {
  class?: Reactive<string>;
  style?: string;
  [key: `on${string}`]: unknown;
  [key: string]: unknown;
};

/** Creates an element. Function-valued props (except on*) are reactive. */
export function h<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  props?: Props | null,
  ...children: Child[]
): HTMLElementTagNameMap[K] {
  const el = document.createElement(tag);
  if (props) {
    for (const [key, value] of Object.entries(props)) {
      if (value === undefined || value === null) continue;
      if (key.startsWith("on") && typeof value === "function") {
        el.addEventListener(key.slice(2).toLowerCase(), value as EventListener);
      } else if (key === "ref" && typeof value === "function") {
        (value as (el: HTMLElement) => void)(el);
      } else if (typeof value === "function") {
        effect(() => setProp(el, key, (value as () => unknown)()));
      } else {
        setProp(el, key, value);
      }
    }
  }
  append(el, children);
  return el;
}

function setProp(el: HTMLElement, key: string, value: unknown) {
  if (key === "class") el.className = String(value ?? "");
  else if (key === "text") el.textContent = String(value ?? "");
  else if (key === "value" || key === "checked" || key === "disabled" || key === "hidden") {
    (el as unknown as Record<string, unknown>)[key] = value;
  } else if (value === false || value === null || value === undefined) el.removeAttribute(key);
  else el.setAttribute(key, value === true ? "" : String(value));
}

export function append(el: Node, children: Child[]) {
  for (const c of children) {
    if (c === null || c === undefined || c === false) continue;
    if (Array.isArray(c)) append(el, c);
    else el.appendChild(typeof c === "object" ? c : document.createTextNode(String(c)));
  }
}

/** A text node bound to a signal. */
export function text(fn: () => string | number): Text {
  const node = document.createTextNode("");
  effect(() => {
    node.data = String(fn());
  });
  return node;
}

/** Parses trusted markup (our own icons, QR SVGs from Go) into a node. */
export function html(markup: string): Element {
  const t = document.createElement("template");
  t.innerHTML = markup.trim();
  return t.content.firstElementChild ?? document.createElement("span");
}

/**
 * Renders the result of view(key) into a container, re-rendering when key
 * changes. The previous view fades out (unless reduced motion) and its
 * effects are disposed.
 */
export function swap<K>(
  host: HTMLElement,
  key: () => K,
  view: (k: K) => Node,
  opts: { animate?: boolean } = {},
) {
  let dispose: (() => void) | null = null;
  let current: Node | null = null;
  effect(() => {
    const k = key();
    const prevDispose = dispose;
    const prevNode = current;
    const [node, d] = root(() => view(k));
    dispose = d;
    current = node;
    prevDispose?.();
    if (prevNode && opts.animate && prevNode instanceof HTMLElement && !reducedMotion()) {
      prevNode.classList.add("view-leave");
      prevNode.style.pointerEvents = "none";
      setTimeout(() => prevNode.parentNode && prevNode.remove(), 200);
    } else if (prevNode) prevNode.parentNode?.removeChild(prevNode);
    if (node instanceof HTMLElement && opts.animate) node.classList.add("view-enter");
    host.appendChild(node);
  });
  onCleanup(() => {
    dispose?.();
    current?.parentNode?.removeChild(current);
  });
}

export function reducedMotion(): boolean {
  return matchMedia("(prefers-reduced-motion: reduce)").matches;
}

/** Listens on a target for the current scope's lifetime. */
export function listen<K extends keyof WindowEventMap>(
  target: Window,
  type: K,
  fn: (e: WindowEventMap[K]) => void,
): void;
export function listen(target: EventTarget, type: string, fn: (e: Event) => void): void;
export function listen(target: EventTarget, type: string, fn: (e: never) => void) {
  target.addEventListener(type, fn as EventListener);
  onCleanup(() => target.removeEventListener(type, fn as EventListener));
}

/** setInterval for the current scope's lifetime. */
export function every(ms: number, fn: () => void) {
  const id = setInterval(fn, ms);
  onCleanup(() => clearInterval(id));
}

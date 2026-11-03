import { cubicOut } from "svelte/easing";
import { fade, fly, type TransitionConfig } from "svelte/transition";

export const reduced = () => matchMedia("(prefers-reduced-motion: reduce)").matches;
const d = (ms: number) => (reduced() ? 0 : ms);

export const fadeIn = (node: Element, ms = 200): TransitionConfig => fade(node, { duration: d(ms), easing: cubicOut });
/** A short rise, for views, cards and popovers. */
export const rise = (node: Element, { y = 6, ms = 240 } = {}): TransitionConfig => fly(node, { y, duration: d(ms), easing: cubicOut });
/** Modal entrance: rise plus a hint of scale. */
export function pop(node: Element, { ms = 240 } = {}): TransitionConfig {
  const base = getComputedStyle(node).transform;
  const t0 = base === "none" ? "" : base;
  return {
    duration: d(ms),
    easing: cubicOut,
    css: (t, u) => `opacity:${t};transform:${t0} translateY(${u * 8}px) scale(${0.985 + 0.015 * t})`,
  };
}

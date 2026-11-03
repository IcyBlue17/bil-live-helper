import autoAnimate from "@formkit/auto-animate";
import type { Action } from "svelte/action";

/** Animates children being added, removed and moved (FLIP), via auto-animate. */
export const animateList: Action<HTMLElement> = (node) => {
  const ctl = autoAnimate(node, { duration: 200, easing: "cubic-bezier(.2,.8,.2,1)" });
  return { destroy: () => ctl.destroy?.() };
};

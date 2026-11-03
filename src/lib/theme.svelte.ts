// Colour theme: dark (default), light, or follow the system. Stored in
// localStorage only; the backend does not know about it.
export type ThemePref = "system" | "dark" | "light";
export type Theme = "dark" | "light";

const KEY = "blh-theme";
const query = matchMedia("(prefers-color-scheme: light)");

function stored(): ThemePref {
  try {
    const v = localStorage.getItem(KEY);
    if (v === "system" || v === "dark" || v === "light") return v;
  } catch {
    // Storage unavailable: use the default.
  }
  return "system";
}

class ThemeStore {
  pref = $state<ThemePref>(stored());
  system = $state<Theme>(query.matches ? "light" : "dark");
  resolved = $derived<Theme>(this.pref === "system" ? this.system : this.pref);
}
export const theme = new ThemeStore();
query.addEventListener("change", (e) => (theme.system = e.matches ? "light" : "dark"));

export function setTheme(pref: ThemePref) {
  theme.pref = pref;
  try {
    localStorage.setItem(KEY, pref);
  } catch {
    // Not persisted; still applies for this session.
  }
}

$effect.root(() => {
  $effect(() => {
    const root = document.documentElement;
    root.dataset.theme = theme.resolved;
    root.style.colorScheme = theme.resolved;
  });
});

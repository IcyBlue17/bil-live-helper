// Inline 24×24 stroke icons (currentColor), so they inherit text colour.
const svg = (body: string, extra = "") =>
  `<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" ${extra}>${body}</svg>`;

export const icons = {
  check: svg(`<path d="M5 12.5l4.5 4.5L19 7.5"/>`),
  copy: svg(`<rect x="8.5" y="8.5" width="11" height="11" rx="2.5"/><path d="M15.5 8.5V6.5a2 2 0 0 0-2-2h-7a2 2 0 0 0-2 2v7a2 2 0 0 0 2 2h2"/>`),
  eye: svg(`<path d="M2.5 12S6 5.5 12 5.5 21.5 12 21.5 12 18 18.5 12 18.5 2.5 12 2.5 12z"/><circle cx="12" cy="12" r="3"/>`),
  eyeOff: svg(`<path d="M10.6 5.6A9.7 9.7 0 0 1 12 5.5c6 0 9.5 6.5 9.5 6.5a17 17 0 0 1-2.6 3.4M6.6 6.6C3.9 8.3 2.5 12 2.5 12S6 18.5 12 18.5c1.9 0 3.5-.6 4.9-1.5"/><path d="M9.9 9.9a3 3 0 0 0 4.2 4.2"/><path d="M3 3l18 18"/>`),
  alert: svg(`<path d="M12 3.5l9.5 16.5h-19z"/><path d="M12 10v4.5"/><path d="M12 17.5v.01"/>`),
  info: svg(`<circle cx="12" cy="12" r="9"/><path d="M12 11v5.5"/><path d="M12 7.5v.01"/>`),
  refresh: svg(`<path d="M20 11.5A8 8 0 0 0 6.3 6.3L4 8.5"/><path d="M4 4v4.5h4.5"/><path d="M4 12.5a8 8 0 0 0 13.7 5.2L20 15.5"/><path d="M20 20v-4.5h-4.5"/>`),
  search: svg(`<circle cx="11" cy="11" r="6.5"/><path d="M16 16l4.5 4.5"/>`),
  close: svg(`<path d="M6 6l12 12M18 6L6 18"/>`),
  chevronDown: svg(`<path d="M6 9.5l6 6 6-6"/>`),
  chevronRight: svg(`<path d="M9.5 6l6 6-6 6"/>`),
  arrowDown: svg(`<path d="M12 5v14M6 13l6 6 6-6"/>`),
  send: svg(`<path d="M4 12l16-8-6 16-2.5-6.5z"/><path d="M11.5 13.5L20 4"/>`),
  play: svg(`<path d="M8 5.5v13l10.5-6.5z" fill="currentColor" stroke="none"/>`),
  stop: svg(`<rect x="6.5" y="6.5" width="11" height="11" rx="2" fill="currentColor" stroke="none"/>`),
  phone: svg(`<rect x="6.5" y="2.5" width="11" height="19" rx="2.5"/><path d="M9.5 12.5l2 2 3.5-4"/>`),
  user: svg(`<circle cx="12" cy="8.5" r="4"/><path d="M4.5 20.5a7.5 7.5 0 0 1 15 0"/>`),
  logout: svg(`<path d="M14 4.5h3.5a2 2 0 0 1 2 2v11a2 2 0 0 1-2 2H14"/><path d="M10 16.5L5.5 12 10 7.5"/><path d="M5.5 12H15"/>`),
  keyboard: svg(`<rect x="2.5" y="6" width="19" height="12" rx="2.5"/><path d="M6.5 10h.01M10 10h.01M14 10h.01M17.5 10h.01M8 14h8"/>`),
  terminal: svg(`<path d="M5 8l4 4-4 4"/><path d="M12 16h7"/>`),
  server: svg(`<rect x="3.5" y="4" width="17" height="6.5" rx="2"/><rect x="3.5" y="13.5" width="17" height="6.5" rx="2"/><path d="M7 7.25h.01M7 16.75h.01"/>`),
  key: svg(`<circle cx="8" cy="15" r="4"/><path d="M11 12l8.5-8.5M16 7l2.5 2.5M14 9l1.5 1.5"/>`),
  edit: svg(`<path d="M4 20h4L19.5 8.5a2.8 2.8 0 0 0-4-4L4 16z"/><path d="M13.5 6.5l4 4"/>`),
  grid: svg(`<rect x="4" y="4" width="6.5" height="6.5" rx="1.5"/><rect x="13.5" y="4" width="6.5" height="6.5" rx="1.5"/><rect x="4" y="13.5" width="6.5" height="6.5" rx="1.5"/><rect x="13.5" y="13.5" width="6.5" height="6.5" rx="1.5"/>`),
  chat: svg(`<path d="M4.5 18.5V7a2.5 2.5 0 0 1 2.5-2.5h10A2.5 2.5 0 0 1 19.5 7v7a2.5 2.5 0 0 1-2.5 2.5H8.5z"/><path d="M8.5 9.5h7M8.5 12.5h4.5"/>`),
  users: svg(`<circle cx="9" cy="9" r="3.5"/><path d="M2.5 19.5a6.5 6.5 0 0 1 13 0"/><path d="M15.5 5.6a3.5 3.5 0 0 1 0 6.8M18 14a6.5 6.5 0 0 1 3.5 5.5"/>`),
  flame: svg(`<path d="M12 21a6.5 6.5 0 0 0 6.5-6.5c0-4-3-6-4-9.5-1.5 2-2 3.5-2 5-1.5-1-2.5-2.5-2.5-4C7 8 5.5 11 5.5 14.5A6.5 6.5 0 0 0 12 21z"/>`),
  heart: svg(`<path d="M12 20s-7.5-4.6-7.5-10.2A4.3 4.3 0 0 1 12 7.3a4.3 4.3 0 0 1 7.5 2.5C19.5 15.4 12 20 12 20z"/>`),
  gift: svg(`<rect x="3.5" y="8.5" width="17" height="4" rx="1"/><path d="M5 12.5v7.5h14v-7.5M12 8.5V20"/><path d="M12 8.5S10.5 4 8 4.5 7 8.5 12 8.5zM12 8.5S13.5 4 16 4.5 17 8.5 12 8.5z"/>`),
  external: svg(`<path d="M14 4.5h5.5V10"/><path d="M19.5 4.5L11 13"/><path d="M17.5 14v4a2 2 0 0 1-2 2h-9a2 2 0 0 1-2-2v-9a2 2 0 0 1 2-2h4"/>`),
  dot: svg(`<circle cx="12" cy="12" r="4" fill="currentColor" stroke="none"/>`),
};

/** The app mark: a rounded tile with a stylised TV and live dot. */
export const mark = `<svg class="mark" viewBox="0 0 32 32" aria-hidden="true">
  <defs><linearGradient id="mk" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#ff8db0"/><stop offset="1" stop-color="#f25d8e"/></linearGradient></defs>
  <rect width="32" height="32" rx="9" fill="url(#mk)"/>
  <path d="M11 7.5l3 3M21 7.5l-3 3" stroke="#fff" stroke-width="2" stroke-linecap="round"/>
  <rect x="6.5" y="10.5" width="19" height="14" rx="4" fill="none" stroke="#fff" stroke-width="2"/>
  <path d="M14 15v4.5l4-2.25z" fill="#fff"/>
</svg>`;

<script lang="ts">
  let { size = 32 }: { size?: number } = $props();
</script>

<svg class="mark" viewBox="0 0 32 32" aria-hidden="true" style="width:{size}px;height:{size}px">
  <defs><linearGradient id="mk" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#ff8db0" /><stop offset="1" stop-color="#f25d8e" /></linearGradient></defs>
  <rect width="32" height="32" rx="9" fill="url(#mk)" />
  <path d="M11 7.5l3 3M21 7.5l-3 3" stroke="#fff" stroke-width="2" stroke-linecap="round" />
  <rect x="6.5" y="10.5" width="19" height="14" rx="4" fill="none" stroke="#fff" stroke-width="2" />
  <path d="M14 15v4.5l4-2.25z" fill="#fff" />
</svg>

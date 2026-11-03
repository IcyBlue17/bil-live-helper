<script lang="ts">
  // The app mark: a monochrome TV glyph that inverts with the theme.
  let { size = 32 }: { size?: number } = $props();
</script>

<svg class="mark" viewBox="0 0 32 32" aria-hidden="true" style="width:{size}px;height:{size}px">
  <rect class="mark-bg" width="32" height="32" rx="8" />
  <g class="mark-fg">
    <path d="M11.5 7.5l3 3M20.5 7.5l-3 3" stroke-width="2" stroke-linecap="round" fill="none" />
    <rect x="6.5" y="10.5" width="19" height="14" rx="3.5" stroke-width="2" fill="none" />
    <path d="M14 14.6v5.3l4.6-2.65z" stroke="none" />
  </g>
</svg>

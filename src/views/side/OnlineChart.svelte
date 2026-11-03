<script lang="ts">
  // Online count over time, in the Vercel Analytics style: one blue 2px line
  // over a faint fill, dashed horizontal grid, mono axis labels, a crosshair
  // and a tooltip. Colours come from the theme's CSS variables.
  import uPlot from "uplot";
  import "uplot/dist/uPlot.min.css";
  import type { OnlineSample } from "../../mygo";
  import { dayjs } from "../../lib/format";
  import { theme } from "../../lib/theme.svelte";

  let { samples }: { samples: OnlineSample[] } = $props();

  let box: HTMLDivElement | undefined = $state();
  let tip = $state<{ x: number; y: number; time: string; online: number } | null>(null);
  let plot: uPlot | null = null;

  const data = $derived<uPlot.AlignedData>([
    samples.map((p) => dayjs(p.time).valueOf() / 1000),
    samples.map((p) => p.online),
  ]);

  const fmt = new Intl.NumberFormat("zh-CN", { notation: "compact", maximumFractionDigits: 1 });

  $effect(() => {
    const el = box;
    void theme.resolved;
    if (!el) return;
    const css = getComputedStyle(el);
    const v = (name: string) => css.getPropertyValue(name).trim();
    const line = v("--blue");
    const axis = v("--text-3");
    const grid = v("--line-2");
    const surface = v("--surface");
    const font = `11px ${v("--mono")}`;
    const opts: uPlot.Options = {
      width: el.clientWidth,
      height: el.clientHeight,
      padding: [12, 8, 0, 0],
      legend: { show: false },
      cursor: {
        y: false,
        points: { size: 9, width: 2, stroke: line, fill: surface },
        drag: { x: false, y: false },
      },
      scales: { x: { time: true }, y: { range: (_u, _min, max) => [0, Math.max(10, max * 1.15)] } },
      axes: [
        {
          stroke: axis,
          font,
          size: 28,
          gap: 8,
          grid: { show: false },
          ticks: { show: false },
          values: (_u, ts) => ts.map((t) => dayjs.unix(t).format("HH:mm")),
        },
        {
          stroke: axis,
          font,
          size: 44,
          gap: 8,
          grid: { stroke: grid, width: 1, dash: [3, 3] },
          ticks: { show: false },
          values: (_u, vs) => vs.map((x) => fmt.format(x)),
        },
      ],
      series: [
        {},
        {
          stroke: line,
          width: 2,
          fill: (u) => {
            const g = u.ctx.createLinearGradient(0, u.bbox.top, 0, u.bbox.top + u.bbox.height);
            g.addColorStop(0, "rgba(0,112,243,0.18)");
            g.addColorStop(1, "rgba(0,112,243,0)");
            return g;
          },
          points: { show: false },
          spanGaps: true,
        },
      ],
      hooks: {
        setCursor: [
          (u) => {
            const i = u.cursor.idx;
            const xs = u.data[0];
            const ys = u.data[1];
            if (i == null || u.cursor.left == null || u.cursor.left < 0 || !ys) return void (tip = null);
            const y = ys[i];
            const t = xs[i];
            if (y == null || t == null) return void (tip = null);
            tip = { x: u.valToPos(t, "x"), y: u.valToPos(y, "y"), time: dayjs.unix(t).format("HH:mm"), online: y };
          },
        ],
      },
    };
    const p = new uPlot(opts, data, el);
    plot = p;
    const ro = new ResizeObserver(() => p.setSize({ width: el.clientWidth, height: el.clientHeight }));
    ro.observe(el);
    return () => {
      ro.disconnect();
      p.destroy();
      plot = null;
    };
  });
  // New samples: update in place instead of rebuilding.
  $effect(() => {
    plot?.setData(data);
  });
</script>

<div class="chart" role="img" aria-label="在线人数折线图，共 {samples.length} 个采样点">
  <div class="chart-plot" bind:this={box}></div>
  {#if samples.length < 2}
    <div class="chart-empty">数据不足，每分钟采样一次</div>
  {/if}
  {#if tip}
    <div class="chart-tip" style="left:{tip.x + 44}px;top:{tip.y + 12}px">
      <span class="chart-tip-time mono">{tip.time}</span>
      <span class="chart-tip-row"><span class="chart-tip-swatch"></span>在线<b class="num">{tip.online.toLocaleString("zh-CN")}</b></span>
    </div>
  {/if}
</div>

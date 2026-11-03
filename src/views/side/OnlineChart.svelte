<script lang="ts">
  // Online count over time: one pink 2px line, recessive grid, crosshair + tooltip.
  import uPlot from "uplot";
  import "uplot/dist/uPlot.min.css";
  import type { OnlineSample } from "../../mygo";
  import { dayjs } from "../../lib/format";

  let { samples }: { samples: OnlineSample[] } = $props();

  let box: HTMLDivElement | undefined = $state();
  let tip = $state<{ x: number; y: number; time: string; online: number } | null>(null);
  let plot: uPlot | null = null;

  const data = $derived<uPlot.AlignedData>([
    samples.map((p) => dayjs(p.time).valueOf() / 1000),
    samples.map((p) => p.online),
  ]);

  const axisFont = "11px system-ui, sans-serif";
  const fmt = new Intl.NumberFormat("zh-CN", { notation: "compact", maximumFractionDigits: 1 });

  $effect(() => {
    const el = box;
    if (!el) return;
    const opts: uPlot.Options = {
      width: el.clientWidth,
      height: el.clientHeight,
      padding: [8, 4, 0, 0],
      legend: { show: false },
      cursor: {
        y: false,
        points: { size: 8, width: 2, stroke: "#fb7299", fill: "#15171c" },
        drag: { x: false, y: false },
      },
      scales: { x: { time: true }, y: { range: (_u, _min, max) => [0, Math.max(10, max * 1.15)] } },
      axes: [
        {
          stroke: "#6b7280",
          font: axisFont,
          size: 24,
          gap: 4,
          grid: { show: false },
          ticks: { show: false },
          values: (_u, ts) => ts.map((t) => dayjs.unix(t).format("HH:mm")),
        },
        {
          stroke: "#6b7280",
          font: axisFont,
          size: 40,
          gap: 6,
          grid: { stroke: "rgba(255,255,255,0.06)", width: 1 },
          ticks: { show: false },
          values: (_u, vs) => vs.map((v) => fmt.format(v)),
        },
      ],
      series: [
        {},
        {
          stroke: "#fb7299",
          width: 2,
          fill: (u) => {
            const g = u.ctx.createLinearGradient(0, u.bbox.top, 0, u.bbox.top + u.bbox.height);
            g.addColorStop(0, "rgba(251,114,153,0.22)");
            g.addColorStop(1, "rgba(251,114,153,0)");
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
            const v = ys[i];
            const t = xs[i];
            if (v == null || t == null) return void (tip = null);
            tip = { x: u.valToPos(t, "x"), y: u.valToPos(v, "y"), time: dayjs.unix(t).format("HH:mm"), online: v };
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
    <div class="chart-empty muted">数据不足，每分钟采样一次</div>
  {/if}
  {#if tip}
    <div class="chart-tip" style="left:{tip.x + 40}px;top:{tip.y + 8}px">
      <span class="muted num">{tip.time}</span><b class="num">{tip.online.toLocaleString("zh-CN")}</b>
    </div>
  {/if}
</div>

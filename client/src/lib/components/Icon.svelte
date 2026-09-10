<script lang="ts">
  // Inline SVG icon from the shared stroke set (lib/icons.ts). Sized in
  // px, coloured by currentColor, decorative by default — pass `label`
  // when the icon is the only content of a control so screen readers
  // get a name.
  import { ICONS, type IconName } from "../icons";

  interface Props {
    name: IconName;
    size?: number;
    strokeWidth?: number;
    label?: string;
  }
  const { name, size = 16, strokeWidth = 1.75, label }: Props = $props();
  const prims = $derived(ICONS[name]);
</script>

<svg
  class="icon"
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width={strokeWidth}
  stroke-linecap="round"
  stroke-linejoin="round"
  role={label ? "img" : undefined}
  aria-label={label}
  aria-hidden={label ? undefined : "true"}
>
  {#each prims as el, i (i)}
    {#if el.t === "path"}
      <path d={el.d} fill={el.fill ? "currentColor" : "none"} stroke={el.fill ? "none" : undefined}
      ></path>
    {:else if el.t === "circle"}
      <circle
        cx={el.cx}
        cy={el.cy}
        r={el.r}
        fill={el.fill ? "currentColor" : "none"}
        stroke={el.fill ? "none" : undefined}
      ></circle>
    {:else}
      <rect
        x={el.x}
        y={el.y}
        width={el.w}
        height={el.h}
        rx={el.rx ?? 0}
        fill={el.fill ? "currentColor" : "none"}
        stroke={el.fill ? "none" : undefined}
      ></rect>
    {/if}
  {/each}
</svg>

<style>
  .icon {
    display: inline-block;
    vertical-align: -0.15em;
    flex: 0 0 auto;
  }
</style>

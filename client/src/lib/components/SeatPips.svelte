<script lang="ts">
  // SeatPips — the four seat colours as four small squares: C2's mark on
  // site pages, beside a page's eyebrow. Four seats, one table. Decorative.
  import { seatColor } from "../colors";
</script>

<span class="pips" aria-hidden="true">
  {#each [0, 1, 2, 3] as seat (seat)}
    <span class="pip" style="--seat: {seatColor(seat)}"></span>
  {/each}
</span>

<style>
  .pips {
    display: inline-flex;
    gap: 3px;
    vertical-align: middle;
  }
  .pip {
    width: 6px;
    height: 6px;
    border-radius: 1.5px;
    background: var(--seat);
  }
</style>

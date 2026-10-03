<script lang="ts">
  // CoachedBoardStub stands in for Board.svelte in
  // tutorialLayout.render.test.ts (#1079): it reports the one prop that
  // test is about, `coached`, and ignores the rest. Test-only; nothing in
  // the app imports it.
  const { coached = false }: { coached?: boolean } = $props();
</script>

<div data-testid="board-stub" data-coached={coached ? "yes" : "no"}></div>

<script lang="ts">
  // BoardStub stands in for Board.svelte in boardBoundary.render.test.ts
  // (#720). It throws during component initialisation when the test
  // asks it to, which is what `<svelte:boundary>` is meant to catch,
  // and otherwise renders a marker so the test can tell the board came
  // back after a reset.
  //
  // Props are ignored on purpose: Game.svelte decides what Board gets,
  // and the test's business is the boundary, not the prop contract.
  import { boardStub, BOARD_STUB_ERROR } from "./boardStub";

  if (boardStub.throwOnNextRender) {
    boardStub.throwOnNextRender = false;
    boardStub.throws++;
    throw new Error(BOARD_STUB_ERROR);
  }
  boardStub.renders++;
</script>

<div data-testid="board-stub">the board</div>

// boardStub.ts — the switch behind BoardStub.svelte (#720).
//
// A render test needs the board to throw on demand, and no prop on the
// real Board.svelte does that. So the test mocks Board.svelte with
// BoardStub.svelte and flips this flag: module state rather than a
// prop, because Game.svelte owns the props it passes to Board and the
// test never sees them.
//
// Test-only. Nothing in the app imports it; the production bundle
// never reaches it.
export const boardStub = {
  /** Set by the test; the next render throws and clears it. */
  throwOnNextRender: false,
  /** How many times the stub has rendered without throwing. */
  renders: 0,
  /** How many times it has thrown. */
  throws: 0,
};

export function resetBoardStub(): void {
  boardStub.throwOnNextRender = false;
  boardStub.renders = 0;
  boardStub.throws = 0;
}

/** The message BoardStub throws, so the test can match on it. */
export const BOARD_STUB_ERROR = "board blew up while rendering";

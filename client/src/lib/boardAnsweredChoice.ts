// boardAnsweredChoice.ts — which pending-choice kinds are answered by
// clicking the BOARD rather than in a modal.
//
// One list, read by both halves of the decision. Board.svelte hands a
// matching choice to the targeting flow (the banner and the clickable
// permanents); ChoicePromptModal must NOT open for it. They used to
// keep separate lists: the board knew `legend_rule` and
// `choose_protector`, the modal skipped only `pick_target` and
// `retarget`, so a legend-rule prompt opened BOTH — and the modal's
// fallback branch rendered it as "Pick 1 card from opponent's revealed
// hand. opponent will discard your pick", under a full-screen backdrop
// that covered the board the real answer had to be clicked on (#1623).
//
// A new kind that is answered on the board goes here and nowhere else.

const BOARD_ANSWERED = new Set<string>([
  // S20: a triggered ability's target, picked by clicking the board.
  "pick_target",
  // S27: the legend rule (CR 704.5j) — pick the permanent to keep.
  "legend_rule",
  // S27: a battle's protector (CR 310.9) — pick the player.
  "choose_protector",
  // #1196: the CR 115.7 retarget, answered on the board too.
  "retarget",
]);

export function isBoardAnsweredChoice(kind: string): boolean {
  return BOARD_ANSWERED.has(kind);
}

<script lang="ts">
  // BoardGroupAttackStub stands in for Board.svelte in Game-level tests
  // of declaring attackers (#4423): a button per creature that asks Game
  // to select it for combat (what a click on it does on the real board),
  // and one per token group that sends the group's members at a seat
  // (what the group list's "Attack <seat> with N" does).
  let {
    onSelectCombatCard,
    onDeclareAttackers,
    onUnstageAttackers,
    stagedAttackers,
  }: {
    onSelectCombatCard?: (cardID: string) => void;
    onDeclareAttackers?: (attackerIDs: string[], defenderSeatID: string) => void;
    onUnstageAttackers?: (attackerIDs: string[]) => void;
    stagedAttackers?: ReadonlySet<string>;
  } = $props();
</script>

<div data-testid="board-stub">
  <button type="button" data-testid="select-b1" onclick={() => onSelectCombatCard?.("b1")}
    >select b1</button
  >
  <button
    type="button"
    data-testid="group-attack"
    onclick={() => onDeclareAttackers?.(["w1", "w2", "w3"], "p-2")}>group attack</button
  >
  <button type="button" data-testid="group-unstage" onclick={() => onUnstageAttackers?.(["w3"])}
    >group unstage</button
  >
  <span data-testid="staged">{[...(stagedAttackers ?? [])].sort().join(",")}</span>
</div>

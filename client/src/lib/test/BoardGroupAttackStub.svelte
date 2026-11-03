<script lang="ts">
  // BoardGroupAttackStub stands in for Board.svelte in Game-level tests
  // of declaring attackers (#4423): a button per creature that asks Game
  // to select it for combat (what a click on it does on the real board),
  // and one per token group that sends the group's members at a seat
  // (what the group list's "Attack <seat> with N" does). #4326: and a
  // button per seat, what a click on that player does while an attacker
  // is selected.
  let {
    onSelectCombatCard,
    onDeclareAttack,
    onDeclareAttackers,
    onUnstageAttackers,
    stagedAttackers,
    selectedCombatCardID,
  }: {
    onSelectCombatCard?: (cardID: string) => void;
    onDeclareAttack?: (targetID: string) => void;
    selectedCombatCardID?: string | null;
    onDeclareAttackers?: (attackerIDs: string[], defenderSeatID: string) => void;
    onUnstageAttackers?: (attackerIDs: string[]) => void;
    stagedAttackers?: ReadonlySet<string>;
  } = $props();
</script>

<div data-testid="board-stub">
  <button type="button" data-testid="select-b1" onclick={() => onSelectCombatCard?.("b1")}
    >select b1</button
  >
  {#each ["w1", "w2", "w3"] as id (id)}
    <button type="button" data-testid="select-{id}" onclick={() => onSelectCombatCard?.(id)}
      >select {id}</button
    >
  {/each}
  {#each ["p-2", "p-3"] as seat (seat)}
    <button type="button" data-testid="target-{seat}" onclick={() => onDeclareAttack?.(seat)}
      >attack {seat}</button
    >
  {/each}
  <button
    type="button"
    data-testid="group-attack"
    onclick={() => onDeclareAttackers?.(["w1", "w2", "w3"], "p-2")}>group attack</button
  >
  <button type="button" data-testid="group-unstage" onclick={() => onUnstageAttackers?.(["w3"])}
    >group unstage</button
  >
  <span data-testid="selected">{selectedCombatCardID ?? ""}</span>
  <span data-testid="staged">{[...(stagedAttackers ?? [])].sort().join(",")}</span>
</div>

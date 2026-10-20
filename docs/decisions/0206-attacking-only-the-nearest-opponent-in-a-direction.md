# ADR 0206 — Attacking only the nearest opponent in a direction

**Status:** Accepted · 2026-10-20 · S76 — Cards: incubate, face-down and outside the game
**Issue:** [#4197](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4197) (found on the Guffster's Party deck request, [#4184](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4184))
**Numbering:** `git fetch origin`, then every `docs/decisions/` file name on all 67 remote heads (`git ls-tree -r --name-only <head> -- docs/decisions`, including `origin/develop` and `origin/main`) and the recent issue comments. The highest in use is **0205**. Reserved on #4197 before any code was written.
**Builds on:** [ADR 0106](0106-any-player-may-activate-and-cant-attack-its-owner.md) §2 (`canAttackTargetWithLocked`, the one declaration-time predicate), [ADR 0063](0063-durations-and-control.md) and its 2026-10-08 amendment (a CR 508.1c restriction on a PLAYER's attacks, stored as a `PlayerStatic`), [ADR 0199](0199-a-hand-played-revealed.md) (a per-player state with a derived and a granted half), and the CR 614.12 anchor-word choice of [ADR 0071](0071-designations.md)'s 2026-09-27 amendment.

---

## Context

Two cards restrict attacks by where players sit:

- Pramikon, Sky Rampart: "As Pramikon enters, choose left or right. Each player may attack only the nearest opponent in the chosen direction and planeswalkers controlled by that opponent."
- Teyo, Geometric Tactician, −2: "Choose left or right. Until your next turn, each player may attack only the nearest opponent in the last chosen direction and planeswalkers controlled by that opponent."

The engine had no left or right. Its attack-target checks judged a creature's own restrictions (ADR 0106 §2) and one player's "can't attack <player>" grant (#2109), not a table-wide "may attack only" rule. Teyo shipped with the −2 left out and a caveat.

The rulings settle the open questions:

- Two Pramikons, or Teyo and a similar effect, that disagree about the direction leave players in a game of three or more unable to attack at all. So every rule in force applies at once; a newer one does not replace an older one. Teyo's "last chosen direction" is the direction chosen as that ability resolved.
- A Pramikon with no choice (a copy that entered without choosing) has no effect.
- Teyo's −2 does not affect creatures that enter the battlefield attacking (CR 508.4c).
- Only attacks are limited, not targeting or anything else.

## Decision

### 1. Left is the next player in turn order

The table's seats are its turn order: `Game.Seats`, advanced one index at a time. CR 101.4 puts the next player in turn order to a player's left, and CR 103.1 says turn order proceeds clockwise. So `AttackDirectionLeft` is the next seat and `AttackDirectionRight` the previous one. The nearest opponent in a direction (`Game.NearestOpponentForEffect`) is the first player still in the game walking that way, so a player who has left the game is skipped (CR 800.4a). The engine has no teams, so every other player is an opponent (CR 806.1, players compete as individuals). With two players left, both directions name the same opponent.

The prompt says what the words mean at this table ("left is the next player in turn order, right the player before"), because a screen has no left and right of its own and the client draws opponents in its own layout.

### 2. What may be attacked

The nearest opponent and planeswalkers they control. Nothing else, battles included: "may attack only" lists what is allowed, and battles are not on the list. The opponent is worked out live, for the attacking creature's controller, each time the question is asked, so a player leaving the game moves everybody's nearest opponent at once.

### 3. Two sources, one reader

- **Derived**: `Spec.AttackOnlyNearestOpponentInChosenDirection` (`CardDef`, read through the permanent's ability key), with the direction read off the permanent's CR 614.12 answer, `Card.ChosenOption` = `"Left"` or `"Right"`. `effects.ChooseLeftOrRightAsEnters` queues the choice as the existing anchor-word option pick. Before it is answered, and on a copy that never chose, the permanent restricts nothing. One that has lost its abilities restricts nothing (CR 613.1f). It is a rule about players, not a characteristic, so it is not a layer static and nothing is stored.
- **Granted**: `PlayerStatic.AttackDirection`, the thirteenth payload on the player slice, with a CR 611.2 `Duration`. `Game.GrantAttackDirectionForEffect` writes one entry on **every** player still in the game, so the grant's controller leaving does not take it with them (CR 800.4m keeps "until your next turn" running until that turn would have begun). Teyo's −2 asks left or right as it resolves (`effects.ChooseLeftOrRightThenEachPlayerAttacksOnlyThatWayUntilYourNextTurn`, a keyed option pick, so a table waiting on the answer is still a restore point) and grants it with `UntilYourNextTurnDuration` of its controller.

`attackDirectionRefusalLocked` folds both and refuses a target outside any rule in force with an `*AttackDirectionError` (wrapping `ErrIllegalAttackTarget`): "Grizzly Bears may attack only Bob, the nearest opponent to the left, and planeswalkers Bob controls (Pramikon, Sky Rampart)." The duration is tested in the reader as well as the sweep, as `playerCantAttackRefusalLocked` does.

### 4. Where it is read

The reader sits in `canAttackTargetWithLocked`, beside the attacking player's "can't attack" grant, and in `AttackTargetsForAttackerForEffect`. So both declaration verbs, the enumerator's per-attacker list, the ADR 0105 digest's `attack_targets`, the client's board gate (`attackTargetOpen`) and the bot all follow, and none of them learns the rule. CR 508.7b reselection and CR 508.4c's "put onto the battlefield attacking" keep the controller-only `canAttackTargetLocked`, which is the Teyo ruling.

The CR 508.1d requirement search (`attackRequirementCandidatesLocked`) now asks `canAttackTargetWithLocked` for each target instead of two of its parts. "Without disobeying any restrictions" covers this rule: a creature goaded by its controller's left-hand opponent under a left rule owes an attack at the goader, the only player it may attack, and nothing else. The same change makes the search honour #2109's "can't attack <player>" grant, which it had been skipping.

The card menu's per-creature "Declare attacker" rows now use the same gate during the declare attackers step: a defender the digest does not list for that creature is shown greyed with no action. That covers every attack-target restriction, not just this one.

### 5. Snapshot

`PlayerStatic.AttackDirection` is one new `omitempty` string on an existing snapshot struct, an additive change within the current version (#522). A binary from before it reads the grant back as an entry that says nothing, so after a rollback a player may attack anyone: a missing restriction, never an extra one. The derived half is not stored. The option pick's continuation key, `option-pick/teyo-geometric-tactician-direction`, is an on-disk identity.

## Consequences

- Pramikon, Sky Rampart ships complete, and Teyo, Geometric Tactician loses its caveat.
- Another "choose left or right" card is `ChooseLeftOrRightAsEnters` plus the Spec flag, or the resolved-ability pair in `effects/attack_direction.go`.
- The bot picks the first answer offered (left) for both choices: a legality-and-termination policy, like every anchor-word choice. Choosing the direction that keeps the strongest attacker off its own controller is a policy improvement for later.

## Out of scope

- A table badge naming the direction in force. The rule shows today as greyed defenders and the refusal sentence.
- CR 803's attack left / attack right variant options: no table option exists for them.

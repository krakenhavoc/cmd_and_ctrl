---
title: "As this enters, choose two players"
date: 2026-10-14
issues: [3389]
---
**As this enters, choose two players** (#3389, CR 614.12, CR 120.3) — Sower of Discord, from the Shirei deck request (#3173). `Card.ChosenPlayer` (ADR 0072 §7) held one player; the pair needed a second slot.
- **Storage.** A second field, `Card.ChosenSecondPlayer`, beside `ChosenPlayer`. It has the same lifecycle: per instance, in the snapshot (`chosenSecondPlayer`, additive within the schema version, so old restore points read "nobody chosen"), cloned with the card, cleared when the permanent leaves the battlefield (`entry_tail.go`, `zone.go`), and not a copiable value, so a copy chooses its own pair. It holds no func or interface, so the closure ratchet has nothing new to classify; `snapshot_drift_test.go` lists it as carried and `snapshot_shape/v8.txt` records it. `Game.ChosenPlayersOf` / `effects.ChosenPlayersOf` read the pair.
- **The prompt.** `QueueChooseTwoPlayersAsEntersForEffect` (`game/choose_player.go`) is two option-pick prompts in a row from the AsEnters hook (`effects.ChoosePlayersAsEnters`): the first offers every seat the clause admits, the second offers the rest, so the pair is two different players (the controller may be one). Both are built by `seatChoiceOptionsLocked`, so a seat that leaves while a prompt is open is pruned off it (CR 800.4a). With one candidate only the first prompt is asked. Each answer logs `player_chosen`.
- **The trigger.** An ordinary `EventDealDamage` watch: positive damage to either member of a complete pair, from any source, combat or not. A fill-in Build records the other player on the item's payload as it triggers, so a Sower that leaves in response still drains the player it named. The loss is `ChangePlayerLifeForEffect`; a chosen player who has left the game loses nothing.
- **Simplification.** The engine-wide as-enters one: the prompts are queued from the AsEnters hook, so the permanent is on the battlefield with no pair chosen until both are answered, and nothing can act in that window.
- **Cards:** Sower of Discord (Full).

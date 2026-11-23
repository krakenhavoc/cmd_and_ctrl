package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gpt_b2_helpers.go — bodies shared by Guildpact batch 2 (#5599).

// gptWhenEntersIfSpent is Guildpact's off-colour cycle, "When this
// creature enters, if {C} was spent to cast it, <effect>" (Steamcore
// Weird, Tin Street Hooligan, Ogre Savant, Shrieking Grotesque, Revenant
// Patriarch). It is Gruul Scrapper's intervening if (CR 603.4): the
// permanent's ManaSpentToCast record, at least one mana of colour `c`.
// A permanent that was not cast, or was cast without that colour, does
// not trigger. The trigger targets `targets`.
func GptWhenEntersIfSpent(c, label string, targets *game.TargetSpec, effect Effect) game.TriggeredAbility {
	return Targeting(On(game.EventETB, AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return source.ManaSpentToCast().Count(c) > 0
	}), label, effect), targets)
}

// gptTargetSkipsNextCombatPhase is "target player skips their next
// combat phase" as a trigger's body (Revenant Patriarch): Blinding
// Angel's record (game.SkipNextCombatPhaseForEffect) for the still-legal
// player target, so exactly one combat phase, the next one of theirs
// that would begin, is skipped (CR 500.11, 614.10).
func GptTargetSkipsNextCombatPhase(g *game.Game, item *game.StackItem) error {
	player, ok := FirstLegalPlayerTarget(NewContext(g, item))
	if !ok {
		return nil
	}
	g.SkipNextCombatPhaseForEffect(player, item.SourceCardID, item.Label)
	return nil
}

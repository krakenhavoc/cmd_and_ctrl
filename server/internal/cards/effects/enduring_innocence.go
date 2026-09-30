package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enduring Innocence — Enchantment Creature — Sheep Glimmer {1}{W}{W},
// 2/1:
//
//	"Lifelink
//	 Whenever one or more other creatures you control with power 2 or
//	 less enter, draw a card. This ability triggers only once each
//	 turn.
//	 When Enduring Innocence dies, if it was a creature, return it to
//	 the battlefield under its owner's control. It's an enchantment.
//	 (It's not a creature.)"
//
// Lifelink and the once-per-turn ETB draw are ordinary printed
// vocabulary: `OncePerBatch` collapses a simultaneous-entry batch into
// the one trigger CR 603.2c already wants, and `b11TriggeredThisTurn`
// (Morbid Opportunist's own gate) shuts the SECOND batch of the turn
// out entirely, which OncePerBatch alone would not do.
//
// The dies-and-return clause is the Enduring cycle's shared trigger
// (WhenThisDiesReturnItAsAnEnchantment, glimmer_return.go): the card
// comes back once as an enchantment that is not a creature, and a
// second death finds "if it was a creature" false and leaves it in
// the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "98a389f4-2905-47f3-b60e-3d4afb3e5cb0",
		Name:            "Enduring Innocence",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventETB, enduringInnocenceLowPowerCreatureEntered,
				enduringInnocenceDrawLabel, Do(DrawCards{N: 1}))),
			WhenThisDiesReturnItAsAnEnchantment("Enduring Innocence"),
		},
	})
}

const enduringInnocenceDrawLabel = "Enduring Innocence — draw a card"

// enduringInnocenceLowPowerCreatureEntered is "one or more OTHER
// creatures you control with power 2 or less enter", gated to once
// each turn.
func enduringInnocenceLowPowerCreatureEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	if !ok || !c.IsCreature() || c.Effective().Power > 2 {
		return false
	}
	return !b11TriggeredThisTurn(g, source.InstanceID, enduringInnocenceDrawLabel)
}

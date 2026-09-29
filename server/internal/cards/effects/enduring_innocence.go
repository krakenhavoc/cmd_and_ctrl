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
// # The Glimmer half is deliberately left out
//
// This is the Enduring cycle's shared gap — see Enduring Vitality and
// Enduring Tenacity, which dropped the identical clause for the
// identical reason: there is no per-instance "lost its creature type
// permanently" state to hang a returned object's stripped Creature
// type on, and the layer-4 machinery that WOULD strip it is keyed on
// the catalog entry shared by every copy of the card, not on one
// instance.
//
// Shipping the return WITHOUT the type strip would make this a free,
// unkillable, infinitely recursive card-draw engine — far STRONGER
// than printed, which the #259 rule forbids. Dropping the clause
// instead leaves a 2/1 lifelinker that dies once, which is weaker
// than printed and is the right direction.
//
// The clause becomes writable once the engine can carry a
// per-instance type override through a battlefield entry — the same
// state every other Enduring card is waiting on.
func init() {
	Register(Spec{
		OracleID:        "98a389f4-2905-47f3-b60e-3d4afb3e5cb0",
		Name:            "Enduring Innocence",
		Completeness:    CompletenessCaveats,
		PrintedKeywords: []string{"lifelink"},
		Caveats: []string{
			"When this dies, it doesn't return to the battlefield as an enchantment — it goes to the graveyard like any other creature.",
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventETB, enduringInnocenceLowPowerCreatureEntered,
				enduringInnocenceDrawLabel, Do(DrawCards{N: 1}))),
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

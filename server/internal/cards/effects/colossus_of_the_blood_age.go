package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Colossus of the Blood Age — Artifact Creature — Construct {4}{R}{W}, 6/6
// (Secrets of Strixhaven, #4456):
//
//	"When this creature enters, it deals 3 damage to each opponent and you
//	 gain 3 life.
//	 When this creature dies, discard any number of cards, then draw that
//	 many cards plus one."
//
// The enters trigger is one event of 3 damage from the Colossus to every
// opponent, then 3 life (the drain shape, Do(DamageEachOpponent, GainLife)).
// The dies trigger is Into the Night's loot: the count actually discarded
// plus one, drawn once the discards are done (the keyed continuation
// drawOneMoreThanDiscarded); with an empty hand there is nothing to discard
// and it just draws the one card. The controller is whoever controlled the
// Colossus as it died.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0788fdd2-6fab-460c-b054-a736df670f96",
		Name:         "Colossus of the Blood Age",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Colossus of the Blood Age — 3 damage to each opponent and you gain 3 life",
				Do(DamageEachOpponent{N: 3}, GainLife{Amount: 3})),
			WhenThisDies("Colossus of the Blood Age — discard any number of cards, then draw that many plus one",
				colossusOfTheBloodAgeDies),
		},
	})
}

// colossusOfTheBloodAgeDies is "discard any number of cards, then draw that
// many cards plus one".
func colossusOfTheBloodAgeDies(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	hand := len(allHandCardIDs(g, item.Controller))
	if hand == 0 {
		return g.DrawNForEffect(item.Controller, 1)
	}
	return b39MayDiscardThenDraw(hand, false,
		"Colossus of the Blood Age — discard any number of cards, then draw that many plus one",
		drawOneMoreThanDiscarded)(ctx)
}

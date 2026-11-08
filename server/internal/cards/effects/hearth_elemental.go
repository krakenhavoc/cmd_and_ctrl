package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hearth Elemental // Stoke Genius — Creature — Elemental {5}{R}, 4/5 //
// Sorcery — Adventure {1}{R} (Wilds of Eldraine, #4471):
//
//	Hearth Elemental — "This spell costs {X} less to cast, where X is the
//	 number of cards in your graveyard that are instant cards, sorcery
//	 cards, and/or have an Adventure."
//	Stoke Genius — "Discard your hand, then draw two cards. (Then exile
//	 this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The reduction is Fungal Colossus' self cost modifier
// (CostsLessEach), counted as the cost is determined (CR 601.2f) with
// instantSorceryOrAdventureCardsInGraveyard; it is on the creature face
// only, so Stoke Genius is never cheaper, and it takes off generic mana
// only, so {R} is always paid. Stoke Genius is Peer Past the Veil's
// wheel: the whole hand is discarded, then two cards are drawn once any
// prompt the discard asked is answered (CR 608.2c). The Adventure
// lifecycle is the engine's (CR 715.3d).
//
// No simplification.
const hearthElementalOracleID = "f01cf0c5-0ac6-4ad3-8642-dd21c8e8563e"

func init() {
	Register(Spec{
		OracleID:     hearthElementalOracleID,
		Name:         "Hearth Elemental",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(func(q game.CostQuery) int {
				if q.Game == nil {
					return 0
				}
				return instantSorceryOrAdventureCardsInGraveyard(q.Game, q.Controller)
			}, "This spell costs {X} less to cast, where X is the number of cards in your graveyard that are instant cards, sorcery cards, and/or have an Adventure."),
		},
	})
	Register(Spec{
		OracleID:     hearthElementalOracleID + "#1",
		Name:         "Stoke Genius",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				_, err := discardWholeHand(ctx.Game, ctx.Controller())
				return err
			}),
			DrawCards{N: 2},
		),
	})
}

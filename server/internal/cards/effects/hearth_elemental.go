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
// Both faces register under the Adventure keyspace (the bare oracle ID and
// "#1"). The discount is Chitin Gravestalker's: counted as the cost is
// determined (CR 601.2f) over the caster's graveyard, reducing generic
// mana only, each card once (std54InstantSorceryOrAdventureInGraveyard).
// Stoke Genius discards the whole hand, which takes every card with no
// choice (Change of Fortune), then draws two.
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
				return std54InstantSorceryOrAdventureInGraveyard(q.Game, q.Controller)
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
			Instruction(DrawCards{N: 2}.Apply),
		),
	})
}

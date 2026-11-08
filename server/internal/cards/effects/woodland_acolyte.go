package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Woodland Acolyte // Mend the Wilds — Creature — Human Cleric {2}{W},
// 2/2 // Instant — Adventure {G} (Wilds of Eldraine, #4471):
//
//	Woodland Acolyte — "When this creature enters, draw a card."
//	Mend the Wilds — "Put target permanent card from your graveyard on
//	 top of your library. (Then exile this card. You may cast the
//	 creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The Acolyte's entry draw is the footprinted primitive. Mend
// the Wilds is Reclaim narrowed to a permanent card (artifact, battle,
// creature, enchantment, land or planeswalker, CR 110.4b). The Adventure
// lifecycle is the engine's (CR 715.3d).
//
// No simplification.
const woodlandAcolyteOracleID = "ba610f08-61a7-47fe-9a80-32b5ce51b118"

func init() {
	Register(Spec{
		OracleID:     woodlandAcolyteOracleID,
		Name:         "Woodland Acolyte",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Woodland Acolyte — draw a card", Do(DrawCards{N: 1})),
		},
	})
	Register(Spec{
		OracleID:     woodlandAcolyteOracleID + "#1",
		Name:         "Mend the Wilds",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target permanent card from your graveyard", YouOwn(), Permanent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return PutChosenTargetOnTopOfLibrary(ctx.Game, item)
		},
	})
}

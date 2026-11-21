package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
const WoodlandAcolyteOracleID = "ba610f08-61a7-47fe-9a80-32b5ce51b118"

func init() {
	effects.Register(effects.Spec{
		OracleID:     WoodlandAcolyteOracleID,
		Name:         "Woodland Acolyte",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Woodland Acolyte — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     WoodlandAcolyteOracleID + "#1",
		Name:         "Mend the Wilds",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target permanent card from your graveyard", effects.YouOwn(), effects.Permanent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PutChosenTargetOnTopOfLibrary(ctx.Game, item)
		},
	})
}

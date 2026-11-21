package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const mirkwoodChannelerLabel = "Mirkwood Channeler — target Elf you control gains trample and gets +X/+X until end of turn"

// Mirkwood Channeler — Creature — Elf Druid {3}{G}, 3/3 (The Lord of the
// Rings: Tales of Middle-earth Commander):
//
//	"At the beginning of combat on your turn, target Elf you control gains
//	 trample and gets +X/+X until end of turn, where X is the number of
//	 Forests you control."
//
// Brambleguard Captain's targeted combat trigger. X is the number of lands
// with the Forest land type its controller controls as the trigger
// resolves (Dungrove Elder's count, by effective subtype); the Channeler
// may target itself, since it is an Elf. A target that left in response
// gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30f373d8-cd8c-4fb2-9383-9969fc547df9",
		Name:         "Mirkwood Channeler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat(mirkwoodChannelerLabel, mirkwoodChannelerPump),
				effects.TargetCreature("target Elf you control", effects.YouControl(), effects.HasSubtype("Elf"))),
		},
	})
}

func mirkwoodChannelerPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	forest := "Forest"
	if src, ok := g.LookupCardForEffect(item.SourceCardID); ok {
		forest = src.TextSubtype("Forest")
	}
	x := effects.B08LandsWithSubtypeControlled(g, item.Controller, forest)
	return effects.GapB125PumpTrample(ctx, id, x, x, "Mirkwood Channeler")
}

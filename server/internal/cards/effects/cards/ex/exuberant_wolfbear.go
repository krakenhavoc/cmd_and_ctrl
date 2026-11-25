package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exuberant Wolfbear — Creature — Wolf Bear {3}{G}, 4/4 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Whenever this creature attacks, you may have the base power and
//	 toughness of target Human you control become equal to this creature's
//	 power and toughness until end of turn."
//
// Galion's shape: the target is chosen as the trigger goes on the stack,
// the "you may" is asked as it resolves (CR 603.5), and the size is read
// from the Wolfbear as it is then (its last known size if it has left, CR
// 608.2h). The base P/T is a layer 7b effect (CR 613.4b), so counters and
// pumps on the Human still apply on top.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40f5f8b9-5c2c-41d4-8b83-55a2e51ced64",
		Name:         "Exuberant Wolfbear",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WheneverThisAttacks("Exuberant Wolfbear — base power and toughness of target Human you control become equal to this creature's until end of turn", exuberantWolfbearCopyStats),
				effects.TargetCreature("target Human you control", effects.YouControl(), effects.HasSubtype("Human"))),
				"Exuberant Wolfbear — have the base power and toughness of the target Human become equal to its power and toughness until end of turn?"),
		},
	})
}

func exuberantWolfbearCopyStats(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	src, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	return effects.ScopedEffectFor{
		Target:   id,
		Mods:     game.SetBasePTMods(src.Power, src.Toughness),
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Exuberant Wolfbear — base power and toughness equal to the Wolfbear's",
	}.Apply(ctx)
}

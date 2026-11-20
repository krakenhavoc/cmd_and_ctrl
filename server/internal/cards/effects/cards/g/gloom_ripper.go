package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gloom Ripper — Creature — Elf Assassin {3}{B}{B}, 4/4 (Lorwyn
// Eclipsed, #4458):
//
//	"When this creature enters, target creature you control gets +X/+0
//	 until end of turn and up to one target creature an opponent
//	 controls gets -0/-X until end of turn, where X is the number of
//	 Elves you control plus the number of Elf cards in your graveyard."
//
// Two clauses with different predicates (Bite Down's shape), read back
// by slot. X is counted once, as the trigger resolves (CR 608.2h): Elf
// permanents you control (the Ripper itself if it is still there, a
// changeling, a Kindred Elf) plus Elf cards in your graveyard. Each half
// applies only to a target still legal; the first can be the Ripper.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78e1beca-840d-47b2-8759-cc31454fdf5a",
		Name:         "Gloom Ripper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Gloom Ripper — target creature you control gets +X/+0 and up to one target creature an opponent controls gets -0/-X",
					gloomRipperEnters),
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("up to one target creature an opponent controls", effects.OpponentControls()).WithCount(0, 1),
				)),
		},
	})
}

// gloomRipperEnters counts X and applies each half to its target.
func gloomRipperEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	x := gloomRipperX(ctx)
	if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard && x > 0 {
		if err := (effects.BoostUntilEOT{Target: t.ID, Power: x, Label: "Gloom Ripper — +X/+0"}).Apply(ctx); err != nil {
			return err
		}
	}
	if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard && x > 0 {
		return effects.BoostUntilEOT{Target: t.ID, Toughness: -x, Label: "Gloom Ripper — -0/-X"}.Apply(ctx)
	}
	return nil
}

// gloomRipperX is the number of Elves the controller controls plus the
// number of Elf cards in their graveyard, the word read through the
// Ripper's text (CR 612.2).
func gloomRipperX(ctx *effects.Context) int {
	elf := ctx.TextSubtype("Elf")
	n := effects.PermanentsOfTypeYouControl(ctx.Game, ctx.Controller(), elf)
	if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil {
		for _, c := range p.Graveyard.Cards {
			if c.HasSubtype(elf) {
				n++
			}
		}
	}
	return n
}

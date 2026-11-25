package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Joust — Sorcery {1}{R} (Throne of Eldraine, #5516):
//
//	"Choose target creature you control and target creature you don't
//	 control. The creature you control gets +2/+1 until end of turn if it's a
//	 Knight. Then those creatures fight each other."
//
// Blizzard Brawl's two clauses: the Knight check is read as the spell
// resolves, the pump comes before the fight so the fight reads the bigger
// power (CR 701.14), and a missing target skips the fight but not the pump.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8584e96f-a3e7-475c-bfd1-2c9a72bee774",
		Name:         "Joust",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			if c, ok := ctx.Game.LookupCardForEffect(mine.ID); ok && c.HasSubtype(ctx.TextSubtype("Knight")) {
				if err := (effects.BoostUntilEOT{Target: mine.ID, Power: 2, Toughness: 1, Label: "Joust — +2/+1"}).Apply(ctx); err != nil {
					return err
				}
			}
			theirs, ok := ctx.ClauseTarget(1)
			if !ok || theirs.Kind != game.TargetCard {
				return nil
			}
			return effects.B10Fight(ctx, mine.ID, theirs.ID)
		},
	})
}

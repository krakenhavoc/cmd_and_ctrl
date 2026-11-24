package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blizzard Brawl — Snow Sorcery {G} (Kaldheim):
//
//	"Choose target creature you control and target creature you don't control.
//	 If you control three or more snow permanents, the creature you control
//	 gets +1/+0 and gains indestructible until end of turn. Then those
//	 creatures fight each other. (Each deals damage equal to its power to the
//	 other.)"
//
// Two required target clauses. The snow count is read as the spell resolves;
// with three the pump and the keyword come first, so the fight reads the
// bigger power (CR 701.14). If either creature is gone no fight happens, but
// the pump still applies to a creature of yours that remains.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25a0fe3b-59c8-496a-a8d6-fca966ca6602",
		Name:         "Blizzard Brawl",
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
			if effects.S82OldB064SnowPermanentsYouControl(ctx.Game, item.Controller) >= 3 {
				if err := (effects.BoostUntilEOT{Target: mine.ID, Power: 1, Label: "Blizzard Brawl — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: mine.ID, Keywords: []string{"indestructible"}, Label: "Blizzard Brawl — indestructible"}).Apply(ctx); err != nil {
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

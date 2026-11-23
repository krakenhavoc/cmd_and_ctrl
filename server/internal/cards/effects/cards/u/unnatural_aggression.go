package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unnatural Aggression — Instant {2}{G}, devoid:
//
//	"Devoid (This card has no color.)
//	 Target creature you control fights target creature an opponent
//	 controls. If the creature an opponent controls would die this turn,
//	 exile it instead."
//
// Devoid is declared in PrintedKeywords, and the engine reads it as CR
// 702.114a's colour-defining ability (#2152), so the card is colourless
// in every zone. The two
// clauses are checked one by one at resolution (CR 608.2b). The fight
// needs both targets to be legal (b10Fight). The exile rider belongs to
// the spell, not the fight. A legal opponent's creature is marked even
// when your creature has gone and no fight happened (the 2015-08-25
// rulings), and even when the fight dealt it no damage.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "597d1dce-67e8-4c37-9781-eeaeb7c2d7b7",
		Name:            "Unnatural Aggression",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine, okMine := ctx.ClauseTarget(0)
			theirs, okTheirs := ctx.ClauseTarget(1)
			okMine = okMine && mine.Kind == game.TargetCard
			okTheirs = okTheirs && theirs.Kind == game.TargetCard
			if okMine && okTheirs {
				if err := effects.B10Fight(ctx, mine.ID, theirs.ID); err != nil {
					return err
				}
			}
			if !okTheirs {
				return nil
			}
			return effects.ExileIfItWouldDieThisTurn{Target: theirs.ID}.Apply(ctx)
		},
	})
}

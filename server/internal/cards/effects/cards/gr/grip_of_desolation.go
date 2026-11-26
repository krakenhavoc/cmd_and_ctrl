package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grip of Desolation — Instant {4}{B}{B} (Battle for Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Exile target creature and target land."
//
// Two target clauses (a creature and a land), each re-checked as the spell
// resolves (CR 608.2b): a clause whose target has become illegal does nothing
// while the other still resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a7c8d026-b5b0-4912-a4a0-4c35df5f8899",
		Name:            "Grip of Desolation",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetPermanent("target land", effects.Land()),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for slot := 0; slot < 2; slot++ {
				t, ok := ctx.ClauseTarget(slot)
				if !ok || t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

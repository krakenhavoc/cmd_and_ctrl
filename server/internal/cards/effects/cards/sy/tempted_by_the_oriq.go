package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tempted by the Oriq — Sorcery {1}{U}{U}{U} (Strixhaven: School of Mages,
// S82 batch 047, #5502):
//
//	"For each opponent, gain control of up to one target creature or
//	 planeswalker that player controls with mana value 3 or less."
//
// "For each opponent, up to one target … that player controls" is Mass
// Mutiny's per-opponent clauses (PerOpponentClauses): one "up to one"
// clause per opponent, each bound to that opponent, so the CR 608.2b
// re-check asks whether THAT player still controls the pick. Mana value is
// read as it is now (a token is 0). The control change has no stated
// duration (CR 611.2a), Blatant Thievery's shape: it lasts until the game
// ends or the permanent stops being the object it was. Each pick resolves
// on its own, so one that became illegal is skipped and the rest are
// still taken.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "304450b0-bde6-4d69-bcaf-fb9b9bae2e4b",
		Name:         "Tempted by the Oriq",
		Completeness: effects.CompletenessFull,
		Targets: effects.PerOpponentClauses(game.MaxPlayers-1, "creature or planeswalker with mana value 3 or less",
			effects.Or(effects.Creature(), effects.Planeswalker()), effects.ManaValueLE(3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GainControl{
					Target:     t.ID,
					Controller: item.Controller,
					Duration:   game.IndefiniteDuration(),
					Label:      "Tempted by the Oriq — gain control (no stated duration)",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reveille Squad — Creature — Human Rebel {2}{W}{W}, 3/3 (Prophecy, S82 OLD b671,
// #5617):
//
//	"Whenever one or more creatures attack you, if this creature is untapped, you
//	 may untap all creatures you control."
//
// Sabotage Strategist's one-trigger-per-declaration attack watcher (an attack
// declared against YOU, not a planeswalker or battle of yours, CR 508.3e) with
// Reservoir Kraken's intervening "if this creature is untapped" (CR 603.4,
// checked as it triggers and again as it resolves, before the "you may" is
// asked). Untapping is Aurelia's UntapAllCreaturesYouControl, the Squad included.
//
// No simplification.
func init() {
	const label = "Reveille Squad — you may untap all creatures you control"
	effects.Register(effects.Spec{
		OracleID:     "2531eb1a-ae24-450a-b8e8-dbbd65eaad64",
		Name:         "Reveille Squad",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.InterveningIf(effects.CondBoard{Name: "this creature is untapped", Test: func(ctx *effects.Context) bool {
				self, ok := ctx.SourcePermanent()
				return ok && !self.Tapped
			}}, effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.Actor != source.Controller && ev.Target == source.Controller && !source.Tapped
			}, label, effects.Do(effects.UntapAllCreaturesYouControl{})))),
				"Reveille Squad — untap all creatures you control?"),
		},
	})
}

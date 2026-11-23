package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gnarling Targ — Creature — Boar {4}{R}, 4/5 (Star Trek Commander, #4974):
//
//	"Trample
//	 Whenever you attack, Klingons and Warriors you control get +2/+0
//	 and gain trample until end of turn."
//
// "Whenever you attack" is one trigger per attack declaration (Adeline's
// OncePerBatch on the per-creature EventAttack). The Targ is a Boar, so
// it pumps only the Klingons and Warriors, attacking or not. The set is
// locked as the ability resolves (CR 611.2c): a Klingon that enters later
// this turn gets nothing. The +2/+0 and the trample are one effect at
// one timestamp.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1fc59e37-0dc0-4e24-af71-6bab4a2b56ca",
		Name:            "Gnarling Targ",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Gnarling Targ — Klingons and Warriors you control get +2/+0 and gain trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.UntilEndOfTurn(effects.NewContext(g, item), uuid.Nil, effects.TrekB37KlingonsAndWarriorsYouControl(),
						"Gnarling Targ — +2/+0 and trample", game.ModifyPTMod(2, 0), game.AddKeywordsMod("trample"))
				})),
		},
	})
}

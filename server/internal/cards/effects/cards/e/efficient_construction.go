package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Efficient Construction — Enchantment {3}{U} (EDHREC rank 2633):
//
//	"Whenever you cast an artifact spell, create a 1/1 colorless
//	 Thopter artifact creature token with flying."
//
// Sai, Master Thopterist's trigger on an enchantment. The spell is
// read off the stack, where its type line is intact; the Thopter is
// itself an artifact, so it feeds the artifact-ETB payoffs, and it
// is a token entering rather than a spell cast, so it does not
// re-trigger this.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5af48f87-7b94-44de-90e3-91f10ced00d3",
		Name:         "Efficient Construction",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && spell.IsArtifact()
			}, "Efficient Construction — create a Thopter", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
	})
}

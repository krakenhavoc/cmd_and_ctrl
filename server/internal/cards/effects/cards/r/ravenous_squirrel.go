package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Squirrel — Creature — Squirrel {B/G}, 1/1 (EDHREC rank
// 3030):
//
//	"Whenever you sacrifice an artifact or creature, put a +1/+1
//	 counter on this creature.
//	 {1}{B}{G}, Sacrifice an artifact or creature: You gain 1 life
//	 and draw a card."
//
// The one-drop aristocrat. The trigger is
// b28YouSacrificedArtifactOrCreature — the sacrifice event fires
// before the zone move, so the permanent is still on the battlefield
// to be typed — and grows the Squirrel if it is still there when the
// trigger resolves. The activated ability's sacrifice is a COST
// (SacrificeOther over artifacts and creatures, the Squirrel itself a
// legal choice as in paper), paid at announce, so it fires the first
// ability above the second: the counter lands, then the life and the
// card, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fb48d65-406b-4231-83ea-9cf3bb8dff76",
		Name:         "Ravenous Squirrel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B28YouSacrificedArtifactOrCreature(ev, source, g)
			}, "Ravenous Squirrel — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}{G}, Sacrifice an artifact or creature: You gain 1 life and draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: effects.Plus(effects.ManaCost("{1}{B}{G}"), game.AbilityCost{
				SacrificeOther: effects.SacrificeSpec("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
			}),
			// The draw waits for a prompt the life gain asks (CR 608.2c, #4155).
			Effect: effects.Do(effects.GainLife{Amount: 1}, effects.DrawCards{N: 1}),
		}},
	})
}

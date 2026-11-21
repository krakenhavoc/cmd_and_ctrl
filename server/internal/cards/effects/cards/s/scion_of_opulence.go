package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scion of Opulence — Creature — Vampire Noble {2}{R}, 3/1 (Innistrad:
// Crimson Vow):
//
//	"Whenever this creature or another nontoken Vampire you control dies,
//	 create a Treasure token.
//	 {R}, Sacrifice two artifacts: Exile the top card of your library. You
//	 may play that card this turn."
//
// The death trigger reads the creature as it last existed on the
// battlefield (CR 603.10a, leftAsSubtype), so a creature a lord made a
// Vampire counts. The activation pays {R} and two artifacts other than the
// Scion's own (a Treasure it makes is the usual fodder); the exiled card
// carries a this-turn permission to play it, a land counting against the
// land drop.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5818b31e-2e86-4230-937e-478e14534f91",
		Name:         "Scion of Opulence",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				if !ok || effects.LeftUnderControlOf(ev, dead) != source.Controller {
					return false
				}
				if ev.CardID == source.InstanceID {
					return true
				}
				return !effects.IsToken(dead) && effects.LeftAsSubtype(ev, dead, source.TextSubtype("Vampire"))
			},
			Key:     "Scion of Opulence — create a Treasure token",
			Purpose: game.Purpose{DeathPayoff: true, Tokens: 1},
			Effect:  effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1}),
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice two artifacts: Exile the top card of your library. You may play that card this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeN(2, "two artifacts", effects.Artifact())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileTopWithPermission{
					From: item.Controller, GrantTo: item.Controller, N: 1,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thermo-Alchemist — Creature — Human Shaman {1}{R}, 0/3 (EDHREC
// rank 1716):
//
//	"Defender
//	 {T}: This creature deals 1 damage to each opponent.
//	 Whenever you cast an instant or sorcery spell, untap this
//	 creature."
//
// The spellslinger's pinger: a point to every opponent per untap,
// and every instant or sorcery is an untap. The tap ability is a
// CR 602 activation with a tap cost, so summoning sickness applies
// (CR 302.6) and the damage goes on the stack with a response
// window; the untap is a trigger on the cast, which resolves above
// the spell — so the Alchemist can tap again before the spell
// resolves, as printed. The Alchemist is the damage source, so
// Torbran adds to it and a Fog-class shield stops it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "228fbae1-423e-461d-b8c3-55786938a3cb",
		Name:            "Thermo-Alchemist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageToEachOpponent(g, item, 1)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.InstantOrSorceryCastByYou(ev, source, g)
			}, "Thermo-Alchemist — untap", func(g *game.Game, item *game.StackItem) error {
				if !effects.B15OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Convergence of Dominion — Artifact {3} (Warhammer 40,000 Commander, S82
// GAP batch 186, #5948):
//
//	"Dynastic Command Node — As long as you control your commander,
//	 activated abilities of cards in your graveyard cost {2} less to
//	 activate. This effect can't reduce the mana in that ability's
//	 activation cost to less than one mana.
//	 Translocation Protocols — {3}, {T}: Mill three cards."
//
// The reduction is Embalmer's Tools' activation pass over any card in its
// controller's graveyard, with Forensic Gadgeteer's one-mana floor, switched
// on while the controller controls a commander they own
// (B18ControlsYourCommander).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "655aece2-f33a-46e0-8f80-112ae8a457fe",
		Name:         "Convergence of Dominion",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.KeepsOneMana(effects.ActivationCostsLess(2,
				"Dynastic Command Node — As long as you control your commander, activated abilities of cards in your graveyard cost {2} less to activate. This effect can't reduce the mana in that ability's activation cost to less than one mana.",
				func(q game.CostQuery) bool {
					return q.FromZone == game.ZoneGraveyard && q.Card.Owner == q.Source.Controller &&
						effects.B18ControlsYourCommander(q.Game, q.Source.Controller)
				})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Translocation Protocols — {3}, {T}: Mill three cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect:  effects.Do(effects.MillCards{N: 3}),
		}},
	})
}

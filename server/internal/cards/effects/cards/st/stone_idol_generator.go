package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stone Idol Generator — Artifact {5}:
//
//	"Whenever a creature you control attacks, you get {E} (an energy
//	 counter).
//	 {T}, Pay six {E}: Create a 6/12 colorless Construct artifact
//	 creature token with trample. Activate only as a sorcery."
//
// One trigger per attacking creature: EventAttack fires once per
// creature on its declaration, so three attackers are three energy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "411c3636-a5b3-42fa-8020-ceacf581a4f3",
		Name:         "Stone Idol Generator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Stone Idol Generator — you get {E}", effects.Do(effects.GetEnergy{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}, Pay six {E}: Create a 6/12 colorless Construct artifact creature token with trample. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.TapCost(), effects.PayEnergy(6)),
			SorcerySpeed: true,
			Purpose:      game.Purpose{Tokens: 1},
			Effect:       effects.Do(effects.CreateToken{N: 1, Template: effects.TokenCard("6/12 colorless Construct artifact with trample")}),
		}},
	})
}

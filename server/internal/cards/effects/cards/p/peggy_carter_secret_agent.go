package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peggy Carter, Secret Agent — Legendary Creature — Human Spy Hero {1}{W},
// 2/1 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever a creature you control attacks alone, it gains indestructible
//	 until end of turn."
//
// Black Widow, Double Agent's shape: "alone" is game.AttackedAlone read
// when the attack trigger is harvested, and the creature that gains the
// keyword is the lone attacker read off the event, Peggy herself or not. A
// creature removed in response gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f3c6b8e-df89-4f0c-97bb-fbfcda5232b4",
		Name:         "Peggy Carter, Secret Agent",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.MshB1YouAttackAlone,
				"Peggy Carter, Secret Agent — the lone attacker gains indestructible",
				func(g *game.Game, item *game.StackItem) error {
					attacker := item.Trigger.Event.CardID
					if !effects.OnBattlefield(g, attacker) {
						return nil
					}
					return effects.GrantKeywordUntilEOT{
						Target:   attacker,
						Keywords: []string{"indestructible"},
						Label:    "Peggy Carter, Secret Agent",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

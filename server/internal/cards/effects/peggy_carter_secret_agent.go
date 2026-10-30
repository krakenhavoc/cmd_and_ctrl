package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2f3c6b8e-df89-4f0c-97bb-fbfcda5232b4",
		Name:         "Peggy Carter, Secret Agent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, mshB1YouAttackAlone,
				"Peggy Carter, Secret Agent — the lone attacker gains indestructible",
				func(g *game.Game, item *game.StackItem) error {
					attacker := item.Trigger.Event.CardID
					if !onBattlefield(g, attacker) {
						return nil
					}
					return GrantKeywordUntilEOT{
						Target:   attacker,
						Keywords: []string{"indestructible"},
						Label:    "Peggy Carter, Secret Agent",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

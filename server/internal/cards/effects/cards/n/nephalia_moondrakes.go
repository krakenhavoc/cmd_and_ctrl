package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nephalia Moondrakes — Creature — Drake {5}{U}{U}, 5/5 (Shadows over
// Innistrad, #5543):
//
//	"Flying
//	 When this creature enters, target creature gains flying until end of
//	 turn.
//	 {4}{U}{U}, Exile this card from your graveyard: Creatures you control
//	 gain flying until end of turn."
//
// The enters trigger targets as it goes on the stack and grants the keyword
// to the target if it is still legal (CR 608.2b). The graveyard ability
// (ActivatedAbilityShape.Zones) exiles the card as a cost, and the grant
// covers the creatures you control as it resolves, locked in then
// (CR 611.2c): a creature that arrives later does not fly.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "95649140-7978-4684-a1f1-af771e1d345a",
		Name:            "Nephalia Moondrakes",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Nephalia Moondrakes — target creature gains flying until end of turn",
				effects.S82OldB285FlyingToFirstTarget("Nephalia Moondrakes")),
				effects.TargetCreature("target creature")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{U}{U}, Exile this card from your graveyard: Creatures you control gain flying until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Keywords: []string{"flying"}}, Answers: game.AnswerCombatGrant},
			Cost:    effects.Plus(effects.ManaCost("{4}{U}{U}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect: effects.Do(effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"flying"},
				Label:    "Nephalia Moondrakes — creatures you control gain flying",
			}),
		}},
	})
}

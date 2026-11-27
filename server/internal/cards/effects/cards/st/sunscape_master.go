package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sunscape Master — Creature — Human Wizard {2}{W}{W}, 2/2 (Invasion):
//
//	"{G}{G}, {T}: Creatures you control get +2/+2 until end of turn.
//	 {U}{U}, {T}: Return target creature to its owner's hand."
//
// Two abilities with {T} in the cost, so summoning sickness holds both back
// (CR 302.6). The pump is Warrior's Honor's: the creatures you control as it
// resolves are pumped (CR 611.2c). The bounce is Barrin's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3be8e19-cec1-43ae-9e22-d7cb374570bb",
		Name:         "Sunscape Master",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{G}{G}, {T}: Creatures you control get +2/+2 until end of turn.",
				Purpose: game.Purpose{TeamPump: &game.Pump{Power: 2, Toughness: 2}, Answers: game.AnswerPump},
				Cost:    effects.Plus(effects.ManaCost("{G}{G}"), effects.TapCost()),
				Effect: effects.Do(effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 2,
					Label: "Sunscape Master — creatures you control get +2/+2",
				}),
			},
			{
				Label:   "{U}{U}, {T}: Return target creature to its owner's hand.",
				Cost:    effects.Plus(effects.ManaCost("{U}{U}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BounceTheTargetToHand,
			},
		},
	})
}

package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pia Nalaar — Legendary Creature — Human Artificer {2}{R}, 2/2 (Commander
// 2021, #5504):
//
//	"When Pia Nalaar enters, create a 1/1 colorless Thopter artifact
//	 creature token with flying.
//	 {1}{R}: Target artifact creature gets +1/+0 until end of turn.
//	 {1}, Sacrifice an artifact: Target creature can't block this turn."
//
// The pump targets an artifact creature (a Thopter, or Pia herself if she
// is one). The second ability sacrifices an artifact as a cost (it may be
// the Thopter, or any artifact) and the creature it targets gains the
// can't-block restriction until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "715336f5-a708-488b-9f30-5702a9013af0",
		Name:         "Pia Nalaar",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pia Nalaar — create a 1/1 Thopter", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{R}: Target artifact creature gets +1/+0 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)),
				Cost:    effects.ManaCost("{1}{R}"),
				Targets: effects.TargetCreature("target artifact creature", effects.Artifact()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: id, Power: 1, Label: "Pia Nalaar — +1/+0 until end of turn"}.Apply(ctx)
				},
			},
			{
				Label:   "{1}, Sacrifice an artifact: Target creature can't block this turn.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Targets: game.ForTargets(effects.CantBlockUntilEOT(0, 0, 0))},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeN(1, "an artifact", effects.Artifact())),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.TargetCreatureCantBlockThisTurn("Pia Nalaar — can't block this turn"),
			},
		},
	})
}

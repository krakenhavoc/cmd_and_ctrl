package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avalanche Caller — Snow Creature — Human Wizard {1}{U}, 1/1 (Jumpstart
// 2022, #5948):
//
//	"{2}: Target snow land you control becomes a 4/4 Elemental creature with
//	 hexproof and haste until end of turn. It's still a land."
//
// One layer-4 type change (a creature, Elemental, still a land), a 4/4 base
// power and toughness (layer 7b) and hexproof and haste (layer 6), all until
// end of turn on the object the target was when the ability resolved (CR 611.2c);
// the target is re-checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Avalanche Caller — target snow land becomes a 4/4 Elemental creature with hexproof and haste until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "c66dc0d3-b98c-40a2-be35-e455b4658828",
		Name:         "Avalanche Caller",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Target snow land you control becomes a 4/4 Elemental creature with hexproof and haste until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetPermanent("target snow land you control", effects.Land(), effects.Snow(), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				land := effects.FirstLegalBattlefieldTarget(ctx)
				if land == uuid.Nil {
					return nil
				}
				if err := (effects.BecomeCreatureUntilEOT{Target: land, Types: []string{"Creature"}, Subtypes: []string{"Elemental"}, SetPower: 4, SetToughness: 4, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: land, Keywords: []string{"hexproof", "haste"}, Label: label}.Apply(ctx)
			},
		}},
	})
}

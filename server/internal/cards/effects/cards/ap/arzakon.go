package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arzakon — Legendary Planeswalker — Arzakon {W}{U}{B}{R}{G}, loyalty 4
// (Mystery Booster Commander, #5948):
//
//	"+2: Arzakon deals 3 damage to any other target.
//	 −3: Create a Black Lotus token. (It's a {0} artifact with "{T},
//	 Sacrifice this token: Add three mana of any one color.")
//	 −7: Each player shuffles their hand and graveyard into their library,
//	 then draws seven cards.
//	 Arzakon can be your commander."
//
// The +2 targets any target but Arzakon himself (Another). The −7 is Echo of
// Eons' body: every player shuffles hand and graveyard into their library, then
// every player draws seven. The Black Lotus is a catalog token
// (gapB05BlackLotusToken). "Can be your commander" is deck-building and is
// read from the oracle text by the deck validator, not here.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a95d29b9-c30b-4e63-afcb-dec7ed522c5a",
		Name:         "Arzakon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Arzakon deals 3 damage to any other target.",
				Cost:    effects.LoyaltyCost(2),
				Targets: effects.Another(effects.TargetAny()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 3}.Apply(ctx)
				},
			},
			{
				Label:   "−3: Create a Black Lotus token.",
				Cost:    effects.LoyaltyCost(-3),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.GapB05BlackLotusToken(), N: 1}),
			},
			{
				Label:   "−7: Each player shuffles their hand and graveyard into their library, then draws seven cards.",
				Cost:    effects.LoyaltyCost(-7),
				Purpose: game.Purpose{Draws: 7},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.EchoOfEonsResolve(item, effects.NewContext(g, item))
				},
			},
		},
	})
}

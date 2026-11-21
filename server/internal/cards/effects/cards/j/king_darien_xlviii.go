package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// King Darien XLVIII — Legendary Creature — Human Soldier {1}{G}{W}, 2/3
// (Dominaria United):
//
//	"Other creatures you control get +1/+1.
//	 {3}{G}{W}: Put a +1/+1 counter on King Darien and create a 1/1 white
//	 Soldier creature token.
//	 Sacrifice King Darien: Creature tokens you control gain hexproof and
//	 indestructible until end of turn."
//
// The anthem is Angel of Invention's: a Layer 7c static scoped to OTHER
// creatures its controller controls. The activated ability does both halves
// in order on one resolution, the counter on King Darien (nothing if he has
// left, CR 608.2b) and the Soldier regardless. The sacrifice ability is Ainok
// Strike Leader's: the set of creature tokens is fixed as it resolves
// (CR 611.2c), so a token made afterwards gains neither keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e923d956-b036-4789-9311-4763a71fcbf8",
		Name:         "King Darien XLVIII",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}{G}{W}: Put a +1/+1 counter on King Darien and create a 1/1 white Soldier creature token.",
				Cost:    effects.ManaCost("{3}{G}{W}"),
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1}.Apply(ctx)
				},
			},
			{
				Label:   "Sacrifice King Darien: Creature tokens you control gain hexproof and indestructible until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive, Saves: &game.Saves{Matches: game.SweepCreatures, Partial: true}},
				Cost:    game.AbilityCost{SacrificeSelf: true},
				Effect: effects.Do(effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.IsTokenPredicate(), effects.YouControl()),
					Keywords: []string{"hexproof", "indestructible"},
					Label:    "King Darien XLVIII — hexproof and indestructible",
				}),
			},
		},
	})
}

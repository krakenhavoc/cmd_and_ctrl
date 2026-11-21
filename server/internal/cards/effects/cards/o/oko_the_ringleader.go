package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oko, the Ringleader — Legendary Planeswalker — Oko {2}{G}{U}, loyalty 3
// (Outlaws of Thunder Junction, #4468):
//
//	"At the beginning of combat on your turn, Oko becomes a copy of up to
//	 one target creature you control until end of turn, except he has
//	 hexproof.
//	 +1: Draw two cards. If you've committed a crime this turn, discard a
//	 card. Otherwise, discard two cards.
//	 −1: Create a 3/3 green Elk creature token.
//	 −5: For each other nonland permanent you control, create a token
//	 that's a copy of that permanent."
//
// The combat copy is Oko, the Trickster's (BecomeCopy of a creature, until
// end of turn) with the printed exception that the copy also has hexproof;
// a copied creature is no planeswalker for the turn, so it has no loyalty
// abilities (the ruling). The +1 reads the crime tally as it resolves and
// asks for the discard only after the draws have settled (drawThenDiscard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4372b729-dc3d-4469-8fc5-811cc787d653",
		Name:            "Oko, the Ringleader",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat("Oko, the Ringleader — becomes a copy of up to one target creature you control until end of turn, except he has hexproof",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if effects.SourceIsNewObject(g, item) || !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					return effects.BecomeCopy{
						Targets: []uuid.UUID{item.SourceCardID},
						Of:      ts[0].ID,
						Except:  func(v *game.PrintedValues) { v.AddKeyword("hexproof") },
						Label:   "Oko, the Ringleader — becomes a copy",
					}.Apply(ctx)
				}),
				effects.TargetCreature("up to one target creature you control", effects.YouControl()).WithCount(0, 1)),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Draw two cards. If you've committed a crime this turn, discard a card. Otherwise, discard two cards.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 2},
				Effect: func(g *game.Game, item *game.StackItem) error {
					n := 2
					if effects.YouCommittedACrimeThisTurn(g, item.Controller) {
						n = 1
					}
					return effects.DrawThenDiscard(g, item.Controller, item.SourceCardID, 2, n, "")
				},
			},
			{
				Label:   "−1: Create a 3/3 green Elk creature token.",
				Cost:    effects.LoyaltyCost(-1),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.B35ElkToken(), N: 1}),
			},
			{
				Label:   "−5: For each other nonland permanent you control, create a token that's a copy of that permanent.",
				Cost:    effects.LoyaltyCost(-5),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  effects.B35TokenCopyOfEachOtherNonland,
			},
		},
	})
}

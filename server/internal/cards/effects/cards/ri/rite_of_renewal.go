package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rite of Renewal — Sorcery {3}{G} (Tarkir: Dragonstorm, #4463):
//
//	"Return up to two target permanent cards from your graveyard to your
//	 hand. Target player shuffles up to four target cards from their
//	 graveyard into their library. Exile Rite of Renewal."
//
// Three target clauses in printed order: up to two permanent cards of yours
// (slot 0), a player (slot 1), and up to four cards from THAT player's
// graveyard (slot 2, Quandrix Command's FromTargetPlayersGraveyard). The
// sentences resolve in order. The return takes each card still a permanent
// card in your graveyard; one that left stays where it went (CR 608.2b). The
// shuffle is Quandrix Command's TargetPlayerShufflesTheirCards: an illegal
// player does nothing at all, a legal one shuffles their library even when no
// card was picked or every one has left. The last sentence exiles the spell
// itself instead of letting it go to the graveyard (Burning Wish's
// instruction), after both halves; a spell countered or fizzled for illegal
// targets is not exiled, because it never resolved.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6512ddb8-0868-4d32-ab32-8a5250fc571b",
		Name:         "Rite of Renewal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("up to two target permanent cards from your graveyard", effects.YouOwn(), effects.Permanent()).WithCount(0, 2),
			effects.TargetPlayer("target player"),
			effects.FromTargetPlayersGraveyard(
				effects.TargetCardInGraveyard("up to four target cards from their graveyard").WithCount(0, 4), 1),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.ClauseTargets(0) {
					if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
						continue
					}
					if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.TargetPlayerShufflesTheirCards(ctx, ctx.Item.Targets, 1, 2)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				// "Exile Rite of Renewal." The spell moves itself, so the
				// resolution frame does not put it into the graveyard (#489).
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Perch Protection — Instant {4}{W}{W}:
//
//	"Gift an extra turn (You may promise an opponent a gift as you cast
//	 this spell. If you do, they take an extra turn after this one.)
//	 Create four 2/2 blue Bird creature tokens with flying. If the gift
//	 was promised, all permanents you control phase out, and until your
//	 next turn, your life total can't change and you gain protection
//	 from everything.
//	 Exile Perch Protection."
//
// The gift is GiftAnExtraTurn (CR 702.174g), given before anything else
// the spell does (CR 702.174j). The four Birds are made first, so a
// promised gift phases them out with the rest: the list is every
// permanent the caster controls, snapshotted after the tokens exist and
// phased out together (CR 702.26), as Teferi's Protection does. The life
// lock and the protection last until the caster's next turn (CR 611.2b).
// The spell exiles itself in every case (CR 608.2n), as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "295fe873-ecb5-4d5e-a73c-3af6c478afc5",
		Name:         "Perch Protection",
		Completeness: CompletenessFull,
		Gift:         GiftAnExtraTurn(),
		Purpose:      game.Purpose{Tokens: 4, CreatureTokens: 4, Answers: game.AnswerSurvive | game.AnswerProtect},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			me := ctx.Controller()
			if err := (CreateToken{Template: TokenCard("2/2 blue Bird with flying"), N: 4}).Apply(ctx); err != nil {
				return err
			}
			if ctx.GiftPromised() {
				var mine []uuid.UUID
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller == me {
						mine = append(mine, c.InstanceID)
					}
				}
				if err := (PhaseOut{Targets: mine}).Apply(ctx); err != nil {
					return err
				}
				if err := (LockLifeTotal{
					Player:   me,
					Label:    "Perch Protection — your life total can't change",
					Duration: DurationUntilYourNextTurn(ctx, me),
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (GainPlayerKeyword{
					Player:   me,
					Keyword:  ProtectionFromEverything,
					Label:    "Perch Protection — protection from everything",
					Duration: DurationUntilYourNextTurn(ctx, me),
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}

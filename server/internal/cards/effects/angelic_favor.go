package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angelic Favor — Instant {3}{W}:
//
//	"If you control a Plains, you may tap an untapped creature you
//	 control rather than pay this spell's mana cost.
//	 Cast this spell only during combat.
//	 Create a 4/4 white Angel creature token with flying. Exile it at
//	 the beginning of the next end step."
//
// ADR 0135 §1 (#2030): the tap alternative cost (CR 118.9), offered only
// while you control a Plains. The timing restriction holds whichever cost
// is paid, as Mandate of Peace's does. The exile is a CR 603.7 delayed
// trigger over exactly the token this spell made (Kindred Charge's
// shape); a token that has left by then is not exiled.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "37162c6e-ac31-4386-9d1e-76aa66070b35",
		Name:         "Angelic Favor",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			TapInstead(1, "an untapped creature you control", ControlsA("Plains"), Creature()),
		},
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return game.PhaseOf(g.Turn.Step) == game.PhaseCombat
		},
		CastConditionLabel: "Cast this spell only during combat.",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			cursor := b25LastEventSeq(ctx.Game)
			if err := (CreateToken{Controller: item.Controller, Template: TokenCard("4/4 white Angel with flying"), N: 1}).Apply(ctx); err != nil {
				return err
			}
			tokens := b27TokensCreatedByAfter(ctx.Game, item.Controller, cursor)
			if len(tokens) == 0 {
				return nil
			}
			return ScheduleDelayedTrigger{
				Label: "Angelic Favor — exile the Angel",
				Cards: tokens,
				Body:  exileListedCardsBody,
			}.Apply(ctx)
		},
	})
}

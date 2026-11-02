package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spider-Man 2099 — Legendary Creature — Spider Human Hero {U}{R}, 2/3
// (Marvel's Spider-Man, #4460):
//
//	"From the Future — You can't cast Spider-Man 2099 during your first,
//	 second, or third turns of the game.
//	 Double strike, vigilance
//	 At the beginning of your end step, if you've played a land or cast a spell
//	 this turn from anywhere other than your hand, Spider-Man 2099 deals damage
//	 equal to his power to any target."
//
// The cast restriction is Starting Town's "your first three turns" read
// (TurnsBegunFor) as a CastCondition: it refuses only on the caster's own
// turns 1-3, so a flash-style cast on someone else's turn is not touched. The
// end-step trigger is an intervening "if" (CR 603.4): checked as the step
// begins and again as it resolves, by scanning the turn's events for a spell
// cast from a zone other than the hand (Vega, the Watcher's read) or a land
// played from one (b20LandPlayed). The damage is Burning Anger's: his power as
// he last had it (CR 608.2h), with him as the source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "19845456-3d96-48a4-89fa-9b8e583774c9",
		Name:            "Spider-Man 2099",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", "vigilance"},
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			mine := g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) &&
				g.Seats[g.Turn.ActiveSeat].ID == controller
			return !(mine && g.TurnsBegunFor(controller) <= 3)
		},
		CastConditionLabel: "From the Future — You can't cast Spider-Man 2099 during your first, second, or third turns of the game.",
		Triggered: []game.TriggeredAbility{
			Targeting(InterveningIf(CondBoard{
				Name: "you've played a land or cast a spell this turn from anywhere other than your hand",
				Test: func(ctx *Context) bool { return spm8PlayedFromOutsideHand(ctx.Game, ctx.Controller()) },
			}, On(game.EventBeginEndStep, AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return spm8PlayedFromOutsideHand(g, source.Controller)
			}), "Spider-Man 2099 — deals damage equal to his power to any target", sourceDealsDamageEqualToItsPowerToTarget)),
				TargetAny()),
		},
	})
}

// spm8PlayedFromOutsideHand is "you've played a land or cast a spell this turn
// from anywhere other than your hand": the turn's events hold a cast whose
// origin zone was not the hand, or a land play from a zone that was not. A cast
// or play with no recorded origin counts as a hand play, the answer that cannot
// over-fire.
func spm8PlayedFromOutsideHand(g *game.Game, player uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Actor != player || ev.OldZone == "" || ev.OldZone == game.ZoneHand {
			continue
		}
		if ev.Kind == game.EventCast || b20LandPlayed(ev, g) {
			return true
		}
	}
	return false
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b488_helpers.go — shared shapes of S82 OLD batch 488 (#5582):
// Worldwake's Traps and Allies.

// S82OldB488AnOpponentHadAGreenCreatureEnter is Permafrost Trap's condition,
// "an opponent had a green creature enter the battlefield under their control
// this turn": one opponent's entry tally by colour (game.EnteredCreatureOfColorThisTurn),
// recorded as each creature entered, so a token counts and a creature that has since
// left or changed colour still counts.
func S82OldB488AnOpponentHadAGreenCreatureEnter(g *game.Game, controller uuid.UUID) bool {
	for _, p := range g.Seats {
		if p == nil || p.ID == controller {
			continue
		}
		if g.EnteredCreatureOfColorThisTurn(p.ID, "G") > 0 {
			return true
		}
	}
	return false
}

// S82OldB488AnOpponentCastABlueSpell is Ricochet Trap's condition, "an opponent
// cast a blue spell this turn": the colours of the spells each opponent cast this turn,
// as each became cast (game.CastTally).
func S82OldB488AnOpponentCastABlueSpell(g *game.Game, controller uuid.UUID) bool {
	for _, p := range g.Seats {
		if p == nil || p.ID == controller {
			continue
		}
		for _, col := range g.CastTallyFor(p.ID).SpellColorsCast() {
			if col == "U" {
				return true
			}
		}
	}
	return false
}

// S82OldB488ABlackCreatureWithFlyingIsAttacking is Slingbow Trap's condition, "a
// black creature with flying is attacking": any attacking creature, whoever controls it,
// that is black and has flying as the layers leave it now.
func S82OldB488ABlackCreatureWithFlyingIsAttacking(g *game.Game, _ uuid.UUID) bool {
	for _, id := range AttackingCreatures(g) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if c.HasColorWord("B") && game.HasKeyword(&c, "flying") {
			return true
		}
	}
	return false
}

// S82OldB488ExileTokensAtYourNextEndStep is a token creation's keyed tail that
// schedules "exile it at the beginning of YOUR next end step" over the tokens that
// landed (CR 603.7): SocB8ExileTokensAtNextEndStep for the printing that says "your",
// so a trap cast on an opponent's turn waits for your own end step
// (ScheduleDelayedTrigger.ControllerTurnOnly). An on-disk key: never renamed, never reused.
func S82OldB488ExileTokensAtYourNextEndStep(key, label string) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		if len(o.IDs) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:                 game.StepEnd,
			Controller:         ctx.Controller(),
			ControllerTurnOnly: true,
			Label:              label,
			Cards:              append([]uuid.UUID(nil), o.IDs...),
			Body:               ExileListedCardsBody,
		}.Apply(ctx)
	})
}

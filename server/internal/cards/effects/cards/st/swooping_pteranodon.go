package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swooping Pteranodon — Creature — Dinosaur {3}{R}{W}, 3/3 (Jurassic World
// Collection, #5948):
//
//	"Flying, haste
//	 Whenever this creature or another Dinosaur you control with flying
//	 enters, gain control of target creature an opponent controls until end
//	 of turn. Untap that creature. It gains flying and haste until end of
//	 turn. At the beginning of the next end step, target land deals 3 damage
//	 to that creature."
//
// One trigger with two target clauses (CR 115.1): the creature and, as the
// trigger goes on the stack, the land. Act of Treason's body on the creature
// (control, untap, flying and haste until end of turn), then a CR 603.7
// delayed trigger over the land and the creature whose body has the land deal
// 3 damage to the creature at the next end step, if both are still on the
// battlefield. The entering Dinosaur's flying is read through the layered
// view as it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0b1f6e77-55db-4638-9030-3f4e4d37eb1b",
		Name:            "Swooping Pteranodon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventETB && ev.CardID == source.InstanceID {
						return true
					}
					c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
					return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Dinosaur")) && game.HasKeyword(&c, "flying")
				}, "Swooping Pteranodon — gain control of target creature an opponent controls until end of turn, untap it, it gains flying and haste; at the next end step target land deals 3 damage to it",
					swoopingPteranodonSteal),
				effects.Clauses(
					effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
					effects.TargetPermanent("target land", effects.Land()),
				)),
		},
	})
}

func swoopingPteranodonSteal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if len(item.Targets) < 2 || !ctx.IsTargetLegal(item.Targets[0]) {
		return nil
	}
	creature := item.Targets[0].ID
	if err := (effects.GainControl{Target: creature, Duration: effects.DurationUntilEndOfTurn(ctx), Label: "Swooping Pteranodon — gain control until end of turn"}).Apply(ctx); err != nil {
		return err
	}
	if err := (effects.UntapTarget{Target: creature}).Apply(ctx); err != nil {
		return err
	}
	if err := (effects.GrantKeywordUntilEOT{Target: creature, Keywords: []string{"flying", "haste"}, Label: "Swooping Pteranodon — flying and haste until end of turn"}).Apply(ctx); err != nil {
		return err
	}
	if !ctx.IsTargetLegal(item.Targets[1]) {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		Label: "Swooping Pteranodon — target land deals 3 damage to that creature",
		Cards: []uuid.UUID{item.Targets[1].ID, creature},
		Body:  effects.GapB82LandDamagesCreatureBody,
	}.Apply(ctx)
}

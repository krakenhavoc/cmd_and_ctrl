package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Whippoorwill — Creature — Bird {G}, 1/1:
//
//	"{G}{G}, {T}: Target creature can't be regenerated this turn. Damage
//	 that would be dealt to that creature this turn can't be prevented or
//	 dealt instead to another permanent or player. When the creature dies
//	 this turn, exile the creature."
//
// Three effects on the one target, all until cleanup and all pinned to
// the object (CR 400.7), in printed order:
//
//   - "can't be regenerated this turn" is ADR 0108 §2's cantBeRegenerated
//     mark (CR 701.19c);
//   - the damage clause is ADR 0107 §5's pinned damageCantBePrevented +
//     damageCantBeRedirected pair (CR 615.12);
//   - "When the creature dies this turn, exile the creature" is an
//     event-delayed trigger (CR 603.7), not a replacement: the creature
//     dies, its dies triggers see it, and then the trigger exiles the
//     card from the graveyard if it is still there (CR 603.7c).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "84050a10-e1f1-413e-aa21-5c1f47bb2a64",
		Name:         "Whippoorwill",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{G}{G}, {T}: Target creature can't be regenerated this turn. Damage that would be dealt to that creature this turn can't be prevented or dealt instead to another permanent or player. When the creature dies this turn, exile the creature.",
			Cost:    Plus(ManaCost("{G}{G}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  whippoorwillMark,
		}},
	})
}

func whippoorwillMark(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (CantBeRegeneratedThisTurn{Target: id, Label: "Whippoorwill — can't be regenerated this turn"}).Apply(ctx); err != nil {
		return err
	}
	g.DamageToCantBePreventedThisTurnForEffect(ctx.Source(), id, true,
		"Whippoorwill — damage to it can't be prevented or redirected this turn")
	ExileWhenItDiesThisTurn(ctx, id, "Whippoorwill — exile the creature")
	return nil
}

// ExileWhenItDiesThisTurn schedules "When <id> dies this turn, exile
// it" (Whippoorwill): an event-delayed trigger on the object, controlled
// by the resolving object's controller, until cleanup and pinned to the
// object so a flicker ends it (CR 400.7).
func ExileWhenItDiesThisTurn(ctx *Context, id uuid.UUID, label string) {
	d := ctx.Game.PinnedTo(ctx.Game.UntilEndOfTurnDuration(), id)
	ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   ctx.Controller(),
		SourceCardID: ctx.Source(),
		Label:        label,
		On:           []game.EventKind{game.EventLTB},
		Condition:    theListedObjectDiedCondition,
		Cards:        []uuid.UUID{id},
		Duration:     &d,
		Body:         exileListedFromGraveyardBody,
	})
}

// theListedObjectDied is the condition of "when the creature dies": the
// listed object left the battlefield for a graveyard (CR 700.4).
func theListedObjectDied(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
	return ev.Kind == game.EventLTB && len(dt.Cards) > 0 && ev.CardID == dt.Cards[0] && ev.NewZone == game.ZoneGraveyard
}

// exileListedCardsFromGraveyard exiles each listed card that is still in
// a graveyard. One that has left it is not affected (CR 603.7c).
func exileListedCardsFromGraveyard(g *game.Game, item *game.StackItem) error {
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := g.ExileCardForEffect(t.ID); err != nil {
			return err
		}
	}
	return nil
}

package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lynde, Cheerful Tormentor — Legendary Creature — Human Warlock
// {1}{U}{B}{R}, 2/4:
//
//	"Deathtouch
//	 Whenever a Curse is put into your graveyard from the battlefield,
//	 return it to the battlefield attached to you at the beginning of the
//	 next end step.
//	 At the beginning of your upkeep, you may attach a Curse attached to
//	 you to one of your opponents. If you do, draw two cards."
//
// THE RETURN is a trigger that schedules a delayed trigger (CR 603.7).
// "Your graveyard" is the graveyard of its owner, who must be Lynde's
// controller, and the Curse is read as it last existed on the battlefield
// (CR 603.10a: a Curse, put there from the battlefield). When the trigger
// resolves it names the card as the object now in the graveyard (CR 400.7,
// 603.7c): a Curse that has left that graveyard by the next end step is
// not followed. At that end step the Curse enters attached to Lynde's
// controller (ReturnFromGraveyardTogether.AttachedTo), under its owner's
// control as ever, the player named with no targeting (an enchant-player
// restriction is a state-based one, CR 303.4f).
//
// THE UPKEEP ABILITY is "you may", then a choice of which Curse attached to
// you and which opponent, one question over every legal pair (so the
// answer is never an illegal combination), and the two cards are drawn only
// if the Curse really ended up on that opponent (CR 701.3b: an attach the
// Curse can't legally make does nothing). The ability reads the Curses
// attached to you as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dc386f6c-bd4a-462a-8d42-0e7a552a4f17",
		Name:            "Lynde, Cheerful Tormentor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.Owner == source.Controller && effects.LeftAsSubtype(ev, c, "Curse")
			}, "Lynde — return the Curse to the battlefield attached to you at the beginning of the next end step",
				effects.Do(effects.Instruction(lyndeScheduleReturn))),
			effects.Optional(effects.AtYourUpkeep("Lynde — you may attach a Curse attached to you to one of your opponents; if you do, draw two cards",
				effects.Do(effects.Instruction(lyndeAttachOffer))),
				"Attach a Curse attached to you to one of your opponents? If you do, draw two cards."),
		},
	})
}

// lyndeScheduleReturn schedules the end-step return of the Curse that was
// put into the graveyard, naming it as the object it now is.
func lyndeScheduleReturn(ctx *effects.Context) error {
	id := ctx.Item.Trigger.Event.CardID
	z := ctx.Game.FindCardZoneForEffect(id)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:     game.StepEnd,
		Label:  "Lynde — return the Curse to the battlefield attached to you",
		Cards:  []uuid.UUID{id},
		Body:   lyndeReturnBody,
		Params: game.EffectParams{Objects: []game.ObjectRef{{ID: id, Epoch: c.ObjectEpoch}}},
	}.Apply(ctx)
}

// lyndeReturnBody is the delayed return: the Curse, if it is still the same
// object in a graveyard, enters attached to the delayed ability's controller.
var lyndeReturnBody = game.DelayedBody("s82-mid-b14/lynde-return-curse", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	ctx := effects.NewContext(g, item)
	for _, ref := range p.Objects {
		z := g.FindCardZoneForEffect(ref.ID)
		if z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		c, ok := g.LookupCardForEffect(ref.ID)
		if !ok || c.ObjectEpoch != ref.Epoch {
			continue
		}
		if err := (effects.ReturnFromGraveyardTogether{
			Targets:    []uuid.UUID{ref.ID},
			AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: item.Controller},
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// lyndeAttachOffer asks which of the Curses attached to you goes to which
// opponent, over every pair; the first option declines.
func lyndeAttachOffer(ctx *effects.Context) error {
	me := ctx.Controller()
	opponents := ctx.Opponents()
	options := []game.ChoiceOption{{Label: "Don't attach a Curse"}}
	var carry []uuid.UUID
	for _, c := range ctx.Game.Battlefield.Cards {
		if c.AttachedTo.Kind != game.TargetPlayer || c.AttachedTo.ID != me || !c.HasSubtype("Curse") {
			continue
		}
		for _, opp := range opponents {
			name := ""
			if p := ctx.Game.PlayerByIDForEffect(opp); p != nil {
				name = p.Name
			}
			options = append(options, game.ChoiceOption{Label: "Attach " + c.Name + " to " + name, Cards: []uuid.UUID{c.InstanceID}, Player: opp})
			carry = append(carry, c.InstanceID, opp)
		}
	}
	if len(carry) == 0 {
		return nil
	}
	return effects.PickOption{
		Question: "Lynde — attach a Curse attached to you to one of your opponents?",
		Options:  options,
		ThenKey:  lyndeAttached,
		Carry:    carry,
	}.Apply(ctx)
}

// lyndeAttached attaches the chosen Curse and draws two cards if it landed
// on that opponent. An on-disk key.
var lyndeAttached = effects.OptionPickThen("option-pick/s82-mid-b14-lynde-attach", func(ctx *effects.Context, r game.OptionPicked) error {
	i := (r.Index - 1) * 2
	if r.Option == nil || r.Index < 1 || i+1 >= len(r.Carry) {
		return nil
	}
	curse, opp := r.Carry[i], r.Carry[i+1]
	me := ctx.Controller()
	c, ok := ctx.Game.LookupCardForEffect(curse)
	if !ok || !effects.OnBattlefield(ctx.Game, curse) || c.AttachedTo.Kind != game.TargetPlayer || c.AttachedTo.ID != me {
		return nil
	}
	host := game.TargetRef{Kind: game.TargetPlayer, ID: opp}
	if err := ctx.Game.AttachForEffect(curse, host); err != nil {
		return err
	}
	if after, ok := ctx.Game.LookupCardForEffect(curse); !ok || after.AttachedTo != host {
		return nil
	}
	return effects.DrawCards{Player: me, N: 2}.Apply(ctx)
})

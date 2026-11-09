package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b24_helpers.go — shared pieces of Standard batch 24 (Duskmourn,
// #4466). Everything here is prefixed b24.

// b24CondDelirium is the intervening "if there are four or more card types
// among cards in your graveyard" (delirium, CR 207.2c) as a Condition, for
// InterveningIf and If. It declares no footprint: the graveyard's card types
// are not a state the trigger-independence analysis models, so a row that
// carries it is simply not reordered.
type b24CondDelirium struct{}

func (b24CondDelirium) Holds(ctx *Context) bool {
	return b16CardTypesInGraveyard(ctx.Game, ctx.Controller()) >= 4
}

func (b24CondDelirium) condition() (game.FootprintStep, bool) { return game.FootprintStep{}, false }

// b24CreatureThatPlayerControls is the TargetsFrom clause "target creature
// that player controls" for a trigger about damage dealt to a player: "that
// player" is the player the triggering event names (Fear of Burning Alive).
func b24CreatureThatPlayerControls(tc game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
	player := tc.Event.Target
	if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
		return nil
	}
	return &game.TargetSpec{
		Mode: "creature", Label: "target creature that player controls",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsCreature() && c.Controller == player
		},
		Min: 1, Max: 1,
	}
}

// b24DamageToTargetEqualToEvent is "this [permanent] deals that amount of
// damage to target [creature]": the triggering event's damage, dealt by the
// item's source (as it last existed if it has left, CR 608.2h) to the first
// target still legal as the ability resolves (CR 608.2b).
func b24DamageToTargetEqualToEvent(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: item.Trigger.Event.Amount}.Apply(ctx)
	}
	return nil
}

// b24DiscardACardOrPay is the either payment "discard a card or pay <mana>"
// for MayPay.Pay (Anthropede): one prompt whose answer names the branch. The
// engine's either payment carries a mana branch (game.PayActionMana, #4566;
// WardEither builds the same one), but effects.EitherPayment refuses it, so
// the two branches are assembled here. A branch the chooser can't pay is not
// offered (CR 118.3), and with neither payable nobody is asked.
func b24DiscardACardOrPay(mana string) UpkeepPayment {
	discard := game.PayAction{Kind: game.PayActionDiscard, Count: 1, Words: "discard a card"}
	pay := game.PayAction{Kind: game.PayActionMana, Mana: mana, Words: "pay " + mana}
	return UpkeepPayment{
		Action: &game.PayAction{Kind: game.PayActionEither, Either: []game.PayAction{discard, pay}},
		Words:  "Discard a card or pay " + mana,
	}
}

// b24SelfTargetedByAnOpponentOnce is "whenever this permanent becomes the
// target of a spell or ability an opponent controls" for a source whose
// ruling says the ability triggers once however many times that spell or
// ability targets it (Altanak). The engine emits one EventBecomesTarget per
// target slot (CR 115.3), so a later slot of the same stack item that names
// the same object is declined: only the first of the consecutive events
// naming this stack item and this object passes.
func b24SelfTargetedByAnOpponentOnce(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || source == nil || ev.CardID != source.InstanceID {
		return false
	}
	if ev.Actor == uuid.Nil || ev.Actor == source.Controller || g.PlayerByIDForEffect(ev.Actor) == nil {
		return false
	}
	for i := len(g.Events) - 1; i >= 0; i-- {
		earlier := g.Events[i]
		if earlier.Seq >= ev.Seq {
			continue
		}
		switch earlier.Kind {
		case game.EventBecomesTarget:
			if earlier.StackItemID == ev.StackItemID && earlier.CardID == ev.CardID {
				return false
			}
		case game.EventTrigger, game.EventTriggerPlaced:
			continue
		default:
			return true
		}
	}
	return true
}

// b24SpellTargetsOnlyOneOfYourCreatures is "a spell that targets only a
// single creature you control" (Leyline of Resonance; CR 115.9c): the cast
// spell's targets, however many instances of the word "target" it has, are
// one creature on the battlefield controlled by the source's controller.
func b24SpellTargetsOnlyOneOfYourCreatures(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || source == nil {
		return false
	}
	sole, ok := g.StackItemSoleTargetForEffect(ev.CardID)
	if !ok || sole.Kind != game.TargetCard {
		return false
	}
	c, found := g.LookupCardForEffect(sole.ID)
	return found && c.IsCreature() && onBattlefield(g, sole.ID) && c.Controller == source.Controller
}

var (
	// "Whenever a creature you control dies" as a delayed trigger's event
	// condition (Waltz of Rage): a creature that was under the delayed
	// ability's controller's control as it left the battlefield for a
	// graveyard.
	b24ACreatureYouControlDiedCondition = game.DelayedCondition("dies/a-creature-you-control", b24ACreatureYouControlDied)

	// "Exile the top card of your library. You may play it until the end of
	// your next turn." as a delayed trigger's body (Waltz of Rage).
	b24ExileTopYouMayPlayUntilNextTurnBody = game.SimpleDelayedBody("exile-top/play-until-end-of-your-next-turn",
		func(g *game.Game, item *game.StackItem) error {
			return ExileTopNUntilYourNextTurn(NewContext(g, item), 1)
		})
)

func b24ACreatureYouControlDied(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
	dead, ok := diedCreature(ev, g)
	return ok && leftUnderControlOf(ev, dead) == dt.Controller
}

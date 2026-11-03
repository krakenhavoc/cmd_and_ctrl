package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b10_helpers.go — shared pieces of Secrets of Strixhaven Commander,
// batch 10 (#4834).

// socB10EnchantedCreatureControllerLoses is "its controller loses n life" for
// an Aura's attack trigger: the creature that attacked is named by the
// trigger's event. Its controller is read as the trigger resolves (CR 608.2h),
// last-known when the creature has left, and falls back to the attacking
// player the event names.
func socB10EnchantedCreatureControllerLoses(n int) func(ctx *Context) error {
	return func(ctx *Context) error {
		victim := uuid.Nil
		if ctx.Item != nil && ctx.Item.Trigger != nil {
			victim = ctx.Item.Trigger.Event.Actor
		}
		if info, ok := ctx.TriggeringPermanent(); ok {
			victim = info.Controller
		}
		if victim == uuid.Nil {
			return nil
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -n)
	}
}

// socB10PestTokenYouControl reports whether the player controls a creature token
// of the subtype `pest`, which is a word of Pest Rescuer's text (CR 612.2).
func socB10PestTokenYouControl(g *game.Game, player uuid.UUID, pest string) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && IsToken(c) && c.IsCreature() && c.HasSubtype(pest) {
			return true
		}
	}
	return false
}

// socB10EnchantmentSpell passes on an enchantment spell (Pearl-Ear's
// affinity).
func socB10EnchantmentSpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.IsEnchantment() }
}

// socB10AuraSpellTargetsYourModifiedPermanent is "whenever you cast an Aura
// spell that targets a modified permanent you control": the cast is by the
// source's controller, the spell is an Aura, and at least one target it
// announced (CR 601.2c) is a permanent on the battlefield under that player's
// control that is modified (CR 700.9, b07IsModified).
func socB10AuraSpellTargetsYourModifiedPermanent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !spell.IsAura() {
		return false
	}
	item := g.StackItemForEffect(ev.CardID)
	if item == nil || item.Kind != game.StackItemSpell {
		return false
	}
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || !onBattlefield(g, t.ID) {
			continue
		}
		if c, ok := g.LookupCardForEffect(t.ID); ok && c.Controller == source.Controller && b07IsModified(g, c) {
			return true
		}
	}
	return false
}

// socB10LookThreeTakeOne is "look at the top three cards of your library. You
// may reveal a [match] card from among them and put it into your hand. Put the
// rest on the bottom of your library in any order." (Duskwatch Recruiter's
// body, Quandrix Apprentice's magecraft.)
func socB10LookThreeTakeOne(match CardPredicate, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		player := ctx.Controller()
		return TakeFromLibraryToHand{
			Player:   player,
			Cards:    g.LookAtTopOfLibraryForEffect(player, 3),
			Match:    match,
			Max:      1,
			Optional: true,
			Label:    label,
			Key:      takeRevealRestBottomAny,
		}.Apply(ctx)
	}
}

// socB10BasePowerZero is "a creature with base power 0" (CR 208.4b): the power
// the layer pass had after layer 7b, without a pump, an anthem or a counter.
func socB10BasePowerZero(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.BasePower() == 0
}

// socB10BasePowerZeroDamage is the combat damage the trigger controller's
// creatures with base power 0 dealt the damaged player in the damage step that
// fired the trigger (Primo, the Unbounded's "the damage dealt"). It sums the
// batch's damage events to that player (CR 510.2 deals them at the same time),
// counting a creature whose base power is 0 as it stands, or as it last stood
// when it has left (CR 608.2h). A pump on the damage dealt is in the event's
// amount, not in the creature's base power.
func socB10BasePowerZeroDamage(ctx *Context) int {
	if ctx.Item == nil || ctx.Item.Trigger == nil {
		return 0
	}
	trigger := ctx.Item.Trigger.Event
	total := 0
	for _, ev := range ctx.Game.EventsThisTurn() {
		if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 ||
			ev.Batch != trigger.Batch || ev.Target != trigger.Target {
			continue
		}
		ref, ok := ctx.Game.PermanentRefForEffect(ev.Source)
		if !ok {
			continue
		}
		info, ok := ctx.Game.PermanentForEffect(ref)
		if !ok || info.Controller != ctx.Controller() || info.BasePower != 0 {
			continue
		}
		total += ev.Amount
	}
	return total
}

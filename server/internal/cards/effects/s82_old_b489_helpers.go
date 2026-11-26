package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b489_helpers.go — Worldwake batch 489 (S82 older sets, #5582).

// S82OldB489AnyLand is the payment match of "return a land you control to
// its owner's hand": every land is one.
func S82OldB489AnyLand(_ *game.Card, _ game.Card) bool { return true }

// UpkeepTrigger is the "At the beginning of your upkeep, sacrifice this
// <permanent> unless you return a land you control to its owner's hand"
// trigger (Vapor Snare): the choice is made as it resolves, like the enters
// form.
func (r *returnLandsOrSacrifice) UpkeepTrigger(label string) game.TriggeredAbility {
	return AtYourUpkeep(label, r.enters)
}

// S82OldB489ADefenderYouControlBlocked is "whenever a creature you control
// with defender blocks" (Perimeter Captain): one trigger per blocking
// creature (CR 509.3a), the blocker read through the layers as it blocks.
func S82OldB489ADefenderYouControlBlocked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBlock || ev.CardID == uuid.Nil || ev.Amount > 1 {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller && c.IsCreature() && game.HasKeyword(&c, "defender")
}

// s82OldB489NibblersLand is the "If it's a land card, you gain 1 life" of Scrib
// Nibblers' exile: gated on the card having reached exile (CR 400.7), with
// "is it a land card" answered before it moved (CR 608.2h).
var s82OldB489NibblersLand = IfItWasTail("wwk/scrib-nibblers-exiled-a-land-card",
	func(ctx *Context) error { return GainLife{Amount: 1}.Apply(ctx) }, nil)

// S82OldB489ExileTopCardOfTargetPlayerGainIfLand is "Exile the top card of
// target player's library. If it's a land card, you gain 1 life" (Scrib
// Nibblers): the first still-legal player target; an empty library exiles
// nothing and gains nothing.
func S82OldB489ExileTopCardOfTargetPlayerGainIfLand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		p := g.PlayerByIDForEffect(t.ID)
		if p == nil || p.Library == nil || len(p.Library.Cards) == 0 {
			return nil
		}
		// The library's top is the LAST element.
		top := p.Library.Cards[len(p.Library.Cards)-1]
		return ExileThenIfItWas{
			Target: top.InstanceID,
			Was:    func(c game.Card) bool { return c.IsLand() },
			Keyed:  s82OldB489NibblersLand,
		}.Apply(ctx)
	}
	return nil
}

// S82OldB489WheneverThisDealsCombatDamageToAPlayerIfYouControl is "Whenever
// this creature deals combat damage to a player, if you control a <subtype>,
// <effect>" (Slavering Nulls): an intervening "if" (CR 603.4), checked as the
// damage is dealt and again as the trigger resolves, with the subtype read
// through the source's text (CR 612.2).
func S82OldB489WheneverThisDealsCombatDamageToAPlayerIfYouControl(subtype, label string, effect Effect) game.TriggeredAbility {
	t := WheneverThisDealsCombatDamageToAPlayer(label, effect)
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return when(ev, source, lki, g) && ControlsA(source.TextSubtype(subtype))(g, source.Controller)
	}
	return InterveningIf(CondBoard{
		Name: "you control a " + subtype,
		Test: func(ctx *Context) bool { return ControlsA(ctx.TextSubtype(subtype))(ctx.Game, ctx.Controller()) },
	}, t)
}

// S82OldB489KickedTimesOfCast is the number of times the spell `id`, as
// it was cast, was kicked (CR 702.33d).
func S82OldB489KickedTimesOfCast(g *game.Game, id uuid.UUID) int {
	spell, ok := g.LookupCardForEffect(id)
	if !ok {
		return 0
	}
	return game.KickedTimesPaid(spell, g.StackItemPaidForEffect(id).OptionalCosts)
}

// S82OldB489WheneverYouCastAKickedSpell is "Whenever you cast a kicked
// spell" (Rumbling Aftershocks): the trigger is built only for a spell
// that was kicked, and carries the number of times in Params.Amount.
func S82OldB489WheneverYouCastAKickedSpell(label string, effect Effect) game.TriggeredAbility {
	t := On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Kind == game.EventCast && ev.Actor == source.Controller && S82OldB489KickedTimesOfCast(g, ev.CardID) > 0
	}, label, effect)
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = S82OldB489KickedTimesOfCast(g, ev.CardID)
		return item
	}
	return t
}

// S82OldB489DamageToTargetEqualToTimesKicked is the damage of Rumbling
// Aftershocks: the number of times the triggering spell was kicked.
func S82OldB489DamageToTargetEqualToTimesKicked(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: item.Params.Amount}.Apply(ctx)
	}
	return nil
}

// S82OldB489TargetsGainFlying is "up to X target creatures gain flying
// until end of turn": each still-legal target (Voyager Drake).
func S82OldB489TargetsGainFlying(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB489Zendikon is one Worldwake Zendikon: "Enchant land. Enchanted
// land is a P/T <colour> <subtype> creature[ with <keyword>]. It's still a
// land. When enchanted land dies, return that card to its owner's hand."
// The animation is S82OldB100AnimateAttachedLand and the keyword (empty for
// none) is granted in layer 6, all kept while the Aura is on the land.
func S82OldB489Zendikon(oracleID, name, color, subtype string, power, toughness int, keyword string) Spec {
	static := S82OldB100AnimateAttachedLand(color, subtype, power, toughness)
	if keyword != "" {
		grant := GrantToAttached(keyword)
		grant.ContinuesAfterRemoval = true
		static = append(static, grant)
	}
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       static,
		Triggered: []game.TriggeredAbility{
			WhenEnchantedCreatureDies(name+" — return that card to its owner's hand", S82OldB100ReturnTheDiedCardToHand),
		},
	}
}

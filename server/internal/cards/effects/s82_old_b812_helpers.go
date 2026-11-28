package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b812_helpers.go — Mirage counters, phasing and recursion (#5633).

// S82OldB812SpellTargetsAnEnchantment matches a spell on the stack at least
// one of whose targets is an enchantment on the battlefield (Vigilant
// Martyr's "spell that targets an enchantment"). The zone check is
// B14SpellTargetsAPermanentYouControl's: a card in a graveyard keeps its
// types, and without it a spell aimed at an enchantment that has left could
// still be countered.
func S82OldB812SpellTargetsAnEnchantment() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		item := g.StackItemForEffect(c.InstanceID)
		if item == nil {
			return false
		}
		for _, t := range item.Targets {
			if t.Kind != game.TargetCard {
				continue
			}
			if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneBattlefield {
				continue
			}
			if tc, ok := g.LookupCardForEffect(t.ID); ok && tc.IsEnchantment() {
				return true
			}
		}
		return false
	}
}

// S82OldB812CounterThenItsControllerGainsLifeEqualToItsManaValue is
// Illumination's body: "Counter target artifact or enchantment spell. Its
// controller gains life equal to its mana value." The controller and the mana
// value are read before the counter moves the spell (CR 608.2h); a spell that
// is no longer there leaves the whole instruction undone (CR 608.2b).
func S82OldB812CounterThenItsControllerGainsLifeEqualToItsManaValue(_ *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		spell := ctx.Game.StackItemForEffect(t.ID)
		card, ok := ctx.Game.LookupCardForEffect(t.ID)
		if spell == nil || !ok {
			return nil
		}
		controller, mv := spell.Controller, card.ManaValue()
		if err := (CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), controller, mv)
	}
	return nil
}

// S82OldB812LandsYouControlPhaseOut is "all lands you control phase out"
// (Taniwha): every land its controller controls as the trigger resolves,
// together (CR 702.26g takes whatever is attached along).
func S82OldB812LandsYouControlPhaseOut(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsLand() {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.PhaseOutForEffect(item.SourceCardID, ids...)
}

// S82OldB812PhaseOutThisUnlessPaid is the branch of Warping Wurm's "phases
// out unless you pay {2}{G}{U}": the decline phases the source out, if it is
// still the permanent that asked. An on-disk key: never renamed, never reused.
var S82OldB812PhaseOutThisUnlessPaid = PayUnlessBranches("pay-unless/old-b812-warping-wurm-phase-out", nil,
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if ctx.Item == nil || !SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
			return nil
		}
		return ctx.Game.PhaseOutForEffect(ctx.Source(), ctx.Source())
	})

// S82OldB812DelayedDrawAtNextUpkeep is "Draw a card at the beginning of the next
// turn's upkeep" (Bone Harvest, Mishra's Bauble): a delayed trigger of the
// spell's controller (CR 603.7).
func S82OldB812DelayedDrawAtNextUpkeep(ctx *Context, label string) error {
	return ScheduleDelayedTrigger{At: game.StepUpkeep, Label: label, Body: PortentDrawBody}.Apply(ctx)
}

// S82OldB812HasAnAuraAttached reports whether an Aura is attached to the
// permanent (Hakim, "only if Hakim isn't enchanted").
func S82OldB812HasAnAuraAttached(g *game.Game, host uuid.UUID) bool {
	return len(S82OldB812AurasAttachedTo(g, host)) > 0
}

// S82OldB812AurasAttachedTo lists the Auras on the battlefield attached to host.
func S82OldB812AurasAttachedTo(g *game.Game, host uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsAttachedTo(host) && c.IsAura() {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// S82OldB812HakimReturnsAnAura is Hakim's first ability: the target Aura card
// in its controller's graveyard returns to the battlefield attached to Hakim,
// if Hakim is still the permanent that was activated and the Aura can enchant
// it (CR 303.4f, 608.2b).
func S82OldB812HakimReturnsAnAura(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	aura := FirstLegalCardTarget(ctx)
	if aura == uuid.Nil {
		return nil
	}
	if z := g.FindCardZoneForEffect(aura); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	card, ok := g.LookupCardForEffect(aura)
	if !ok || !CanBeAttachedTo(g, item.Controller, card, item.SourceCardID) {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect([]game.BatchEntry{{CardID: aura, From: game.ZoneGraveyard}},
		game.ZoneEntryOptions{Controller: item.Controller, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: item.SourceCardID}},
		game.KeyedTail{})
}

// S82OldB812DestroyAurasAttachedToThis is "Destroy all Auras attached to
// Hakim": the set is read once as the ability resolves (CR 608.2c).
func S82OldB812DestroyAurasAttachedToThis(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	for _, id := range S82OldB812AurasAttachedTo(g, item.SourceCardID) {
		if err := (DestroyTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB812ReturnThisAndAGriffin is Mtenda Griffin's ability: return this
// creature to its owner's hand and return the target Griffin card from its
// controller's graveyard to their hand. Each half is done if it can be (CR
// 608.2b): a Griffin that is no longer a legal target is not returned, and a
// Mtenda Griffin that is a new object is not bounced (CR 400.7).
func S82OldB812ReturnThisAndAGriffin(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if SourceIsStillThisPermanent(g, item) {
		if err := (BounceToHand{Target: item.SourceCardID}).Apply(ctx); err != nil {
			return err
		}
	}
	card := FirstLegalCardTarget(ctx)
	if card == uuid.Nil {
		return nil
	}
	return ReturnFromGraveyard{Target: card, Dest: game.ZoneHand}.Apply(ctx)
}

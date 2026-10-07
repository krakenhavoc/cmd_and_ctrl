package effects

import (
	"fmt"
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prevention_static.go — the card-facing half of ADR 0108 §8 (#1906): a
// prevention STATIC that does something with the damage it prevented,
// and the follow-up bodies the scoped "prevent the next N damage" shields
// share since owner decision 2. The engine half is
// game/prevention_then.go, which says how the apply loop measures what
// was prevented and why CR 615.12 still runs the additional effect.
//
// Append-only, mechanic-named: every member of the family is one call
// here, so the clone gate never sees the same body twice.
//
// Writing a card. The constructor IS the printed subject, which fixes the
// unit the additional effect runs once per (§8 decision 2):
//
//	PreventDamageDealtTo(PreventionStatic{To: ToThisCreature, Then: removeACounterFromThisBody})
//	  — "If damage would be dealt to this creature, prevent that damage.
//	    Remove a +1/+1 counter from this creature." (the Phantoms): once
//	    per recipient, so three blockers remove one counter.
//	PreventDamageASourceWouldDeal(PreventionStatic{To: ToYou, Then: …})
//	  — "If a source would deal damage to you, prevent that damage and put
//	    an incarnation counter on this enchantment." (Nine Lives): once per
//	    source, so three attackers put on three.
//
// THE TWO AMOUNTS. A body is handed "that damage" — the damage the static
// was applied to — as its item's Trigger.Event.Amount (thatDamage), and
// "the damage prevented this way" as its params' Amount. Under damage
// that can't be prevented (CR 615.12) the first is the whole event and
// the second is zero, and the rulings follow the printed words:
// Polukranos still removes "that many" counters, and Phyrexian Hydra puts
// on no -1/-1 counter "for each 1 damage prevented this way". Pick the one
// the card says.

// StaticRecipient is what a prevention static protects: the X of "if
// damage would be dealt to X".
type StaticRecipient uint8

const (
	// ToThisCreature is "to this creature" / "to <its name>".
	ToThisCreature StaticRecipient = iota
	// ToYou is "to you": the static's controller.
	ToYou
	// ToEquippedCreature is "to equipped creature" (Panther Habit).
	ToEquippedCreature
	// ToAnOpponent is "to an opponent" of the static's controller.
	ToAnOpponent
	// ToACreature is "to a creature", any creature (Weeping Angel).
	ToACreature
)

// StaticDamage narrows which damage a prevention static sees.
type StaticDamage uint8

const (
	// AnyDamage is "damage".
	AnyDamage StaticDamage = iota
	// CombatDamage is "combat damage" (Gloom Surgeon, Ironscale Hydra).
	CombatDamage
	// NoncombatDamage is "noncombat damage" (Stormwild Capridor).
	NoncombatDamage
)

// PreventionStatic is one printed "If [<a source>] [combat] damage would
// be dealt to X, prevent that damage. <Then>".
type PreventionStatic struct {
	// To is what it protects.
	To StaticRecipient
	// Damage narrows it to combat or noncombat damage.
	Damage StaticDamage
	// From narrows the SOURCE, read off the source as the damage would be
	// dealt (its last-known information): FromACreature, FromASourceYou
	// Control, FromThis. Nil is any source.
	From DamageSourceFilter
	// While is the static's condition, read as the damage would be dealt
	// ("while it has a +1/+1 counter on it"). Nil is always.
	While func(g *game.Game, src *game.Card) bool
	// Then is the additional effect (CR 615.5): a registered body.
	Then game.BodyRef
	// Label is the CR 616 ordering prompt's line.
	Label string
}

// DamageSourceFilter is a prevention static's "a creature" / "a source
// you control" / "this creature": `lki` is the source as the damage
// would be dealt, `src` the permanent with the static.
type DamageSourceFilter func(ev *game.ReplacementEvent, lki *game.Characteristic, src *game.Card) bool

// FromACreature is "if a creature would deal damage".
func FromACreature() DamageSourceFilter {
	return func(_ *game.ReplacementEvent, lki *game.Characteristic, _ *game.Card) bool {
		return lki != nil && slices.Contains(lki.Types, "Creature")
	}
}

// FromASourceYouControl is "if a source you control would deal damage".
func FromASourceYouControl() DamageSourceFilter {
	return func(_ *game.ReplacementEvent, lki *game.Characteristic, src *game.Card) bool {
		return lki != nil && lki.Controller == src.Controller
	}
}

// FromThis is "if this creature would deal damage".
func FromThis() DamageSourceFilter {
	return func(ev *game.ReplacementEvent, _ *game.Characteristic, src *game.Card) bool {
		return ev.DamageSource == src.InstanceID
	}
}

// WhileItHasAPlusOneCounter is "while it has a +1/+1 counter on it"
// (Undergrowth Champion, Polukranos).
func WhileItHasAPlusOneCounter(_ *game.Game, src *game.Card) bool {
	return src.Counters[game.CounterPlusOne] > 0
}

// PreventDamageDealtTo is "If damage would be dealt to X, prevent that
// damage. <Then>": the additional effect runs once per RECIPIENT in one
// damage instance (the Phantom Centaur ruling).
func PreventDamageDealtTo(p PreventionStatic) game.ReplacementEffect {
	return p.replacement(game.ThenPerRecipient)
}

// PreventDamageASourceWouldDeal is "If a source would deal damage to X,
// prevent that damage and <Then>": the additional effect runs once per
// SOURCE in one damage instance (the Nine Lives ruling).
func PreventDamageASourceWouldDeal(p PreventionStatic) game.ReplacementEffect {
	return p.replacement(game.ThenPerSource)
}

// PreventAllDamageDealtTo is "Prevent all damage that would be dealt to
// X" with no additional effect (Glittering Lion and Lynx, #1859): the
// plain CR 615.1a prevention, which PreventDamageDealtTo cannot express
// because it demands a Then. Then is refused here.
func PreventAllDamageDealtTo(p PreventionStatic) game.ReplacementEffect {
	if p.Then.Key() != "" {
		panic("effects: PreventAllDamageDealtTo with a Then: use PreventDamageDealtTo")
	}
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true, // CR 615.1a — "prevent"; CR 615.12 reads it
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			return src != nil && p.applies(ev, g, src)
		},
		Replace:    preventThatDamage,
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      p.Label,
	}
}

func (p PreventionStatic) replacement(per game.PreventionUnit) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true, // CR 615.1a — "prevent"; CR 615.12 reads it
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			return src != nil && p.applies(ev, g, src)
		},
		Replace:    preventThatDamage,
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Then:       p.Then,
		ThenPer:    per,
		Label:      p.Label,
	}
}

func (p PreventionStatic) applies(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 {
		return false
	}
	switch p.Damage {
	case CombatDamage:
		if !ev.IsCombatDamage {
			return false
		}
	case NoncombatDamage:
		if ev.IsCombatDamage {
			return false
		}
	}
	if p.While != nil && !p.While(g, src) {
		return false
	}
	if p.From != nil && !p.From(ev, ev.SourceLKI, src) {
		return false
	}
	return staticProtects(p.To, ev.DamageTarget, g, src)
}

// staticProtects reports whether `target` is what the static protects.
func staticProtects(to StaticRecipient, target uuid.UUID, g *game.Game, src *game.Card) bool {
	switch to {
	case ToThisCreature:
		return target == src.InstanceID
	case ToYou:
		return target == src.Controller
	case ToEquippedCreature:
		return src.IsAttachedTo(target)
	case ToAnOpponent:
		p := g.PlayerByIDForEffect(target)
		return p != nil && p.ID != src.Controller
	case ToACreature:
		c, ok := g.LookupCardForEffect(target)
		return ok && c.IsCreature() && g.Battlefield.Contains(target)
	}
	return false
}

// preventThatDamage is every prevention static's Replace: "prevent that
// damage" (CR 615.1). It changes the event and nothing else, so the apply
// loop's measure of it is the whole of what was prevented.
func preventThatDamage(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
	ev.Cancel()
	return nil
}

// replacementThenProblem is Register's check on a replacement's
// additional effect (ADR 0108 §8 decision 1): Then is a prevention
// effect's (CR 615.5), and its unit is required with it. "" when sound.
func replacementThenProblem(r game.ReplacementEffect) string {
	hasThen := r.Then.Key() != ""
	switch {
	case hasThen && !r.Prevention:
		return "declares Then without Prevention: an additional effect belongs to a prevention effect (CR 615.5)"
	case hasThen && !r.ThenPer.Valid():
		return fmt.Sprintf("declares Then with no unit (ThenPer %q): \"damage would be dealt to X\" is per recipient, \"a source would deal damage to X\" per source", r.ThenPer)
	case !hasThen && r.ThenPer != "":
		return "declares ThenPer without Then"
	}
	return ""
}

// --- reading a follow-up ----------------------------------------------

// thatDamage is "that damage" / "that many": the damage the prevention
// effect was applied to, prevented or not (CR 615.12).
func thatDamage(item *game.StackItem) int {
	if item == nil || item.Trigger == nil {
		return 0
	}
	return item.Trigger.Event.Amount
}

// followUpRecipient is "it" / "that creature": the permanent the damage
// would have been dealt to, while it is still that object.
func followUpRecipient(g *game.Game, item *game.StackItem) (uuid.UUID, bool) {
	if item == nil || item.Trigger == nil {
		return uuid.Nil, false
	}
	id := item.Trigger.Event.Target
	ref, ok := g.PermanentRefForEffect(id)
	if !ok {
		return uuid.Nil, false
	}
	info, ok := g.PermanentForEffect(ref)
	return id, ok && !info.Left
}

// followUpThis is "this creature" in a static's follow-up: the permanent
// whose static it is, while it is still that object (CR 400.7). Last-known
// information is something to read, not to put counters on.
func followUpThis(g *game.Game, item *game.StackItem) (game.PermanentInfo, bool) {
	info, ok := NewContext(g, item).SourcePermanent()
	return info, ok && !info.Left
}

// --- the shared follow-ups ----------------------------------------------

var (
	// "Remove a +1/+1 counter from this creature" (the Phantoms,
	// Oathsworn Knight, Unbreathing Horde, Undergrowth Champion): one per
	// application, whatever the amount, and still under CR 615.12.
	removeACounterFromThisBody = game.DelayedBody("prevention/remove-a-counter-from-this", followUpRemoveACounterFromThis)

	// "Remove that many +1/+1 counters from it" (Polukranos, Ugin's
	// Conjurant): that damage, or every counter it has when that is fewer
	// (the rulings), and still under CR 615.12.
	removeThatManyCountersFromThisBody = game.DelayedBody("prevention/remove-that-many-counters-from-this", followUpRemoveThatManyCountersFromThis)

	// "Put that many +1/+1 counters on him" (Anti-Venom).
	thatManyCountersOnThisBody = game.DelayedBody("prevention/that-many-counters-on-this", followUpThatManyCountersOnThis)

	// "Put a +1/+1 counter on this creature for each 1 damage prevented
	// this way" (Stormwild Capridor): none under CR 615.12.
	countersOnThisPerPreventedBody = game.DelayedBody("prevention/counters-on-this-per-prevented", followUpCountersOnThisPerPrevented)

	// "Put a +1/+1 counter on this creature" (Ironscale Hydra): one per
	// application.
	aCounterOnThisBody = game.DelayedBody("prevention/a-counter-on-this", followUpACounterOnThis)

	// "Put that many +1/+1 counters on it" (Panther Habit): on the
	// creature the damage would have been dealt to.
	thatManyCountersOnItBody = game.DelayedBody("prevention/that-many-counters-on-it", followUpThatManyCountersOnIt)

	// "For each 1 damage prevented this way, put a +1/+1 counter on that
	// creature" (Test of Faith, Temper).
	countersOnItPerPreventedBody = game.DelayedBody("prevention/counters-on-it-per-prevented", followUpCountersOnItPerPrevented)

	// "You gain life equal to the damage prevented this way", on a
	// charged shield (Candles' Glow). The same rule as the source shields'
	// prevention/gain-life-equal, under its own key: a binary from before
	// ADR 0108 PR 8 knows that key but not that a preventDamage shield
	// has a follow-up, and would restore the shield with its follow-up
	// silently dropped. A key it does not know makes it refuse the file.
	shieldGainLifeEqualBody = game.DelayedBody("prevention/shield-gain-life-equal", preventedGainLife)

	// "<This> deals that much damage to <the chosen target>" (Acolyte's
	// Reward, Vengeful Archon): the amount prevented, to the follow-up's
	// recipient (game.ShieldFollowUp.To), from the shield's source. Not a
	// redirection: the shield's source deals it, and it is not combat
	// damage (the rulings).
	dealThatMuchToTheChosenTargetBody = game.DelayedBody("prevention/deal-that-much-to-the-chosen-target", followUpDealThatMuchToTheChosenTarget)

	// "At the beginning of the next end step, put a +0/+1 counter on that
	// creature for each 1 damage prevented this way" (Sacred Boon, Scars
	// of the Veteran): ONE delayed trigger, joined by every prevention of
	// the shield before it fires (ScheduleOrJoinDelayedTriggerForEffect).
	endStepToughnessCountersBody = game.DelayedBody("prevention/end-step-toughness-counters-per-prevented", followUpEndStepToughnessCounters)
	// The delayed trigger itself: Params.Amount +0/+1 counters on its
	// card, while it is still on the battlefield.
	putToughnessCountersBody = game.DelayedBody("prevention/put-toughness-counters", putToughnessCounters)
)

func followUpRemoveACounterFromThis(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	info, ok := followUpThis(g, item)
	if !ok || info.Counters[game.CounterPlusOne] <= 0 {
		return nil
	}
	return g.AddCounterForEffect(item.SourceCardID, game.CounterPlusOne, -1)
}

func followUpRemoveThatManyCountersFromThis(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	info, ok := followUpThis(g, item)
	if !ok {
		return nil
	}
	n := min(thatDamage(item), info.Counters[game.CounterPlusOne])
	if n <= 0 {
		return nil
	}
	return g.AddCounterForEffect(item.SourceCardID, game.CounterPlusOne, -n)
}

func followUpThatManyCountersOnThis(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	if _, ok := followUpThis(g, item); !ok || thatDamage(item) <= 0 {
		return nil
	}
	return g.AddCounterByForEffect(item.Controller, item.SourceCardID, game.CounterPlusOne, thatDamage(item))
}

func followUpCountersOnThisPerPrevented(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if _, ok := followUpThis(g, item); !ok || p.Amount <= 0 {
		return nil
	}
	return g.AddCounterByForEffect(item.Controller, item.SourceCardID, game.CounterPlusOne, p.Amount)
}

func followUpACounterOnThis(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	if _, ok := followUpThis(g, item); !ok {
		return nil
	}
	return g.AddCounterByForEffect(item.Controller, item.SourceCardID, game.CounterPlusOne, 1)
}

func followUpThatManyCountersOnIt(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	it, ok := followUpRecipient(g, item)
	if !ok || thatDamage(item) <= 0 {
		return nil
	}
	return g.AddCounterByForEffect(item.Controller, it, game.CounterPlusOne, thatDamage(item))
}

func followUpCountersOnItPerPrevented(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	it, ok := followUpRecipient(g, item)
	if !ok || p.Amount <= 0 {
		return nil
	}
	return g.AddCounterByForEffect(item.Controller, it, game.CounterPlusOne, p.Amount)
}

func followUpDealThatMuchToTheChosenTarget(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 || len(item.Targets) != 1 {
		return nil
	}
	return g.DealMarkedDamageForEffect(item.SourceCardID, nil, item.Targets[0].ID, p.Amount, game.DamageMarks{})
}

// countersPerPreventedShield is "Prevent the next N damage that would be
// dealt to target creature this turn. For each 1 damage prevented this
// way, put a +1/+1 counter on that creature." (Test of Faith, Temper). A
// target gone by resolution gets no shield (CR 608.2b).
func countersPerPreventedShield(_ *game.StackItem, ctx *Context, n int, name string) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok || n < 1 {
		return nil
	}
	return PreventNextDamage{
		Target: id,
		Amount: n,
		Then:   countersOnItPerPreventedBody,
		Label:  fmt.Sprintf("%s — prevent the next %d damage", name, n),
	}.Apply(ctx)
}

func followUpEndStepToughnessCounters(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	it, ok := followUpRecipient(g, item)
	if !ok || p.Amount <= 0 {
		return nil
	}
	if c, found := g.LookupCardForEffect(it); !found || !c.IsCreature() {
		// Scars of the Veteran: "If it's a creature".
		return nil
	}
	g.ScheduleOrJoinDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   item.Controller,
		SourceCardID: item.SourceCardID,
		Label:        item.Label,
		At:           game.StepEnd,
		Cards:        []uuid.UUID{it},
		Body:         putToughnessCountersBody,
		Params:       game.EffectParams{Amount: p.Amount},
	})
	return nil
}

func putToughnessCounters(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 || len(item.Targets) == 0 {
		return nil
	}
	id := item.Targets[0].ID
	if !g.Battlefield.Contains(id) {
		return nil
	}
	return g.AddCounterByForEffect(item.Controller, id, "+0/+1", p.Amount)
}

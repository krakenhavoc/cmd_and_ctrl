package effects

import (
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// redirect_damage.go — the card-facing half of ADR 0108 §9 (#1905):
// damage dealt to something else instead (CR 614.9). The engine half is
// game/redirect_damage.go: the scoped ModRedirectDamage record, and the
// one primitive every redirection goes through.
//
// Append-only, mechanic-named. Two shapes:
//
//   - a resolved spell or ability's redirection is a RedirectDamage
//     value — the source as on every shield (Choose, FromThis, From,
//     Queries), what it protects (a ShieldTarget), how long (Next,
//     Amount, or all this turn) and where to (a RedirectTo):
//
//     RedirectDamage{Choose: true, Protect: ShieldYou, Next: true, To: RedirectToThis}
//     — Beacon of Destiny
//     RedirectDamage{Choose: true, Protect: ShieldYouAndPermanentsYouControl, Amount: 2, To: RedirectToClause(0)}
//     — Harm's Way
//
//   - a static one ("All damage that would be dealt to you is dealt to
//     enchanted creature instead", Pariah) is staticRedirection, a
//     ReplacementEffect that declares RedirectsDamage and calls
//     game.RedirectDamageEventForEffect.

// RedirectTo is where a RedirectDamage deals the damage instead.
type RedirectTo struct {
	kind   redirectToKind
	clause int
	id     uuid.UUID
}

type redirectToKind uint8

const (
	redirectToNone redirectToKind = iota
	redirectToThis
	redirectToYou
	redirectToSourceController
	redirectToClause
	redirectToEnchanted
	redirectToObject
)

var (
	// RedirectToThis is "is dealt to this creature instead": the object
	// the ability came from, as it is now. A source that has left (or is
	// a new object) leaves nowhere to redirect to, so nothing is made.
	RedirectToThis = RedirectTo{kind: redirectToThis}
	// RedirectToYou is "is dealt to you instead" — the resolving effect's
	// controller (Jade Monolith, Vassal's Duty).
	RedirectToYou = RedirectTo{kind: redirectToYou}
	// RedirectToSourceController is "that damage is dealt to that
	// source's controller instead" (Reflect Damage), read as the damage
	// would be dealt.
	RedirectToSourceController = RedirectTo{kind: redirectToSourceController}
	// RedirectToEnchanted is "is dealt to enchanted creature instead" for
	// an Aura's own ability (Saving Grace), the creature it is attached to
	// as the ability resolves, or as the Aura last existed (CR 608.2h).
	RedirectToEnchanted = RedirectTo{kind: redirectToEnchanted}
)

// RedirectToClause is "is dealt to <the target that answered clause i>
// instead" (Harm's Way's "any target", Kor Chant's "another target
// creature"). A target that is illegal as the effect resolves leaves
// nowhere to redirect to (CR 608.2b).
func RedirectToClause(i int) RedirectTo { return RedirectTo{kind: redirectToClause, clause: i} }

// RedirectToObject is "is dealt to <that player or permanent> instead"
// for an object the card has already named.
func RedirectToObject(id uuid.UUID) RedirectTo { return RedirectTo{kind: redirectToObject, id: id} }

// RedirectDamage is the redirection a resolving spell or ability makes.
// The source fields mean what they mean on PreventNextDamageFromSource;
// with none of them, any source matching Queries is the one, and with no
// Queries either, every source.
type RedirectDamage struct {
	Choose   bool
	FromThis bool
	From     uuid.UUID
	Queries  []game.PermanentQuery

	// Protect is whose damage is redirected. AlsoYou adds "and/or you" to
	// a protected permanent ("this creature and/or you", Glarecaster).
	// Opponents is "to an opponent" (Soltari Guerrillas), in place of
	// Protect.
	Protect   ShieldTarget
	AlsoYou   bool
	Opponents bool

	// CombatOnly is "combat damage".
	CombatOnly bool

	// Next is "the next time"; Amount is "the next N damage" (CR 615.7).
	// Neither is all damage this turn.
	Next   bool
	Amount int

	// To is where the damage is dealt instead. The zero RedirectTo, with
	// Then, is Eye for an Eye's shape: nothing is dealt instead, and Then
	// runs with the amount.
	To   RedirectTo
	Then game.BodyRef

	// UntilYourNextTurn is "until your next turn"; false is "this turn".
	UntilYourNextTurn bool

	// Question is the source prompt's header; Label the record's. Both
	// default to the card's name.
	Question string
	Label    string
}

func (r RedirectDamage) Apply(ctx *Context) error {
	label := r.Label
	if label == "" {
		label = shieldSourceName(ctx)
	}
	red := game.DamageRedirection{
		EffectSource:      ctx.Source(),
		Controller:        ctx.Controller(),
		Queries:           r.Queries,
		CombatOnly:        r.CombatOnly,
		Next:              r.Next,
		Amount:            r.Amount,
		Then:              r.Then,
		UntilYourNextTurn: r.UntilYourNextTurn,
		Label:             label,
	}
	if r.Opponents {
		red.ProtectRecipients = game.DamageRecipientsOpponents
	} else {
		protected := game.NextDamageShield{Controller: ctx.Controller()}
		if !(PreventNextDamageFromSource{Protect: r.Protect}).protect(ctx, &protected) {
			return nil
		}
		red.ProtectPlayer, red.ProtectTypes, red.ProtectPermanent = protected.ProtectPlayer, protected.ProtectTypes, protected.ProtectPermanent
		if r.AlsoYou {
			red.ProtectPlayer = ctx.Controller()
		}
	}
	if !r.To.resolve(ctx, &red) {
		return nil
	}
	pick := shieldSourcePick{Choose: r.Choose, FromThis: r.FromThis, From: r.From, Queries: r.Queries, Question: r.Question}
	return pick.resolve(ctx, registerRedirection(red))
}

// resolve writes the destination onto the redirection, and reports
// whether there is one to write: a "this creature" that has left, or a
// target that is no longer legal, is nowhere to redirect to.
func (t RedirectTo) resolve(ctx *Context, red *game.DamageRedirection) bool {
	switch t.kind {
	case redirectToNone:
		return red.Then != (game.BodyRef{})
	case redirectToThis:
		src := ctx.Source()
		if src == uuid.Nil || ctx.isNewSourceObjectAsThis(src) {
			return false
		}
		red.To = src
	case redirectToYou:
		red.To = ctx.Controller()
	case redirectToSourceController:
		red.ToSourceController = true
	case redirectToClause:
		ref, ok := ctx.ClauseTarget(t.clause)
		if !ok || (ref.Kind != game.TargetCard && ref.Kind != game.TargetPlayer) {
			return false
		}
		red.To = ref.ID
	case redirectToEnchanted:
		info, ok := ctx.SourcePermanent()
		if !ok || info.AttachedTo.Kind != game.TargetCard {
			return false
		}
		red.To = info.AttachedTo.ID
	case redirectToObject:
		red.To = t.id
	}
	return red.To != uuid.Nil || red.ToSourceController
}

// registerRedirection finishes a redirection once its source is pinned.
// It captures the redirection's plain description and nothing else.
func registerRedirection(red game.DamageRedirection) func(*game.Game, game.ObjectRef, game.ZoneKind) error {
	return func(g *game.Game, ref game.ObjectRef, zone game.ZoneKind) error {
		r := red
		r.Source, r.SourceZone = ref, zone
		g.RedirectDamageThisTurnForEffect(r)
		return nil
	}
}

// redirectRow is an activated ability whose whole effect is a
// redirection.
func redirectRow(label string, cost game.AbilityCost, targets *game.TargetSpec, r RedirectDamage) ActivatedAbility {
	return nextDamageShieldRow(label, cost, targets, r)
}

// redirectSpell is an instant's OnResolve whose whole effect is a
// redirection.
func redirectSpell(r RedirectDamage) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		return r.Apply(ctx)
	}
}

// enKorRow is the en-Kor cycle's "{0}: The next 1 damage that would be
// dealt to this creature this turn is dealt to target creature you
// control instead." It can redirect to itself, which does nothing (the
// rulings).
func enKorRow() ActivatedAbility {
	return redirectRow("{0}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature you control instead.",
		ManaCost("{0}"), TargetCreature("target creature you control", YouControl()),
		RedirectDamage{Protect: ShieldThis, Amount: 1, To: RedirectToClause(0)})
}

// --- the statics -----------------------------------------------------

// redirectWhere names a static redirection's two halves, judged as the
// damage would be dealt: whether it applies (`applies`, with the
// permanent whose static it is) and where the damage goes (`to`).
type redirectWhere struct {
	applies func(g *game.Game, src *game.Card, ev *game.ReplacementEvent) bool
	to      func(g *game.Game, src *game.Card, ev *game.ReplacementEvent) (uuid.UUID, bool)
}

// staticRedirection is a static "… is dealt to <something> instead"
// (CR 614.9): a replacement that declares RedirectsDamage, applies only
// when the redirection would do something (CR 614.9's test, so a
// redirection to a creature that has gone is never offered in a CR 616
// ordering prompt), and redirects through the one primitive. Two of
// them are two different modifications — each writes its own object into
// the event — so the affected player orders them (the Pariah and
// Protector of the Crown rulings).
func staticRedirection(label string, w redirectWhere) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:         []game.EventKind{game.EventDealDamage},
		RedirectsDamage: true,
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || !w.applies(g, src, ev) {
				return false
			}
			to, ok := w.to(g, src, ev)
			return ok && g.CanRedirectDamageForEffect(ev, to)
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			if to, ok := w.to(g, src, ev); ok {
				g.RedirectDamageEventForEffect(ev, to)
			}
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// damageToYou is "damage that would be dealt to you": the static's
// controller.
func damageToYou(_ *game.Game, src *game.Card, ev *game.ReplacementEvent) bool {
	return ev.DamageTarget == src.Controller
}

// toThisPermanent is "is dealt to this creature instead".
func toThisPermanent(_ *game.Game, src *game.Card, _ *game.ReplacementEvent) (uuid.UUID, bool) {
	return src.InstanceID, true
}

// toAttachedHost is "is dealt to enchanted (equipped) creature instead":
// the creature the static's permanent is attached to. Unattached, there
// is nowhere to redirect to (the Pariah's Shield ruling).
func toAttachedHost(g *game.Game, src *game.Card, _ *game.ReplacementEvent) (uuid.UUID, bool) {
	host := g.AttachedHostOf(src)
	if host == nil {
		return uuid.Nil, false
	}
	return host.InstanceID, true
}

// redirectYourDamageToAttached is Pariah's "All damage that would be dealt
// to you is dealt to enchanted creature instead" (and Pariah's Shield's
// "equipped creature").
func redirectYourDamageToAttached(label string) game.ReplacementEffect {
	return staticRedirection(label, redirectWhere{applies: damageToYou, to: toAttachedHost})
}

// redirectYourDamageToThis is "All damage that would be dealt to you is
// dealt to this creature instead" (Empyrial Archangel, Protector of the
// Crown).
func redirectYourDamageToThis(label string) game.ReplacementEffect {
	return staticRedirection(label, redirectWhere{applies: damageToYou, to: toThisPermanent})
}

// damageToYouOrYourOtherPermanents is "damage that would be dealt to you
// and other permanents you control" (Palisade Giant).
func damageToYouOrYourOtherPermanents(g *game.Game, src *game.Card, ev *game.ReplacementEvent) bool {
	if ev.DamageTarget == src.Controller {
		return true
	}
	if ev.DamageTarget == src.InstanceID {
		return false
	}
	if !onBattlefield(g, ev.DamageTarget) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.DamageTarget)
	return ok && c.Controller == src.Controller
}

// untappedAnd is "as long as this creature is untapped, …" around a
// static redirection's test.
func untappedAnd(applies func(*game.Game, *game.Card, *game.ReplacementEvent) bool) func(*game.Game, *game.Card, *game.ReplacementEvent) bool {
	return func(g *game.Game, src *game.Card, ev *game.ReplacementEvent) bool {
		return !src.Tapped && applies(g, src, ev)
	}
}

// damageToYouFromAnUnblockedCreature is "damage that would be dealt to you
// by unblocked creatures" (CR 509.1h) — with combat, "combat damage".
func damageToYouFromAnUnblockedCreature(combat bool) func(*game.Game, *game.Card, *game.ReplacementEvent) bool {
	return func(g *game.Game, src *game.Card, ev *game.ReplacementEvent) bool {
		if !damageToYou(g, src, ev) || (combat && !ev.IsCombatDamage) {
			return false
		}
		return ev.SourceLKI != nil && lkiHasType(ev.SourceLKI, "Creature") && g.UnblockedAttackerForEffect(ev.DamageSource)
	}
}

// damageToTheAttachedHost is "damage that would be dealt to enchanted
// creature".
func damageToTheAttachedHost(g *game.Game, src *game.Card, ev *game.ReplacementEvent) bool {
	host := g.AttachedHostOf(src)
	return host != nil && ev.DamageTarget == host.InstanceID
}

// toTheAttachedHostsController is "is dealt to its controller instead",
// "it" being the enchanted creature.
func toTheAttachedHostsController(g *game.Game, src *game.Card, _ *game.ReplacementEvent) (uuid.UUID, bool) {
	host := g.AttachedHostOf(src)
	if host == nil || host.Controller == uuid.Nil {
		return uuid.Nil, false
	}
	return host.Controller, true
}

// damageToYouFromAnInstantOrSorceryOfTheChosenColor is Harsh Judgment's
// "if an instant or sorcery spell of the chosen color would deal damage to
// you", judged as the spell is when it would deal the damage.
func damageToYouFromAnInstantOrSorceryOfTheChosenColor(g *game.Game, src *game.Card, ev *game.ReplacementEvent) bool {
	ch := ev.SourceLKI
	if !damageToYou(g, src, ev) || src.ChosenColor == "" || ch == nil {
		return false
	}
	if !lkiHasType(ch, "Instant") && !lkiHasType(ch, "Sorcery") {
		return false
	}
	return slices.Contains(ch.Colors, src.ChosenColor)
}

// toTheSourcesController is "it deals that damage to its controller
// instead", the source as it is when the damage would be dealt.
func toTheSourcesController(_ *game.Game, _ *game.Card, ev *game.ReplacementEvent) (uuid.UUID, bool) {
	if ev.SourceLKI == nil || ev.SourceLKI.Controller == uuid.Nil {
		return uuid.Nil, false
	}
	return ev.SourceLKI.Controller, true
}

// lkiHasType reports whether a source's characteristics have the card
// type t.
func lkiHasType(ch *game.Characteristic, t string) bool {
	return slices.ContainsFunc(ch.Types, func(have string) bool { return strings.EqualFold(have, t) })
}

// targetCreaturePlaneswalkerOrPlayer is "target creature, planeswalker, or
// player" (Captain's Maneuver): any target but a battle.
func targetCreaturePlaneswalkerOrPlayer(label string) *game.TargetSpec {
	spec := TargetAny()
	spec.Label = label
	spec.CardOK = func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
		return c.IsCreature() || c.IsPlaneswalker()
	}
	return spec
}

// korRedirection is Kor Chant and Kor Dirge: "All damage that would be
// dealt this turn to target creature you control by a source of your
// choice is dealt to another target creature instead."
func korRedirection(oracleID, name string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			Distinct(TargetCreature("another target creature")),
		),
		OnResolve: redirectSpell(RedirectDamage{Choose: true, Protect: ShieldClause(0), To: RedirectToClause(1)}),
	}
}

// thatMuchToTheSourcesControllerBody is Eye for an Eye's follow-up: "Eye
// for an Eye deals that much damage to that source's controller" — that
// much being the damage the source dealt you (the handed event's Amount),
// the controller being the source's as it dealt it (the params' Player).
var thatMuchToTheSourcesControllerBody = game.DelayedBody("redirection/that-much-to-the-sources-controller", thatMuchToTheSourcesController)

func thatMuchToTheSourcesController(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if item == nil || item.Trigger == nil || p.Player == uuid.Nil {
		return nil
	}
	n := item.Trigger.Event.Amount
	if n <= 0 {
		return nil
	}
	return g.DealDamageToPlayerForEffect(item.SourceCardID, p.Player, n)
}

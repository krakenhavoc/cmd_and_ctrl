package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prevent_next_from_source.go — the card-facing half of ADR 0107 §6
// (#1860): "The next time a [red] source of your choice would deal
// damage to you this turn, prevent that damage" and the shapes around
// it. The engine half is game/prevent_next_from_source.go, which says
// why the shield is spent by the next INSTANCE of damage (CR 615.8) and
// how the source is pinned (CR 400.7, 609.7a).
//
// Append-only, mechanic-named: the Circles of Protection, the Runes of
// Protection and every "prevented this way" card build on the helpers
// here, so a new member of the family is one call and the clone gate
// never sees the same body twice.
//
// Writing a card:
//
//	PreventNextDamageFromChosenSource(ShieldYou, game.PermanentQuery{Colors: []string{"R"}})
//	  — "a red source of your choice … to you"
//	PreventNextDamageFromThis(ShieldYou)
//	  — Mercenaries' "the next time this creature would deal damage to you"
//	PreventNextDamageFromSource{From: id, Protect: ShieldAnything}
//	  — Awe Strike's "the next time target creature would deal damage"
//	PreventNextDamageFromSource{Queries: …, Protect: ShieldYou}
//	  — Circle of Solace's "a creature of the chosen type", no choice
//
// and set Then to one of the follow-up bodies below for "the damage
// prevented this way" (CR 615.5).

// ShieldTarget is what a next-damage shield protects.
type ShieldTarget struct {
	kind shieldKind
	id   uuid.UUID
	// ids are ShieldObjects' permanents; permTypes are
	// ShieldYouAndYourPermanents' card types.
	ids       []uuid.UUID
	permTypes []string
	clause    int
	// recipients is a recipient set's description (#2045); only the
	// not-one-use shield reads it.
	recipients game.DamageRecipientFilter
}

// types is a copy of the card types "you and <those> you control"
// names.
func (t ShieldTarget) types() []string { return append([]string(nil), t.permTypes...) }

type shieldKind uint8

const (
	shieldYou shieldKind = iota
	shieldYouAndYourCreatures
	shieldAnything
	shieldTheTarget
	shieldEnchanted
	shieldObject
	// ADR 0108 Delivery PR 7: the recipient families. The two below
	// protect several permanents in one record, so only the not-one-use
	// shield (PreventDamageFromSource) reads them.
	shieldYouAndYourPermanentsOf
	shieldTheTargetPermanents
	shieldObjects
	// ADR 0108 PR 9: the redirection's "target clause i" and "this
	// creature".
	shieldClause
	shieldThisObject
	// #2045: a set read as the damage would be dealt — "to creatures",
	// "to creatures you control", "to players". Only the not-one-use
	// shield (PreventDamageFromSource) reads it.
	shieldRecipientSet
)

var (
	// ShieldYou is "would deal damage to you": the resolving effect's
	// controller — for an any-player ability, the activator.
	ShieldYou = ShieldTarget{kind: shieldYou}
	// ShieldYouAndYourCreatures is "to you and/or creatures you
	// control" (Shadowbane).
	ShieldYouAndYourCreatures = ShieldTarget{kind: shieldYouAndYourCreatures}
	// ShieldAnything is "would deal damage this turn", to anything
	// (Awe Strike, the Pilgrims).
	ShieldAnything = ShieldTarget{kind: shieldAnything}
	// ShieldTheTarget is "to any target" / "to target creature": the
	// item's first legal target, player or permanent. A target gone by
	// resolution leaves nothing to shield (CR 608.2b).
	ShieldTheTarget = ShieldTarget{kind: shieldTheTarget}
	// ShieldEnchantedCreature is "to enchanted creature", read as the
	// Aura last existed when its own sacrifice paid for the ability
	// (Kithkin Armor; CR 608.2h).
	ShieldEnchantedCreature = ShieldTarget{kind: shieldEnchanted}
)

// ShieldThis is "to this creature": the object the ability came from, as
// it is now. A source that left and came back is a new object, and is
// not shielded (CR 400.7).
var ShieldThis = ShieldTarget{kind: shieldThisObject}

// ShieldClause protects the target that answered clause i ("target
// creature you control" of Kor Chant's two). A target that is illegal as
// the effect resolves leaves nothing to shield (CR 608.2b).
func ShieldClause(i int) ShieldTarget { return ShieldTarget{kind: shieldClause, clause: i} }

// ShieldObject protects one named player or permanent.
func ShieldObject(id uuid.UUID) ShieldTarget { return ShieldTarget{kind: shieldObject, id: id} }

// PreventNextDamageFromSource is the shield. Exactly one of Choose,
// FromThis and From names the source; with none of them, any source
// matching Queries is the one (Circle of Solace).
type PreventNextDamageFromSource struct {
	// Choose is "a source of your choice": the controller chooses as
	// the shield is made (CR 609.7a), among the sources matching
	// Queries.
	Choose bool
	// FromThis is "this creature": the object the ability came from.
	FromThis bool
	// From is a named object — "the next time target creature would
	// deal damage".
	From uuid.UUID

	// Queries is the property the source must have ("a red source"),
	// offered by and rechecked as the damage would be dealt (CR 615.9).
	Queries []game.PermanentQuery

	// Protect is what the shield protects.
	Protect ShieldTarget

	// Then is the CR 615.5 follow-up: a body run with the damage
	// prevented. Zero is none.
	Then game.BodyRef

	// Half is Dark Sphere's "prevent half that damage, rounded down"
	// (ADR 0108 §7 decision 4).
	Half bool
	// CombatOnly is "the next time <source> would deal combat damage"
	// (Impulsive Maneuvers): other damage passes it by, unspent.
	CombatOnly bool

	// Question is the source prompt's header; Label the shield's. Both
	// default to the card's name.
	Question string
	Label    string
}

// PreventNextDamageFromChosenSource is "The next time a [<queries>]
// source of your choice would deal damage to <protect> this turn,
// prevent that damage" — the Circle of Protection family.
func PreventNextDamageFromChosenSource(protect ShieldTarget, queries ...game.PermanentQuery) PreventNextDamageFromSource {
	return PreventNextDamageFromSource{Choose: true, Protect: protect, Queries: queries}
}

// PreventNextDamageFromThis is "The next time this creature would deal
// damage to <protect> this turn, prevent that damage" (Mercenaries).
func PreventNextDamageFromThis(protect ShieldTarget) PreventNextDamageFromSource {
	return PreventNextDamageFromSource{FromThis: true, Protect: protect}
}

// WithThen returns the shield with a CR 615.5 follow-up.
func (p PreventNextDamageFromSource) WithThen(then game.BodyRef) PreventNextDamageFromSource {
	p.Then = then
	return p
}

func (p PreventNextDamageFromSource) Apply(ctx *Context) error {
	label := p.Label
	if label == "" {
		label = shieldSourceName(ctx) + " — prevent the next damage from a source"
	}
	shield := game.NextDamageShield{
		EffectSource: ctx.Source(),
		Controller:   ctx.Controller(),
		Queries:      p.Queries,
		Then:         p.Then,
		Half:         p.Half,
		CombatOnly:   p.CombatOnly,
		Label:        label,
	}
	if !p.protect(ctx, &shield) {
		return nil
	}
	return p.pick().resolve(ctx, registerChosenShield(shield))
}

// pick is the shield's "which source" half.
func (p PreventNextDamageFromSource) pick() shieldSourcePick {
	return shieldSourcePick{Choose: p.Choose, FromThis: p.FromThis, From: p.From, Queries: p.Queries, Question: p.Question}
}

// shieldSourcePick is the "which source" half every shield against a
// source shares (ADR 0107 §6, ADR 0108 §7): this object, a named object,
// "a source of your choice" (CR 609.7a), or none — any source with the
// queried properties.
type shieldSourcePick struct {
	Choose   bool
	FromThis bool
	From     uuid.UUID
	Queries  []game.PermanentQuery
	Question string
}

// resolve pins the source and hands it to `register` — at once, or once
// the controller has chosen. A choice that finds no source (no legal
// candidate, a chooser who left) makes no shield (CR 609.7a).
func (s shieldSourcePick) resolve(ctx *Context, register func(*game.Game, game.ObjectRef, game.ZoneKind) error) error {
	g := ctx.Game
	switch {
	case s.FromThis:
		ref, ok := ctx.SourceRef()
		if !ok {
			return nil
		}
		return register(g, ref, game.ZoneBattlefield)
	case s.From != uuid.Nil:
		ref, zone, ok := g.DamageSourceRefLocked(s.From)
		if !ok {
			return nil
		}
		return register(g, ref, zone)
	case s.Choose:
		question := s.Question
		if question == "" {
			question = shieldSourceName(ctx) + " — choose a source of damage"
		}
		_, err := g.ChooseDamageSourceThenForEffect(game.ChooseSourcePrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: question,
			Queries:  s.Queries,
			Then:     registerPickedSource(register),
		})
		return err
	}
	return register(g, game.ObjectRef{}, "")
}

// registerPickedSource is a choose_source continuation: no source chosen
// is no shield.
func registerPickedSource(register func(*game.Game, game.ObjectRef, game.ZoneKind) error) func(*game.Game, game.ObjectRef, game.ZoneKind) error {
	return func(g *game.Game, ref game.ObjectRef, zone game.ZoneKind) error {
		if ref.ID == uuid.Nil {
			return nil
		}
		return register(g, ref, zone)
	}
}

// shieldSourceName is the card making the shield, for its labels.
func shieldSourceName(ctx *Context) string {
	if c, ok := ctx.Game.LookupCardForEffect(ctx.Source()); ok && c.Name != "" {
		return c.Name
	}
	return "Shield"
}

// registerChosenShield finishes a next-damage shield once its source is
// pinned. It captures the shield's plain description and nothing else.
func registerChosenShield(shield game.NextDamageShield) func(*game.Game, game.ObjectRef, game.ZoneKind) error {
	return func(g *game.Game, ref game.ObjectRef, zone game.ZoneKind) error {
		s := shield
		s.Source, s.SourceZone = ref, zone
		g.PreventNextDamageFromSourceForEffect(s)
		return nil
	}
}

// protect fills in what the shield protects. False when there is
// nothing left to protect.
func (p PreventNextDamageFromSource) protect(ctx *Context, s *game.NextDamageShield) bool {
	one := func(id uuid.UUID) bool {
		if id == uuid.Nil {
			return false
		}
		if ctx.Game.PlayerByIDForEffect(id) != nil {
			s.ProtectPlayer = id
			return true
		}
		if ctx.isNewSourceObject(id) {
			return false
		}
		s.ProtectPermanent = id
		return true
	}
	switch p.Protect.kind {
	case shieldYou:
		s.ProtectPlayer = s.Controller
	case shieldYouAndYourCreatures:
		s.ProtectPlayer = s.Controller
		s.ProtectTypes = []string{"creature"}
	case shieldAnything:
	case shieldTheTarget:
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard || t.Kind == game.TargetPlayer {
				return one(t.ID)
			}
		}
		return false
	case shieldEnchanted:
		info, ok := ctx.SourcePermanent()
		if !ok || info.AttachedTo.Kind != game.TargetCard {
			return false
		}
		return one(info.AttachedTo.ID)
	case shieldObject:
		return one(p.Protect.id)
	case shieldYouAndYourPermanentsOf:
		s.ProtectPlayer = s.Controller
		s.ProtectTypes = p.Protect.types()
	case shieldTheTargetPermanents, shieldObjects, shieldRecipientSet:
		// Several permanents, or a set: only PreventDamageFromSource
		// reads these (protectMany, recipients); a next-damage shield
		// protects one thing.
		return false
	case shieldClause:
		t, ok := ctx.ClauseTarget(p.Protect.clause)
		if !ok || (t.Kind != game.TargetCard && t.Kind != game.TargetPlayer) {
			return false
		}
		return one(t.ID)
	case shieldThisObject:
		src := ctx.Source()
		if src == uuid.Nil || ctx.isNewSourceObjectAsThis(src) {
			return false
		}
		s.ProtectPermanent = src
		return true
	}
	return true
}

// QueryColors is "a <colour> [or <colour>] source": a query over colours.
func QueryColors(colors ...string) game.PermanentQuery {
	return game.PermanentQuery{Colors: colors}
}

// QueryTypes is "an artifact source", "a land source".
func QueryTypes(types ...string) game.PermanentQuery {
	return game.PermanentQuery{Types: types}
}

// --- the follow-ups (CR 615.5) ----------------------------------------
//
// Each body is handed the item the engine builds for a follow-up: its
// Controller is the shield's controller ("you"), its SourceCardID the
// card whose effect made the shield (the source of any damage it deals),
// and its Trigger the prevented damage — Amount, the damage Source, its
// controller as Actor and its Colors, all as they were when the damage
// would have been dealt. The params carry the amount (Amount) and the
// damage source's controller (Player). Every body does nothing for zero:
// a follow-up after unpreventable damage (CR 615.12) prevented nothing.

var (
	// "You gain life equal to the damage prevented this way" (Reverse
	// Damage, Awe Strike, Cho-Arrim Alchemist, Intervention Pact).
	preventedGainLifeBody = game.DelayedBody("prevention/gain-life-equal", preventedGainLife)

	// "<This> deals that much damage to that source's controller"
	// (Deflecting Palm).
	preventedDamageSourceControllerBody = game.DelayedBody("prevention/damage-source-controller", preventedDamageSourceController)

	// "If damage from a red source is prevented this way, <this> deals
	// that much damage to the source's controller" (Honorable Passage).
	preventedRedDamageSourceControllerBody = game.DelayedBody("prevention/red-damage-source-controller", preventedRedDamageSourceController)

	// "If damage from a black source is prevented this way, you gain
	// that much life" (Shadowbane).
	preventedBlackGainLifeBody = game.DelayedBody("prevention/black-gain-life", preventedBlackGainLife)

	// "Exile cards from the top of your library equal to the damage
	// prevented this way" (Bone Mask).
	preventedExileLibraryTopBody = game.DelayedBody("prevention/exile-library-top", preventedExileLibraryTop)
)

func preventedGainLife(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 {
		return nil
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, p.Amount)
}

func preventedDamageSourceController(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 || p.Player == uuid.Nil {
		return nil
	}
	return g.DealDamageToPlayerForEffect(item.SourceCardID, p.Player, p.Amount)
}

func preventedRedDamageSourceController(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if !preventedFromColor(item, "R") {
		return nil
	}
	return preventedDamageSourceController(g, item, p)
}

func preventedBlackGainLife(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if !preventedFromColor(item, "B") {
		return nil
	}
	return preventedGainLife(g, item, p)
}

func preventedExileLibraryTop(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 {
		return nil
	}
	who := g.PlayerByIDForEffect(item.Controller)
	if who == nil || who.Library == nil {
		return nil
	}
	cards := who.Library.Cards
	n := min(p.Amount, len(cards))
	top := make([]uuid.UUID, 0, n)
	// The library's top is its LAST element.
	for i := len(cards) - 1; i >= len(cards)-n; i-- {
		top = append(top, cards[i].InstanceID)
	}
	g.ExileCardsForEffect(top)
	return nil
}

// preventedFromColor reports whether the prevented damage came from a
// source of `color` as it was when it would have dealt it.
func preventedFromColor(item *game.StackItem, color string) bool {
	return item != nil && item.Trigger != nil && slices.Contains(item.Trigger.Event.Colors, color)
}

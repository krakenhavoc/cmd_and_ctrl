package game

import (
	"github.com/google/uuid"
)

// cast_gate.go — S42, ADR 0073 §7: the one announce-time answer to
// "may this player cast this spell at all?" (CR 101.2).
//
// Before this file nothing could say no. `restrictions.go` carries
// per-PERMANENT bits — CantAttack, CantBlock, CantActivate — and a
// "can't cast" restriction is not about a permanent: it is about a
// CAST, and it has to see the caster, the card, the face and the zone
// together. `CastPermissionForLocked` (ADR 0066) answers the opposite
// question and says so in its own doc comment: "a permission that
// carries a 'cast only if' condition is checked by the one
// announce-time cast gate that issue designs, not here."
//
// TWO SOURCES, and the split is the CR's own:
//
//  1. A STATIC ON A PERMANENT. Rule of Law, Grafdigger's Cage,
//     Rakdos, Lord of Riots. Collected from the battlefield through
//     CatalogAbilityKey, exactly as CostModifiersForCard collects a
//     cost modifier — so a permanent that has lost its abilities
//     (CR 613.1f) stops restricting, and one whose designation gate
//     is unsatisfied is not there at all. Nothing is stored, which
//     IS the duration: the source leaving stops the restriction on
//     the next query.
//
//  2. THE SPELL'S OWN CONDITION. CR 307.6's legendary sorcery ("you
//     may cast this spell only if you control a legendary creature
//     or planeswalker"), and the "cast only if" family generally. It
//     is declared on the card because it is printed on the card, and
//     it is read off the CardDef under CatalogKey so a multi-face
//     card answers for the face being cast.
//
// What is deliberately NOT here:
//
//   - SPLIT SECOND. It restricts taking an ACTION, blocks activations
//     too, and correctly does not block a land play (CR 702.61b, a
//     special action). Folding it into a per-card cast gate would
//     either change the land rule or make this function carry an
//     exception with nothing to do with casting. CastSpell and the
//     enumerator keep their own check, beside the gate call.
//   - BANS WITH A DURATION. Silence's "this turn" and Reflector
//     Mage's "until your next turn" want the turn-scoped and
//     permanent-duration registries (#755). A third source slots into
//     castRestrictionsLocked without changing this function's
//     signature; that is the extension point.
//   - A BAN ON A CHOSEN CARD NAME. Meddling Mage and Nevermore need a
//     choose-a-card-name prompt, and the engine has NamedTribe and
//     ChosenColor but no name.

// CastQuery is everything a cast restriction may look at. Passed by
// value for the reason CostQuery is: a restriction is consulted
// inside the cast path under the write lock, and handing it pointers
// into the battlefield slice would invite a card file to mutate the
// board while deciding whether a spell may be cast.
//
// Game is a pointer because the interesting predicates are board-wide
// or tally reads — "more than one spell each turn", "unless an
// opponent lost life this turn". Treat it as READ-ONLY: a *ForEffect
// accessor is fine, a mutator is a bug and a public locking mutator
// is a deadlock.
type CastQuery struct {
	// Game is the game the cast is being announced in. Read-only.
	Game *Game

	// Card is the spell being cast, as it stands at announce, with
	// the chosen face already materialised (ADR 0034). The face rides
	// the card rather than a separate parameter so the two can never
	// disagree about what is being cast.
	Card Card

	// Controller is the player casting the spell — the "you" in
	// "YOU can't cast creature spells", the "each player" of Rule of
	// Law measured one seat at a time.
	Controller uuid.UUID

	// Source is the permanent contributing the restriction. Its
	// Controller is the "you" of the ABILITY, as against Controller
	// above which is the "you" of the CAST: Rakdos restricts its own
	// controller, Archon of Emeria restricts everyone else. Zero for
	// the spell's own condition, which has no source permanent.
	Source Card

	// FromZone is where the spell is being cast from — hand, the
	// command zone, a graveyard, exile, the top of a library.
	// Grafdigger's Cage is a predicate on nothing else.
	FromZone ZoneKind

	// OptionalCosts are the optional additional costs the caster
	// announced (CR 601.2b, ADR 0073). Present because CR 601.3a lets
	// a player begin casting when a choice made while proposing the
	// spell could lift a ban, and the optional-cost choice is made
	// before this gate runs. Nothing in the catalog reads it yet.
	OptionalCosts []int
}

// CastRestriction is one "can't cast" static ability contributed by a
// permanent on the battlefield.
//
// Declared as a struct of hooks rather than an interface for the
// reason StaticAbility, ReplacementEffect and CostModifier are: a
// card file writes a literal, not a type.
type CastRestriction struct {
	// Label is the clause as printed ("Each player can't cast more
	// than one spell each turn"). It is not decoration: a refused
	// cast returns it to the client, so the toast names the card that
	// said no rather than saying "invalid parameter".
	Label string

	// Forbids decides whether this restriction refuses THIS cast.
	// Nil forbids nothing, which makes an under-declared card file a
	// no-op rather than a lockout.
	//
	// Evaluated once per cast per restriction, under the write lock.
	// Read-only.
	Forbids func(q CastQuery) bool

	// ActiveWhen is the CR 716 / 719 / 721 designation gate, the same
	// field CostModifier carries and evaluated in the same accessor,
	// so a gated-off restriction never reaches the gate and the
	// engine, the enumerator and the view cannot disagree about
	// whether it exists (ADR 0071).
	ActiveWhen Designation
}

// CatalogCastRestrictions is the catalog hook the effects package
// wires at init, mirroring CatalogCostModifiers. Nil, or a nil
// return, means the card restricts nobody's casts — which is nearly
// every card.
var CatalogCastRestrictions func(oracleID string) []CastRestriction

// CastRestrictionsForCard returns the restrictions a permanent
// contributes right now: none for a permanent under an
// ability-removing effect (CatalogAbilityKey), and none for one whose
// designation gate is unsatisfied.
func CastRestrictionsForCard(c Card) []CastRestriction {
	if CatalogCastRestrictions == nil {
		return nil
	}
	key := CatalogAbilityKey(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogCastRestrictions(key), func(r CastRestriction) Designation {
		return r.ActiveWhen
	})
}

// CastConditionFor returns the card's own "you may cast this only
// if …" condition (CR 307.6 and the "cast only if" family), or nil.
// Read off the CardDef under CatalogKey, so a multi-face card answers
// for the face being cast and no other.
func CastConditionFor(card Card) func(g *Game, controller uuid.UUID, card Card) bool {
	if d := catalogDef(CatalogKey(card)); d != nil {
		return d.CastCondition
	}
	return nil
}

// CastConditionLabelFor is the printed clause behind CastConditionFor
// ("Cast this spell only if you control a legendary creature or
// planeswalker"), for the refusal message and the client's tooltip.
func CastConditionLabelFor(card Card) string {
	if d := catalogDef(CatalogKey(card)); d != nil {
		return d.CastConditionLabel
	}
	return ""
}

// CantCastError is what the gate returns: the sentinel plus the
// printed clause that refused the cast, so the client can say which
// card said no. Wraps ErrCantCast, so callers doing
// errors.Is(err, ErrCantCast) still match — the same shape
// InsufficientManaError has.
type CantCastError struct {
	// Reason is the printed clause ("Each player can't cast more than
	// one spell each turn"), never an engine phrase.
	Reason string

	// Source is the permanent whose static refused the cast, or
	// uuid.Nil when the spell's own condition did.
	Source uuid.UUID
}

func (e *CantCastError) Error() string {
	if e.Reason == "" {
		return "game: an effect prevents casting this spell"
	}
	return "game: " + e.Reason
}

// Unwrap lets errors.Is(err, ErrCantCast) match a *CantCastError.
func (e *CantCastError) Unwrap() error { return ErrCantCast }

// CastGateLocked is CR 101.2's "can't beats may", asked once, at
// announce. It reports nil when `caster` may cast `card` out of
// `zone` right now, and a *CantCastError naming the clause that
// refused it otherwise.
//
// THREE CALLERS, ONE FUNCTION, and that is the entire point:
//
//   - CastSpell, after the face, the source zone, the permission and
//     the announce-time choices are settled and BEFORE any cost is
//     paid. Every free cast ends up in CastSpell — cascade, a granted
//     permission, an impulse grant — so "can't beats may" falls out
//     of the placement and needs no rule of its own.
//   - legal.castMovesForCard, so a bot is never offered a banned cast.
//   - protocol.stampLegalTargets, so the client greys the card from
//     server data rather than from a rule it reimplemented.
//
// `card` carries the chosen face (ADR 0034 materialises it before any
// announce gate runs). `params` is the announcement so far; the
// enumerator and the view pass a zero value, which is honest — they
// are asking whether ANY cast of this card is open, and no catalog
// restriction reads the announcement today.
//
// A land PLAY is not a cast (CR 305.1, CR 116.2a) and is not gated
// here. CastSpell's land branch runs after this call and is untouched
// by it, because every restriction the catalog can express is written
// about casting.
//
// Caller must hold g.mu (read or write).
func (g *Game) CastGateLocked(caster uuid.UUID, card Card, zone ZoneKind, params CastSpellParams) error {
	q := CastQuery{
		Game:          g,
		Card:          card,
		Controller:    caster,
		FromZone:      zone,
		OptionalCosts: params.OptionalCosts,
	}
	// The statics first: a restriction from the board beats the
	// card's own permission to exist, and it is the cheaper check on
	// the overwhelmingly common board where nothing restricts
	// anything.
	if g.Battlefield != nil && CatalogCastRestrictions != nil {
		for i := range g.Battlefield.Cards {
			src := g.Battlefield.Cards[i]
			for _, r := range CastRestrictionsForCard(src) {
				if r.Forbids == nil {
					continue
				}
				q.Source = src
				if r.Forbids(q) {
					return &CantCastError{Reason: r.Label, Source: src.InstanceID}
				}
			}
		}
	}
	// Then the spell's own condition (CR 307.6). Last because a card
	// that is legal to cast on its own terms is still stopped by the
	// board, and reporting the board's reason is the more useful
	// message when both apply.
	if cond := CastConditionFor(card); cond != nil && !cond(g, caster, card) {
		return &CantCastError{Reason: CastConditionLabelFor(card)}
	}
	return nil
}

// AnyCastRestrictionsForEffect reports whether anything on the
// battlefield restricts casting at all. The fast negative the view
// and the enumerator take before walking a zone card by card — in
// almost every game nothing restricts anything, and the gate is
// otherwise pure cost on a hand of seven.
//
// Deliberately ignores the spell's own condition: that is a per-card
// read the caller is already doing when it looks the card up.
//
// Caller must hold g.mu (read or write).
func (g *Game) AnyCastRestrictionsForEffect() bool {
	if g == nil || g.Battlefield == nil || CatalogCastRestrictions == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		if len(CastRestrictionsForCard(g.Battlefield.Cards[i])) > 0 {
			return true
		}
	}
	return false
}

package game

import "github.com/google/uuid"

// exhaust_permission.go — #1184. "You may activate exhaust abilities
// as though they haven't been activated" (Elvish Refueler).
//
// # Why the gate needed a second argument
//
// #1181 built the exhaust gate as one predicate over one record:
// exhaustedLocked(source, exhaust, label) answered "has this object's
// ability been activated", and that was the whole rule. It is a fact
// about the OBJECT, so nobody had to say who was asking.
//
// A permission is not a fact about the object. Elvish Refueler does
// not refresh anything — the record still says the ability was
// activated, and every other player still reads it that way — it lets
// ONE player act as though the record said otherwise. So the question
// stopped being "is this ability spent" and became "is this ability
// spent FOR YOU", and every reader had to start naming the asker:
// ActivateCatalogAbility names the activator, internal/legal names
// the seat it is enumerating for, the view names the controller whose
// menu it is stamping. That is the same set of three (five, with the
// mana half) that #1181 pointed at one reader so they could not
// disagree; they still read one reader, it just takes one more
// argument now.
//
// # Why it is not a one-shot that clears the record
//
// The tempting shape is an entry point that deletes one entry of
// Activations.Ever, and it is wrong twice over. The card grants a
// permission continuously while its conditions hold ("During your
// turn, as long as you haven't activated an exhaust ability this
// turn"), so there is no moment to run a clear AT; and a clear would
// be visible to every player and survive the Refueler dying, which
// "as though" never does (CR 609.4 — an effect that lets you do
// something as though a rule were different changes nothing else).
// Keeping the record honest and reading it through a permission is
// what makes both of those fall out instead of being coded.
//
// # What the permission suspends, and what it does not
//
// The gate only. An exhaust ability under this permission still costs
// what it costs, still checks its Condition, still needs an untapped
// source for a {T}; and activating it WRITES the record again, which
// is what makes Elvish Refueler's own condition self-limiting — the
// second activation is the exhaust ability you activated this turn,
// so the permission is off for the rest of the turn without anything
// having to turn it off.

// ExhaustPermission is one "you may activate exhaust abilities as
// though they haven't been activated" static contributed by a
// permanent on the battlefield.
//
// A struct of hooks rather than an interface, for the reason
// StaticAbility, CostModifier and TriggeredAbility are: a card file
// writes a literal.
type ExhaustPermission struct {
	// Label is the clause as printed, for the event log and for
	// debugging an exhaust ability that came back when it should not
	// have.
	Label string

	// Applies decides whether this permanent's permission reaches the
	// player asking. `asker` is the player the gate is being answered
	// for; `source` is the permanent contributing the permission, so
	// source.Controller is the "you" of the ability and asker is the
	// "you" of the activation — Elvish Refueler needs both equal, and
	// a hypothetical "each player may" would not.
	//
	// Nil means "every player, always", which no printed card wants
	// and which is therefore refused at Register rather than
	// silently granted.
	//
	// Evaluated under g.mu. READ-ONLY: a *ForEffect accessor is fine,
	// a mutator is a bug and a public locking one is a deadlock.
	Applies func(g *Game, asker uuid.UUID, source Card) bool

	// ActiveWhen is the ADR 0071 designation gate, the same slot
	// CostModifier and TriggeredAbility carry. The zero value is "no
	// gate". Evaluated in ExhaustPermissionsForCard and nowhere else,
	// so a gated-off permission never reaches the walk below and the
	// engine, the enumerator and the view cannot disagree.
	ActiveWhen Designation
}

// CatalogExhaustPermissions is the catalog hook, mirroring
// CatalogCostModifiers. Nil, or a nil return, means the card grants
// nobody anything — which is every card but one.
var CatalogExhaustPermissions func(key string) []ExhaustPermission

// ExhaustPermissionsForCard is the permissions a permanent
// contributes right now, ability removal and designation gate
// applied. CatalogAbilityKey and not CatalogKey, because this is a
// static ability and a permanent under a CR 613.1f ability-removing
// effect has stopped granting it (ADR 0046).
func ExhaustPermissionsForCard(c Card) []ExhaustPermission {
	if CatalogExhaustPermissions == nil {
		return nil
	}
	key := CatalogAbilityKey(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogExhaustPermissions(key), func(p ExhaustPermission) Designation {
		return p.ActiveWhen
	})
}

// exhaustPermittedLocked reports whether any permanent on the
// battlefield lets `asker` activate exhaust abilities as though they
// had not been activated.
//
// CR 113.6: a permanent's static ability works from the battlefield,
// which is why this is a battlefield walk and not a wider one.
//
// Consulted ONLY after the record has already said "spent", so the
// walk costs nothing on the ordinary path: an ability that does not
// print the keyword, or whose key nothing has written, never reaches
// here. Caller must hold g.mu.
func (g *Game) exhaustPermittedLocked(asker uuid.UUID) bool {
	if g == nil || g.Battlefield == nil || asker == uuid.Nil || CatalogExhaustPermissions == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		src := g.Battlefield.Cards[i]
		for _, perm := range ExhaustPermissionsForCard(src) {
			if perm.Applies != nil && perm.Applies(g, asker, src) {
				return true
			}
		}
	}
	return false
}

// ExhaustAbilitiesActivatedThisTurn counts the exhaust abilities
// `player` has activated this turn, both kinds — the CR 602
// announcement and CR 605's mana abilities, which is the whole set of
// things "you haven't activated an exhaust ability this turn" is about.
//
// An event scan rather than a tally field, for the reason AGENTS.md
// gives for one: this is a FILTERED question (whose activation, and
// was it an exhaust one) that no counter carries, and
// Game.Activations is keyed by object rather than by player so it
// cannot answer "you". EventsThisTurn is bounded at the real turn
// boundary, and this runs only while a permission is on the
// battlefield — exhaustPermittedLocked's walk has already found one
// by the time a card's Applies asks.
//
// Caller must hold g.mu.
func (g *Game) ExhaustAbilitiesActivatedThisTurn(player uuid.UUID) int {
	if g == nil || player == uuid.Nil {
		return 0
	}
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if !ev.Exhaust || ev.Actor != player {
			continue
		}
		switch ev.Kind {
		case EventActivateAbility, EventManaAbilityActivated:
			n++
		}
	}
	return n
}

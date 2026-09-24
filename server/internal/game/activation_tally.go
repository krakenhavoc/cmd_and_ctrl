package game

import "github.com/google/uuid"

// activation_tally.go is the engine's record of what has been
// ACTIVATED — CR 602.2b's announcement, not CR 608's resolution
// (#1181).
//
// # Why it is not TurnTally
//
// TurnTally already counts what a source's abilities RESOLVED or
// TRIGGERED this turn (turn_tally.go, #586), and neither number
// answers the two questions the catalog keeps asking:
//
//   - an activation is spent whether or not it resolves. An exhaust
//     ability countered on the stack, or fizzled for want of a legal
//     target, has still been activated, so a record written at
//     resolution would hand the player a second use of it. The write
//     is therefore at the announce, beside StackItem.SourceEpoch, on
//     the one path that pays for an activation.
//   - exhaust never refreshes. "Activate each exhaust ability only
//     once" is a fact about the whole GAME, and TurnTally is emptied
//     on every turn advance by construction.
//
// So the record carries BOTH scopes of one count. `Ever` is the
// game-lifetime half that exhaust reads; `Turn` is the per-turn half
// the "Per-source activations-this-turn count" seam row
// (docs/engine-seams.md — Quirion Ranger, Wirewood Symbiote, boast)
// will read when its cards are written, and it is here rather than in
// a later PR because the two are one write at one call site and
// splitting them would mean writing the call site twice.
//
// # The key is an OBJECT and an ability, not a card and an ability
//
// ObjectTallyKey(source, Card.ObjectEpoch, label) — the same key
// ResolvedThisTurn and TriggeredThisTurn already use, for the same
// reason (CR 400.7: "an object that moves from one zone to another
// becomes a new object with no memory of its previous existence").
// The consequences are the printed ones and they fall out rather than
// being coded:
//
//   - a FLICKER refreshes an exhaust ability. Exile and return is two
//     zone changes, so the returning permanent reads a key nothing has
//     written.
//   - a PHASE-OUT does not. Phasing is not a zone change (CR 702.26d),
//     so the epoch is untouched and the permanent comes back spent.
//     Built in #1199 (game/phasing.go, ADR 0084), and it keeps that
//     promise the only way it could: phaseOutLocked and phaseInLocked
//     move the Card value between g.Battlefield and g.PhasedOut
//     without going through MoveCard, which is the one function that
//     bumps the epoch.
//
//     (The citation used to read CR 702.25f, which is a rule that does
//     not exist: 702.25 is Flanking. 702.26d is the sentence it was
//     reaching for.)
//   - a COPY of the permanent has its own exhausts (CR 707.2: a copy
//     takes the copiable values, and what a permanent has already done
//     is not one of them). A token copy is a different instance ID, so
//     it is a different key with nothing in it.
//
// The LABEL rather than the ability's index, because the index is a
// position in ActivatedAbilitiesForCard's FILTERED list and an
// ADR 0071 designation gate switching an ability on renumbers the ones
// behind it. A label is what the card printed.
//
// # Nothing is deleted at the battlefield exit
//
// Deliberately, and for the reason battlefield_exit.go gives about
// TurnTally's per-ability counts: the epoch is already the forgetting,
// so the entries the old object wrote are unreachable rather than
// absent. Unlike TurnTally's, `Ever`'s entries are never flushed by a
// turn boundary either — that is what "for the whole game" means, and
// the cost is one map entry per ability ever activated.
//
// # Mana abilities write here too (#1183)
//
// They take the other entry point — ActivateManaAbility, CR 605.3b: no
// stack, no priority, no announcement to hang a record on — so #1181
// left them out and `effects.ManaAbility` carried no Exhaust marker at
// all. Loot, the Pathfinder prints "Exhaust — {G}, {T}: Add three mana
// of any one color" beside two ordinary exhaust abilities, so the
// combination had to become spellable.
//
// TWO write sites for one activation kind, which is one more than the
// CR 602 path needs and is the whole difficulty:
//
//   - ActivateManaAbility, the hand click, beside the same gate;
//   - materializePlanLocked, the AUTO-TAPPER's executor, which taps a
//     permanent and mints its mana directly rather than routing
//     through ActivateManaAbility. A plan that spent an exhaust
//     ability without recording it would hand the player the ability
//     back, and the planner refuses to plan an exhausted one for the
//     same reason it refuses a gated or a sick source.
//
// Both take the key with g.activationTallyKeyLocked before anything is
// paid, for the reason spelled out below.

// ActivationTally counts activations of one printed ability of one
// object, in the two scopes the rules ask about.
//
// Both maps are keyed by ObjectTallyKey(source, epoch, label). A nil
// map reads as zero everywhere, which is why nothing allocates one
// until the first activation.
type ActivationTally struct {
	// Ever is the whole-game count and is never reset. Exhaust
	// ("Activate each exhaust ability only once") is the reader: one
	// is already too many.
	Ever map[string]int `json:"ever,omitempty"`

	// Turn is the same count for the current turn, emptied on the
	// turn advance beside TurnTally. No card reads it yet — it is the
	// half "Activate only once each turn" (Quirion Ranger, Wirewood
	// Symbiote) and boast (CR 702.142a) want.
	Turn map[string]int `json:"turn,omitempty"`
}

// noteAbilityActivationLocked records one activation of `key` in both
// scopes. `key` is an ObjectTallyKey, built by
// activationTallyKeyLocked before any cost is paid — see the comment
// there on why it cannot be built afterwards.
//
// Caller must hold g.mu.
func (g *Game) noteAbilityActivationLocked(key string) {
	if key == "" {
		return
	}
	if g.Activations.Ever == nil {
		g.Activations.Ever = map[string]int{}
	}
	if g.Activations.Turn == nil {
		g.Activations.Turn = map[string]int{}
	}
	g.Activations.Ever[key]++
	g.Activations.Turn[key]++
}

// activationTallyKeyLocked is the key for one printed ability of the
// object `source` names RIGHT NOW.
//
// It has to be taken BEFORE the costs are paid. A cost that moves the
// source — SacrificeSelf, DiscardSelf, a self-bounce — has already
// ended the object by the time the announcement is built, and
// Card.ObjectEpoch has moved on with it, so a key taken at the
// announce would be written against an object that never had the
// ability. Nothing could then read it back, which for exhaust means
// the wrong answer rather than a harmless one.
//
// Caller must hold g.mu.
func (g *Game) activationTallyKeyLocked(source uuid.UUID, label string) string {
	return g.objectTallyKeyLocked(source, label)
}

// ActivatedThisGame reports how many times the ability labelled
// `label` of the OBJECT `source` names has been activated in this
// game. An empty label sums every ability of that object.
//
// Caller must hold g.mu.
func (g *Game) ActivatedThisGame(source uuid.UUID, label string) int {
	return sumTally(g.Activations.Ever, g.objectTallyPrefixLocked(source), label)
}

// ActivatedThisTurn is the same count for the current turn. Caller
// must hold g.mu.
func (g *Game) ActivatedThisTurn(source uuid.UUID, label string) int {
	return sumTally(g.Activations.Turn, g.objectTallyPrefixLocked(source), label)
}

// AbilityExhausted is the WHOLE exhaust gate, and it is one function
// so that the three places the rule has to be enforced cannot
// disagree: ActivateCatalogAbility (which refuses with
// ErrAbilityExhausted), internal/legal (which does not enumerate the
// move) and protocol's ActivatedAbilityView (which stamps `exhausted`
// and greys the row). #544's rule — a bot is never offered a move the
// engine refuses — is a property of there being one reader, not of
// three copies agreeing.
//
// False for every ability that is not an exhaust ability, which is all
// of them but the marked ones.
//
// `asker` is the player the question is being answered FOR — the
// activator, the seat the enumerator is working on, the controller
// whose menu the view is stamping (#1184). It matters because a
// permission can suspend the gate for one player and nobody else; see
// exhaust_permission.go on why that could not be a fact about the
// object. uuid.Nil is legal and means "nobody in particular", which
// reads the record straight.
//
// Caller must hold g.mu.
func (g *Game) AbilityExhausted(asker, source uuid.UUID, ab ActivatedAbilityShape) bool {
	return g.exhaustedLocked(asker, source, ab.Exhaust, ab.Label)
}

// ManaAbilityExhausted is AbilityExhausted for the OTHER ability kind
// (#1183). The same record, the same key, the same sentence — and one
// more reader than its twin has, because a mana ability can be spent
// by the auto-tapper as well as by a click:
//
//  1. ActivateManaAbility — refuses with ErrAbilityExhausted before
//     any cost is validated, so a second click taps nothing;
//  2. legal.manaMoves — does not enumerate the move;
//  3. protocol's ManaAbilityView — stamps `exhausted`, and the client
//     greys the row with the string it already has;
//  4. the AUTO-TAPPER, both halves. gatherTapSources will not plan a
//     spent ability as a mana source (through autoTapAbilityFor, the
//     one picker the planner and the executor share) and
//     materializePlanLocked re-asks before it taps, exactly as it
//     re-asks the gate and the counter cost — a plan can arrive stale;
//  5. ProducibleManaLocked — CR 106.7's "could produce", so a
//     Reflecting Pool next to a spent Loot is not priced on mana Loot
//     can never make again.
//
// The shapes are two structs and not one because the two ability kinds
// are two structs everywhere else in the engine; the PREDICATE is one
// function (exhaustedLocked), which is the part that must not drift.
//
// Caller must hold g.mu.
func (g *Game) ManaAbilityExhausted(asker, source uuid.UUID, ab ManaAbilityShape) bool {
	return g.exhaustedLocked(asker, source, ab.Exhaust, ab.Label)
}

// exhaustedLocked is the rule itself: an ability that prints the
// keyword and whose (object, label) key has been written once is spent
// for the rest of this object's life. "Activate each exhaust ability
// only once" — one is already too many.
//
// False for every ability that is not an exhaust ability, which is all
// of them but the marked ones, and it is the FIRST test so that the
// ordinary permanent costs nothing but a bool read.
//
// #1184 adds the third clause, and its ORDER is the point: the record
// is read first and the permission only afterwards, so an ability
// nobody has spent never walks the battlefield looking for a
// permission it does not need, and a permission can only ever turn a
// "yes" into a "no" — never the other way.
//
// Caller must hold g.mu.
func (g *Game) exhaustedLocked(asker, source uuid.UUID, exhaust bool, label string) bool {
	if !exhaust {
		return false
	}
	if g.ActivatedThisGame(source, label) == 0 {
		return false
	}
	return !g.exhaustPermittedLocked(asker)
}

// resetActivationTurnTallyLocked empties the per-turn half and leaves
// the game-lifetime half alone. Called from the turn advance beside
// resetTurnTallyLocked. Caller must hold g.mu.
func (g *Game) resetActivationTurnTallyLocked() {
	g.Activations.Turn = nil
}

// cloneActivationTally deep-copies both maps, so an undo snapshot and
// the live game cannot share an entry.
func cloneActivationTally(t ActivationTally) ActivationTally {
	return ActivationTally{
		Ever: copyStringIntMap(t.Ever),
		Turn: copyStringIntMap(t.Turn),
	}
}

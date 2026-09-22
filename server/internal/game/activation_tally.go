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
//   - a PHASE-OUT does not. Phasing is not a zone change (CR 702.25f),
//     so the epoch is untouched and the permanent comes back spent.
//     Phasing is not modelled in this engine at all; when it is, the
//     rule it must respect to keep this true is "phasing does not bump
//     ObjectEpoch".
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
// # What it does not count yet
//
// Mana abilities. They take the other entry point (ActivateManaAbility,
// CR 605.3a — no stack, no priority) and nothing writes here from it,
// so ManaAbility carries no Exhaust marker and the combination is
// unspellable rather than silently ignored. One printed card wants it
// — Loot, the Pathfinder's "Exhaust — {G}, {T}: Add three mana of any
// one color" — and it is named in ADR 0020's exhaust addendum.

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
// Caller must hold g.mu.
func (g *Game) AbilityExhausted(source uuid.UUID, ab ActivatedAbilityShape) bool {
	if !ab.Exhaust {
		return false
	}
	return g.ActivatedThisGame(source, ab.Label) > 0
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

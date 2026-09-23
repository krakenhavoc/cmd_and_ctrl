package game

import (
	"strings"

	"github.com/google/uuid"
)

// producible_mana.go — CR 106.7's "the type of mana a permanent could
// produce", the one question Exotic Orchard, Reflecting Pool and
// Fellwar Stone ask of another permanent (#782).
//
// CR 106.7: "The type of mana a permanent could produce at any time
// includes any type of mana that an ability of that permanent would
// produce if the ability were to resolve at that time." Two halves of
// that sentence do the work:
//
//   - "WOULD produce … AT THAT TIME" — the current answer, not the
//     printed one. A Thriving Isle that chose red would produce {U}
//     or {R} and nothing else; before its colour is chosen it would
//     produce {U}; an Uncharted Haven with no colour chosen would
//     produce nothing at all. So this asks each ability what it would
//     add NOW, through exactly the seam the activation uses —
//     ProducedFunc for the derived and scaled shapes, then
//     manaPickOptions for the identity narrowing (#844) and ordering
//     (#875). Whatever a tap would offer is what "could produce"
//     answers, by construction.
//   - "IF THE ABILITY WERE TO RESOLVE" — costs and timing are not
//     asked about. A tapped opposing Island still offers {U}; a
//     Temple of the False God its controller could not activate still
//     offers {C}; summoning sickness, an empty pool for a Signet's
//     {1} and a Condition that is false today are all irrelevant.
//
// # The Scryfall fallback, and why it is a fallback
//
// Before #782 this unioned the card's mana abilities with
// Card.ProducedMana — the `produced_mana` array Scryfall stamps at
// deck import. The union was wrong in both directions at once:
//
//   - TOO MANY colours. Scryfall lists all five for every
//     chosen-colour land, because all five are printable outcomes. An
//     Exotic Orchard facing an opponent's Thriving Isle that chose red
//     could tap for white, black or green — stronger than printed,
//     the direction #259 refuses.
//   - NONE AT ALL. A catalog-built card with no Scryfall record — a
//     token, a demo seed, a test fixture — got nothing, because the
//     whole mana ability of a Thriving Isle is a ProducedFunc and
//     every ProducedFunc was skipped by the recursion guard.
//
// So the catalog answer WINS when there is one, and `produced_mana`
// answers only for a card the catalog has never heard of: an imported
// land with no spec and no basic land type, where Scryfall's array is
// the only thing that knows anything. A card whose abilities all
// answer "nothing" answers nothing — falling through to Scryfall
// there would defeat both the chosen-colour read and the recursion
// guard.
//
// # Exhaust is the one restriction this asks about (#1183)
//
// The second bullet above says costs and timing are not asked about: a
// tapped Island still offers {U}, a Temple of the False God its
// controller could not activate still offers {C}, a Condition that is
// false today is irrelevant. A SPENT EXHAUST ability is the one
// exception, and it is deliberate.
//
// Read strictly, CR 106.7 would count it: "Activate each exhaust
// ability only once" is an activation restriction of the same family
// as a Condition, and the rule asks what the ability WOULD produce if
// it were to resolve. Counting it is the stronger-than-printed
// direction for the reader that matters — a cast priced on a
// Reflecting Pool that is deriving a colour from a Loot, the
// Pathfinder whose exhaust is already gone is a cast the auto-tapper
// cannot actually pay for, and the executor would tap the Pool for
// nothing on the way to it. So this answers no, which is one colour
// short in exactly the direction the recursion guard below is already
// short in, and it keeps the CR 106.7 reader and the auto-tapper (which
// refuses the same ability through autoTapAbilityFor) saying the same
// thing about the same permanent.
//
// The narrowing is worth one sentence out loud because it is the only
// place in this file that a card can be "could produce nothing" for a
// reason that is not about its output: ADR 0020's exhaust addendum
// (#1183) is where the decision lives.
//
// # The recursion guard is CR 106.7 itself, exactly (#1323)
//
// There is no CR 106.6b in the pinned edition (MagicCompRules
// 20260819) — the circular case is the last sentence of 106.7 itself:
// "If that permanent wouldn't produce any mana under these
// conditions, OR NO TYPE OF MANA CAN BE DEFINED THIS WAY, there's no
// type of mana it could produce." A chain of "could produce" reads
// that loops back on itself is exactly a type that "can't be defined
// this way" — asking for it never bottoms out — so the rule's own
// answer for it is the same "no type of mana" as every other
// undefined case. Every prior comment on this file and on the three
// card files that cited "CR 106.6b" was citing a rule number that
// does not exist in the pinned edition; this amendment corrects them
// to 106.7.
//
// The card's own official ruling (Exotic Orchard, WotC, 2009-02-01)
// gives the worked example this engine now matches move for move:
//
//	"Lands that produce mana based only on what other lands 'could
//	 produce' won't help each other unless SOME OTHER LAND allows one
//	 of them to actually produce some type of mana. For example, if
//	 you control an Exotic Orchard and your opponent controls an
//	 Exotic Orchard and a Reflecting Pool, none of those lands would
//	 produce mana if their mana abilities were activated. On the other
//	 hand, if you control a Forest and an Exotic Orchard, and your
//	 opponent controls an Exotic Orchard and a Reflecting Pool, then
//	 each of those lands can be tapped to produce {G}."
//
// That is a REACHABILITY question over the "derives from" graph — a
// permanent produces a colour iff there is a path from it, through
// derived sources, to a REAL (non-derived) source of that colour,
// and a loop with no real source anywhere along it produces nothing
// because no such path exists (there IS no such path — every route
// out of it leads back into the loop). `producibleManaVisitingLocked`
// marks `c.InstanceID` on entry and (`defer`) UNMARKS it on return,
// so `visiting` tracks the ANCESTOR PATH rather than "every permanent
// this query has ever touched" — the textbook shape for cycle
// detection in a reachability walk (a node ON the current path is a
// real cycle; a node already fully resolved on a SIBLING branch is
// not, and gets walked again rather than being told it doesn't
// exist). A permanently-growing set gives the same TOP-LEVEL answer
// here too, because the aggregation is a monotone union with nothing
// ever discarded — a colour found via any one path is retained by
// that path's own return value regardless of what a redundant,
// masked reference elsewhere would have found — but the ancestor-path
// version is the one that is correct BY CONSTRUCTION rather than by
// an argument about this specific aggregation, so it is what ships.
//
// A `DerivesFromOtherSources` ability (ManaAbilityShape — Exotic
// Orchard, Reflecting Pool, Fellwar Stone, and nothing else) declares
// DerivedMatch instead of ProducedFunc, and this file walks the
// battlefield and recurses ITSELF rather than going through the
// generic ProducedFunc closure, because the path set has to thread
// the whole way down and a `func(*Game, controller, source uuid.UUID)
// string` closure (the shape every OTHER mana ability's ProducedFunc
// uses) has no fourth argument to carry it.
//
// Every other ProducedFunc IS evaluated, which is the half #782
// added: a chosen colour (the Thriving lands, the Gates, Uncharted
// Haven, Crossroads Village, Mirage Mesa, Valgavoth's Lair), a board
// count (Cabal Coffers, Gaea's Cradle), a devotion (Nyx Lotus), a
// creature's power (Marwyn). A Reflecting Pool therefore now sees a
// Cabal Coffers' {B} — one of the simplifications ADR 0040 declared.
//
// Read-only, and it runs under g.mu: the callers are ProducedFuncs,
// which ActivateManaAbility reaches holding the write lock and the
// auto-tapper holding the read lock.

// ProducibleManaLocked is the mana `c` could produce right now, as
// distinct symbols in WUBRGC order. Includes "C" — a caller that
// means COLOURS filters it out, which is the whole difference between
// Exotic Orchard's "any color" and Reflecting Pool's "any type".
//
// Caller must hold g.mu.
func (g *Game) ProducibleManaLocked(c Card) []string {
	return g.producibleManaVisitingLocked(c, map[uuid.UUID]bool{})
}

// producibleManaVisitingLocked is ProducibleManaLocked's real body,
// carrying the set of instance IDs currently ON THE PATH from the
// original top-level query (#1323). A card already on that path
// contributes nothing — CR 106.7's "no type of mana can be defined
// this way" — and everything else works exactly as
// ProducibleManaLocked always did.
//
// The `defer` unmark keeps `visiting` an ANCESTOR path rather than
// "every permanent this query has ever asked about" — the shape that
// is correct by construction for a reachability walk with cycles,
// rather than by an argument specific to this file's aggregation (see
// the recursion-guard comment above the file's imports for why a
// permanently-growing set would answer the same TOP-LEVEL question
// here too, and why the ancestor-path version ships anyway).
//
// Caller must hold g.mu.
func (g *Game) producibleManaVisitingLocked(c Card, visiting map[uuid.UUID]bool) []string {
	if visiting[c.InstanceID] {
		return nil
	}
	visiting[c.InstanceID] = true
	defer delete(visiting, c.InstanceID)

	abilities := ManaAbilitiesForCard(c)
	if len(abilities) == 0 {
		// Nothing in the catalog, no instance ability and no basic
		// land type: an imported card the engine knows only through
		// its Scryfall record.
		return scryfallProducibleMana(c)
	}
	// The identity is computed at most once per card, and only for a
	// card that actually narrows — manaPickOptions only REORDERS an
	// ordinary pipe, and a set has no order. A derivation walks every
	// land on the battlefield, so the four narrowing cards should not
	// make the other thirty-six re-scan a player's zones.
	var identity commanderIdentity
	var haveIdentity bool
	seen := map[string]bool{}
	for _, ab := range abilities {
		// #1183: a SPENT exhaust ability. The one place this function
		// looks past "if the ability were to resolve" at something
		// that stops it being activated, and it is a deliberate,
		// declared narrowing rather than an oversight — see the
		// "Exhaust is the one restriction this asks about" section
		// above.
		if g.ManaAbilityExhausted(c.Controller, c.InstanceID, ab) {
			continue
		}
		if ab.DerivesFromOtherSources {
			// #1323: recurse through the SAME visiting set, rather
			// than the blanket skip this used to be. A card already
			// in `visiting` (the circular case) contributes nothing
			// from inside derivedManaLocked; anything else is a
			// one-way chain and resolves normally.
			for _, m := range g.derivedManaLocked(c, &ab, visiting) {
				seen[m] = true
			}
			continue
		}
		// #789: an ability whose output depends on what its cost paid
		// is asked with the LARGEST payment it could make right now.
		// "Could produce" is about the possible (CR 106.7), so a
		// Mage-Ring Network holding three storage counters could
		// produce {C} and one holding none could not — and a
		// Reflecting Pool next to it has to see the same answer the
		// activation would give.
		produced := manaAbilityProducedLocked(g, c.Controller, c.InstanceID, &ab,
			g.maxCounterPaymentLocked(c.Controller, c.InstanceID, ab.RemoveCounters))
		slots, err := ParseProducedMana(produced)
		if err != nil {
			continue
		}
		if ab.NarrowToCommanderIdentity && !haveIdentity {
			identity, haveIdentity = commanderIdentityFor(g, g.playerByIDLocked(c.Controller)), true
		}
		for _, slot := range slots {
			// The same narrowing the activation, the enumerator and
			// the auto-tapper read, so "could produce" and "does
			// produce" cannot disagree. An empty answer is CR 903.4f's
			// "adds no mana" and contributes nothing.
			for _, opt := range manaPickOptions(slot.Options, identity, ab.NarrowToCommanderIdentity) {
				seen[strings.ToUpper(opt)] = true
			}
		}
	}
	return orderedManaSymbols(seen)
}

// derivedManaLocked is the CR 106.7 union for a DerivesFromOtherSources
// ability: every permanent on the battlefield `ab.DerivedMatch`
// accepts, asked with the SAME visited set the caller is already
// carrying, then filtered by DerivedColorsOnly. Shared by the CR 106.7
// reader above and by a real activation (manaAbilityProducedLocked,
// ActivateManaAbility), so the two can never disagree about what an
// ability adds — the same guarantee #782 built for every other
// ProducedFunc.
//
// Caller must hold g.mu.
func (g *Game) derivedManaLocked(c Card, ab *ManaAbilityShape, visiting map[uuid.UUID]bool) []string {
	if ab.DerivedMatch == nil {
		return nil
	}
	seen := map[string]bool{}
	for _, cand := range g.BattlefieldCardsForEffect() {
		if cand.InstanceID == c.InstanceID || !ab.DerivedMatch(cand, c.Controller) {
			continue
		}
		for _, m := range g.producibleManaVisitingLocked(cand, visiting) {
			if ab.DerivedColorsOnly && m == "C" {
				continue
			}
			seen[m] = true
		}
	}
	return orderedManaSymbols(seen)
}

// pipeString renders a symbol set as a single produced-mana slot —
// ["W","U"] becomes "{W|U}". Empty input returns "", which
// ActivateManaAbility and manaAbilityProducedLocked both treat as
// "this ability produced no mana": the printed outcome for Exotic
// Orchard facing no opposing lands.
func pipeString(symbols []string) string {
	if len(symbols) == 0 {
		return ""
	}
	return "{" + strings.Join(symbols, "|") + "}"
}

// scryfallProducibleMana is the import-only fallback: the
// `produced_mana` array, filtered to the six symbols the engine has a
// token for.
func scryfallProducibleMana(c Card) []string {
	seen := map[string]bool{}
	for _, m := range c.ProducedMana {
		switch u := strings.ToUpper(m); u {
		case "W", "U", "B", "R", "G", "C":
			seen[u] = true
		}
	}
	return orderedManaSymbols(seen)
}

// orderedManaSymbols flattens a symbol set into canonical WUBRGC
// order, so a derived pipe string is stable across activations and a
// test can assert on it.
func orderedManaSymbols(seen map[string]bool) []string {
	var out []string
	for _, m := range []string{"W", "U", "B", "R", "G", "C"} {
		if seen[m] {
			out = append(out, m)
		}
	}
	return out
}

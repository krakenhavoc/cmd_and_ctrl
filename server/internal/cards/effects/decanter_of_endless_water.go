package effects

// Decanter of Endless Water — Artifact for {3}:
//
//	"You have no maximum hand size."
//	"{T}: Add one mana of any color."
//
// Thought Vessel one mana later, trading colourless for fixing. Both
// halves are staples of the same decks: a hand-size lock is what lets
// a draw-heavy deck actually keep what it draws.
//
// # Why this is writable now
//
// The batch-02 triage (#295) filed this under "player / game-rule
// statics — no tracking issue yet", which was accurate when it was
// written. #338 closed it: Spec.NoMaxHandSize landed with Thought
// Vessel, and the engine DERIVES the answer at cleanup by asking the
// battlefield whether the active player controls any permanent with
// the bit set (EffectiveMaxHandSizeLocked via the
// game.CatalogNoMaxHandSize hook).
//
// That derivation is what makes two hand-size permanents, and one of
// two leaving, come out right with no bookkeeping — nothing is
// written to Player.MaxHandSize, so nothing has to be restored. It is
// deliberately NOT a Static entry: every CR 613 layer modifies a
// characteristic of an OBJECT, and "you have no maximum hand size"
// modifies a PLAYER, which has no characteristic and no layer to sit
// in.
//
// # The mana ability
//
// "Add one mana of any color" as pipe syntax, one picker rather than
// five menu entries. IgnoreCommanderIdentity is left FALSE to match
// Birds of Paradise, the catalog's existing "{T}: Add one mana of any
// color" — the engine narrows the pipe to the controller's commander
// colour identity, which can only remove a colour the controller
// could not spend on anything in their deck. (The flag exists for
// City of Brass and the painlands, whose printed text names colours
// for a reason and whose damage makes an off-identity pick
// meaningful.) Two cards with identical printed text should not
// disagree about this, and Birds came first.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:      "8ae98ef8-8f52-4877-a08c-1fae5514184e",
		Name:          "Decanter of Endless Water",
		Completeness:  CompletenessFull,
		NoMaxHandSize: true,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}

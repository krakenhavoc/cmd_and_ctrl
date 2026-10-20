package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hama_chain_test.go — #4177: a spell cast through Hama's waterbend
// permission keeps the convoke or improvise it has (CR 702.51a, 702.126a).
// The waterbend's taps pay only its {N} (CR 701.67b); a convoke or
// improvise tap pays the rest of the total cost, a tax for instance.

const (
	chordOfCallingOracle = "6789a170-f2c5-4fc0-8a45-2b2361e67410"
	arcReactorOracle     = "55f7b005-31ea-4865-a8f3-460d15bed795"
)

// become re-prints the exiled card as another catalog card, keeping its
// identity and its place in exile, so a Hama cast can be tried on a
// spell with convoke or improvise of its own.
func (h *hamaTable) become(name, typeLine, oracle, cost string) {
	h.g.WithWriteLock(func() {
		for i := range h.g.Exile.Cards {
			if c := &h.g.Exile.Cards[i]; c.InstanceID == h.rite {
				c.Name, c.TypeLine, c.OracleID, c.ManaCost = name, typeLine, oracle, cost
			}
		}
	})
}

// thalia seats an opponent's Thalia, Guardian of Thraben: the cast costs
// {1} more.
func (h *hamaTable) thalia() {
	b12Push(h.g, h.opp.ID, "Thalia, Guardian of Thraben", "Legendary Creature — Human Soldier", thaliaGuardianOracle, 2, 1)
	h.g.BumpLayerVersionForTest()
}

func (h *hamaTable) rocks(n int) []uuid.UUID {
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		out = append(out, b12Permanent(h.g, h.me.ID, "Rock", "Artifact"))
	}
	h.g.BumpLayerVersionForTest()
	return out
}

func cat(parts ...[]uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// Chord of Calling ({X}{G}{G}{G}) is a waterbend {3} at X 0; under
// Thalia the total is {4}. Three taps pay the waterbend and a creature
// tapped for convoke pays the tax, with no mana at all.
func TestHamaCastOfAConvokeSpellMayAlsoConvoke(t *testing.T) {
	h := hamaExiling(t)
	h.become("Chord of Calling", "Instant", chordOfCallingOracle, "{X}{G}{G}{G}")
	h.thalia()
	if err := h.cast(game.CastSpellParams{TapIDs: h.helpers}); err != nil {
		t.Fatalf("three waterbend taps and one convoke tap for a taxed {4}: %v", err)
	}
	if got := h.untapped(); got != 0 {
		t.Errorf("%d helpers untapped, want 0", got)
	}
	if !h.g.Stack.Contains(h.rite) {
		t.Fatal("the spell is not on the stack")
	}
}

// Convoke is creatures only: artifacts pay the waterbend's {3} but not
// the tax, and a tap past the total pays for nothing.
func TestHamaConvokeDoesNotWidenTheWaterbendPool(t *testing.T) {
	h := hamaExiling(t)
	h.become("Chord of Calling", "Instant", chordOfCallingOracle, "{X}{G}{G}{G}")
	h.thalia()
	rocks := h.rocks(4)
	if err := h.cast(game.CastSpellParams{TapIDs: rocks}); err == nil {
		t.Fatal("four artifacts paid {4}: the fourth has no convoke to pay the tax")
	}
	if err := h.cast(game.CastSpellParams{TapIDs: cat(h.helpers, rocks[:1])}); err == nil {
		t.Fatal("five taps accepted for a total of {4}")
	}
	if got := h.untapped(); got != 4 {
		t.Fatalf("a refused cast tapped creatures (%d untapped)", got)
	}
	if err := h.cast(game.CastSpellParams{TapIDs: cat(rocks[:3], h.helpers[:1])}); err != nil {
		t.Fatalf("three artifacts for the waterbend and a creature for the tax: %v", err)
	}
}

// Arc Reactor has improvise: under Thalia its {5} waterbend and {1} tax
// are paid by four creatures and two artifacts, the artifact left over
// improvising the tax.
func TestHamaCastOfAnImproviseSpellMayAlsoImprovise(t *testing.T) {
	h := hamaExiling(t)
	h.become("Arc Reactor", "Artifact", arcReactorOracle, "{5}")
	h.thalia()
	rocks := h.rocks(2)
	if err := h.cast(game.CastSpellParams{TapIDs: cat(h.helpers, rocks)}); err != nil {
		t.Fatalf("six taps for a waterbend {5} and a {1} tax with improvise: %v", err)
	}
	if got := h.untapped(); got != 0 {
		t.Errorf("%d helpers untapped, want 0", got)
	}
}

// Improvise is artifacts only: five creatures and nothing else cannot
// pay the tax of an improvise spell.
func TestHamaImproviseDoesNotWidenToCreatures(t *testing.T) {
	h := hamaExiling(t)
	h.become("Arc Reactor", "Artifact", arcReactorOracle, "{5}")
	h.thalia()
	extra := b12Creature(h.g, h.me.ID, "Helper E", "Creature — Human", 1, 1)
	h.g.BumpLayerVersionForTest()
	if err := h.cast(game.CastSpellParams{TapIDs: cat(h.helpers, []uuid.UUID{extra, b12Creature(h.g, h.me.ID, "Helper F", "Creature — Human", 1, 1)})}); err == nil {
		t.Fatal("six creatures paid a waterbend {5} and a tax with only improvise")
	}
}

// A spell with neither keyword takes the waterbend's taps and no more.
func TestHamaPlainSpellStillTakesOnlyTheWaterbendTaps(t *testing.T) {
	h := hamaExiling(t)
	h.thalia()
	if err := h.cast(game.CastSpellParams{TapIDs: h.helpers}); err == nil {
		t.Fatal("a fourth tap paid the tax of a spell with no convoke")
	}
}

// A board grant of convoke (Hoarding Broodlord: spells cast from exile)
// reaches a Hama cast too.
func TestHamaCastOfAGrantedConvokeSpell(t *testing.T) {
	h := hamaExiling(t)
	h.thalia()
	pushCatalogPermanent(h.g, h.me.ID, "Hoarding Broodlord", "Creature — Dragon", mcBroodlordOracle, false)
	h.g.BumpLayerVersionForTest()
	if err := h.cast(game.CastSpellParams{TapIDs: h.helpers}); err != nil {
		t.Fatalf("a granted convoke should pay the tax: %v", err)
	}
}

// Waterbender's Restoration prints a waterbend {X} additional cost. Cast
// through Hama it owes both: the {2} waterbend (its mana value) and the
// {X}, and no longer drops the X.
func TestHamaCastOfWaterbendersRestorationStillOwesItsX(t *testing.T) {
	h := hamaExiling(t)
	h.become("Waterbender's Restoration", "Instant — Lesson", "285046f6-b3c4-4eb7-8712-9dffebabc762", "{U}{U}")
	// {2} waterbend plus X 2 = {4}: the four helpers pay all of it.
	if err := h.cast(game.CastSpellParams{XValue: 2, TapIDs: h.helpers[:3]}); err == nil {
		t.Fatal("three taps and no mana paid a waterbend {2} and a waterbend {2}")
	}
	if err := h.cast(game.CastSpellParams{XValue: 2, TapIDs: h.helpers, Targets: h.helperTargets(2)}); err != nil {
		t.Fatalf("four taps for waterbend {2} and waterbend {X=2}: %v", err)
	}
}

func (h *hamaTable) helperTargets(n int) []game.TargetRef {
	var out []game.TargetRef
	for _, id := range h.helpers[:n] {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

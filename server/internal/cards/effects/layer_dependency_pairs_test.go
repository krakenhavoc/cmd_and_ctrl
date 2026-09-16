package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// layer_dependency_pairs_test.go pins the catalog pairs whose outcome
// depends on entry order today and should not.
//
// CR 613.8a: an effect depends on another in the same layer when
// applying the other changes what the first applies to. Dependent
// effects apply after the effects they depend on, whatever their
// timestamps. The layer engine applies each layer strictly in
// timestamp order (layers.go), so every pair below comes out right in
// one entry order and wrong in the other.
//
// Each pair is two tests. The one in the order that already works is
// live and guards against a regression. The one in the order that
// does not is skipped, and asserts the rules answer, so the change
// that implements CR 613.8 ordering only has to delete the t.Skip
// lines. The cards carry player-facing caveats for the same gaps;
// clear those in the same change.
//
// Song of the Dryads also has a gap that is not about ordering
// (CR 305.7: it removes only the permanent's own abilities). It is
// pinned at the bottom, skipped for the same reason.

const layerDependencySkip = "CR 613.8 dependency ordering is not implemented: layer 4 applies strictly by timestamp — tracked in the CR 613.8 dependency issue (split from #159)"

// assertForestSwamp checks subtypes and the mana the permanent makes.
func assertForestSwamp(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	sub := layeredCard(t, g, id).Effective().Subtypes
	if !slices.Contains(sub, "Forest") || !slices.Contains(sub, "Swamp") {
		t.Errorf("subtypes = %v, want Forest and Swamp", sub)
	}
	produced := manaAbilityProduced(t, g, id)
	if !slices.Contains(produced, "{G}") || !slices.Contains(produced, "{B}") {
		t.Errorf("mana abilities produce %v, want {G} and {B}", produced)
	}
}

// --- Urborg, Tomb of Yawgmoth + Song of the Dryads ------------------

// Urborg's "each land" reads the land type Song writes, so Urborg
// depends on Song: the Sol Ring is a Forest Swamp either way.
func TestLayerDependencyUrborgThenSongOfTheDryadsIsAForestSwamp(t *testing.T) {
	t.Skip(layerDependencySkip)
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", urborgOracle)
	solRing := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: solRingOracle, Owner: me.ID, Controller: me.ID,
	})
	enchant(t, g, "Song of the Dryads", songOfTheDryadsOracle, solRing)
	assertForestSwamp(t, g, solRing)
}

func TestLayerDependencySongOfTheDryadsThenUrborgIsAForestSwamp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	solRing := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: solRingOracle, Owner: me.ID, Controller: me.ID,
	})
	enchant(t, g, "Song of the Dryads", songOfTheDryadsOracle, solRing)
	seedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", urborgOracle)
	assertForestSwamp(t, g, solRing)
}

// --- Urborg, Tomb of Yawgmoth + a slumbering Arixmethes -------------

// castArixmethes casts Arixmethes from hand and lets it resolve, so
// it enters with its five slumber counters and is a land.
func castArixmethes(t *testing.T, g *game.Game, owner *game.Player) uuid.UUID {
	t.Helper()
	arix := uuid.New()
	owner.Hand.PushTop(game.Card{
		InstanceID: arix, Name: "Arixmethes, Slumbering Isle", TypeLine: "Legendary Creature — Kraken",
		OracleID: b18ArixmethesOracle, Power: 12, Toughness: 12, Owner: owner.ID, Controller: owner.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(owner.ID, arix, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Arixmethes: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, arix); c.IsCreature() || !c.IsLand() {
		t.Fatalf("slumbering Arixmethes: creature=%v land=%v, want a noncreature land", c.IsCreature(), c.IsLand())
	}
	return arix
}

func assertSlumberingSwamp(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	if sub := layeredCard(t, g, id).Effective().Subtypes; !slices.Contains(sub, "Swamp") {
		t.Errorf("subtypes = %v, want Swamp", sub)
	}
	produced := manaAbilityProduced(t, g, id)
	if !slices.Contains(produced, "{G}{U}") || !slices.Contains(produced, "{B}") {
		t.Errorf("mana abilities produce %v, want {G}{U} and {B}", produced)
	}
}

// Arixmethes' "it's a land" changes what Urborg applies to, so Urborg
// depends on it: a slumbering Arixmethes is a Swamp either way.
func TestLayerDependencyUrborgThenArixmethesIsASwamp(t *testing.T) {
	t.Skip(layerDependencySkip)
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", urborgOracle)
	arix := castArixmethes(t, g, me)
	assertSlumberingSwamp(t, g, arix)
}

func TestLayerDependencyArixmethesThenUrborgIsASwamp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	arix := castArixmethes(t, g, me)
	seedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", urborgOracle)
	assertSlumberingSwamp(t, g, arix)
}

// --- Maskwood Nexus + a crewed Vehicle ------------------------------

// crewCopterUnderChieftain puts a Goblin Chieftain and a Smuggler's
// Copter on the battlefield, with Maskwood Nexus entering before or
// after the crew ability resolves.
func crewCopterUnderChieftain(t *testing.T, nexusFirst bool) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	pushNexus := func() {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Maskwood Nexus", TypeLine: "Artifact",
			OracleID: maskwoodNexusOracle, Owner: owner.ID, Controller: owner.ID,
		})
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Chieftain",
		TypeLine: "Creature — Goblin", Power: 2, Toughness: 2,
		OracleID: goblinChieftainOracle, Owner: owner.ID, Controller: owner.ID,
	})
	if nexusFirst {
		pushNexus()
	}
	copter := pushVehicleForTest(g, owner.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	crewer := pushCrewerForTest(g, owner.ID, "Crewer", 1)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if err := g.ActivateCatalogAbility(owner.ID, copter, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !nexusFirst {
		pushNexus()
	}
	return g, copter
}

func assertCopterIsAPumpedGoblin(t *testing.T, g *game.Game, copter uuid.UUID) {
	t.Helper()
	v := layeredCard(t, g, copter)
	if !v.IsCreature() {
		t.Fatal("the crewed Copter is not a creature")
	}
	if !v.HasSubtype("Goblin") {
		t.Error("the crewed Copter is not a Goblin under Maskwood Nexus")
	}
	if got := v.CurrentPower(); got != 4 {
		t.Errorf("power = %d, want 4 (3 printed + Goblin Chieftain)", got)
	}
	if !game.HasKeyword(&v, "haste") {
		t.Error("the crewed Copter has no haste from Goblin Chieftain")
	}
}

// The Nexus's "creatures you control" reads the creature type crew
// adds, so the Nexus depends on the crew effect. Crew resolves after
// the Nexus entered in nearly every real game.
func TestLayerDependencyMaskwoodThenCrewedVehicleIsEveryType(t *testing.T) {
	t.Skip(layerDependencySkip)
	g, copter := crewCopterUnderChieftain(t, true)
	assertCopterIsAPumpedGoblin(t, g, copter)
}

func TestLayerDependencyCrewedVehicleThenMaskwoodIsEveryType(t *testing.T) {
	g, copter := crewCopterUnderChieftain(t, false)
	assertCopterIsAPumpedGoblin(t, g, copter)
}

// --- Maskwood Nexus + a permanent that turns itself off -------------

func pushNexusFor(g *game.Game, owner uuid.UUID) {
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Maskwood Nexus", TypeLine: "Artifact",
		OracleID: maskwoodNexusOracle, Owner: owner, Controller: owner,
	})
}

func assertNotEveryCreatureType(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	c := layeredCard(t, g, id)
	if c.IsCreature() {
		t.Fatal("the permanent is a creature; the pair needs it switched off")
	}
	if c.HasSubtype("Goblin") {
		t.Errorf("a noncreature permanent is a Goblin under Maskwood Nexus (abilities %v)", c.Effective().Abilities)
	}
}

// The Triad's "isn't a creature" changes what the Nexus applies to,
// so the Nexus depends on it: with a short graveyard the Triad has no
// creature types, whichever entered first.
func TestLayerDependencyMaskwoodThenWarringTriadIsNotEveryType(t *testing.T) {
	t.Skip(layerDependencySkip)
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushNexusFor(g, me.ID)
	triad := b31Push(g, me.ID, "The Warring Triad", "Legendary Artifact Creature — God", b31TheWarringTriadOracle, "{3}", 5, 5)
	assertNotEveryCreatureType(t, g, triad)
}

func TestLayerDependencyWarringTriadThenMaskwoodIsNotEveryType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	triad := b31Push(g, me.ID, "The Warring Triad", "Legendary Artifact Creature — God", b31TheWarringTriadOracle, "{3}", 5, 5)
	pushNexusFor(g, me.ID)
	assertNotEveryCreatureType(t, g, triad)
}

func TestLayerDependencyMaskwoodThenSlumberingArixmethesIsNotEveryType(t *testing.T) {
	t.Skip(layerDependencySkip)
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushNexusFor(g, me.ID)
	arix := castArixmethes(t, g, me)
	assertNotEveryCreatureType(t, g, arix)
}

func TestLayerDependencySlumberingArixmethesThenMaskwoodIsNotEveryType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	arix := castArixmethes(t, g, me)
	pushNexusFor(g, me.ID)
	assertNotEveryCreatureType(t, g, arix)
}

// --- Song of the Dryads and abilities other effects granted ---------

// CR 305.7 removes only the abilities from the permanent's own rules
// text, and Song's 2014-11-07 ruling says the permanent "will still
// have any abilities it gained from other effects." The engine models
// the loss as a full layer-6 ability wipe, so Boros Charm's
// indestructible is gone too and a wrath destroys the Forest.
func TestSongOfTheDryadsKeepsAnAbilityAnotherEffectGranted(t *testing.T) {
	t.Skip("Song of the Dryads removes every ability, not only the permanent's own (CR 305.7) — tracked in the CR 613.8 dependency issue (split from #159), which moves the removal to layer 4")
	g := newCatalogGame(t)
	me := g.Seats[0]
	solRing := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: solRingOracle, Owner: me.ID, Controller: me.ID,
	})
	castModal(t, g, "Boros Charm", "Instant", borosCharmOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, solRing); !game.HasKeyword(&c, "indestructible") {
		t.Fatal("Boros Charm did not make the Sol Ring indestructible")
	}
	enchant(t, g, "Song of the Dryads", songOfTheDryadsOracle, solRing)
	if c := layeredCard(t, g, solRing); !game.HasKeyword(&c, "indestructible") {
		t.Errorf("the Song'd Sol Ring lost Boros Charm's indestructible (abilities %v)", c.Effective().Abilities)
	}
}

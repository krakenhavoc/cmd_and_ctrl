package testsl

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// layer_dependency_pairs_test.go pins the catalog pairs whose outcome
// must NOT depend on entry order.
//
// CR 613.8a: an effect depends on another in the same layer when
// applying the other changes what the first applies to. Dependent
// effects apply after the effects they depend on, whatever their
// timestamps. Each pair is therefore two tests, one per entry order,
// and both assert the same rules answer. They were pinned as skips by
// #644 while the layer engine ordered layer 4 by timestamp alone; ADR
// 0067 implemented CR 613.8 and the skips came off with it.
//
// The negative case at the bottom is as load-bearing as the pairs.
// Urborg and Song of the Dryads on a permanent that was ALREADY a
// land are independent — the Song changes neither what Urborg applies
// to nor what it does — so timestamp order stands and the newer Song
// wins. A detector that diffed Urborg's output instead of its
// instruction would call that a dependency and get it wrong.
//
// Song of the Dryads' other pinned gap was not about ordering at all
// (CR 305.7: it removes only the permanent's own abilities). It is at
// the bottom too, and it went with the same ADR.

// assertForestSwamp checks subtypes and the mana the permanent makes.
func assertForestSwamp(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	sub := effectstest.LayeredCard(t, g, id).Effective().Subtypes
	if !slices.Contains(sub, "Forest") || !slices.Contains(sub, "Swamp") {
		t.Errorf("subtypes = %v, want Forest and Swamp", sub)
	}
	produced := effectstest.ManaAbilityProduced(t, g, id)
	if !slices.Contains(produced, "{G}") || !slices.Contains(produced, "{B}") {
		t.Errorf("mana abilities produce %v, want {G} and {B}", produced)
	}
}

// --- Urborg, Tomb of Yawgmoth + Song of the Dryads ------------------

// Urborg's "each land" reads the land type Song writes, so Urborg
// depends on Song: the Sol Ring is a Forest Swamp either way.
func TestLayerDependencyUrborgThenSongOfTheDryadsIsAForestSwamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", effectstest.UrborgOracle)
	solRing := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: effectstest.SolRingOracle, Owner: me.ID, Controller: me.ID,
	})
	effectstest.Enchant(t, g, "Song of the Dryads", effectstest.SongOfTheDryadsOracle, solRing)
	assertForestSwamp(t, g, solRing)
}

func TestLayerDependencySongOfTheDryadsThenUrborgIsAForestSwamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	solRing := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: effectstest.SolRingOracle, Owner: me.ID, Controller: me.ID,
	})
	effectstest.Enchant(t, g, "Song of the Dryads", effectstest.SongOfTheDryadsOracle, solRing)
	effectstest.SeedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", effectstest.UrborgOracle)
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
		OracleID: effectstest.B18ArixmethesOracle, Power: 12, Toughness: 12, Owner: owner.ID, Controller: owner.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(owner.ID, arix, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Arixmethes: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, arix); c.IsCreature() || !c.IsLand() {
		t.Fatalf("slumbering Arixmethes: creature=%v land=%v, want a noncreature land", c.IsCreature(), c.IsLand())
	}
	return arix
}

func assertSlumberingSwamp(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	if sub := effectstest.LayeredCard(t, g, id).Effective().Subtypes; !slices.Contains(sub, "Swamp") {
		t.Errorf("subtypes = %v, want Swamp", sub)
	}
	produced := effectstest.ManaAbilityProduced(t, g, id)
	if !slices.Contains(produced, "{G}{U}") || !slices.Contains(produced, "{B}") {
		t.Errorf("mana abilities produce %v, want {G}{U} and {B}", produced)
	}
}

// Arixmethes' "it's a land" changes what Urborg applies to, so Urborg
// depends on it: a slumbering Arixmethes is a Swamp either way.
func TestLayerDependencyUrborgThenArixmethesIsASwamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", effectstest.UrborgOracle)
	arix := castArixmethes(t, g, me)
	assertSlumberingSwamp(t, g, arix)
}

func TestLayerDependencyArixmethesThenUrborgIsASwamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arix := castArixmethes(t, g, me)
	effectstest.SeedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", effectstest.UrborgOracle)
	assertSlumberingSwamp(t, g, arix)
}

// --- Maskwood Nexus + a crewed Vehicle ------------------------------

// crewCopterUnderChieftain puts a Goblin Chieftain and a Smuggler's
// Copter on the battlefield, with Maskwood Nexus entering before or
// after the crew ability resolves.
func crewCopterUnderChieftain(t *testing.T, nexusFirst bool) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	pushNexus := func() {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Maskwood Nexus", TypeLine: "Artifact",
			OracleID: effectstest.MaskwoodNexusOracle, Owner: owner.ID, Controller: owner.ID,
		})
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Chieftain",
		TypeLine: "Creature — Goblin", Power: 2, Toughness: 2,
		OracleID: effectstest.GoblinChieftainOracle, Owner: owner.ID, Controller: owner.ID,
	})
	if nexusFirst {
		pushNexus()
	}
	copter := effectstest.PushVehicleForTest(g, owner.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	crewer := effectstest.PushCrewerForTest(g, owner.ID, "Crewer", 1)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if err := g.ActivateCatalogAbility(owner.ID, copter, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !nexusFirst {
		pushNexus()
	}
	return g, copter
}

func assertCopterIsAPumpedGoblin(t *testing.T, g *game.Game, copter uuid.UUID) {
	t.Helper()
	v := effectstest.LayeredCard(t, g, copter)
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
	g, copter := crewCopterUnderChieftain(t, true)
	assertCopterIsAPumpedGoblin(t, g, copter)
}

func TestLayerDependencyCrewedVehicleThenMaskwoodIsEveryType(t *testing.T) {
	g, copter := crewCopterUnderChieftain(t, false)
	assertCopterIsAPumpedGoblin(t, g, copter)
}

func assertNotEveryCreatureType(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	c := effectstest.LayeredCard(t, g, id)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushNexusFor(g, me.ID)
	triad := effectstest.B31Push(g, me.ID, "The Warring Triad", "Legendary Artifact Creature — God", effectstest.B31TheWarringTriadOracle, "{3}", 5, 5)
	assertNotEveryCreatureType(t, g, triad)
}

func TestLayerDependencyWarringTriadThenMaskwoodIsNotEveryType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	triad := effectstest.B31Push(g, me.ID, "The Warring Triad", "Legendary Artifact Creature — God", effectstest.B31TheWarringTriadOracle, "{3}", 5, 5)
	effectstest.PushNexusFor(g, me.ID)
	assertNotEveryCreatureType(t, g, triad)
}

func TestLayerDependencyMaskwoodThenSlumberingArixmethesIsNotEveryType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushNexusFor(g, me.ID)
	arix := castArixmethes(t, g, me)
	assertNotEveryCreatureType(t, g, arix)
}

func TestLayerDependencySlumberingArixmethesThenMaskwoodIsNotEveryType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arix := castArixmethes(t, g, me)
	effectstest.PushNexusFor(g, me.ID)
	assertNotEveryCreatureType(t, g, arix)
}

// --- Song of the Dryads and abilities other effects granted ---------

// CR 305.7 removes only the abilities from the permanent's own rules
// text, and Song's 2014-11-07 ruling says the permanent "will still
// have any abilities it gained from other effects." The Song's loss
// is a LAYER 4 removal (SetsBasicLandType), which is what makes that
// true: Boros Charm's grant lands in layer 6, after it.
func TestSongOfTheDryadsKeepsAnAbilityAnotherEffectGranted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	solRing := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: effectstest.SolRingOracle, Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastModal(t, g, "Boros Charm", "Instant", effectstest.BorosCharmOracle, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, solRing); !game.HasKeyword(&c, "indestructible") {
		t.Fatal("Boros Charm did not make the Sol Ring indestructible")
	}
	effectstest.Enchant(t, g, "Song of the Dryads", effectstest.SongOfTheDryadsOracle, solRing)
	if c := effectstest.LayeredCard(t, g, solRing); !game.HasKeyword(&c, "indestructible") {
		t.Errorf("the Song'd Sol Ring lost Boros Charm's indestructible (abilities %v)", c.Effective().Abilities)
	}
}

// CR 305.7's other two clauses on the Song: the permanent's own mana
// ability goes, and the Forest's intrinsic {G} arrives with the type.
// A Sol Ring that used to add {C}{C} adds {G}.
func TestSongOfTheDryadsSwapsTheIntrinsicManaAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	solRing := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: effectstest.SolRingOracle, Owner: me.ID, Controller: me.ID,
	})
	if got := effectstest.ManaAbilityProduced(t, g, solRing); len(got) != 1 || got[0] != "{C}{C}" {
		t.Fatalf("fixture is wrong: Sol Ring produces %v, want {C}{C}", got)
	}
	effectstest.Enchant(t, g, "Song of the Dryads", effectstest.SongOfTheDryadsOracle, solRing)
	if got := effectstest.ManaAbilityProduced(t, g, solRing); len(got) != 1 || got[0] != "{G}" {
		t.Errorf("the Song'd Sol Ring produces %v, want only the Forest's {G} (CR 305.6, CR 305.7)", got)
	}
}

// --- the negative case: independent effects keep timestamp order ----

// Urborg and a Song of the Dryads on a Command Tower are INDEPENDENT.
// The Tower is a land whether or not the Song has applied, so the
// Song changes neither what Urborg applies to nor what "add Swamp"
// does — CR 613.8a's test is about the instruction, not about the
// result. Timestamp order stands, the newer Song's type SET wins, and
// the Tower is a Forest with no Swamp.
//
// This is the assertion that fails if dependency detection is built
// by diffing an effect's output instead of its instruction.
func TestUrborgAndSongOnAnExistingLandStayIndependent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", effectstest.UrborgOracle)
	tower := effectstest.SeedLand(g, me.ID, "Command Tower", "Land", "")
	effectstest.Enchant(t, g, "Song of the Dryads", effectstest.SongOfTheDryadsOracle, tower)

	sub := effectstest.LayeredCard(t, g, tower).Effective().Subtypes
	if !slices.Contains(sub, "Forest") {
		t.Errorf("subtypes = %v, want Forest", sub)
	}
	if slices.Contains(sub, "Swamp") {
		t.Errorf("subtypes = %v, want no Swamp: the two effects are independent, so the newer Song wins", sub)
	}
}

// --- #670: every creature type is a layer-4 fact --------------------

// Maskwood Nexus' 2021-02-05 ruling, for a GRANT rather than a
// printed changeling: Kenrith's Transformation removes all abilities
// in layer 6, and the Nexus' grant is a layer-4 type effect, so the
// Elk is still every creature type. The two are independent — the
// Elk is a creature either way — so timestamp order decides which
// SUBTYPES it has, and the every-type fact rides with the newer of
// the two.
func TestMaskwoodNexusAndKenrithsTransformationByTimestamp(t *testing.T) {
	// Nexus newer: its grant applies after the Elk set, so the Elk is
	// also every creature type.
	t.Run("Nexus newer", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.SeedBear(g, me.ID)
		effectstest.Enchant(t, g, "Kenrith's Transformation", effectstest.KenrithTransformOracle, bear)
		effectstest.Settle(t, g)
		effectstest.PushNexusFor(g, me.ID)

		c := effectstest.LayeredCard(t, g, bear)
		if !c.HasLostAllAbilities() {
			t.Fatal("fixture is wrong: the Bear was not silenced")
		}
		if !slices.Contains(c.Effective().Subtypes, "Elk") {
			t.Errorf("subtypes = %v, want Elk", c.Effective().Subtypes)
		}
		if !c.HasSubtype("Goblin") {
			t.Error("the silenced Elk is not every creature type: a layer-6 removal deleted a layer-4 grant")
		}
	})
	// Nexus older: "is a green Elk creature" is a type SET newer than
	// the grant, so it replaces every creature type with Elk.
	t.Run("Nexus older", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.SeedBear(g, me.ID)
		effectstest.PushNexusFor(g, me.ID)
		effectstest.Enchant(t, g, "Kenrith's Transformation", effectstest.KenrithTransformOracle, bear)
		effectstest.Settle(t, g)

		c := effectstest.LayeredCard(t, g, bear)
		if !slices.Contains(c.Effective().Subtypes, "Elk") {
			t.Errorf("subtypes = %v, want Elk", c.Effective().Subtypes)
		}
		if c.HasSubtype("Goblin") {
			t.Error("a later layer-4 subtype SET must overwrite every creature type (CR 205.1b)")
		}
	})
}

// The other direction of the same rule, on a card that sets its own
// types: a Duplicant that imprinted a Human Wizard is a Human Wizard
// Shapeshifter and NOT every creature type, even with an older
// Maskwood Nexus out. Predicted from the code by #670 and not
// reachable before it: the marker used to ride the ability list,
// which a subtype set does not clear.
func TestMaskwoodNexusThenDuplicantSetsOnlyTheExiledTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushNexusFor(g, me.ID)
	wizard := effectstest.B27Push(g, opp.ID, "Archmage", "Legendary Creature — Human Wizard", "", "{3}{U}{U}", 5, 6, "U")
	dup := effectstest.B20CastCreature(t, g, me, "Duplicant", "Artifact Creature — Shapeshifter", effectstest.B27DuplicantOracle, 2, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, wizard)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.LayeredCard(t, g, dup)
	if !slices.Contains(c.Effective().Subtypes, "Wizard") {
		t.Fatalf("fixture is wrong: Duplicant subtypes = %v", c.Effective().Subtypes)
	}
	if c.HasSubtype("Goblin") {
		t.Error("Duplicant is still every creature type: a layer-4 subtype SET must overwrite the Nexus' grant")
	}
}

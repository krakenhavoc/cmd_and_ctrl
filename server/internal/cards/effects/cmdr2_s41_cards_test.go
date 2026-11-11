package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s41_cards_test.go — cmdr2 batch 41, spellslinger: storm, cascade and
// spell copies (#5368). Each test plays the real catalog entry through the
// engine and asserts what the printed text says.

const (
	cmdr2S41ErisOracle        = "918991a6-7321-419d-8f3d-150fe32d02d2"
	cmdr2S41StellaOracle      = "8d0da2f9-0e05-4a94-8d42-fd57d06b09ba"
	cmdr2S41StagecoachOracle  = "83f87884-f0f9-482e-887e-35a8a5c20eb0"
	cmdr2S41SlingerOracle     = "ed7ba558-5341-4c7b-a9c5-b382dee88a13"
	cmdr2S41OctaviaOracle     = "65681a2a-8395-4390-845f-9be974994dcb"
	cmdr2S41FoundryOracle     = "118da256-d1ea-44e8-9026-317e49694d29"
	cmdr2S41ArchmageOracle    = "f1ecb3d7-5ea0-45b2-b1b0-d7ae304db2a1"
	cmdr2S41FeatherOracle     = "957413ae-e756-4afe-886a-b57bf59d5f8d"
	cmdr2S41MurmurationOracle = "84e7be36-b8cd-4111-a91b-3628f421d5d5"
	cmdr2S41WildsearOracle    = "9ceac4c6-e62e-4445-852c-b2c621e4a83b"
	cmdr2S41CaseOracle        = "409697d5-ce3c-46aa-a97d-969d143a2a65"
)

func TestCmdr2S41EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		cmdr2S41ErisOracle:        "Eris, Roar of the Storm",
		cmdr2S41StellaOracle:      "Stella Lee, Wild Card",
		cmdr2S41StagecoachOracle:  "Smoldering Stagecoach",
		cmdr2S41SlingerOracle:     "Crackling Spellslinger",
		cmdr2S41OctaviaOracle:     "Octavia, Living Thesis",
		cmdr2S41FoundryOracle:     "Forger's Foundry",
		cmdr2S41ArchmageOracle:    "Archmage of Echoes",
		cmdr2S41FeatherOracle:     "Feather, Radiant Arbiter",
		cmdr2S41MurmurationOracle: "Murmuration",
		cmdr2S41WildsearOracle:    "Wildsear, Scouring Maw",
		cmdr2S41CaseOracle:        "Case of the Shifting Visage",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Wildsear, Scouring Maw ----------------------------------------------------

func TestCmdr2S41WildsearCascadesEnchantmentsCastFromHandOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 8)
	pushCatalogPermanent(g, me.ID, "Wildsear, Scouring Maw", "Legendary Creature — Elemental Wolf", cmdr2S41WildsearOracle, false)

	castWithCost(t, g, "A Sorcery", "Sorcery", "{3}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 0 {
		t.Errorf("a sorcery got %d cascade offers, want 0", got)
	}

	castWithCost(t, g, "An Enchantment", "Enchantment", "{3}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 1 {
		t.Fatalf("an enchantment cast from hand got %d cascade offers, want 1", got)
	}
	answerAllMayCast(t, g, me.ID, false)
}

func TestCmdr2S41WildsearDoesNotCascadeAnEnchantmentCastFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 8)
	pushCatalogPermanent(g, me.ID, "Wildsear, Scouring Maw", "Legendary Creature — Elemental Wolf", cmdr2S41WildsearOracle, false)
	advanceToMain(t, g)
	id := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: id, Name: "Exiled Enchantment", TypeLine: "Enchantment", ManaCost: "{3}{G}",
		Owner: me.ID, Controller: me.ID})
	g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{Player: me.ID, Zone: game.ZoneExile,
		Duration: game.WhileInZoneDuration(), Source: uuid.New()})
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "GGGG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: string(game.ZoneExile)}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 0 {
		t.Errorf("an enchantment cast from exile got %d cascade offers, want 0", got)
	}
}

// --- Crackling Spellslinger ----------------------------------------------------

func TestCmdr2S41SpellslingerStormsTheNextInstantOrSorceryOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stormFiller(t, g) // one spell before the Spellslinger

	castCatalogSpell(t, g, "Crackling Spellslinger", "Creature — Human Wizard", cmdr2S41SlingerOracle, nil)
	passPriorityAroundTable(t, g)

	// A creature spell neither storms nor uses the watch up.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)

	before := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	// Three spells came before it this turn (the filler, the Spellslinger, the Bear).
	answerCopyPrompts(t, g, me.ID, opp.ID, 3)
	if got := lifeOf(g, opp.ID); got != before-12 {
		t.Errorf("life %d -> %d, want -12 (the spell and three copies, 3 each)", before, got)
	}

	// The watch was used up: the next one does not storm.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	before = lifeOf(g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != before-3 {
		t.Errorf("the second spell dealt %d, want 3: it does not have storm", before-got)
	}
}

func TestCmdr2S41SpellslingerDoesNothingIfItWasNotCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Crackling Spellslinger", "Creature — Human Wizard", cmdr2S41SlingerOracle, false)
	passPriorityAroundTable(t, g)
	stormFiller(t, g)
	before := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != before-3 {
		t.Errorf("a Spellslinger that was not cast gave storm: dealt %d, want 3", before-got)
	}
}

var _ = uuid.Nil

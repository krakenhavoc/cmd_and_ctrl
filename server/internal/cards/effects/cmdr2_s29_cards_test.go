package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s29_cards_test.go — batch 29, Delirium (Commander decks, #5368). Each
// test plays the real catalog entry and asserts what its text says. Delirium
// is seeded with b24SeedTypes (four is delirium, three is not).

const (
	s29SlimeOracle     = "b4429e71-434e-480c-a8d1-2ad11de8b2d9"
	s29SpawnOracle     = "17bdd827-2e89-4559-939a-fc5c153189ed"
	s29CultivatorOracl = "26e555f1-e0ed-4b84-8783-1c0c968f462d"
	s29IshkanahOracle  = "6f4861f4-1b3b-4dbc-b55b-196075d0a693"
	s29HeatOracle      = "b23cba46-6ed2-442d-bc39-7c4933f6b083"
	s29UrsineOracle    = "647feda9-2876-4430-902f-10dc1f4f99a9"
	s29DescendOracle   = "9cbf95f4-3510-420b-9355-645450f4493d"
	s29DryadOracle     = "e823639f-1308-47f6-9b51-8fd9a682489a"
	s29DelugeOracle    = "a81fcdf8-59e7-471d-96ce-f87e656572b1"
	s29WinterOracle    = "e70a01dd-4dbe-4811-ac93-1c6d0142e06e"
	s29SkinnerOracle   = "e7e59437-cdd0-476a-b1ec-231be4139d31"
	s29FlayerOracle    = "21254bdb-7b1f-4ac6-8fdd-f0ac2c35f47d"
)

func TestCmdr2S29EveryBuiltCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s29CultivatorOracl: "Deathcap Cultivator",
		s29DryadOracle:     "Gnarlwood Dryad",
		s29FlayerOracle:    "Grim Flayer",
	} {
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

// s29Bump makes the layers re-read the graveyard after a test seeds it.
func s29Bump(g *game.Game) { g.WithWriteLock(func() { g.BumpLayerVersionForTest() }) }

func s29Size(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectivePower(t, g, id), effectiveToughness(t, g, id)
}

func TestCmdr2S29GnarlwoodDryadGrowsOnlyWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dryad := b12Push(g, me.ID, "Gnarlwood Dryad", "Creature — Dryad Horror", s29DryadOracle, 1, 1)
	b24SeedTypes(me, 3)
	if p, tou := s29Size(t, g, dryad); p != 1 || tou != 1 {
		t.Errorf("three card types: %d/%d, want 1/1", p, tou)
	}
	gyCard(me, "Old Forest", "Basic Land — Forest", "", "")
	s29Bump(g)
	if p, tou := s29Size(t, g, dryad); p != 3 || tou != 3 {
		t.Errorf("four card types: %d/%d, want 3/3", p, tou)
	}
}

func TestCmdr2S29DeathcapCultivatorHasDeathtouchOnlyWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c := b12Push(g, me.ID, "Deathcap Cultivator", "Creature — Human Druid", s29CultivatorOracl, 2, 1)
	b24SeedTypes(me, 3)
	if hasKeywordOf(g, c, "deathtouch") {
		t.Error("three card types: has deathtouch")
	}
	gyCard(me, "Old Forest", "Basic Land — Forest", "", "")
	s29Bump(g)
	if !hasKeywordOf(g, c, "deathtouch") {
		t.Error("four card types: no deathtouch")
	}
}

func TestCmdr2S29DeathcapCultivatorTapsForBlackOrGreen(t *testing.T) {
	spec, ok := Lookup(s29CultivatorOracl)
	if !ok || len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{B|G}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Fatalf("Deathcap Cultivator mana abilities: %+v", spec.ManaAbilities)
	}
}

func TestCmdr2S29GrimFlayerGrowsAndSurveilsOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flayer := b12Push(g, me.ID, "Grim Flayer", "Creature — Human Warrior", s29FlayerOracle, 2, 2)
	b24SeedTypes(me, 4)
	s29Bump(g)
	if p, tou := s29Size(t, g, flayer); p != 4 || tou != 4 {
		t.Fatalf("four card types: %d/%d, want 4/4", p, tou)
	}
	attackWith(t, g, opp.ID, flayer)
	if opp.Life != 40-4 {
		t.Errorf("opponent life %d, want 36", opp.Life)
	}
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("no surveil prompt after combat damage to a player")
	}
}

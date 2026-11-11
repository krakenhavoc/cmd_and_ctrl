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
		s29HeatOracle:      "Unholy Heat",
		s29DelugeOracle:    "Deluge of Doom",
		s29DescendOracle:   "Descend upon the Sinful",
		s29SlimeOracle:     "Convert to Slime",
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

func s29Token(g *game.Game, name string) (game.Card, bool) {
	var out game.Card
	var ok bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.IsToken() {
				out, ok = c, true
			}
		}
	})
	return out, ok
}

func s29OnBattlefield(g *game.Game, id uuid.UUID) bool {
	var on bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				on = true
			}
		}
	})
	return on
}

// --- Unholy Heat ---------------------------------------------------------------------

func s29Heat(t *testing.T, seed int) (*game.Game, int) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b24SeedTypes(me, seed)
	target := b12Creature(g, opp.ID, "Big Wall", "Creature — Wall", 0, 9)
	castCatalogSpell(t, g, "Unholy Heat", "Instant", s29HeatOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)
	return g, b24Damage(t, g, target)
}

func TestCmdr2S29UnholyHeatDealsTwoWithoutDelirium(t *testing.T) {
	if _, dmg := s29Heat(t, 3); dmg != 2 {
		t.Errorf("three card types: %d damage, want 2", dmg)
	}
}

func TestCmdr2S29UnholyHeatDealsSixWithDelirium(t *testing.T) {
	if _, dmg := s29Heat(t, 4); dmg != 6 {
		t.Errorf("four card types: %d damage, want 6", dmg)
	}
}

func TestCmdr2S29UnholyHeatRefusesAPlayerTarget(t *testing.T) {
	g := newCatalogGame(t)
	err := castCatalogSpellErr(t, g, "Unholy Heat", "Instant", s29HeatOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	if err == nil {
		t.Error("Unholy Heat accepted a player target")
	}
}

// --- Deluge of Doom ------------------------------------------------------------------

func TestCmdr2S29DelugeOfDoomShrinksEveryCreatureByTheTypeCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b24SeedTypes(me, 3)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 5, 5)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	small := b12Creature(g, opp.ID, "Their Mouse", "Creature — Mouse", 1, 3)
	castCatalogSpell(t, g, "Deluge of Doom", "Sorcery", s29DelugeOracle, nil)
	passPriorityAroundTable(t, g)
	// Three types in the graveyard; the resolving spell is not counted.
	if p, tou := s29Size(t, g, mine); p != 2 || tou != 2 {
		t.Errorf("my bear %d/%d, want 2/2", p, tou)
	}
	if s29OnBattlefield(g, theirs) || s29OnBattlefield(g, small) {
		t.Error("a creature with toughness 3 survived -3/-3")
	}
}

// --- Descend upon the Sinful ---------------------------------------------------------

func TestCmdr2S29DescendUponTheSinfulExilesAllCreaturesAndMakesAnAngelWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b24SeedTypes(me, 4)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Descend upon the Sinful", "Sorcery", s29DescendOracle, nil)
	passPriorityAroundTable(t, g)
	if s29OnBattlefield(g, mine) || s29OnBattlefield(g, theirs) {
		t.Error("a creature survived")
	}
	for _, c := range g.Exile.Cards {
		if c.InstanceID == mine {
			goto exiled
		}
	}
	t.Error("my bear was not exiled")
exiled:
	angel, ok := s29Token(g, "Angel")
	if !ok {
		t.Fatal("no Angel token with four card types")
	}
	if p, tou := s29Size(t, g, angel.InstanceID); p != 4 || tou != 4 || !hasKeywordOf(g, angel.InstanceID, "flying") {
		t.Errorf("Angel %d/%d, want a 4/4 flier", p, tou)
	}
}

func TestCmdr2S29DescendUponTheSinfulMakesNoAngelWithoutDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b24SeedTypes(me, 3)
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Descend upon the Sinful", "Sorcery", s29DescendOracle, nil)
	passPriorityAroundTable(t, g)
	if _, ok := s29Token(g, "Angel"); ok {
		t.Error("an Angel token without delirium")
	}
}

// --- Convert to Slime ----------------------------------------------------------------

func s29Slime(t *testing.T, seed int) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	b24SeedTypes(me, seed)
	artifact := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Old Relic",
		TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	creature := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Old Giant",
		TypeLine: "Creature — Giant", ManaCost: "{3}{R}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Convert to Slime", "Sorcery", s29SlimeOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: artifact, Slot: 0},
		{Kind: game.TargetCard, ID: creature, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	return g, artifact, creature
}

func TestCmdr2S29ConvertToSlimeDestroysAndMakesAnOozeOfTheTotalManaValue(t *testing.T) {
	g, artifact, creature := s29Slime(t, 3)
	if s29OnBattlefield(g, artifact) || s29OnBattlefield(g, creature) {
		t.Fatal("a target survived")
	}
	// Three seeded types plus the destroyed artifact: four card types, so
	// delirium counts the destroyed cards themselves ("Then").
	ooze, ok := s29Token(g, "Ooze")
	if !ok {
		t.Fatal("no Ooze token")
	}
	if p, tou := s29Size(t, g, ooze.InstanceID); p != 6 || tou != 6 {
		t.Errorf("Ooze %d/%d, want 6/6 (2 + 4)", p, tou)
	}
}

func TestCmdr2S29ConvertToSlimeMakesNoOozeWithoutDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	artifact := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Old Relic",
		TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Convert to Slime", "Sorcery", s29SlimeOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: artifact, Slot: 0}})
	passPriorityAroundTable(t, g)
	if s29OnBattlefield(g, artifact) {
		t.Error("the artifact survived")
	}
	if _, ok := s29Token(g, "Ooze"); ok {
		t.Error("an Ooze with only creature and artifact in the graveyard")
	}
}

func TestCmdr2S29ConvertToSlimeRefusesAPermanentNoClauseTakes(t *testing.T) {
	g := newCatalogGame(t)
	land := b12Permanent(g, g.Seats[1].ID, "Old Forest", "Basic Land — Forest")
	err := castCatalogSpellErr(t, g, "Convert to Slime", "Sorcery", s29SlimeOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: land, Slot: 0}})
	if err == nil {
		t.Error("a land was accepted as a target")
	}
}

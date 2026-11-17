package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_dmu_b12_test.go — Dominaria United batch 12 (S82, #5487).

const (
	s82DmuKingDarienOracle  = "e923d956-b036-4789-9311-4763a71fcbf8"
	s82DmuQueenOracle       = "5f04d4b0-2c75-4f2d-97d5-96220b624b17"
	s82DmuSqueeOracle       = "124a5dd7-c4f8-479c-9eb0-669aca01d045"
	s82DmuHordemasterOracle = "1e080c7d-102c-4d28-bcb9-62b36d400d13"
	s82DmuPriestOracle      = "cf42a7e3-71ee-40b8-a6ca-0368c746e128"
	s82DmuToriOracle        = "6d51ebd1-ceb9-42c0-8c1e-34dbbbc7bf8a"
	s82DmuHellkiteOracle    = "0166996d-941f-49fe-bf4f-64b6fd4b69e9"
	s82DmuChickOracle       = "40bb7b6b-a75d-4b59-ba45-a5801dcd59a1"
	s82DmuRavenManOracle    = "19918740-f6e3-40cd-9848-3c7a852a6c2a"
	s82DmuNemataOracle      = "f72558f5-ac5c-4efa-b01b-439bc0bbf18d"
	s82DmuHumanSoldierType  = "Legendary Creature — Human Soldier"
	s82DmuGoblinWarriorType = "Creature — Goblin Warrior"
	s82DmuHumanClericType   = "Creature — Human Cleric"
	s82DmuNemataType        = "Legendary Creature — Treefolk"
	s82DmuRavenManType      = "Legendary Creature — Human Wizard"
	s82DmuQueenType         = "Legendary Creature — Elf Noble"
	s82DmuHellkiteType      = "Creature — Dragon"
	s82DmuToriType          = "Legendary Creature — Human Knight"
	s82DmuSqueeType         = "Legendary Creature — Goblin Noble"
	s82DmuChickType         = "Creature — Phoenix"
)

func TestS82DmuB12EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82DmuKingDarienOracle:  "King Darien XLVIII",
		s82DmuQueenOracle:       "Queen Allenal of Ruadach",
		s82DmuSqueeOracle:       "Squee, Dubious Monarch",
		s82DmuHordemasterOracle: "Rundvelt Hordemaster",
		s82DmuPriestOracle:      "Shadow-Rite Priest",
		s82DmuToriOracle:        "Tori D'Avenant, Fury Rider",
		s82DmuHellkiteOracle:    "Ragefire Hellkite",
		s82DmuChickOracle:       "Phoenix Chick",
		s82DmuRavenManOracle:    "The Raven Man",
		s82DmuNemataOracle:      "Nemata, Primeval Warden",
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

// s82DmuPush puts a coloured catalog creature on seat `owner`'s side.
func s82DmuPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Colors: colors, Owner: owner, Controller: owner,
	})
}

func s82DmuActivate(t *testing.T, g *game.Game, who, card uuid.UUID, index int, sacrifice ...uuid.UUID) error {
	t.Helper()
	return g.ActivateCatalogAbility(who, card, index, game.ActivateAbilityParams{Strict: true, SacrificeIDs: sacrifice})
}

// --- King Darien XLVIII -------------------------------------------------

func TestS82DmuB12KingDarienPumpsOnlyTheOthersYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	king := s82DmuPush(g, me.ID, "King Darien XLVIII", s82DmuHumanSoldierType, s82DmuKingDarienOracle, 2, 3, "G", "W")
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if p := effectivePower(t, g, king); p != 2 {
		t.Errorf("King Darien is %d power, want 2: the anthem says other creatures", p)
	}
	if p := effectivePower(t, g, mine); p != 3 {
		t.Errorf("my creature is %d power, want 3", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("their creature is %d power, want 2", p)
	}
}

func TestS82DmuB12KingDarienGrowsAndMakesASoldier(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	king := s82DmuPush(g, me.ID, "King Darien XLVIII", s82DmuHumanSoldierType, s82DmuKingDarienOracle, 2, 3, "G", "W")
	if err := s82DmuActivate(t, g, me.ID, king, 0); err == nil {
		t.Fatal("activated with no mana to pay {3}{G}{W}")
	}
	floatMana(t, g, me, "{C}{C}{C}{G}{W}")
	if err := s82DmuActivate(t, g, me.ID, king, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Soldier", me.ID); got != 1 {
		t.Fatalf("%d Soldier tokens, want 1", got)
	}
	// 2/3 base + one +1/+1 counter.
	if p := effectivePowerWithCounters(t, g, king); p != 3 {
		t.Errorf("King Darien is %d power, want 3 after the counter", p)
	}
	soldier := findBattlefieldByName(g, "Soldier")
	if p := effectivePower(t, g, soldier); p != 2 {
		t.Errorf("the Soldier is %d power, want 2 (1/1 plus the anthem)", p)
	}
}

func TestS82DmuB12KingDarienSacrificeProtectsTokensOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	king := s82DmuPush(g, me.ID, "King Darien XLVIII", s82DmuHumanSoldierType, s82DmuKingDarienOracle, 2, 3, "G", "W")
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	if err := s82DmuActivate(t, g, me.ID, king, 1); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(king) {
		t.Fatal("King Darien was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"hexproof", "indestructible"} {
		if !hasEffectiveKeyword(t, g, token, kw) {
			t.Errorf("the token lacks %s", kw)
		}
		if hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the nontoken creature gained %s", kw)
		}
	}
}

// --- Queen Allenal of Ruadach ------------------------------------------

func TestS82DmuB12QueenAllenalSizeIsYourCreatureCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	queen := s82DmuPush(g, me.ID, "Queen Allenal of Ruadach", s82DmuQueenType, s82DmuQueenOracle, 0, 0, "G", "W")
	pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if p := effectivePower(t, g, queen); p != 2 {
		t.Errorf("the Queen is %d power, want 2 (herself and one bear)", p)
	}
}

func TestS82DmuB12QueenAllenalAddsASoldierToCreatureTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s82DmuPush(g, me.ID, "Queen Allenal of Ruadach", s82DmuQueenType, s82DmuQueenOracle, 0, 0, "G", "W")
	g.WithWriteLock(func() {
		if err := (CreateToken{Controller: me.ID, Template: TokenCard("1/1 red Goblin"), N: 3}).Apply(
			&Context{Game: g, Item: &game.StackItem{Controller: me.ID}}); err != nil {
			t.Fatalf("create: %v", err)
		}
	})
	if got := tokensNamedControlledBy(g, "Goblin", me.ID); got != 3 {
		t.Errorf("%d Goblins, want the 3 that were made", got)
	}
	if got := tokensNamedControlledBy(g, "Soldier", me.ID); got != 1 {
		t.Errorf("%d Soldiers, want exactly 1 however many tokens", got)
	}
	// An opponent's tokens are not under the Queen's controller.
	g.WithWriteLock(func() {
		_ = (CreateToken{Controller: opp.ID, Template: TokenCard("1/1 red Goblin"), N: 1}).Apply(
			&Context{Game: g, Item: &game.StackItem{Controller: opp.ID}})
	})
	if got := tokensNamedControlledBy(g, "Soldier", opp.ID); got != 0 {
		t.Errorf("an opponent's token made %d Soldiers", got)
	}
}

func TestS82DmuB12QueenAllenalIgnoresANoncreatureToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	s82DmuPush(g, me.ID, "Queen Allenal of Ruadach", s82DmuQueenType, s82DmuQueenOracle, 0, 0, "G", "W")
	g.WithWriteLock(func() {
		_ = (CreateToken{Controller: me.ID, Template: TreasureToken(), N: 1}).Apply(
			&Context{Game: g, Item: &game.StackItem{Controller: me.ID}})
	})
	if got := tokensNamedControlledBy(g, "Soldier", me.ID); got != 0 {
		t.Errorf("a lone Treasure made %d Soldiers", got)
	}
}

package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_sz_cards_test.go — Foundations, S–Z (#4465): the cards from that
// stretch of the alphabet that needed no new seam. Each test plays the
// real catalog entry through the engine and asserts what its text says.

const (
	fdnSZSanguineSyphonerOracle     = "b81a748b-f7a6-4ec4-8b18-40a1819745f6"
	fdnSZSavannahLionsOracle        = "60ba93eb-39e6-4af2-9c66-cd38f72daff2"
	fdnSZSeismicRuptureOracle       = "3b1f1a57-d83a-443f-b190-03df63d2bfe6"
	fdnSZShivanDragonOracle         = "711eea87-0fa3-46e0-a42b-fa5a86455f04"
	fdnSZSkyrakerGiantOracle        = "d481630c-83c4-47d2-a770-db331154d2b9"
	fdnSZStabOracle                 = "b4ee6843-bb79-4665-9763-83f656863dc5"
	fdnSZSwabGoblinOracle           = "343ecf2f-0bbc-4840-8fb3-f5f61b26e76f"
	fdnSZSwiftbladeVindicatorOracle = "44240b1e-c104-4eb4-8ab9-0d56607ba241"
	fdnSZTajuruPathwardenOracle     = "d8c7cd66-d8bf-4a99-add5-1d5651d4e025"
	fdnSZThornwealdArcherOracle     = "b5ba87f5-3589-4a8b-ad65-2ae2523e98e8"
	fdnSZVampireSpawnOracle         = "33cee4b4-3d80-46f1-9223-49ed785891a1"
	fdnSZZetalpaOracle              = "7da0e5de-3e4c-420a-8685-991206100b9d"
)

func TestFdnSZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnSZSanguineSyphonerOracle:     "Sanguine Syphoner",
		fdnSZSavannahLionsOracle:        "Savannah Lions",
		fdnSZSeismicRuptureOracle:       "Seismic Rupture",
		fdnSZShivanDragonOracle:         "Shivan Dragon",
		fdnSZSkyrakerGiantOracle:        "Skyraker Giant",
		fdnSZStabOracle:                 "Stab",
		fdnSZSwabGoblinOracle:           "Swab Goblin",
		fdnSZSwiftbladeVindicatorOracle: "Swiftblade Vindicator",
		fdnSZTajuruPathwardenOracle:     "Tajuru Pathwarden",
		fdnSZThornwealdArcherOracle:     "Thornweald Archer",
		fdnSZVampireSpawnOracle:         "Vampire Spawn",
		fdnSZZetalpaOracle:              "Zetalpa, Primal Dawn",
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

// --- Keyword-only and vanilla bodies -----------------------------------------

func TestFdnSZKeywordBodiesCarryTheirPrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Skyraker Giant", "Creature — Giant", fdnSZSkyrakerGiantOracle, 4, 3, []string{"reach"}},
		{"Swiftblade Vindicator", "Creature — Human Soldier", fdnSZSwiftbladeVindicatorOracle, 1, 1, []string{"double strike", "vigilance", "trample"}},
		{"Tajuru Pathwarden", "Creature — Elf Warrior Ally", fdnSZTajuruPathwardenOracle, 5, 4, []string{"vigilance", "trample"}},
		{"Thornweald Archer", "Creature — Elf Archer", fdnSZThornwealdArcherOracle, 2, 1, []string{"reach", "deathtouch"}},
		{"Zetalpa, Primal Dawn", "Legendary Creature — Elder Dinosaur", fdnSZZetalpaOracle, 4, 8, []string{"flying", "double strike", "vigilance", "trample", "indestructible"}},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		assertKeywords(t, g, id, c.keywords...)
	}
}

func TestFdnSZVanillaBodiesHaveNoAbilities(t *testing.T) {
	for _, c := range []struct{ name, oracle string }{
		{"Savannah Lions", fdnSZSavannahLionsOracle},
		{"Swab Goblin", fdnSZSwabGoblinOracle},
	} {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Fatalf("%s is not registered", c.name)
		}
		if len(spec.PrintedKeywords)+len(spec.Triggered)+len(spec.Activated)+len(spec.Static)+len(spec.Replacements) != 0 || spec.OnResolve != nil {
			t.Errorf("%s has abilities; it is a vanilla creature", c.name)
		}
	}
}

func TestFdnSZThornwealdArcherKillsWhatItBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushTribalCreature(g, me.ID, "Big Bird", "Creature — Bird", 5, 5, "flying")
	archer := b12Push(g, opp.ID, "Thornweald Archer", "Creature — Elf Archer", fdnSZThornwealdArcherOracle, 2, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(archer, attacker); err != nil {
		t.Fatalf("a reach creature could not block a flier: %v", err)
	}
	advanceTo(t, g, game.StepEndCombat)
	if g.Battlefield.Contains(attacker) {
		t.Error("a 5/5 blocked by a deathtouch creature survived its damage")
	}
}

// --- Sanguine Syphoner --------------------------------------------------------

func TestFdnSZSanguineSyphonerDrainsEachOpponentWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	syphoner := b12Push(g, me.ID, "Sanguine Syphoner", "Creature — Vampire Warlock", fdnSZSanguineSyphonerOracle, 1, 3)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(syphoner, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		switch s.ID {
		case me.ID:
			if s.Life != lives[s.ID]+1 {
				t.Errorf("my life %d, want %d", s.Life, lives[s.ID]+1)
			}
		case opp.ID:
			// The attacked player may also have taken the combat damage by
			// now; never less than the drain.
			if s.Life > lives[s.ID]-1 {
				t.Errorf("the attacked opponent is at %d, want at most %d", s.Life, lives[s.ID]-1)
			}
		default:
			if s.Life != lives[s.ID]-1 {
				t.Errorf("seat %s life %d, want %d", s.ID, s.Life, lives[s.ID]-1)
			}
		}
	}
}

// --- Seismic Rupture ----------------------------------------------------------

func TestFdnSZSeismicRuptureHitsEachCreatureWithoutFlyingOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	bird := pushTribalCreature(g, opp.ID, "Bird", "Creature — Bird", 2, 2, "flying")
	mine := pushTribalCreature(g, me.ID, "My Ox", "Creature — Ox", 3, 3)
	before := b27Lives(g)
	castCatalogSpell(t, g, "Seismic Rupture", "Sorcery", fdnSZSeismicRuptureOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a 2/2 without flying survived 2 damage")
	}
	if !g.Battlefield.Contains(bird) || damageMarkedOn(g, bird) != 0 {
		t.Error("a flier was dealt damage")
	}
	if !g.Battlefield.Contains(mine) || damageMarkedOn(g, mine) != 2 {
		t.Error("your own creature without flying was not dealt 2 damage")
	}
	for i, l := range b27Lives(g) {
		if l != before[i] {
			t.Errorf("seat %d life %d, want an unchanged %d", i, l, before[i])
		}
	}
}

// --- Shivan Dragon ------------------------------------------------------------

func TestFdnSZShivanDragonFliesAndFirebreathes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dragon := b12Push(g, me.ID, "Shivan Dragon", "Creature — Dragon", fdnSZShivanDragonOracle, 5, 5)
	assertKeywords(t, g, dragon, "flying")
	advanceToMain(t, g)

	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without red mana")
	}
	for range 2 {
		b06AddMana(me, "R")
		b16Activate(t, g, me.ID, dragon, 0, game.ActivateAbilityParams{})
		passPriorityAroundTable(t, g)
	}
	if p, tough := effectivePower(t, g, dragon), effectiveToughness(t, g, dragon); p != 7 || tough != 5 {
		t.Errorf("the Dragon is %d/%d after two activations, want 7/5", p, tough)
	}
}

// --- Stab ---------------------------------------------------------------------

func TestFdnSZStabKillsATwoToughnessCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Stab", "Instant", fdnSZStabOracle, cardRefs(small))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a 2/2 survived -2/-2")
	}
}

func TestFdnSZStabShrinksALargerCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 4, 4)
	castCatalogSpell(t, g, "Stab", "Instant", fdnSZStabOracle, cardRefs(big))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, big), effectiveToughness(t, g, big); p != 2 || tough != 2 {
		t.Errorf("the target is %d/%d, want 2/2", p, tough)
	}
}

// --- Vampire Spawn ------------------------------------------------------------

func TestFdnSZVampireSpawnDrainsEachOpponentForTwo(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	castHoldCreature(t, g, "Vampire Spawn", "Creature — Vampire", fdnSZVampireSpawnOracle, 2, 3)
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := lives[s.ID] - 2
		if s.ID == active.ID {
			want = lives[s.ID] + 2
		}
		if s.Life != want {
			t.Errorf("seat %s life %d, want %d", s.ID, s.Life, want)
		}
	}
}

package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b633_cards_test.go — Apocalypse, batch 633 (#5614). Each test plays
// the real catalog entry through the engine and asserts what its text says.

const (
	s82OldB633RakaSanctuaryOracle     = "0b085c90-95df-4823-ba67-9eb398b1ec94"
	s82OldB633MaskOfIntoleranceOracle = "990d4798-3f59-462d-952c-777da5af1c41"
	s82OldB633BloodfireKavuOracle     = "a12880c4-1987-42af-8595-0b88981e1091"
	s82OldB633BloodfireColossusOracle = "cc59e28a-eda3-468c-a31b-73e4b615f953"
	s82OldB633MinotaurOracle          = "ee5fc6e9-0c10-4015-8853-dd1f33dcc27e"
	s82OldB633BloodfireDwarfOracle    = "ff42551a-a08e-4d0c-a5b1-d1c1bbff4915"
)

func TestS82OldB633EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82OldB633RakaSanctuaryOracle:     "Raka Sanctuary",
		s82OldB633MaskOfIntoleranceOracle: "Mask of Intolerance",
		s82OldB633BloodfireKavuOracle:     "Bloodfire Kavu",
		s82OldB633BloodfireColossusOracle: "Bloodfire Colossus",
		s82OldB633MinotaurOracle:          "Minotaur Illusionist",
		s82OldB633BloodfireDwarfOracle:    "Bloodfire Dwarf",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// s82OldB633Settle answers a target pick with `ref` and passes priority until
// the stack is empty.
func s82OldB633Settle(t *testing.T, g *game.Game, chooser uuid.UUID, ref game.TargetRef) {
	t.Helper()
	for i := 0; i < 48; i++ {
		if p := effectstest.LatestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser, ref); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if len(g.PendingTriggers) > 0 {
			g.RunStateChecksForTest()
			if effectstest.LatestPickTarget(g, chooser) != nil {
				continue
			}
		}
		if len(g.PendingChoices) > 0 || effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				continue
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack did not settle")
}

func s82OldB633Raka(t *testing.T, mine ...string) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	raka := effectstest.B12Push(g, me.ID, "Raka Sanctuary", "Enchantment", s82OldB633RakaSanctuaryOracle, 0, 0)
	for _, c := range mine {
		effectstest.TmpB1Push(g, me.ID, "Rock "+c, "Artifact", "", 0, 0, c)
	}
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	return g, me, raka, bear
}

func TestS82OldB633RakaSanctuaryOneDamageWithWhiteOrBlue(t *testing.T) {
	for _, colour := range []string{"W", "U"} {
		g, me, _, bear := s82OldB633Raka(t, colour)
		s82OldB633Settle(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear})
		c, ok := effectstest.BattlefieldCard(g, bear)
		if !ok || c.DamageMarked != 1 {
			t.Errorf("with a %s permanent the 3/3 has %d damage, want exactly 1", colour, c.DamageMarked)
		}
	}
}

func TestS82OldB633RakaSanctuaryThreeDamageWithBoth(t *testing.T) {
	g, me, _, bear := s82OldB633Raka(t, "W", "U")
	s82OldB633Settle(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if g.Battlefield.Contains(bear) {
		t.Error("the 3/3 survived 3 damage from a white and a blue permanent")
	}
}

func TestS82OldB633RakaSanctuaryDoesNothingWithoutWhiteOrBlue(t *testing.T) {
	g, me, _, bear := s82OldB633Raka(t, "R", "G")
	s82OldB633Settle(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if c, ok := effectstest.BattlefieldCard(g, bear); !ok || c.DamageMarked != 0 {
		t.Errorf("the 3/3 was damaged without a white or blue permanent")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the trigger asked for a target with the intervening if false")
	}
}

func s82OldB633Lands(g *game.Game, owner uuid.UUID, types ...string) {
	for _, ty := range types {
		effectstest.B12Permanent(g, owner, ty, "Basic Land — "+ty)
	}
}

func TestS82OldB633MaskOfIntoleranceHitsAPlayerWithFourBasicTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Mask of Intolerance", "Artifact", s82OldB633MaskOfIntoleranceOracle, 0, 0)
	s82OldB633Lands(g, opp.ID, "Plains", "Island", "Swamp", "Mountain")
	s82OldB633Lands(g, me.ID, "Plains", "Island", "Swamp")
	life, mine := opp.Life, me.Life
	effectstest.BeginUpkeep(t, g, opp.ID)
	if opp.Life != life-3 {
		t.Errorf("the opponent's life %d -> %d, want -3 in their own upkeep", life, opp.Life)
	}
	effectstest.BeginUpkeep(t, g, me.ID)
	if me.Life != mine {
		t.Errorf("the controller took damage with only three basic land types (%d -> %d)", mine, me.Life)
	}
}

func TestS82OldB633MaskOfIntoleranceHitsItsControllerToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Mask of Intolerance", "Artifact", s82OldB633MaskOfIntoleranceOracle, 0, 0)
	s82OldB633Lands(g, me.ID, "Plains", "Island", "Swamp", "Mountain", "Forest")
	life := me.Life
	effectstest.BeginUpkeep(t, g, me.ID)
	if me.Life != life-3 {
		t.Errorf("life %d -> %d, want -3", life, me.Life)
	}
}

func s82OldB633Activate(t *testing.T, name, typeLine, oracle string, p, tough int, mana string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	src := effectstest.B12Push(g, me.ID, name, typeLine, oracle, p, tough)
	effectstest.B06AddMana(me, mana)
	return g, me, src
}

func TestS82OldB633BloodfireKavuDamagesEveryCreature(t *testing.T) {
	g, me, kavu := s82OldB633Activate(t, "Bloodfire Kavu", "Creature — Kavu", s82OldB633BloodfireKavuOracle, 2, 2, "R")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 3, 3)
	effectstest.B16Activate(t, g, me.ID, kavu, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(kavu) {
		t.Error("the Kavu was not sacrificed")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the controller's own 2/2 survived 2 damage")
	}
	if c, ok := effectstest.BattlefieldCard(g, big); !ok || c.DamageMarked != 2 {
		t.Errorf("the 3/3 should carry exactly 2 damage")
	}
}

func TestS82OldB633BloodfireDwarfSparesFliers(t *testing.T) {
	g, me, dwarf := s82OldB633Activate(t, "Bloodfire Dwarf", "Creature — Dwarf", s82OldB633BloodfireDwarfOracle, 1, 1, "R")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	walker := effectstest.B12Creature(g, opp.ID, "Walker", "Creature — Elf", 1, 1)
	effectstest.B16Activate(t, g, me.ID, dwarf, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(dwarf) {
		t.Error("the Dwarf was not sacrificed")
	}
	if g.Battlefield.Contains(walker) {
		t.Error("the 1/1 without flying survived")
	}
	if !g.Battlefield.Contains(flier) {
		t.Error("the flier was dealt damage")
	}
}

func TestS82OldB633BloodfireColossusHitsEveryCreatureAndPlayer(t *testing.T) {
	g, me, colossus := s82OldB633Activate(t, "Bloodfire Colossus", "Creature — Giant", s82OldB633BloodfireColossusOracle, 6, 6, "R")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 6, 6)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.B16Activate(t, g, me.ID, colossus, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(colossus) {
		t.Error("the Colossus was not sacrificed")
	}
	if g.Battlefield.Contains(big) {
		t.Error("the 6/6 survived 6 damage")
	}
	for _, p := range g.Seats {
		if p.Life != lives[p.ID]-6 {
			t.Errorf("seat %v life %d -> %d, want -6", p.ID, lives[p.ID], p.Life)
		}
	}
}

func TestS82OldB633MinotaurIllusionistGainsShroudAndSacrificesForItsPower(t *testing.T) {
	g, me, mino := s82OldB633Activate(t, "Minotaur Illusionist", "Creature — Minotaur Wizard", s82OldB633MinotaurOracle, 3, 4, "U")
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, mino, 0, game.ActivateAbilityParams{})
	if !effectstest.HasKeywordOnBattlefield(t, g, mino, "shroud") {
		t.Error("the Illusionist has no shroud after {1}{U}")
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
	effectstest.B06AddMana(me, "R")
	effectstest.B16Activate(t, g, me.ID, mino, 1, game.ActivateAbilityParams{Targets: effectstest.CardTarget(small)})
	if g.Battlefield.Contains(mino) {
		t.Error("the Illusionist was not sacrificed")
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 3/3 survived damage equal to the Illusionist's power 3")
	}
	if c, ok := effectstest.BattlefieldCard(g, big); !ok || c.DamageMarked != 0 {
		t.Error("the untargeted creature was damaged")
	}
}

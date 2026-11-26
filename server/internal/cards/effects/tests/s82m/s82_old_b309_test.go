package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b309_test.go — Battle for Zendikar landfall cards (#5547). Every
// test plays the real catalog entry through the engine.

const (
	oldB309Guardian   = "06432957-decc-4ba2-95b1-cc9536a576d4"
	oldB309Firebird   = "10ab3102-43a3-4c60-8641-2c5a7c1e29c2"
	oldB309Hydra      = "1b065a17-14a4-433f-a6c8-2c8212f505c7"
	oldB309Stonewaker = "36d1a46e-44ac-4a65-a93c-8d6ab5c31cbf"
	oldB309Retreat    = "91108ed7-5b24-4b51-b01d-85005998c69d"
	oldB309Overseer   = "9eff1ce5-41e4-4ae6-9480-2b549c8546bd"
	oldB309Whiptail   = "cf0a15f5-5f41-4eea-81f7-02059b9586a2"
)

func oldB309PlayLand(t *testing.T, g *game.Game, name, typeLine string) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: active.ID, Controller: active.ID})
	effectstest.AdvanceToMain(t, g)
	active.LandDropsPerTurn++
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
}

func TestOldB309EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{oldB309Guardian, oldB309Firebird, oldB309Hydra, oldB309Stonewaker, oldB309Retreat, oldB309Overseer, oldB309Whiptail} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestOldB309GuardianOfTazeemTapsAndOnlyAnIslandFreezesIt(t *testing.T) {
	for _, tc := range []struct {
		land, typeLine string
		frozen         bool
	}{
		{"Forest", "Basic Land — Forest", false},
		{"Island", "Basic Land — Island", true},
	} {
		t.Run(tc.land, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.B12Push(g, me.ID, "Guardian of Tazeem", "Creature — Sphinx", oldB309Guardian, 4, 5)
			victim := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
			oldB309PlayLand(t, g, tc.land, tc.typeLine)
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.PickCard(t, g, me.ID, victim)
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.B16Tapped(t, g, victim) {
				t.Fatal("the target was not tapped")
			}
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			if got := effectstest.B16Tapped(t, g, victim); got != tc.frozen {
				t.Errorf("tapped in its controller's next upkeep = %t, want %t", got, tc.frozen)
			}
		})
	}
}

func TestOldB309GuardianOfTazeemIgnoresAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guardian := effectstest.B12Push(g, me.ID, "Guardian of Tazeem", "Creature — Sphinx", oldB309Guardian, 4, 5)
	effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	if n := effectstest.TriggersOnStackFrom(g, guardian); n != 0 {
		t.Errorf("%d triggers for a land I don't control", n)
	}
}

func TestOldB309OranRiefHydraGrowsByOneOrTwoForAForest(t *testing.T) {
	for _, tc := range []struct {
		land, typeLine string
		want           int
	}{
		{"Island", "Basic Land — Island", 1},
		{"Forest", "Basic Land — Forest", 2},
	} {
		t.Run(tc.land, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			hydra := effectstest.B12Push(g, me.ID, "Oran-Rief Hydra", "Creature — Hydra", oldB309Hydra, 5, 5)
			oldB309PlayLand(t, g, tc.land, tc.typeLine)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.Pr8Counters(g, hydra, game.CounterPlusOne); got != tc.want {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.want)
			}
		})
	}
}

func TestOldB309BelligerentWhiptailGainsFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	whip := effectstest.B12Push(g, me.ID, "Belligerent Whiptail", "Creature — Wurm", oldB309Whiptail, 4, 2)
	oldB309PlayLand(t, g, "Forest", "Basic Land — Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, whip, "first strike") {
		t.Error("the Whiptail did not gain first strike")
	}
}

func TestOldB309GuulDrazOverseerPumpsOthersByOneOrTwoForASwamp(t *testing.T) {
	for _, tc := range []struct {
		land, typeLine string
		want           int
	}{
		{"Forest", "Basic Land — Forest", 1},
		{"Swamp", "Basic Land — Swamp", 2},
	} {
		t.Run(tc.land, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			overseer := effectstest.B12Push(g, me.ID, "Guul Draz Overseer", "Creature — Vampire", oldB309Overseer, 3, 4)
			ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
			foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
			oldB309PlayLand(t, g, tc.land, tc.typeLine)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.EffectivePower(t, g, ally); got != 2+tc.want {
				t.Errorf("ally power %d, want %d", got, 2+tc.want)
			}
			if got := effectstest.EffectivePower(t, g, overseer); got != 3 {
				t.Errorf("the Overseer itself got pumped: power %d", got)
			}
			if got := effectstest.EffectivePower(t, g, foe); got != 2 {
				t.Errorf("an opponent's creature got pumped: power %d", got)
			}
		})
	}
}

func TestOldB309AkoumStonewakerMayPayForAHastyElementalThatIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Akoum Stonewaker", "Creature — Human Shaman", oldB309Stonewaker, 2, 1)
	oldB309PlayLand(t, g, "Mountain", "Basic Land — Mountain")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "R", "R", "R")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Elemental"); got != 1 {
		t.Fatalf("%d Elemental tokens, want 1", got)
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Elemental"); got != 0 {
		t.Errorf("%d Elemental tokens after the end step, want 0 (exiled)", got)
	}
}

func TestOldB309AkoumStonewakerDeclinedMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Akoum Stonewaker", "Creature — Human Shaman", oldB309Stonewaker, 2, 1)
	oldB309PlayLand(t, g, "Mountain", "Basic Land — Mountain")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Elemental"); got != 0 {
		t.Errorf("%d Elemental tokens after declining, want 0", got)
	}
}

func TestOldB309AkoumFirebirdReturnsFromTheGraveyardForSixMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bird := effectstest.SeedGraveyardCard(t, g, "Akoum Firebird", "Creature — Phoenix", oldB309Firebird)
	oldB309PlayLand(t, g, "Mountain", "Basic Land — Mountain")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "R", "R", "R", "R", "R", "R")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bird); !ok || me.Graveyard.Contains(bird) {
		t.Error("paying did not return the Firebird to the battlefield")
	}
}

func TestOldB309AkoumFirebirdStaysWhenDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bird := effectstest.SeedGraveyardCard(t, g, "Akoum Firebird", "Creature — Phoenix", oldB309Firebird)
	oldB309PlayLand(t, g, "Mountain", "Basic Land — Mountain")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bird) {
		t.Error("declining moved the Firebird out of the graveyard")
	}
}

func TestOldB309AkoumFirebirdDoesNothingOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Akoum Firebird", "Creature — Phoenix", oldB309Firebird, 3, 3)
	oldB309PlayLand(t, g, "Mountain", "Basic Land — Mountain")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a battlefield Firebird offered a payment")
	}
}

func TestOldB309RetreatToValakutPumpsOrStopsABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Retreat to Valakut", "Enchantment", oldB309Retreat, 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	oldB309PlayLand(t, g, "Forest", "Basic Land — Forest")
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1}, 0)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d after +2/+0, want 4", got)
	}
}

func TestOldB309RetreatToValakutCantBlockMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Retreat to Valakut", "Enchantment", oldB309Retreat, 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	oldB309PlayLand(t, g, "Forest", "Basic Land — Forest")
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1}, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("power %d, the can't-block mode must not pump", got)
	}
	if !effectstest.SbCantBlock(t, g, bear) {
		t.Error("the target can still block")
	}
}

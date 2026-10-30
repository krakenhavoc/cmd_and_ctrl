package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_winter_typhoid_test.go — Winter Soldier, Reborn Avenger and Typhoid
// Mary, Fractured (Marvel Super Heroes Commander, #4455). Each test plays
// the real catalog entry through the engine and asserts what its text says.

const (
	mscWinterSoldierOracle = "df81ecf2-b49e-4be2-8757-ee2050784c9e"
	mscTyphoidMaryOracle   = "867b3923-692c-4444-a071-518a290e43d8"
)

func TestMscWinterTyphoidEveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mscWinterSoldierOracle: "Winter Soldier, Reborn Avenger",
		mscTyphoidMaryOracle:   "Typhoid Mary, Fractured",
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

// --- Winter Soldier, Reborn Avenger -----------------------------------------------------

const mscWinterTypeLine = "Legendary Creature — Human Soldier Hero"

// mscWinterTable puts a 3/5 Winter Soldier on seat 0's battlefield.
func mscWinterTable(t *testing.T) (g *game.Game, me, opp *game.Player, ws uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	ws = b12Push(g, me.ID, "Winter Soldier, Reborn Avenger", mscWinterTypeLine, mscWinterSoldierOracle, 3, 5)
	return g, me, opp, ws
}

func TestMscWinterSoldierOffersOnlyYourCreatureCardsWithinHisPower(t *testing.T) {
	g, me, opp, ws := mscWinterTable(t)
	small := gyCard(me, "Small Bear", "Creature — Bear", "{2}", "")
	equal := gyCard(me, "Equal Bear", "Creature — Bear", "{1}{G}{G}", "")
	big := gyCard(me, "Big Bear", "Creature — Bear", "{3}{G}", "")
	notCreature := gyCard(me, "Rock", "Artifact", "{1}", "")
	theirs := gyCard(opp, "Their Bear", "Creature — Bear", "{1}", "")

	declareAttack(t, g, opp.ID, ws)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("Winter Soldier asked for no target")
	}
	for _, id := range []uuid.UUID{small, equal} {
		if !slices.Contains(p.PickTargetCards, id) {
			t.Errorf("offered %v, want it to include the card %s", p.PickTargetCards, id)
		}
	}
	for name, id := range map[string]uuid.UUID{"mana value 4": big, "noncreature": notCreature, "opponent's": theirs} {
		if slices.Contains(p.PickTargetCards, id) {
			t.Errorf("offered the %s card", name)
		}
		err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: id})
		if !errors.Is(err, game.ErrIllegalTarget) {
			t.Errorf("picking the %s card = %v, want ErrIllegalTarget (CR 601.2c)", name, err)
		}
	}
}

func TestMscWinterSoldierReturnsAHeroWithAnExtraCounterAndOthersWithNone(t *testing.T) {
	g, me, opp, ws := mscWinterTable(t)
	hero := gyCard(me, "Fallen Hero", "Creature — Human Hero", "{3}", "")
	declareAttack(t, g, opp.ID, ws)
	pickCard(t, g, me.ID, hero)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hero) {
		t.Fatal("the Hero did not return")
	}
	if got := plusOnes(g, hero); got != 1 {
		t.Errorf("the Hero entered with %d +1/+1 counters, want 1", got)
	}

	g, me, opp, ws = mscWinterTable(t)
	bear := gyCard(me, "Fallen Bear", "Creature — Bear", "{3}", "")
	declareAttack(t, g, opp.ID, ws)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the Bear did not return")
	}
	if got := plusOnes(g, bear); got != 0 {
		t.Errorf("a non-Hero entered with %d +1/+1 counters, want 0", got)
	}
}

func TestMscWinterSoldierPumpedBeforeTheAttackReachesBiggerCards(t *testing.T) {
	g, me, opp, ws := mscWinterTable(t)
	big := gyCard(me, "Big Bear", "Creature — Bear", "{3}{G}", "")
	findBattlefieldCardByID(g, ws).Counters = map[string]int{game.CounterPlusOne: 1}
	declareAttack(t, g, opp.ID, ws)
	pickCard(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) {
		t.Error("a mana value 4 creature did not return for a 4 power Winter Soldier")
	}
}

// Shrunk below the target's mana value in response, the target is illegal as
// the ability resolves (CR 608.2b) and nothing returns.
func TestMscWinterSoldierShrunkInResponseReturnsNothing(t *testing.T) {
	g, me, opp, ws := mscWinterTable(t)
	bear := gyCard(me, "Fallen Bear", "Creature — Bear", "{3}", "")
	declareAttack(t, g, opp.ID, ws)
	pickCard(t, g, me.ID, bear)
	findBattlefieldCardByID(g, ws).Counters = map[string]int{game.CounterMinusOne: 1}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Error("the card returned although Winter Soldier's power fell below its mana value")
	}
}

// Winter Soldier, with two +1/+1 counters (power 5), dies in response. His
// graveyard card prints power 3, so a mana value 5 target comes back only if
// the check reads his last-known power (CR 608.2h).
func TestMscWinterSoldierGoneInResponseUsesLastKnownPower(t *testing.T) {
	g, me, opp, ws := mscWinterTable(t)
	findBattlefieldCardByID(g, ws).Counters = map[string]int{game.CounterPlusOne: 2}
	five := gyCard(me, "Five Drop", "Creature — Bear", "{4}{G}", "")
	declareAttack(t, g, opp.ID, ws)
	pickCard(t, g, me.ID, five)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ws, game.DestroyOptions{}) })
	if findBattlefieldCardByID(g, ws) != nil {
		t.Fatal("Winter Soldier did not die")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(five) {
		t.Error("the mana value 5 creature did not return, although Winter Soldier's last-known power was 5 (CR 608.2h)")
	}
}

func TestMscWinterSoldierGoneInResponseStillRefusesAnOverlargeCard(t *testing.T) {
	g, me, opp, ws := mscWinterTable(t)
	bear := gyCard(me, "Fallen Bear", "Creature — Bear", "{2}{G}", "")
	declareAttack(t, g, opp.ID, ws)
	pickCard(t, g, me.ID, bear)
	// Shrunk to power 2 and then killed: the last-known power is 2, below the target's 3.
	findBattlefieldCardByID(g, ws).Counters = map[string]int{game.CounterMinusOne: 1}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ws, game.DestroyOptions{}) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the card returned although his last-known power (2) is below its mana value (3)")
	}
}

// --- Typhoid Mary, Fractured -------------------------------------------------------------

const mscMaryTypeLine = "Legendary Creature — Mutant Villain"

// mscMaryTable puts a 3/3 Typhoid Mary on seat 0's battlefield with a
// library to draw from and a seeded random stream.
func mscMaryTable(t *testing.T, key byte) (g *game.Game, me, opp *game.Player, mary uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	g.SetRNGKeyForTest(randomModesKey(key))
	me, opp = g.Seats[0], g.Seats[1]
	mary = b12Push(g, me.ID, "Typhoid Mary, Fractured", mscMaryTypeLine, mscTyphoidMaryOracle, 3, 3)
	seedLibrary(me, "A", "B", "C")
	return g, me, opp, mary
}

func mscMaryTreasures(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == owner && c.Name == "Treasure" {
			n++
		}
	}
	return n
}

// With no discard this turn the engine picks the bullet and nobody is asked.
func TestMscTyphoidMaryPicksAtRandomWithoutAskingAndRunsTheDrawnBullet(t *testing.T) {
	seen := map[int]bool{}
	for key := byte(0); key < 64 && len(seen) < 3; key++ {
		g, me, opp, mary := mscMaryTable(t, key)
		hand, life, oppLife := me.Hand.Size(), me.Life, opp.Life
		declareAttack(t, g, opp.ID, mary)
		if c := modePickChoiceFor(g, me.ID); c != nil {
			t.Fatalf("key %d: a random mode asks nobody to choose: %+v", key, c)
		}
		item := triggerOnStack(g, mary)
		if item == nil || len(item.Modes) != 1 {
			t.Fatalf("key %d: trigger %+v, want one drawn mode", key, item)
		}
		mode := item.Modes[0]
		seen[mode] = true
		passPriorityAroundTable(t, g)
		switch mode {
		case 0:
			if got := mscMaryTreasures(g, me.ID); got != 1 {
				t.Errorf("Mary: %d Treasure tokens, want 1", got)
			}
		case 1:
			if got := me.Hand.Size(); got != hand+1 {
				t.Errorf("Typhoid Mary: hand %d, want %d", got, hand+1)
			}
		case 2:
			if opp.Life != oppLife-2 || me.Life != life+2 {
				t.Errorf("Bloody Mary: life %d/%d, want %d/%d", me.Life, opp.Life, life+2, oppLife-2)
			}
		default:
			t.Fatalf("key %d: mode %d is not one of the three bullets", key, mode)
		}
		if mode != 0 && mscMaryTreasures(g, me.ID) != 0 {
			t.Errorf("key %d mode %d: a Treasure appeared that the mode does not make", key, mode)
		}
		if mode != 1 && me.Hand.Size() != hand {
			t.Errorf("key %d mode %d: a card was drawn that the mode does not draw", key, mode)
		}
		if mode != 2 && opp.Life != oppLife {
			t.Errorf("key %d mode %d: the opponent lost life that the mode does not take", key, mode)
		}
	}
	if len(seen) != 3 {
		t.Errorf("64 draws reached modes %v; a uniform pick reaches all three", seen)
	}
}

func mscMaryDiscards(g *game.Game, p *game.Player) {
	p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Creature — Bear", Owner: p.ID, Controller: p.ID})
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(p.ID, 1) })
}

// After a discard the controller is asked, is offered all three bullets, and
// gets the bullet they pick.
func TestMscTyphoidMaryAfterADiscardYouChooseInstead(t *testing.T) {
	for pick, want := range []string{"Mary", "Typhoid Mary", "Bloody Mary"} {
		g, me, opp, mary := mscMaryTable(t, 7)
		mscMaryDiscards(g, me)
		hand, life, oppLife := me.Hand.Size(), me.Life, opp.Life
		declareAttack(t, g, opp.ID, mary)

		c := modePickChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("%s: a discard this turn should hand the choice to the controller, no prompt was asked", want)
		}
		if len(c.ModeOptionIndex) != 3 {
			t.Fatalf("%s: offered %v, want all three bullets", want, c.ModeOptionIndex)
		}
		if c.Reason != "Choose one" {
			t.Errorf("prompt = %q, want the ordinary %q", c.Reason, "Choose one")
		}
		if triggerOnStack(g, mary) != nil {
			t.Fatalf("%s: the trigger went on the stack before the player chose", want)
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{pick}); err != nil {
			t.Fatalf("%s: ResolveModePick: %v", want, err)
		}
		passPriorityAroundTable(t, g)
		switch pick {
		case 0:
			if got := mscMaryTreasures(g, me.ID); got != 1 {
				t.Errorf("Mary: %d Treasure tokens, want 1", got)
			}
		case 1:
			if got := me.Hand.Size(); got != hand+1 {
				t.Errorf("Typhoid Mary: hand %d, want %d", got, hand+1)
			}
		case 2:
			if opp.Life != oppLife-2 || me.Life != life+2 {
				t.Errorf("Bloody Mary: life %d/%d, want %d/%d", me.Life, opp.Life, life+2, oppLife-2)
			}
		}
	}
}

// Only YOUR discard turns the draw into a choice (it reads the controller's
// own tally).
func TestMscTyphoidMaryAnOpponentsDiscardLeavesThePickRandom(t *testing.T) {
	g, me, opp, mary := mscMaryTable(t, 3)
	mscMaryDiscards(g, opp)
	declareAttack(t, g, opp.ID, mary)
	if c := modePickChoiceFor(g, me.ID); c != nil {
		t.Fatalf("an opponent's discard made Typhoid Mary's controller choose: %+v", c)
	}
	if item := triggerOnStack(g, mary); item == nil || len(item.Modes) != 1 {
		t.Errorf("trigger %+v, want one randomly drawn mode", item)
	}
}

// The discard has to be this turn's: the next turn's attack is random again.
func TestMscTyphoidMaryAnEarlierTurnsDiscardIsForgotten(t *testing.T) {
	g, me, opp, mary := mscMaryTable(t, 5)
	mscMaryDiscards(g, me)
	nextTurnOfMine(t, g)
	declareAttack(t, g, opp.ID, mary)
	if c := modePickChoiceFor(g, me.ID); c != nil {
		t.Fatalf("a discard from an earlier turn made Typhoid Mary's controller choose: %+v", c)
	}
	if item := triggerOnStack(g, mary); item == nil || len(item.Modes) != 1 {
		t.Errorf("trigger %+v, want one randomly drawn mode", item)
	}
}

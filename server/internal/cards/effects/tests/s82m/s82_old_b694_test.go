package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b694_test.go — the ten Mercadian Masques life cards (#5619),
// each played through the real catalog and the engine.

const (
	mmqPiousWarrior   = "163e5482-700c-4d91-b70c-9170cc8f1ad2"
	mmqSoothingBalm   = "1f493e26-a556-4ca0-bb44-5408800ccf82"
	mmqDeepwoodTantiv = "3b31a9a7-6c97-4e82-bf72-de83034f81a9"
	mmqHighwayRobber  = "5de0769e-5141-42d5-aaa6-549cfc154de8"
	mmqLiability      = "65f1bebb-6441-4189-b6de-993ec9d24a59"
	mmqRenounce       = "72f9a6f7-ea26-438d-aafa-758239fd3681"
	mmqJeweledTorque  = "91f36a03-6031-4834-98ee-20b36cdf5569"
	mmqMoonlitWake    = "c93965a8-c527-4b1f-8322-b043f910ffb7"
	mmqSacredPrey     = "e24110e8-72cf-4dd0-bca0-665758daca88"
	mmqEmbargo        = "edadd0bf-15c8-4e9c-810e-2edfd77a9d01"
)

func TestMmqB694Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		mmqPiousWarrior: "Pious Warrior", mmqSoothingBalm: "Soothing Balm",
		mmqDeepwoodTantiv: "Deepwood Tantiv", mmqHighwayRobber: "Highway Robber",
		mmqLiability: "Liability", mmqRenounce: "Renounce",
		mmqJeweledTorque: "Jeweled Torque", mmqMoonlitWake: "Moonlit Wake",
		mmqSacredPrey: "Sacred Prey", mmqEmbargo: "Embargo",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestMmqPiousWarriorGainsTheCombatDamageDealtToIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warrior := effectstest.PushCatalogPermanent(g, opp.ID, "Pious Warrior", "Creature — Human Rebel Warrior", mmqPiousWarrior, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == warrior {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 3
			}
		}
	})
	brute := effectstest.B12Creature(g, me.ID, "Brute", "Creature — Ogre", 3, 3)
	before := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, brute)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(warrior, brute); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before+3 {
		t.Errorf("life %d, want %d (3 combat damage dealt to the Warrior)", opp.Life, before+3)
	}
}

func TestMmqPiousWarriorIgnoresNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	warrior := effectstest.PushCatalogPermanent(g, me.ID, "Pious Warrior", "Creature — Human Rebel Warrior", mmqPiousWarrior, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == warrior {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 3
			}
		}
		_ = g.DealDamageToCreatureForEffect(uuid.New(), warrior, 2)
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 40 {
		t.Errorf("life %d, want 40: only combat damage counts", me.Life)
	}
}

func TestMmqSoothingBalmGainsFiveForTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Soothing Balm", "Instant", mmqSoothingBalm, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before+5 {
		t.Errorf("life %d, want %d", opp.Life, before+5)
	}
}

func TestMmqBlockedCreaturesGainLifeOnceWhenBlocked(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		life         int
	}{
		{"Deepwood Tantiv", mmqDeepwoodTantiv, 2},
		{"Sacred Prey", mmqSacredPrey, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			atk := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Creature — Beast", tc.oracle, false)
			w1 := effectstest.B12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 4)
			w2 := effectstest.B12Creature(g, opp.ID, "Wall B", "Creature — Wall", 0, 4)
			before := me.Life
			effectstest.DeclareAttack(t, g, opp.ID, atk)
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			for _, w := range []uuid.UUID{w1, w2} {
				if err := g.DeclareBlocker(w, atk); err != nil {
					t.Fatalf("DeclareBlocker: %v", err)
				}
			}
			effectstest.FinishBlockDeclarations(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if me.Life != before+tc.life {
				t.Errorf("life %d, want %d (once, however many block)", me.Life, before+tc.life)
			}
		})
	}
}

func TestMmqHighwayRobberDrainsTheTargetOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	myLife, theirLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Highway Robber", "Creature — Human Mercenary", mmqHighwayRobber, nil)
	pick := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	effectstest.AssistCardSettle(t, g, me.ID, &pick, true)
	if me.Life != myLife+2 || opp.Life != theirLife-2 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, myLife+2, theirLife-2)
	}
}

func TestMmqMoonlitWakeGainsForEveryCreatureThatDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Moonlit Wake", "Enchantment", mmqMoonlitWake, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	before := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 2) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear survived")
	}
	if me.Life != before+1 {
		t.Errorf("life %d, want %d", me.Life, before+1)
	}
}

func TestMmqLiabilityDrainsTheOwnerOfANontokenPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Liability", "Enchantment", mmqLiability, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	token := effectstest.B12Creature(g, opp.ID, "Their Token", "Token Creature — Bear", 2, 2)
	before := opp.Life
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.New(), token, 2)
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Fatalf("a token died and life went %d -> %d", before, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 2) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("life %d, want %d (the owner loses 1)", opp.Life, before-1)
	}
	if me.Life != 40 {
		t.Errorf("Liability's controller life changed to %d", me.Life)
	}
}

func TestMmqEmbargoStopsNonlandsUntappingAndCostsTwoEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Embargo", "Enchantment", mmqEmbargo, false)
	bear := effectstest.B12Creature(g, me.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	for _, id := range []uuid.UUID{bear, land} {
		if err := g.TapCard(id, true); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.UntapStaticTapped(t, g, bear) {
		t.Error("a nonland permanent untapped")
	}
	if effectstest.UntapStaticTapped(t, g, land) {
		t.Error("a land stayed tapped")
	}
	if me.Life != 40 {
		t.Errorf("life %d: Embargo's controller is the other player", me.Life)
	}
}

func TestMmqEmbargoUpkeepCostsItsControllerTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Embargo", "Enchantment", mmqEmbargo, false)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 38 {
		t.Errorf("life %d, want 38", me.Life)
	}
}

func TestMmqJeweledTorqueGainsOnlyWhenPaidForTheChosenColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushChosenColorPermanent(t, g, me.ID, "Jeweled Torque", "Artifact", mmqJeweledTorque, "R")
	effectstest.AdvanceToMain(t, g)
	before := me.Life
	blue := game.Card{Name: "Blue Thing", TypeLine: "Instant", ManaCost: "{U}", Colors: []string{"U"}}
	if err := effectstest.SocB19CastWithPool(g, me, blue, "U", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) || me.Life != before {
		t.Fatalf("a blue spell offered the Torque's payment (life %d)", me.Life)
	}
	red := game.Card{Name: "Red Bolt", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}}
	if err := effectstest.SocB19CastWithPool(g, me, red, "R", 3, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a red spell did not offer the payment")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d, want %d after paying {2}", me.Life, before+2)
	}
	// Declining gains nothing.
	if err := effectstest.SocB19CastWithPool(g, me, red, "R", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d after declining, want %d", me.Life, before+2)
	}
}

func TestMmqRenounceSacrificesTheChosenPermanentsForTwoLifeEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	keep := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Renounce", "Instant", mmqRenounce, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || !g.Battlefield.Contains(keep) {
		t.Fatal("the wrong permanents were sacrificed")
	}
	if me.Life != before+4 {
		t.Errorf("life %d, want %d (2 for each of two)", me.Life, before+4)
	}
}

func TestMmqRenounceMayChooseNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	keep := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Renounce", "Instant", mmqRenounce, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(keep) || me.Life != before {
		t.Errorf("choosing none changed the board or life (%d)", me.Life)
	}
}

package testsru

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b35_cards_test.go — S82 BNS batch 35 (Foundations Jumpstart,
// #5929): Militant Inquisitor, Minotaur Sureshot, Mogg Flunkies, Moodmark
// Painter, Morkrut Necropod, Murasa Ranger, Muse Drake, Mystic Skyfish,
// Nantuko Husk, Ojutai Exemplars, Oran-Rief Ooze.

const (
	bnsB35InquisitorOracle = "d1e326e8-6e9e-4dbd-8c2f-0bef90db2bb7"
	bnsB35SureshotOracle   = "cbb6c481-bcf1-459d-a6fb-bb46cc521368"
	bnsB35FlunkiesOracle   = "b92fc9c1-7264-4894-90b8-a193039489c5"
	bnsB35PainterOracle    = "f5068b54-b1f8-4751-9a09-fc1b554e8115"
	bnsB35NecropodOracle   = "ed3990cd-fa67-427b-bbe4-a2b1942e9be2"
	bnsB35RangerOracle     = "84ba3598-a2d2-46d4-b73c-ab139cac37c4"
	bnsB35MuseDrakeOracle  = "04c9b3d7-83ba-400e-9837-2ba0119e9445"
	bnsB35SkyfishOracle    = "645b5784-a6f7-4cf3-966a-e1a51420b96b"
	bnsB35HuskOracle       = "0dcefc00-9425-43e3-bd47-1e34fff8b0e2"
	bnsB35ExemplarsOracle  = "4405db13-7b36-4dc2-9a5d-1d31d568abf5"
	bnsB35OozeOracle       = "94a7dcef-d1c1-4ffa-869e-b911e84ed66c"
)

func bnsB35Counters(g *game.Game, id uuid.UUID, n int) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				if g.Battlefield.Cards[i].Counters == nil {
					g.Battlefield.Cards[i].Counters = map[string]int{}
				}
				g.Battlefield.Cards[i].Counters[game.CounterPlusOne] = n
			}
		}
	})
}

func TestBnsB35MilitantInquisitorCountsYourEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inq := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Militant Inquisitor", bnsB35InquisitorOracle, 2, 3))
	if v := effectstest.S58p6View(t, g, inq); v.CurrentPower() != 2 {
		t.Fatalf("power with no Equipment = %d, want 2", v.CurrentPower())
	}
	effectstest.FinABEquipment(g, me.ID, "Sword")
	effectstest.FinABEquipment(g, me.ID, "Axe")
	effectstest.FinABEquipment(g, opp.ID, "Their Sword")
	effectstest.PushTypedCard(g, me.ID, "Rock", "Artifact", "{1}")
	if v := effectstest.S58p6View(t, g, inq); v.CurrentPower() != 4 || v.CurrentToughness() != 3 {
		t.Errorf("with two Equipment of mine = %d/%d, want 4/3", v.CurrentPower(), v.CurrentToughness())
	}
}

func TestBnsB35MinotaurSureshotFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	s := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Minotaur Sureshot", bnsB35SureshotOracle, 2, 3))
	if !effectstest.HasKeywordOf(g, s, "reach") {
		t.Error("no reach")
	}
	effectstest.B06AddMana(me, "R", "C", "R", "C")
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, s); p != 4 {
		t.Errorf("power after two activations = %d, want 4", p)
	}
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with an empty pool")
	}
}

func TestBnsB35MoggFlunkiesCantAttackAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fl := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Mogg Flunkies", bnsB35FlunkiesOracle, 3, 3))
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(fl, opp.ID); err != nil {
		t.Fatalf("staging the Flunkies: %v", err)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackLimit) {
		t.Fatalf("passing with the Flunkies alone: %v, want ErrAttackLimit", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("the companion was refused: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, fl); c == nil || c.AttackingTarget == uuid.Nil {
		t.Error("the Flunkies are not attacking beside the Bear")
	}
}

func TestBnsB35MoggFlunkiesCantBlockAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a1 := effectstest.PushVanillaCreature(g, me.ID, "Raider One", 2, 2)
	a2 := effectstest.PushVanillaCreature(g, me.ID, "Raider Two", 2, 2)
	fl := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Mogg Flunkies", bnsB35FlunkiesOracle, 3, 3))
	bear := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Bear", "", 2, 2))
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: a1, Target: opp.ID}, {Attacker: a2, Target: opp.ID}}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	err := g.DeclareBlocker(fl, a1)
	var re *game.BlockRefusedError
	if !errors.As(err, &re) || re.Reason != game.BlockReasonCantBlockAlone {
		t.Fatalf("blocking alone: %v, want cant_block_alone", err)
	}
	if err := g.DeclareBlocker(bear, a1); err != nil {
		t.Fatalf("the Bear block: %v", err)
	}
	if err := g.DeclareBlocker(fl, a2); err != nil {
		t.Fatalf("the Flunkies beside the Bear: %v", err)
	}
}

func TestBnsB35MoodmarkPainterPumpsByCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.P1g2Graveyard(me, "Creature — Zombie", 3)
	effectstest.P1g2Graveyard(me, "Sorcery", 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastAndResolveCreature(t, g, "Moodmark Painter", "Creature — Human Shaman", bnsB35PainterOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("Bear = %d/%d, want 5/2", p, tough)
	}
	if !effectstest.HasKeywordOf(g, bear, "menace") {
		t.Error("the Bear did not gain menace")
	}
}

func bnsB35NecropodTable(t *testing.T) (g *game.Game, me *game.Player, pod uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	pod = effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Morkrut Necropod", bnsB35NecropodOracle, 7, 7))
	return g, me, pod
}

func TestBnsB35MorkrutNecropodAttackSacrificesAnotherCreatureOrLand(t *testing.T) {
	g, me, pod := bnsB35NecropodTable(t)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	land := effectstest.PushTypedCard(g, me.ID, "Forest", "Basic Land — Forest", "")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(pod, g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	for _, id := range c.ChooseCards {
		if id == pod {
			t.Error("the Necropod itself is offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if g.Battlefield.Contains(land) || !me.Graveyard.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	if !g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(pod) {
		t.Error("something else was sacrificed")
	}
}

func TestBnsB35MorkrutNecropodAloneSacrificesNothing(t *testing.T) {
	g, me, pod := bnsB35NecropodTable(t)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(pod, g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a sacrifice prompt with nothing to sacrifice")
	}
	if !g.Battlefield.Contains(pod) {
		t.Error("the Necropod sacrificed itself")
	}
}

func TestBnsB35MorkrutNecropodBlockSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.PushVanillaCreature(g, me.ID, "Raider", 2, 2)
	pod := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Morkrut Necropod", bnsB35NecropodOracle, 7, 7))
	fodder := effectstest.B12Creature(g, opp.ID, "Fodder", "Creature — Goblin", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: atk, Target: opp.ID}}); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(pod, atk); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, opp.ID) == nil {
		t.Fatal("no sacrifice prompt on a block")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, fodder)
	if g.Battlefield.Contains(fodder) {
		t.Error("the Fodder was not sacrificed")
	}
}

func TestBnsB35MurasaRangerLandfallPaysForCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ranger := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Murasa Ranger", bnsB35RangerOracle, 3, 3))
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("landfall did not offer {3}{G}")
	}
	effectstest.B06AddMana(me, "G", "C", "C", "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, ranger, game.CounterPlusOne); n != 2 {
		t.Errorf("counters after paying = %d, want 2", n)
	}

	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		effectstest.AnswerPayUnless(t, g, me.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.B12Counter(t, g, ranger, game.CounterPlusOne); n != 2 {
		t.Errorf("counters after declining = %d, want still 2", n)
	}
}

func TestBnsB35MuseDrakeDrawsOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	before := effectstest.HandSize(me)
	effectstest.CastAndResolveCreature(t, g, "Muse Drake", "Creature — Drake", bnsB35MuseDrakeOracle)
	effectstest.PassPriorityAroundTable(t, g)
	// The Drake left the hand and a card was drawn.
	// The fixture puts the Drake in hand and casts it: net one drawn card.
	if got := effectstest.HandSize(me); got != before+1 {
		t.Errorf("hand = %d, want %d (drew exactly one card)", got, before+1)
	}
}

func TestBnsB35MysticSkyfishFliesOnTheSecondDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	fish := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Mystic Skyfish", bnsB35SkyfishOracle, 3, 1))
	draw := func() {
		g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
		effectstest.PassPriorityAroundTable(t, g)
	}
	// The turn's draw step was the first card; the next draw is the second.
	if n := g.TurnTallyFor(me.ID).CardsDrawn; n != 1 {
		t.Fatalf("draw tally = %d before the test draw, want 1", n)
	}
	if effectstest.HasKeywordOf(g, fish, "flying") {
		t.Fatal("flying before the second draw")
	}
	draw()
	if !effectstest.HasKeywordOf(g, fish, "flying") {
		t.Error("no flying after the second draw")
	}
	// A third draw is not the second: a fish that arrives later stays grounded.
	later := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Mystic Skyfish", bnsB35SkyfishOracle, 3, 1))
	draw()
	if effectstest.HasKeywordOf(g, later, "flying") {
		t.Error("the third draw gave flying")
	}
}

func TestBnsB35NantukoHuskSacrificesForPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	husk := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Nantuko Husk", bnsB35HuskOracle, 2, 2))
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, husk, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, husk); p != 4 || tough != 4 {
		t.Errorf("Husk = %d/%d, want 4/4", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, husk, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated without sacrificing a creature")
	}
	rock := effectstest.PushTypedCard(g, me.ID, "Rock", "Artifact", "{1}")
	if err := g.ActivateCatalogAbility(me.ID, husk, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Error("sacrificed a noncreature permanent")
	}
}

func bnsB35Exemplars(t *testing.T) (g *game.Game, me *game.Player, ex uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me = g.Seats[g.Turn.ActiveSeat]
	ex = effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Ojutai Exemplars", bnsB35ExemplarsOracle, 4, 4))
	return g, me, ex
}

func TestBnsB35OjutaiExemplarsModes(t *testing.T) {
	t.Run("tap target creature", func(t *testing.T) {
		g, me, _ := bnsB35Exemplars(t)
		theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
		effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
		effectstest.PickCard(t, g, me.ID, theirs)
		effectstest.PassPriorityAroundTable(t, g)
		if c := effectstest.FindBattlefieldCardForTest(g, theirs); c == nil || !c.Tapped {
			t.Error("the target was not tapped")
		}
	})
	t.Run("first strike and lifelink", func(t *testing.T) {
		g, me, ex := bnsB35Exemplars(t)
		effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
		effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasKeywordOf(g, ex, "first strike") || !effectstest.HasKeywordOf(g, ex, "lifelink") {
			t.Error("missing first strike or lifelink")
		}
	})
	t.Run("blink returns tapped", func(t *testing.T) {
		g, me, ex := bnsB35Exemplars(t)
		effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
		effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 2)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(ex) {
			t.Fatal("the old object is still on the battlefield")
		}
		var c *game.Card
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].OracleID == bnsB35ExemplarsOracle {
				c = &g.Battlefield.Cards[i]
			}
		}
		if c == nil {
			t.Fatal("the Exemplars did not return")
		}
		if c.Controller != me.ID {
			t.Error("the Exemplars returned under another controller")
		}
		if !c.Tapped {
			t.Error("the Exemplars returned untapped")
		}
	})
	t.Run("creature spell does not trigger", func(t *testing.T) {
		g, me, _ := bnsB35Exemplars(t)
		effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
		effectstest.NoModePick(t, g, me.ID, "a creature spell is not a noncreature spell")
	})
}

func TestBnsB35OranRiefOozeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastAndResolveCreature(t, g, "Oran-Rief Ooze", "Creature — Ooze", bnsB35OozeOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, bear, game.CounterPlusOne); n != 1 {
		t.Fatalf("Bear counters after the enter trigger = %d, want 1", n)
	}

	// A later combat: the Ooze (no counter) and two creatures attack.
	ooze := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Oran-Rief Ooze", bnsB35OozeOracle, 2, 2))
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain", 2, 2)
	home := effectstest.PushVanillaCreature(g, me.ID, "Stays Home", 2, 2)
	bnsB35Counters(g, home, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{ooze, bear, plain} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("attacking Bear with a counter = %d, want 2", n)
	}
	if n := effectstest.B12Counter(t, g, plain, game.CounterPlusOne); n != 0 {
		t.Errorf("attacking creature without a counter = %d, want 0", n)
	}
	if n := effectstest.B12Counter(t, g, ooze, game.CounterPlusOne); n != 0 {
		t.Errorf("the Ooze without a counter = %d, want 0", n)
	}
	if n := effectstest.B12Counter(t, g, home, game.CounterPlusOne); n != 1 {
		t.Errorf("a creature that did not attack = %d, want 1", n)
	}
}

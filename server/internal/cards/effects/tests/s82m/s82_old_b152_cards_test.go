package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b152_cards_test.go — S82 older sets, Commander 2019 batch 152
// (#5517): Thousand Winds; Icefeather Aven; Fresh Meat; Runic Repetition;
// Boneyard Parley; Wingmate Roc; From Under the Floorboards; Ghired, Conclave
// Exile; Ghastly Conscription. (Gift of Doom and Hedonist's Trove are skipped;
// see the commit message.)

const (
	c19B152WindsOracle    = "81b50d34-d778-4606-80de-0e0780610d1f"
	c19B152AvenOracle     = "9fb901db-1ff7-4a8d-8233-fe7e3dee7fd1"
	c19B152MeatOracle     = "3be6e78f-f9ec-4bf4-b70f-40ba677aaa48"
	c19B152RepetitionOrac = "8ab99b5e-c9e0-45c2-8df2-864ab0ad1caf"
	c19B152ParleyOracle   = "a3d048c8-e8df-4b50-a422-fc94a69f44bb"
	c19B152RocOracle      = "5badc2e8-a7d6-48c7-b594-65db953d9d57"
	c19B152FloorboardsOra = "a1fc4269-f3e4-4a59-849d-aa1727eef23a"
	c19B152GhiredOracle   = "ce08fbfb-5d51-4dfc-80ae-87d5d9c7726f"
	c19B152ConscriptOracl = "3bf20e68-0e19-4242-bfba-a8a1bb244685"
)

func TestC19B152CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		c19B152WindsOracle:    "Thousand Winds",
		c19B152AvenOracle:     "Icefeather Aven",
		c19B152MeatOracle:     "Fresh Meat",
		c19B152RepetitionOrac: "Runic Repetition",
		c19B152ParleyOracle:   "Boneyard Parley",
		c19B152RocOracle:      "Wingmate Roc",
		c19B152FloorboardsOra: "From Under the Floorboards",
		c19B152GhiredOracle:   "Ghired, Conclave Exile",
		c19B152ConscriptOracl: "Ghastly Conscription",
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

func c19B152Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
}

func TestC19B152ThousandWindsBouncesEveryOtherTappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirTapped := effectstest.B12Creature(g, opp.ID, "Their Tapped", "Creature — Bear", 2, 2)
	theirUntapped := effectstest.B12Creature(g, opp.ID, "Their Untapped", "Creature — Bear", 2, 2)
	myTapped := effectstest.B12Creature(g, me.ID, "My Tapped", "Creature — Bear", 2, 2)
	c19B152Tap(g, theirTapped)
	c19B152Tap(g, myTapped)
	id := effectstest.S36CastFaceDown(t, g, "Thousand Winds", "Creature — Elemental", c19B152WindsOracle, "morph", 5, 6)
	effectstest.S36TurnUp(t, g, me, id, "U", "U", "C", "C", "C", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirTapped) || !opp.Hand.Contains(theirTapped) {
		t.Error("their tapped creature was not returned to its owner's hand")
	}
	if g.Battlefield.Contains(myTapped) || !me.Hand.Contains(myTapped) {
		t.Error("my own tapped creature was not returned too")
	}
	if !g.Battlefield.Contains(theirUntapped) {
		t.Error("an untapped creature was returned")
	}
	if !g.Battlefield.Contains(id) {
		t.Error("Thousand Winds returned itself")
	}
}

func TestC19B152IcefeatherAvenMayBounceAnotherCreatureWhenTurnedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.S36CastFaceDown(t, g, "Icefeather Aven", "Creature — Bird Shaman", c19B152AvenOracle, "morph", 2, 2)
	effectstest.S36TurnUp(t, g, me, id, "G", "U", "C")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the target was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the Aven left")
	}
}

func TestC19B152FreshMeatMakesABeastForEachCreatureThatWentToYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "My A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "My B", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	survivor := effectstest.B12Creature(g, me.ID, "Survivor", "Creature — Bear", 3, 3)
	if err := (effects.CreateToken{Controller: me.ID, Template: effects.TokenCard("2/2 green Bear"), N: 1}).Apply(effects.KeyedContext(g, me.ID, uuid.Nil)); err != nil {
		t.Fatalf("token: %v", err)
	}
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && effects.IsToken(c) {
			token = c.InstanceID
		}
	}
	for _, id := range []uuid.UUID{a, b, token, theirs} {
		effectstest.DealDamageTo(t, g, survivor, id, 5)
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || g.Battlefield.Contains(token) {
		t.Fatal("setup: my creatures did not die")
	}
	effectstest.B06AddMana(me, "G", "C", "C", "C")
	effectstest.CastFromHandForTest(t, g, me, "Fresh Meat", "Instant", "{3}{G}", c19B152MeatOracle, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Beast"); n != 3 {
		t.Errorf("%d Beasts, want 3 (two nontokens and a token of mine; the opponent's creature does not count)", n)
	}
}

func TestC19B152RunicRepetitionReturnsOnlyAnExiledCardWithFlashback(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	flash := uuid.New()
	plain := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: flash, Name: "Faithless Looting", TypeLine: "Sorcery", ManaCost: "{R}", OracleID: effectstest.FaithlessLootingOracl, Owner: me.ID, Controller: me.ID})
		g.Exile.PushTop(game.Card{InstanceID: plain, Name: "Plain", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	effectstest.B06AddMana(me, "U", "C", "C")
	if err := g.CastSpell(me.ID, effectstest.PushCatalogHandCard(me, "Runic Repetition", "Sorcery", c19B152RepetitionOrac), game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(plain)}); err == nil {
		t.Fatal("a card without flashback was accepted")
	}
	effectstest.B06AddMana(me, "U", "C", "C")
	id := effectstest.PushCatalogHandCard(me, "Runic Repetition", "Sorcery", c19B152RepetitionOrac)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(flash)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(flash) || g.Exile.Contains(flash) {
		t.Error("the exiled flashback card did not return to hand")
	}
	if !g.Exile.Contains(plain) {
		t.Error("the plain card left exile")
	}
}

func TestC19B152WingmateRocMakesABirdOnlyAfterYouAttacked(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if attacked {
			attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
			effectstest.DeclareAttack(t, g, opp.ID, attacker)
			effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
		} else {
			effectstest.AdvanceToMain(t, g)
		}
		roc := effectstest.CastAndResolveCreature(t, g, "Wingmate Roc", "Creature — Bird", c19B152RocOracle)
		effectstest.PassPriorityAroundTable(t, g)
		birds := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && effects.IsToken(c) && c.Name == "Bird" && c.Power == 3 && c.Toughness == 4 {
				birds++
			}
		}
		want := 0
		if attacked {
			want = 1
		}
		if birds != want {
			t.Errorf("attacked=%v: %d Bird tokens, want %d (roc %v)", attacked, birds, want, roc)
		}
	}
}

func TestC19B152WingmateRocGainsLifeForEachAttackingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	roc := effectstest.PushCatalogPermanent(g, me.ID, "Wingmate Roc", "Creature — Bird", c19B152RocOracle, false)
	friend := effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Stays Home", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, roc, friend)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d (one for each of two attackers)", me.Life, life+2)
	}
}

func TestC19B152FromUnderTheFloorboardsMakesThreeTappedZombiesAndGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "B", "C", "C", "C")
	life := me.Life
	effectstest.CastFromHandForTest(t, g, me, "From Under the Floorboards", "Sorcery", "{3}{B}{B}", c19B152FloorboardsOra, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	n, tapped := effectstest.CountTokensNamed(g, me.ID, "Zombie")
	if n != 3 || tapped != 3 {
		t.Errorf("%d Zombies, %d tapped; want 3 and 3", n, tapped)
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestC19B152FromUnderTheFloorboardsMadnessMakesXZombiesAndGainsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	card := effectstest.TorB1LootIntoExile(t, g, "From Under the Floorboards", "Sorcery", c19B152FloorboardsOra)
	effectstest.DcMana(t, g, me.ID, "{C}{C}{B}{B}")
	life := me.Life
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness, XValue: 2}); err != nil {
		t.Fatalf("the madness cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	n, tapped := effectstest.CountTokensNamed(g, me.ID, "Zombie")
	if n != 2 || tapped != 2 {
		t.Errorf("%d Zombies, %d tapped; want 2 and 2", n, tapped)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestC19B152GhiredMakesARhinoAndPopulatesTappedAndAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ghired := effectstest.CastAndResolveCreature(t, g, "Ghired, Conclave Exile", "Legendary Creature — Human Shaman", c19B152GhiredOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Rhino"); n != 1 {
		t.Fatalf("%d Rhinos after entering, want 1", n)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ghired {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	effectstest.DeclareAttack(t, g, opp.ID, ghired)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestOptionPickFor(g, me.ID); p != nil {
		if err := g.ResolveOptionPick(p.ID, me.ID, 0); err != nil {
			t.Fatalf("keep attacking: %v", err)
		}
	}
	rhinos, attacking, tapped := 0, 0, 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && effects.IsToken(c) && c.Name == "Rhino" {
			rhinos++
			if c.AttackingTarget != uuid.Nil {
				attacking++
			}
			if c.Tapped {
				tapped++
			}
		}
	}
	if rhinos != 2 || attacking != 1 || tapped < 1 {
		t.Errorf("%d Rhinos, %d attacking, %d tapped; want 2, the new one attacking and tapped", rhinos, attacking, tapped)
	}
}

func TestC19B152GhastlyConscriptionManifestsEveryCreatureCardFromAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	c1 := effectstest.B38Graveyard(g, opp.ID, "Dead One", "Creature — Bear", "{1}", 3, 3)
	c2 := effectstest.B38Graveyard(g, opp.ID, "Dead Two", "Creature — Bear", "{2}", 4, 4)
	spell := effectstest.B38Graveyard(g, opp.ID, "Old Spell", "Sorcery", "{1}", 0, 0)
	effectstest.B06AddMana(me, "B", "B", "C", "C", "C", "C", "C")
	effectstest.CastFromHandForTest(t, g, me, "Ghastly Conscription", "Sorcery", "{5}{B}{B}", c19B152ConscriptOracl,
		game.CastSpellParams{Strict: true, Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	manifested := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.FaceDownIsPermanent() && c.CurrentPower() == 2 && c.CurrentToughness() == 2 {
			manifested++
		}
	}
	if manifested != 2 {
		t.Errorf("%d face-down 2/2s under my control, want 2", manifested)
	}
	if opp.Graveyard.Contains(c1) || opp.Graveyard.Contains(c2) {
		t.Error("a creature card is still in the graveyard")
	}
	if !opp.Graveyard.Contains(spell) {
		t.Error("the noncreature card left the graveyard")
	}
}

func TestC19B152BoneyardParleySplitsExiledCreaturesIntoTwoPiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B38Graveyard(g, me.ID, "My Dead", "Creature — Bear", "{1}", 2, 2)
	theirs1 := effectstest.B38Graveyard(g, opp.ID, "Their Dead 1", "Creature — Bear", "{1}", 3, 3)
	theirs2 := effectstest.B38Graveyard(g, opp.ID, "Their Dead 2", "Creature — Bear", "{1}", 4, 4)
	effectstest.B06AddMana(me, "B", "B", "C", "C", "C", "C", "C")
	effectstest.CastFromHandForTest(t, g, me, "Boneyard Parley", "Sorcery", "{5}{B}{B}", c19B152ParleyOracle,
		game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(mine, theirs1, theirs2)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	}
	split := effectstest.LatestRevealPickFor(g, opp.ID)
	if split == nil {
		t.Fatalf("the opponent was not asked to separate: %+v", g.PendingChoices)
	}
	// Pile one: my own card and one of theirs; pile two: the other of theirs.
	if err := g.ResolveRevealPick(split.ID, opp.ID, []uuid.UUID{mine, theirs1}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	p := effectstest.LatestOptionPickFor(g, me.ID)
	if p == nil {
		t.Fatalf("no pile pick: %+v", g.PendingChoices)
	}
	if err := g.ResolveOptionPick(p.ID, me.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	names := map[string]bool{"My Dead": true, "Their Dead 1": true, "Their Dead 2": true}
	taken, left := 0, 0
	for _, c := range g.Battlefield.Cards {
		if names[c.Name] {
			if c.Controller != me.ID {
				t.Errorf("%s entered under %v, want me", c.Name, c.Controller)
			}
			taken++
		}
	}
	for _, p := range []*game.Player{me, opp} {
		for _, c := range p.Graveyard.Cards {
			if names[c.Name] {
				if c.Owner != p.ID {
					t.Errorf("%s is in %v's graveyard but owned by %v", c.Name, p.ID, c.Owner)
				}
				left++
			}
		}
	}
	if taken != 2 || left != 1 {
		t.Errorf("%d on the battlefield and %d in graveyards, want the two-card pile in play and the one-card pile in its owner's graveyard", taken, left)
	}
}

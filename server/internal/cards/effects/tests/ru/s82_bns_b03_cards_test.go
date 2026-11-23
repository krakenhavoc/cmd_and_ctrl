package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 BNS batch 3 (Breaking News, #5929): twelve cards, engine-driven.

const (
	bnsTyrantsScorn         = "afb278f1-0e21-4f6d-96bb-c34aa3ef96b0"
	bnsArchiveTrap          = "d412b175-2365-4b2c-a8f5-3e4e9d06280d"
	bnsOkoThiefOfCrowns     = "60c60923-ff1b-43f7-8768-731499fcffc9"
	bnsCruelUltimatum       = "01294ac9-1a06-4ee8-b3b1-db1f07c9d94e"
	bnsHeartlessPillage     = "8e5e8bca-b52a-4178-ab53-bdd0e0e11dc1"
	bnsIndomitableCreativty = "e2161ae2-e5d2-4a56-85d4-d2214c5b3e2e"
	bnsOverwhelmingForces   = "bdccceb6-ad3c-4a75-a8a6-ca796ede4185"
	bnsPrimalCommand        = "350f5c0a-563f-41c5-8f7b-10f409bb4d3b"
	bnsSavageSmash          = "676c0315-1cb3-4d80-b136-16d88fd612cb"
	bnsSkewerTheCritics     = "c43b616e-ef87-4c12-810b-a1a3260168fa"
	bnsTerminalAgony        = "b69e68e7-0612-40f8-a40a-9c4e77653442"
	bnsCrimePunishment      = "c3df3574-c4da-415f-a56b-807ed8e41d73"
)

func TestBNSBatch3CardsAreRegisteredFull(t *testing.T) {
	for _, id := range []string{
		bnsTyrantsScorn, bnsArchiveTrap, bnsOkoThiefOfCrowns, bnsCruelUltimatum, bnsHeartlessPillage,
		bnsIndomitableCreativty, bnsOverwhelmingForces, bnsPrimalCommand, bnsSavageSmash,
		bnsSkewerTheCritics, bnsTerminalAgony, bnsCrimePunishment, bnsCrimePunishment + "#1",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: spec = %+v, %v", id, spec, ok)
		}
	}
}

// bnsPutInHand puts a card in the active seat's hand and moves to a main phase.
func bnsPutInHand(t *testing.T, g *game.Game, name, typeLine, oracle, mana string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		ManaCost: mana, Owner: me.ID, Controller: me.ID})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id
}

func bnsCast(t *testing.T, g *game.Game, name, typeLine, oracle, mana string, p game.CastSpellParams) error {
	t.Helper()
	id := bnsPutInHand(t, g, name, typeLine, oracle, mana)
	me := g.Seats[g.Turn.ActiveSeat]
	return g.CastSpell(me.ID, id, p)
}

func bnsCreature(g *game.Game, owner uuid.UUID, name, mana string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", ManaCost: mana,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func bnsCard(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// --- Tyrant's Scorn ------------------------------------------------

func TestBNSTyrantsScornDestroysAThreeOrLessAndRefusesAFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	small := bnsCreature(g, opp.ID, "Small", "{1}{G}", 2, 2)
	big := bnsCreature(g, opp.ID, "Big", "{3}{G}", 4, 4)
	if err := bnsCast(t, g, "Tyrant's Scorn", "Instant", bnsTyrantsScorn, "{U}{B}",
		game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: []game.TargetRef{effectstest.ModeRef(game.TargetCard, big, 0, 0)}}); err == nil {
		t.Fatal("destroyed a mana value 4 creature")
	}
	if err := bnsCast(t, g, "Tyrant's Scorn", "Instant", bnsTyrantsScorn, "{U}{B}",
		game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{effectstest.ModeRef(game.TargetCard, small, 0, 0)}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || !g.Battlefield.Contains(big) {
		t.Errorf("small on table %v, big on table %v", g.Battlefield.Contains(small), g.Battlefield.Contains(big))
	}
}

func TestBNSTyrantsScornBouncesAnyCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	big := bnsCreature(g, opp.ID, "Big", "{5}{G}", 6, 6)
	if err := bnsCast(t, g, "Tyrant's Scorn", "Instant", bnsTyrantsScorn, "{U}{B}",
		game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{effectstest.ModeRef(game.TargetCard, big, 0, 0)}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !opp.Hand.Contains(big) {
		t.Error("the creature did not return to its owner's hand")
	}
}

// --- Archive Trap ----------------------------------------------------

func bnsSearched(g *game.Game, actor, target uuid.UUID, label string) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventSearchLibrary, Actor: actor, Target: target, Label: label})
	})
}

func TestBNSArchiveTrapIsFreeOnlyAfterAnOpponentSearchedTheirOwnLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(opp, "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t")
	opps := game.CastSpellParams{Strict: true, AlternativeCost: "free",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}

	// None of these is "an opponent searched their library".
	bnsSearched(g, opp.ID, opp.ID, "shuffle")
	bnsSearched(g, me.ID, me.ID, "")
	bnsSearched(g, me.ID, opp.ID, "")
	if err := bnsCast(t, g, "Archive Trap", "Instant — Trap", bnsArchiveTrap, "{3}{U}{U}", opps); err == nil {
		t.Fatal("cast for {0} with no opponent search this turn")
	}
	bnsSearched(g, opp.ID, opp.ID, "")
	before := opp.Graveyard.Size()
	if err := bnsCast(t, g, "Archive Trap", "Instant — Trap", bnsArchiveTrap, "{3}{U}{U}", opps); err != nil {
		t.Fatalf("cast for {0}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size() - before; got != 13 {
		t.Errorf("milled %d, want 13", got)
	}
}

func TestBNSArchiveTrapMillsWhatALibraryHolds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	if opp.Library.Size() > 5 {
		opp.Library.Cards = opp.Library.Cards[:5]
	}
	if err := bnsCast(t, g, "Archive Trap", "Instant — Trap", bnsArchiveTrap, "{3}{U}{U}",
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != 0 || opp.Eliminated {
		t.Errorf("library %d, eliminated %v: milling the library empty loses nothing (CR 701.17b)", opp.Library.Size(), opp.Eliminated)
	}
}

// --- Heartless Pillage -----------------------------------------------

func TestBNSHeartlessPillageDiscardsTwoAndRaidMakesATreasure(t *testing.T) {
	for _, raid := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if raid {
			atk := bnsCreature(g, me.ID, "Raider", "{1}", 1, 1)
			effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(atk, opp.ID); err != nil {
				t.Fatal(err)
			}
			effectstest.LockInAttacks(t, g)
			effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
		}
		hand := opp.Hand.Size()
		if err := bnsCast(t, g, "Heartless Pillage", "Sorcery", bnsHeartlessPillage, "{2}{B}",
			game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.DiscardFromHand(t, g, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if got := hand - opp.Hand.Size(); got != 2 {
			t.Errorf("raid=%v: discarded %d, want 2", raid, got)
		}
		want := 0
		if raid {
			want = 1
		}
		if n := effectstest.RfdCountNamed(g, me.ID, "Treasure"); n != want {
			t.Errorf("raid=%v: %d Treasures, want %d", raid, n, want)
		}
	}
}

// --- Skewer the Critics ----------------------------------------------

func TestBNSSkewerTheCriticsSpectacleNeedsAnOpponentToHaveLostLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	effectstest.B06AddMana(me, "R")
	if err := bnsCast(t, g, "Skewer the Critics", "Sorcery", bnsSkewerTheCritics, "{2}{R}",
		game.CastSpellParams{Strict: true, AlternativeCost: "spectacle", Targets: target}); err == nil {
		t.Fatal("spectacle offered with no opponent having lost life")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })
	before := opp.Life
	effectstest.B06AddMana(me, "R")
	if err := bnsCast(t, g, "Skewer the Critics", "Sorcery", bnsSkewerTheCritics, "{2}{R}",
		game.CastSpellParams{Strict: true, AlternativeCost: "spectacle", Targets: target}); err != nil {
		t.Fatalf("spectacle after life loss: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("life %d → %d, want 3 damage", before, opp.Life)
	}
}

// --- Savage Smash ----------------------------------------------------

func TestBNSSavageSmashPumpsThenFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := bnsCreature(g, me.ID, "Mine", "{1}{G}", 2, 2)
	theirs := bnsCreature(g, opp.ID, "Theirs", "{2}{R}", 3, 3)
	other := bnsCreature(g, me.ID, "Other", "{1}", 1, 1)
	if err := bnsCast(t, g, "Savage Smash", "Sorcery", bnsSavageSmash, "{1}{R}{G}",
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: other, Slot: 1}}}); err == nil {
		t.Fatal("fought a creature I control")
	}
	effectstest.B06AddMana(me, "R", "G", "C")
	if err := bnsCast(t, g, "Savage Smash", "Sorcery", bnsSavageSmash, "{1}{R}{G}",
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the 4/4 did not kill the 3/3")
	}
	if !g.Battlefield.Contains(mine) || effectstest.EffectivePower(t, g, mine) != 4 {
		t.Errorf("my creature should survive as a 4/4 until end of turn")
	}
}

// --- Terminal Agony --------------------------------------------------

func TestBNSTerminalAgonyDestroysAndHasMadness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := bnsCreature(g, opp.ID, "Victim", "{2}", 2, 2)
	if err := bnsCast(t, g, "Terminal Agony", "Sorcery", bnsTerminalAgony, "{2}{B}{R}",
		game.CastSpellParams{Targets: []game.TargetRef{bnsCard(victim)}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature survived")
	}
	if reps, trigs := effectstest.MadnessAbilitiesFor(bnsTerminalAgony); reps != 1 || trigs != 1 {
		t.Errorf("madness halves: %d replacement, %d trigger", reps, trigs)
	}
}

// --- Overwhelming Forces ---------------------------------------------

func TestBNSOverwhelmingForcesDestroysTheirCreaturesAndDrawsPerDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	a := bnsCreature(g, opp.ID, "A", "{1}", 1, 1)
	b := bnsCreature(g, opp.ID, "B", "{1}", 1, 1)
	c := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Tough", TypeLine: "Creature — Golem",
		Power: 2, Toughness: 2, Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID})
	mine := bnsCreature(g, me.ID, "Mine", "{1}", 1, 1)
	theirsElsewhere := bnsCreature(g, third.ID, "Third", "{1}", 1, 1)
	hand := me.Hand.Size()
	if err := bnsCast(t, g, "Overwhelming Forces", "Sorcery", bnsOverwhelmingForces, "{6}{B}{B}",
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a creature of the target opponent survived")
	}
	if !g.Battlefield.Contains(c) || !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirsElsewhere) {
		t.Error("the indestructible creature or another player's creature was destroyed")
	}
	// The spell was put in hand and cast, so the hand delta is the draws.
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2: one per creature really destroyed", got)
	}
}

// --- Cruel Ultimatum -------------------------------------------------

func TestBNSCruelUltimatumDoesAllSevenThings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "a", "b", "c", "d")
	victim := bnsCreature(g, opp.ID, "Victim", "{1}", 1, 1)
	spare := bnsCreature(g, opp.ID, "Spare", "{1}", 1, 1)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Fallen", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	oppHand, oppLife, myLife, myHand := opp.Hand.Size(), opp.Life, me.Life, me.Hand.Size()
	if err := bnsCast(t, g, "Cruel Ultimatum", "Sorcery", bnsCruelUltimatum, "{U}{U}{B}{B}{B}{R}{R}",
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); c != nil {
		if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{dead}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(victim) || !g.Battlefield.Contains(spare) {
		t.Error("the opponent should sacrifice exactly the one creature they chose")
	}
	if got := oppHand - opp.Hand.Size(); got != 3 {
		t.Errorf("opponent discarded %d, want 3", got)
	}
	if opp.Life != oppLife-5 || me.Life != myLife+5 {
		t.Errorf("life: opponent %d→%d, me %d→%d", oppLife, opp.Life, myLife, me.Life)
	}
	if !me.Hand.Contains(dead) {
		t.Error("the creature card was not returned to my hand")
	}
	// +1 returned, +3 drawn (the spell itself went in and out of the hand).
	if got := me.Hand.Size() - myHand; got != 4 {
		t.Errorf("my hand grew by %d, want 4 (return +1, draw +3)", got)
	}
}

// --- Primal Command --------------------------------------------------

func TestBNSPrimalCommandChoosesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	creature := bnsCreature(g, opp.ID, "Bear", "{1}", 2, 2)
	if err := bnsCast(t, g, "Primal Command", "Sorcery", bnsPrimalCommand, "{3}{G}{G}",
		game.CastSpellParams{Strict: true, Modes: []int{0, 1}, Targets: []game.TargetRef{
			effectstest.ModeRef(game.TargetPlayer, me.ID, 0, 0), effectstest.ModeRef(game.TargetCard, creature, 1, 0)}}); err == nil {
		t.Fatal("put a creature on top of a library")
	}
	life := me.Life
	libTop := opp.Library.Size()
	if err := bnsCast(t, g, "Primal Command", "Sorcery", bnsPrimalCommand, "{3}{G}{G}",
		game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{
			effectstest.ModeRef(game.TargetPlayer, me.ID, 0, 0), effectstest.ModeRef(game.TargetCard, rock, 1, 0)}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+7 {
		t.Errorf("life %d → %d, want +7", life, me.Life)
	}
	if g.Battlefield.Contains(rock) || opp.Library.Size() != libTop+1 {
		t.Error("the noncreature permanent did not go on top of its owner's library")
	}
}

func TestBNSPrimalCommandShufflesAGraveyardAndTutorsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gy := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: gy, Name: "Old", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	want := effectstest.SeedLibraryTop(me, "Tutor Target", "Creature — Elf", "{G}")
	if err := bnsCast(t, g, "Primal Command", "Sorcery", bnsPrimalCommand, "{3}{G}{G}",
		game.CastSpellParams{Modes: []int{2, 3}, Targets: []game.TargetRef{effectstest.ModeRef(game.TargetPlayer, me.ID, 0, 0)}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(gy) && !me.Library.Contains(gy) {
		t.Error("the graveyard card was not shuffled into the library")
	}
	search := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
	if search == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{want}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(want) {
		t.Error("the creature card did not reach my hand")
	}
	if me.Graveyard.Contains(gy) {
		t.Error("the graveyard still holds its card")
	}
}

// --- Oko, Thief of Crowns --------------------------------------------

func bnsOko(t *testing.T, g *game.Game, loyalty int) uuid.UUID {
	t.Helper()
	seat := g.Turn.ActiveSeat
	oko := effectstest.PushWalkerForTest(g, g.Seats[seat].ID, "Oko, Thief of Crowns", bnsOkoThiefOfCrowns, loyalty)
	effectstest.AdvanceToMainOf(t, g, seat)
	return oko
}

func TestBNSOkoPlusTwoMakesFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oko := bnsOko(t, g, 4)
	if err := g.ActivateCatalogAbility(me.ID, oko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RfdCountNamed(g, me.ID, "Food") != 1 || effectstest.LoyaltyOf(g, oko) != 6 {
		t.Errorf("foods %d loyalty %d", effectstest.RfdCountNamed(g, me.ID, "Food"), effectstest.LoyaltyOf(g, oko))
	}
}

func TestBNSOkoPlusOneMakesAnythingAGreenElk(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, mana string
		keywords             []string
	}{
		{"Dragon", "Legendary Creature — Dragon", "{4}{R}{R}", []string{"flying"}},
		{"Rock", "Artifact — Equipment", "{2}", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			oko := bnsOko(t, g, 4)
			id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: tc.name, TypeLine: tc.typeLine,
				ManaCost: tc.mana, Power: 7, Toughness: 7, Keywords: tc.keywords, Owner: opp.ID, Controller: opp.ID})
			if err := g.ActivateCatalogAbility(me.ID, oko, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{bnsCard(id)}}); err != nil {
				t.Fatal(err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 3 || tough != 3 {
				t.Errorf("%s is %d/%d, want 3/3", tc.name, p, tough)
			}
			if ty := effectstest.EffectiveTypes(t, g, id); len(ty) != 1 || ty[0] != "Creature" {
				t.Errorf("types %v, want only Creature", ty)
			}
			if st := effectstest.EffectiveSubtypes(t, g, id); len(st) != 1 || st[0] != "Elk" {
				t.Errorf("subtypes %v, want only Elk", st)
			}
			if cs := effectstest.EffectiveColors(t, g, id); len(cs) != 1 || cs[0] != "G" {
				t.Errorf("colors %v, want [G]", cs)
			}
			if ab := effectstest.EffectiveAbilities(t, g, id); len(ab) != 0 {
				t.Errorf("abilities %v, want none", ab)
			}
			if effectstest.LoyaltyOf(g, oko) != 5 {
				t.Errorf("loyalty %d, want 5", effectstest.LoyaltyOf(g, oko))
			}
		})
	}
}

func TestBNSOkoMinusFiveExchangesControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	oko := bnsOko(t, g, 5)
	mine := bnsCreature(g, me.ID, "Mine", "{1}", 1, 1)
	small := bnsCreature(g, opp.ID, "Small", "{4}", 3, 3)
	big := bnsCreature(g, opp.ID, "Big", "{4}", 4, 4)
	if err := g.ActivateCatalogAbility(me.ID, oko, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: big, Slot: 1}}}); err == nil {
		t.Fatal("exchanged for a creature with power 4")
	}
	if err := g.ActivateCatalogAbility(me.ID, oko, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: small, Slot: 1}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, mine) != opp.ID || effectstest.ControllerOf(t, g, small) != me.ID {
		t.Error("control was not exchanged")
	}
	if effectstest.LoyaltyOf(g, oko) != 0 && g.Battlefield.Contains(oko) {
		t.Errorf("loyalty %d", effectstest.LoyaltyOf(g, oko))
	}
}

// --- Indomitable Creativity ------------------------------------------

func TestBNSIndomitableCreativityDestroysXAndRebuildsEachControllerBoard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := bnsCreature(g, opp.ID, "A", "{1}", 1, 1)
	b := bnsCreature(g, opp.ID, "B", "{1}", 1, 1)
	m := bnsCreature(g, me.ID, "M", "{1}", 1, 1)
	keep := bnsCreature(g, opp.ID, "Keep", "{1}", 1, 1)
	// Their library, top first: a land, a creature, an artifact. Two permanents
	// died, so they reveal twice: creature then artifact (the land is revealed
	// and left in the library).
	art := effectstest.SeedLibraryTop(opp, "Gem", "Artifact", "{2}")
	cre := effectstest.SeedLibraryTop(opp, "Cub", "Creature — Bear", "{1}")
	land := effectstest.SeedLibraryTop(opp, "Forest", "Basic Land — Forest", "")
	mine := effectstest.SeedLibraryTop(me, "Mine", "Creature — Elf", "{1}")
	if err := bnsCast(t, g, "Indomitable Creativity", "Sorcery", bnsIndomitableCreativty, "{X}{R}{R}{R}",
		game.CastSpellParams{XValue: 3, Targets: []game.TargetRef{bnsCard(a), bnsCard(b), bnsCard(m)}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, dead := range []uuid.UUID{a, b, m} {
		if g.Battlefield.Contains(dead) {
			t.Error("a target survived")
		}
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("an untargeted creature died")
	}
	// A card that changed zones is a new object (CR 400.7), so look by name.
	for name, owner := range map[string]uuid.UUID{"Cub": opp.ID, "Gem": opp.ID, "Mine": me.ID} {
		if n := effectstest.RfdCountNamed(g, owner, name); n != 1 {
			t.Errorf("%d %s under its owner's control, want 1", n, name)
		}
	}
	if !opp.Library.Contains(land) {
		t.Error("the revealed land must stay in the library")
	}
	if opp.Library.Contains(cre) || opp.Library.Contains(art) || me.Library.Contains(mine) {
		t.Error("an exiled card is still in the library")
	}
}

func TestBNSIndomitableCreativityRefusesTheWrongNumberOfTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := bnsCreature(g, opp.ID, "A", "{1}", 1, 1)
	if err := bnsCast(t, g, "Indomitable Creativity", "Sorcery", bnsIndomitableCreativty, "{X}{R}{R}{R}",
		game.CastSpellParams{Strict: true, XValue: 2, Targets: []game.TargetRef{bnsCard(a)}}); err == nil {
		t.Fatal("cast X=2 with one target")
	}
}

// --- Crime // Punishment ---------------------------------------------

func TestBNSCrimePutsAnOpponentsCreatureOrEnchantmentOntoTheBattlefieldUnderMyControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Fallen", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	mine := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "Mine", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	land := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	for name, id := range map[string]uuid.UUID{"my own graveyard": mine, "a land card": land} {
		if err := bnsCast(t, g, "Crime", "Sorcery", bnsCrimePunishment, "{3}{W}{B}",
			game.CastSpellParams{Strict: true, Targets: []game.TargetRef{bnsCard(id)}}); err == nil {
			t.Errorf("targeted a card from %s", name)
		}
	}
	if err := bnsCast(t, g, "Crime", "Sorcery", bnsCrimePunishment, "{3}{W}{B}",
		game.CastSpellParams{Targets: []game.TargetRef{bnsCard(theirs)}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) || effectstest.ControllerOf(t, g, theirs) != me.ID {
		t.Error("the creature card did not come back under my control")
	}
}

func TestBNSPunishmentDestroysEveryPermanentWithManaValueX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	two := bnsCreature(g, opp.ID, "Two", "{1}{G}", 2, 2)
	mine := bnsCreature(g, me.ID, "MyTwo", "{2}", 2, 2)
	ench := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Aura", TypeLine: "Enchantment", ManaCost: "{1}{W}", Owner: opp.ID, Controller: opp.ID})
	art := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Gem", TypeLine: "Artifact", ManaCost: "{2}", Owner: opp.ID, Controller: opp.ID})
	three := bnsCreature(g, opp.ID, "Three", "{3}", 3, 3)
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Wastes", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID})
	effectstest.B19CastXModal(t, g, "Punishment", "Sorcery", bnsCrimePunishment+"#1", "{X}{B}{G}", 2, nil, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, dead := range []uuid.UUID{two, mine, ench, art} {
		if g.Battlefield.Contains(dead) {
			t.Error("a mana value 2 permanent survived")
		}
	}
	if !g.Battlefield.Contains(three) || !g.Battlefield.Contains(land) {
		t.Error("a mana value 3 creature or a land was destroyed")
	}
}

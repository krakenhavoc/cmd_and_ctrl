package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Crew Captain ---------------------------------------------------------

func TestSncB13CrewCaptainIsIndestructibleOnlyOnTheTurnItEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Crew Captain", "Creature — Human Warrior", sncB13CrewCaptain, true)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id}) })
	if !effectstest.HasEffectiveKeyword(t, g, id, "indestructible") {
		t.Fatal("not indestructible on the turn it entered")
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Error("no haste")
	}
	effectstest.NextTurnOfMine(t, g)
	if effectstest.HasEffectiveKeyword(t, g, id, "indestructible") {
		t.Error("still indestructible on a later turn")
	}
}

// --- Freelance Muscle -----------------------------------------------------

func TestSncB13FreelanceMuscleGrowsByTheGreatestPowerOrToughnessOnAttackAndBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	muscle := effectstest.B12Push(g, me.ID, "Freelance Muscle", "Creature — Rhino Warrior", sncB13Muscle, 4, 4)
	effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 7) // toughness 7 is the greatest number
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 2)
	effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 9, 9) // not yours
	effectstest.DeclareAttack(t, g, opp.ID, muscle)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, muscle), effectstest.EffectiveToughness(t, g, muscle); p != 11 || tough != 11 {
		t.Fatalf("attacking: %d/%d, want 11/11", p, tough)
	}
}

func TestSncB13FreelanceMuscleAloneGetsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	muscle := effectstest.B12Push(g, me.ID, "Freelance Muscle", "Creature — Rhino Warrior", sncB13Muscle, 4, 4)
	effectstest.DeclareAttack(t, g, opp.ID, muscle)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, muscle); p != 4 {
		t.Errorf("alone: power %d, want 4", p)
	}
}

// --- Mr. Orfeo, the Boulder -----------------------------------------------

func TestSncB13OrfeoDoublesTheTargetsPowerOncePerAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Mr. Orfeo, the Boulder", "Legendary Creature — Rhino Warrior", sncB13Orfeo, 2, 4)
	a := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	b := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target pick for the trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: a}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, a); p != 6 {
		t.Errorf("target power %d, want 6", p)
	}
	if p := effectstest.EffectivePower(t, g, b); p != 2 {
		t.Errorf("a non-target changed: power %d, want 2", p)
	}
}

// --- Cutthroat Contender --------------------------------------------------

func TestSncB13ContenderPaysLifeOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Cutthroat Contender", "Creature — Vampire Warrior", sncB13Contender, 1, 1)
	life := me.Life
	if err := sncB13Activate(t, g, me, id, 0, "", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := sncB13Activate(t, g, me, id, 0, "", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a second activation in the same turn was allowed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, me.Life)
	}
	if p := effectstest.EffectivePower(t, g, id); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
}

func TestSncB13ContenderCantActivateWithoutLifeToPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Cutthroat Contender", "Creature — Vampire Warrior", sncB13Contender, 1, 1)
	me.Life = 0
	if err := sncB13Activate(t, g, me, id, 0, "", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("paid life it did not have")
	}
}

// --- Incriminate ----------------------------------------------------------

func TestSncB13IncriminateTheControllerChoosesWhichOfTheTwoToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 3, 3)
	c := effectstest.B12Creature(g, opp.ID, "Cat", "Creature — Cat", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Incriminate", "Sorcery", sncB13Incriminate, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, opp.ID) == nil {
		t.Fatal("the creatures' controller was not asked")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(b) || !g.Battlefield.Contains(a) || !g.Battlefield.Contains(c) {
		t.Errorf("zones: a on bf %v, b on bf %v, c on bf %v", g.Battlefield.Contains(a), g.Battlefield.Contains(b), g.Battlefield.Contains(c))
	}
}

func TestSncB13IncriminateRefusesTargetsControlledByDifferentPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 3, 3)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Incriminate", TypeLine: "Sorcery", OracleID: sncB13Incriminate, ManaCost: "{1}{B}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b, Slot: 1},
	}}); err == nil {
		t.Fatal("cast with creatures of two different players")
	}
}

func TestSncB13IncriminateIndestructibleStillGoes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Darksteel", TypeLine: "Artifact Creature — Golem",
		Power: 2, Toughness: 2, Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID})
	b := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Incriminate", "Sorcery", sncB13Incriminate, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Error("the indestructible creature was not sacrificed")
	}
}

// --- Dusk Mangler ---------------------------------------------------------

// sncB13CastMangler casts Dusk Mangler for its mana and the named additional
// cost branch (0 sacrifice, 1 discard, 2 pay life).
func sncB13CastMangler(t *testing.T, g *game.Game, me *game.Player, branch int, params game.CastSpellParams) error {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Dusk Mangler", TypeLine: "Creature — Horror", OracleID: sncB13Mangler,
		ManaCost: "{5}{B}{B}", Power: 5, Toughness: 4, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "C", 5)
	effectstest.FillPoolColored(me, "B", 2)
	params.CostBranch = &branch
	params.Strict = true
	return g.CastSpell(me.ID, id, params)
}

func TestSncB13ManglerEntryMakesEachOpponentSacrificeThenDiscardThenLoseLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	opps := g.Seats[1:]
	victims := map[uuid.UUID][]uuid.UUID{}
	lives := map[uuid.UUID]int{}
	hands := map[uuid.UUID]int{}
	for _, p := range opps {
		victims[p.ID] = []uuid.UUID{
			effectstest.B12Creature(g, p.ID, "Bear", "Creature — Bear", 2, 2),
			effectstest.B12Creature(g, p.ID, "Wolf", "Creature — Wolf", 3, 3),
		}
		effectstest.HandCardForTest(p, "Spare", "Sorcery", "")
		effectstest.HandCardForTest(p, "Spare2", "Sorcery", "")
		lives[p.ID] = p.Life
		hands[p.ID] = p.Hand.Size()
	}
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	if err := sncB13CastMangler(t, g, me, 0, game.CastSpellParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice branch did not sacrifice the creature")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Printed order: every opponent sacrifices first, before any discard.
	for _, p := range opps {
		if c := effectstest.DiscardChoiceFor(g, p.ID); c != nil && !effectstest.IsSacrificePrompt(c) {
			t.Fatalf("a discard prompt for %s was open before every sacrifice was answered", p.ID)
		}
		effectstest.AnswerSacrifice(t, g, p.ID, victims[p.ID][0])
		effectstest.PassPriorityAroundTable(t, g)
	}
	for _, p := range opps {
		if c := effectstest.DiscardChoiceFor(g, p.ID); c != nil {
			effectstest.DiscardFromHand(t, g, p.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	for _, p := range opps {
		if g.Battlefield.Contains(victims[p.ID][0]) || !g.Battlefield.Contains(victims[p.ID][1]) {
			t.Errorf("seat %s did not sacrifice exactly the chosen creature", p.ID)
		}
		if p.Hand.Size() != hands[p.ID]-1 {
			t.Errorf("seat %s hand %d -> %d, want -1", p.ID, hands[p.ID], p.Hand.Size())
		}
		if p.Life != lives[p.ID]-4 {
			t.Errorf("seat %s life %d -> %d, want -4", p.ID, lives[p.ID], p.Life)
		}
	}
}

func TestSncB13ManglerPaysLifeOrDiscardsWhenChosen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	if err := sncB13CastMangler(t, g, me, 2, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast paying life: %v", err)
	}
	if me.Life != life-4 {
		t.Errorf("life %d -> %d, want -4", life, me.Life)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	pitch := effectstest.HandCardForTest(me2, "Pitch", "Sorcery", "")
	if err := sncB13CastMangler(t, g2, me2, 1, game.CastSpellParams{DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("cast discarding: %v", err)
	}
	for _, c := range me2.Graveyard.Cards {
		if c.InstanceID == pitch {
			return
		}
	}
	t.Error("the discard branch did not discard the card")
}

func TestSncB13ManglerRefusesWithoutAPayableBranchChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// Branch 0 (sacrifice) with nothing to sacrifice.
	if err := sncB13CastMangler(t, g, me, 0, game.CastSpellParams{}); err == nil {
		t.Fatal("cast with the sacrifice branch and no creature named")
	}
}

// --- Bellowing Mauler -----------------------------------------------------

func sncB13PassToEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSncB13MaulerEachPlayerLosesLifeUnlessTheySacrificeANontokenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Bellowing Mauler", "Creature — Ogre Warrior", sncB13Mauler, 4, 6)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 3, 3)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	sncB13PassToEndStep(t, g)
	// You first: sacrifice the Bear instead of losing life.
	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		t.Fatal("the controller was not asked")
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.AnswerChooseCards(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	// The opponent takes the life loss.
	if effectstest.LatestOptionPickFor(g, opp.ID) == nil {
		t.Fatal("the next player was not asked after the first answered")
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the Bear was not sacrificed")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the opponent's Ox went")
	}
	if me.Life != lives[me.ID] {
		t.Errorf("the controller lost life after sacrificing: %d -> %d", lives[me.ID], me.Life)
	}
	if opp.Life != lives[opp.ID]-4 {
		t.Errorf("the opponent's life %d -> %d, want -4", lives[opp.ID], opp.Life)
	}
}

func TestSncB13MaulerTokensCantPayAndAPlayerWithNoNontokenCreatureJustLoses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	// The Mauler is itself a nontoken creature, so it could pay; make it a
	// token's worth of board for the others by leaving only tokens.
	effectstest.B12Push(g, me.ID, "Bellowing Mauler", "Creature — Ogre Warrior", sncB13Mauler, 4, 6)
	for _, p := range g.Seats[1:] {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
			Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
	}
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	sncB13PassToEndStep(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Life != lives[p.ID]-4 {
			t.Errorf("seat %s life %d -> %d, want -4 with no nontoken creature to sacrifice", p.ID, lives[p.ID], p.Life)
		}
	}
}

// --- Industrial Advancement -----------------------------------------------

func TestSncB13AdvancementSacrificesThenPutsACreatureFromTheTopXOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Industrial Advancement", "Enchantment", sncB13Advancement, 0, 0)
	fodder := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 4, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == fodder {
				g.Battlefield.Cards[i].ManaCost = "{2}{R}"
			}
		}
	})
	// Top of library, top first: a land, a creature, then a card outside X = 3.
	deep := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: deep, Name: "Deep", TypeLine: "Creature — Elf", ManaCost: "{G}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	creature := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: creature, Name: "Target Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	filler := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: filler, Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	sncB13PassToEndStep(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no creature pick after the sacrifice")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if !offered[creature] || offered[deep] || offered[land] || offered[filler] {
		t.Errorf("offered %v, want only the creature among the top three", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, creature)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(creature) {
		t.Error("the creature did not enter")
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still there")
	}
	// The rest went to the bottom: the deep card is now on top of the
	// library, the land and the filler beneath it.
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != deep {
		t.Errorf("top of the library is %s, want the card outside X", top.Name)
	}
}

func TestSncB13AdvancementDeclineSacrificesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Industrial Advancement", "Enchantment", sncB13Advancement, 0, 0)
	fodder := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 4, 4)
	sncB13PassToEndStep(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fodder) {
		t.Error("declined, but the creature was sacrificed")
	}
}

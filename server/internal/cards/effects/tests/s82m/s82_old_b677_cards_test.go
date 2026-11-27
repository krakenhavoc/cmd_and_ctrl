package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b677_cards_test.go — S82 OLD batch 677 (#5618, Nemesis: card draw
// and library). Each test plays the real catalog entry.

const (
	oldB677RathiAssassin = "1ea0c775-fde0-4a21-b4c2-8bcec0e4f4f8"
	oldB677NestingWurm   = "3160dada-d59f-4f23-8420-ba14f9df7514"
	oldB677PackHunt      = "3ee41993-9d14-43a8-bf92-433c4f3aab01"
	oldB677Sentinel      = "4f437f42-1ac5-4681-9b16-4eae9f49bb22"
	oldB677Vanguard      = "680766e2-637b-40bc-a15f-0b4dc55bd619"
	oldB677Eye           = "e0af613a-05ba-438d-ba74-3f6a9958576c"
	oldB677Witch         = "ebd77c52-97ea-421f-99eb-c6e2263cdecf"
)

func TestOldB677EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB677RathiAssassin: "Rathi Assassin",
		oldB677NestingWurm:   "Nesting Wurm",
		oldB677PackHunt:      "Pack Hunt",
		oldB677Sentinel:      "Skyshroud Sentinel",
		oldB677Vanguard:      "Defiant Vanguard",
		oldB677Eye:           "Eye of Yawgmoth",
		oldB677Witch:         "Divining Witch",
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

// --- Rathi Assassin ---------------------------------------------------------------

func oldB677Tapped(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
}

func TestOldB677RathiAssassinDestroysATappedNonblackCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	assassin := effectstest.B12Push(g, me.ID, "Rathi Assassin", "Creature — Phyrexian Zombie Mercenary Assassin", oldB677RathiAssassin, 2, 2)
	green := effectstest.B12Creature(g, opp.ID, "Green Bear", "Creature — Bear", 2, 2)
	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Black Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID})
	oldB677Tapped(g, black)
	effectstest.FloatForTest(g, me, "BBB")
	activate := func(target uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, assassin, 0, game.ActivateAbilityParams{Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
	}
	if err := activate(green); err == nil {
		t.Fatal("an untapped creature was a legal target")
	}
	if err := activate(black); err == nil {
		t.Fatal("a black creature was a legal target")
	}
	oldB677Tapped(g, green)
	if err := activate(green); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("the tapped nonblack creature was not destroyed")
	}
}

func TestOldB677RecruitersFindTheirSubtypeUpToTheirBound(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana, subtype string
		maxMV                                 int
	}{
		{"Rathi Assassin", "Creature — Phyrexian Zombie Mercenary Assassin", oldB677RathiAssassin, "CCC", "Mercenary", 3},
		{"Defiant Vanguard", "Creature — Human Rebel", oldB677Vanguard, "CCCCC", "Rebel", 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.SeedSearchLibrary(me,
				game.Card{Name: "Edge " + tc.subtype, TypeLine: "Creature — " + tc.subtype, ManaCost: "{2}{B}{B}"},
				game.Card{Name: "Big " + tc.subtype, TypeLine: "Creature — " + tc.subtype, ManaCost: "{5}{B}"},
				game.Card{Name: "Small " + tc.subtype, TypeLine: "Creature — " + tc.subtype, ManaCost: "{1}"},
				game.Card{Name: "Small Soldier", TypeLine: "Creature — Human Soldier", ManaCost: "{W}"},
			)
			src := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			idx := 1
			if tc.name == "Defiant Vanguard" {
				idx = 0
			}
			effectstest.FloatForTest(g, me, tc.mana)
			if err := g.ActivateCatalogAbility(me.ID, src, idx, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.SearchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("the search asks")
			}
			edge := effectstest.SearchOptionNamed(g, c, "Edge "+tc.subtype) != uuid.Nil
			if edge != (tc.maxMV >= 4) {
				t.Errorf("a mana value 4 card offered = %v, bound %d", edge, tc.maxMV)
			}
			if effectstest.SearchOptionNamed(g, c, "Big "+tc.subtype) != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Small Soldier") != uuid.Nil {
				t.Error("a too-expensive or off-type card was offered")
			}
			effectstest.AnswerSearchNamed(t, g, me.ID, "Small "+tc.subtype)
			if effectstest.FindBattlefieldByName(g, "Small "+tc.subtype) == uuid.Nil {
				t.Error("the found card did not enter the battlefield")
			}
		})
	}
}

// --- Nesting Wurm and Skyshroud Sentinel ------------------------------------------

func TestOldB677NamedSearchFetchesUpToThreeCopies(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Nesting Wurm", "Creature — Wurm", oldB677NestingWurm},
		{"Skyshroud Sentinel", "Creature — Elf", oldB677Sentinel},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			var cards []game.Card
			for i := 0; i < 4; i++ {
				cards = append(cards, game.Card{Name: tc.name, TypeLine: tc.typeLine, OracleID: tc.oracle})
			}
			cards = append(cards, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
			effectstest.SeedSearchLibrary(me, cards...)
			hand := len(me.Hand.Cards)
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.SearchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no search prompt")
			}
			if effectstest.SearchOptionNamed(g, c, "Forest") != uuid.Nil {
				t.Error("a card with another name was offered")
			}
			if len(c.SearchCards) != 4 {
				t.Fatalf("%d cards offered, want the four copies", len(c.SearchCards))
			}
			effectstest.BeginSearchIfAsked(t, g, c)
			if err := g.ResolveSearchLibrary(c.ID, me.ID, c.SearchCards[:4]); err == nil {
				t.Error("four cards were accepted for an up-to-three search")
			}
			if err := g.ResolveSearchLibrary(c.ID, me.ID, c.SearchCards[:3]); err != nil {
				t.Fatalf("taking three copies: %v", err)
			}
			if len(me.Hand.Cards) != hand+3 {
				t.Errorf("hand %d → %d, want three copies in", hand, len(me.Hand.Cards))
			}
		})
	}
}

func TestOldB677NestingWurmHasTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Nesting Wurm", "Creature — Wurm", oldB677NestingWurm, 4, 3)
	if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Error("Nesting Wurm does not have trample")
	}
}

// --- Pack Hunt --------------------------------------------------------------------

func TestOldB677PackHuntFetchesCardsNamedLikeTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	var cards []game.Card
	for i := 0; i < 4; i++ {
		cards = append(cards, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"})
	}
	cards = append(cards, game.Card{Name: "Runeclaw Bear", TypeLine: "Creature — Bear"})
	effectstest.SeedSearchLibrary(me, cards...)
	var target uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Grizzly Bears" {
			target = c.InstanceID
		}
	}
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Pack Hunt", "Sorcery", oldB677PackHunt, effectstest.CardTarget(target))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if len(c.SearchCards) != 4 || effectstest.SearchOptionNamed(g, c, "Runeclaw Bear") != uuid.Nil {
		t.Fatalf("%d cards offered, want the four Grizzly Bears only", len(c.SearchCards))
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, c.SearchCards[:4]); err == nil {
		t.Error("four cards were accepted for an up-to-three search")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, c.SearchCards[:3]); err != nil {
		t.Fatalf("taking three copies: %v", err)
	}
	if len(me.Hand.Cards) != hand+3 {
		t.Errorf("hand %d → %d, want three copies in", hand, len(me.Hand.Cards))
	}
}

func TestOldB677PackHuntTargetsOnlyACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Pack Hunt", "Sorcery", oldB677PackHunt, effectstest.PlayerRefs(me.ID)); err == nil {
		t.Error("a player was a legal target")
	}
}

// --- Defiant Vanguard -------------------------------------------------------------

func TestOldB677DefiantVanguardDestroysItselfAndWhatItBlockedAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 5)
	bystander := effectstest.B12Creature(g, me.ID, "Bystander", "Creature — Bear", 1, 1)
	vanguard := effectstest.PushCatalogPermanent(g, opp.ID, "Defiant Vanguard", "Creature — Human Rebel", oldB677Vanguard, false)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, vanguard)
	effectstest.PassPriorityAroundTable(t, g)
	if zone, _ := effectstest.CardWhere(g, opp, vanguard); zone != "battlefield" {
		t.Fatalf("the Vanguard is in %s before end of combat", zone)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if zone, _ := effectstest.CardWhere(g, opp, vanguard); zone != "graveyard" {
		t.Errorf("the Vanguard is in %s after end of combat, want the graveyard", zone)
	}
	if g.Battlefield.Contains(attacker) {
		t.Error("the creature it blocked survived")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("a creature it did not block was destroyed")
	}
}

func TestOldB677DefiantVanguardStaysWhenItDoesNotBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	vanguard := effectstest.PushCatalogPermanent(g, opp.ID, "Defiant Vanguard", "Creature — Human Rebel", oldB677Vanguard, false)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if zone, _ := effectstest.CardWhere(g, opp, vanguard); zone != "battlefield" {
		t.Errorf("the Vanguard that never blocked is in %s", zone)
	}
	if !g.Battlefield.Contains(attacker) {
		t.Error("the unblocked attacker was destroyed")
	}
}

// --- Eye of Yawgmoth --------------------------------------------------------------

func TestOldB677EyeOfYawgmothRevealsPowerCardsKeepsOneAndExilesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	eye := effectstest.B12Push(g, me.ID, "Eye of Yawgmoth", "Artifact", oldB677Eye, 0, 0)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Ogre", 3, 3)
	below := effectstest.SweepFTop(me, "Below", "Sorcery")
	c1 := effectstest.SweepFTop(me, "C1", "Sorcery")
	c2 := effectstest.SweepFTop(me, "C2", "Sorcery")
	c3 := effectstest.SweepFTop(me, "C3", "Sorcery")
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, eye, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	rp := effectstest.LatestRevealPickFor(g, me.ID)
	if rp == nil {
		t.Fatalf("no pick for the controller: %+v", g.PendingChoices)
	}
	if len(rp.ChooseCards) != 3 {
		t.Fatalf("%d cards revealed, want the sacrificed creature's power, 3", len(rp.ChooseCards))
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{c2}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, c2) {
		t.Error("the picked card is not in hand")
	}
	if !effectstest.InExile(g, c1) || !effectstest.InExile(g, c3) {
		t.Error("the other revealed cards were not exiled")
	}
	if !me.Library.Contains(below) {
		t.Error("a card below the revealed ones left the library")
	}
}

func TestOldB677EyeOfYawgmothWithAZeroPowerCreatureRevealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	eye := effectstest.B12Push(g, me.ID, "Eye of Yawgmoth", "Artifact", oldB677Eye, 0, 0)
	fodder := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	top := effectstest.SweepFTop(me, "Top", "Sorcery")
	hand, n := len(me.Hand.Cards), len(me.Library.Cards)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, eye, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand || len(me.Library.Cards) != n || !me.Library.Contains(top) {
		t.Error("a power-0 sacrifice revealed or moved cards")
	}
}

// --- Divining Witch ---------------------------------------------------------------

func TestOldB677DiviningWitchExilesSixThenFindsTheNamedCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	witch := effectstest.B12Push(g, me.ID, "Divining Witch", "Creature — Human Spellshaper", oldB677Witch, 1, 1)
	pitch := effectstest.HandCardAtMain(t, g, "Pitch", "Sorcery", "")
	g.WithWriteLock(func() { me.Library.Cards = nil })
	deep := effectstest.SweepFTop(me, "Deep", "Sorcery")
	wanted := effectstest.SweepFTop(me, "Wanted", "Sorcery")
	r1 := effectstest.SweepFTop(me, "Rev1", "Sorcery")
	r2 := effectstest.SweepFTop(me, "Rev2", "Sorcery")
	var six []uuid.UUID
	for i := 0; i < 6; i++ {
		six = append(six, effectstest.SweepFTop(me, "Six", "Sorcery"))
	}
	effectstest.FloatForTest(g, me, "CB")
	if err := g.ActivateCatalogAbility(me.ID, witch, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "Wanted")
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range six {
		if !effectstest.InExile(g, id) {
			t.Error("one of the top six was not exiled")
		}
	}
	if !effectstest.InHandOf(me, wanted) {
		t.Error("the named card is not in hand")
	}
	if !effectstest.InExile(g, r1) || !effectstest.InExile(g, r2) {
		t.Error("a revealed card that was not named was not exiled")
	}
	if !me.Library.Contains(deep) {
		t.Error("a card below the named one left the library")
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discard cost was not paid")
	}
}

func TestOldB677DiviningWitchWithAnAbsentNameExilesTheWholeLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	witch := effectstest.B12Push(g, me.ID, "Divining Witch", "Creature — Human Spellshaper", oldB677Witch, 1, 1)
	pitch := effectstest.HandCardAtMain(t, g, "Pitch", "Sorcery", "")
	effectstest.FloatForTest(g, me, "CB")
	if err := g.ActivateCatalogAbility(me.ID, witch, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "A Card Nobody Printed")
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Library.Cards) != 0 {
		t.Errorf("%d cards left in the library, want it all exiled", len(me.Library.Cards))
	}
}

package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b546_cards_test.go — S82 OLD batch 546 (#5593, Lorwyn: card draw
// and library): Peppersmoke, Sylvan Echoes, Giant's Ire, Lammastide Weave,
// Surge of Thoughtweft, Fathom Trawl, Needle Drop, Kithkin Mourncaller,
// Colfenor's Plans, Rootgrapple and Wanderer's Twig. Each test plays the real
// catalog entry through the engine.

const (
	oldB546Peppersmoke = "071733a2-6dc6-4b6e-8ed6-a130d14657fd"
	oldB546Echoes      = "6d19e2bb-2e5f-4e55-bb86-de01ea35263d"
	oldB546Ire         = "73d175b3-f0dc-4599-8702-2967a8ead06a"
	oldB546Weave       = "74250f3e-ad88-486f-a10a-90d9a2f0097f"
	oldB546Surge       = "76aea8c6-3bab-45e9-afb4-65e5597bb2fa"
	oldB546Trawl       = "7b002cae-ee7d-400d-af9c-445eb10f1f0c"
	oldB546NeedleDrop  = "8622a2b2-fe07-41bd-bef3-e60731e0d701"
	oldB546Mourncaller = "a21e773e-1f23-42af-a1bf-b92c84b7436e"
	oldB546Plans       = "ad786230-4045-46af-bc89-078d45891e5b"
	oldB546Rootgrapple = "d6cd8c51-dd5b-4037-b6ae-1776e3743338"
	oldB546Twig        = "fd15a4b0-abd7-4bae-8331-81b85c94fce8"
	oldB546Greed       = "60040023-7bf0-42e0-ab30-46a730675564"
)

func TestOldB546EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB546Peppersmoke: "Peppersmoke", oldB546Echoes: "Sylvan Echoes", oldB546Ire: "Giant's Ire",
		oldB546Weave: "Lammastide Weave", oldB546Surge: "Surge of Thoughtweft", oldB546Trawl: "Fathom Trawl",
		oldB546NeedleDrop: "Needle Drop", oldB546Mourncaller: "Kithkin Mourncaller", oldB546Plans: "Colfenor's Plans",
		oldB546Rootgrapple: "Rootgrapple", oldB546Twig: "Wanderer's Twig",
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

func oldB546Card(target uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: target}}
}

func oldB546Player(target uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: target}}
}

// oldB546Changeling puts a changeling creature onto the battlefield.
func oldB546Changeling(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Changeling", TypeLine: "Creature — Shapeshifter", Power: 1, Toughness: 1,
		Keywords: []string{game.KeywordChangeling}, Owner: owner, Controller: owner,
	})
}

// --- Peppersmoke / Giant's Ire / Surge of Thoughtweft / Rootgrapple --------------

func TestOldB546PeppersmokeDrawsOnlyWithAFaerie(t *testing.T) {
	for _, withFaerie := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
		if withFaerie {
			effectstest.B12Creature(g, me.ID, "Faerie", "Creature — Faerie Rogue", 1, 1)
		}
		hand := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Peppersmoke", "Kindred Instant — Faerie", oldB546Peppersmoke, oldB546Card(bear))
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
			t.Errorf("faerie=%v: the Bear is %d/%d, want 2/2", withFaerie, p, tough)
		}
		want := hand
		if withFaerie {
			want++
		}
		if me.Hand.Size() != want {
			t.Errorf("faerie=%v: hand %d, want %d", withFaerie, me.Hand.Size(), want)
		}
	}
}

func TestOldB546PeppersmokeCountsAChangeling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	oldB546Changeling(g, me.ID)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Peppersmoke", "Kindred Instant — Faerie", oldB546Peppersmoke, oldB546Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a changeling is a Faerie: hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestOldB546GiantsIreDamagesAPlayerAndDrawsWithAGiant(t *testing.T) {
	for _, withGiant := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if withGiant {
			effectstest.B12Creature(g, me.ID, "Giant", "Creature — Giant", 3, 3)
		}
		life, hand := opp.Life, me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Giant's Ire", "Kindred Sorcery — Giant", oldB546Ire, oldB546Player(opp.ID))
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != life-4 {
			t.Errorf("giant=%v: life %d, want %d", withGiant, opp.Life, life-4)
		}
		want := hand
		if withGiant {
			want++
		}
		if me.Hand.Size() != want {
			t.Errorf("giant=%v: hand %d, want %d", withGiant, me.Hand.Size(), want)
		}
	}
}

func TestOldB546GiantsIreCannotTargetACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 3, 3)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Giant's Ire", "Kindred Sorcery — Giant", oldB546Ire, oldB546Card(bear)); err == nil {
		t.Fatal("Giant's Ire targeted a creature")
	}
}

func TestOldB546SurgePumpsYourCreaturesAndDrawsWithAKithkin(t *testing.T) {
	for _, withKithkin := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		if withKithkin {
			effectstest.B12Creature(g, me.ID, "Kithkin", "Creature — Kithkin Soldier", 1, 1)
		}
		hand := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Surge of Thoughtweft", "Kindred Instant — Kithkin", oldB546Surge, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, mine); p != 3 || tough != 3 {
			t.Errorf("my Bear is %d/%d, want 3/3", p, tough)
		}
		if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
			t.Errorf("their Bear is %d/%d, want 2/2", p, tough)
		}
		want := hand
		if withKithkin {
			want++
		}
		if me.Hand.Size() != want {
			t.Errorf("kithkin=%v: hand %d, want %d", withKithkin, me.Hand.Size(), want)
		}
	}
}

func TestOldB546RootgrappleDestroysANoncreaturePermanentAndDrawsWithATreefolk(t *testing.T) {
	for _, withTreefolk := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rock := effectstest.B12Creature(g, opp.ID, "Rock", "Artifact", 0, 0)
		if withTreefolk {
			effectstest.B12Creature(g, me.ID, "Treefolk", "Creature — Treefolk", 3, 3)
		}
		hand := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Rootgrapple", "Kindred Instant — Treefolk", oldB546Rootgrapple, oldB546Card(rock))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) {
			t.Error("the artifact survived")
		}
		want := hand
		if withTreefolk {
			want++
		}
		if me.Hand.Size() != want {
			t.Errorf("treefolk=%v: hand %d, want %d", withTreefolk, me.Hand.Size(), want)
		}
	}
}

func TestOldB546RootgrappleCannotTargetACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 3, 3)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Rootgrapple", "Kindred Instant — Treefolk", oldB546Rootgrapple, oldB546Card(bear)); err == nil {
		t.Fatal("Rootgrapple targeted a creature")
	}
}

// --- Needle Drop -----------------------------------------------------------------

func TestOldB546NeedleDropNeedsATargetDealtDamageThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hurt := effectstest.B12Creature(g, opp.ID, "Hurt Bear", "Creature — Bear", 2, 5)
	fresh := effectstest.B12Creature(g, opp.ID, "Fresh Bear", "Creature — Bear", 2, 5)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Needle Drop", "Instant", oldB546NeedleDrop, oldB546Card(hurt)); err == nil {
		t.Fatal("Needle Drop targeted a creature that was not dealt damage")
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Needle Drop", "Instant", oldB546NeedleDrop, oldB546Player(opp.ID)); err == nil {
		t.Fatal("Needle Drop targeted a player who was not dealt damage")
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.BoltOracleCombat, oldB546Card(hurt))
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.BoltOracleCombat, oldB546Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Needle Drop", "Instant", oldB546NeedleDrop, oldB546Card(hurt))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: Needle Drop draws", me.Hand.Size(), hand+1)
	}
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Needle Drop", "Instant", oldB546NeedleDrop, oldB546Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Needle Drop", "Instant", oldB546NeedleDrop, oldB546Card(fresh)); err == nil {
		t.Fatal("Needle Drop targeted an undamaged creature after the others were hit")
	}
}

// --- Wanderer's Twig -------------------------------------------------------------

func TestOldB546TwigSacrificesToFetchABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	twig := effectstest.PushCatalogPermanent(g, me.ID, "Wanderer's Twig", "Artifact", oldB546Twig, false)
	forest := game.NewCard("Forest", me.ID)
	forest.TypeLine = "Basic Land — Forest"
	me.Library.PushTop(forest)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, twig, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	effectstest.B06AddMana(me, "G")
	effectstest.ApaActivate(t, g, me, twig, 0, game.ActivateAbilityParams{})
	effectstest.AnswerSearchByID(t, g, me.ID, forest.InstanceID)
	if g.Battlefield.Contains(twig) {
		t.Error("the Twig was not sacrificed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

// --- Fathom Trawl ----------------------------------------------------------------

func TestOldB546TrawlTakesThreeNonlandsAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	// Top down: land, spell A, land, spell B, spell C, spell D (untouched).
	d := effectstest.TopOfLibrary(me, "Spell D", "Sorcery", "{1}")
	c := effectstest.TopOfLibrary(me, "Spell C", "Sorcery", "{1}")
	b := effectstest.TopOfLibrary(me, "Spell B", "Sorcery", "{1}")
	l2 := effectstest.TopOfLibrary(me, "Land 2", "Basic Land — Forest", "")
	a := effectstest.TopOfLibrary(me, "Spell A", "Sorcery", "{1}")
	l1 := effectstest.TopOfLibrary(me, "Land 1", "Basic Land — Forest", "")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Fathom Trawl", "Sorcery", oldB546Trawl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if order := effectstest.PutInLibraryChoiceFor(g, me.ID); order != nil {
		// The order of the bottomed lands is the player's to choose.
		if err := g.ResolvePutInLibrary(order.ID, me.ID, order.ScryCards, nil); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
	} else {
		t.Fatal("no any-order prompt for the cards put on the bottom")
	}
	inHand := func(id uuid.UUID) bool {
		for _, hc := range me.Hand.Cards {
			if hc.InstanceID == id {
				return true
			}
		}
		return false
	}
	if me.Hand.Size() != hand+3 || !inHand(a) || !inHand(b) || !inHand(c) {
		t.Errorf("hand %d, want %d with Spells A, B and C", me.Hand.Size(), hand+3)
	}
	if inHand(d) || inHand(l1) || inHand(l2) {
		t.Error("a land, or the fourth spell, reached the hand")
	}
	if me.Library.Size() != 3 {
		t.Fatalf("library has %d cards, want Spell D and the two lands", me.Library.Size())
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != d {
		t.Error("Spell D is not on top: the unrevealed card moved")
	}
}

// --- Kithkin Mourncaller ---------------------------------------------------------

func TestOldB546MourncallerDrawsForAnAttackingKithkinOrElfThatDies(t *testing.T) {
	for _, tc := range []struct {
		name      string
		typeLine  string
		keywords  []string
		attacking bool
		theirs    bool
		draws     bool
	}{
		{"attacking Kithkin", "Creature — Kithkin Soldier", nil, true, false, true},
		{"attacking Elf", "Creature — Elf Warrior", nil, true, false, true},
		{"attacking changeling", "Creature — Shapeshifter", []string{game.KeywordChangeling}, true, false, true},
		{"attacking Bear", "Creature — Bear", nil, true, false, false},
		{"Kithkin that is not attacking", "Creature — Kithkin Soldier", nil, false, false, false},
		{"opponent's attacking Kithkin", "Creature — Kithkin Soldier", nil, true, true, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Kithkin Mourncaller", "Creature — Kithkin Scout", oldB546Mourncaller, 2, 2)
		owner := me
		if tc.theirs {
			owner = opp
		}
		id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: tc.name, TypeLine: tc.typeLine, Power: 1, Toughness: 1,
			Keywords: tc.keywords, Owner: owner.ID, Controller: owner.ID,
		})
		if tc.attacking {
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == id {
						g.Battlefield.Cards[i].AttackingTarget = opp.ID
						if tc.theirs {
							g.Battlefield.Cards[i].AttackingTarget = me.ID
						}
					}
				}
			})
		}
		hand := me.Hand.Size()
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner.ID}, id); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		g.RunStateChecksForTest()
		if tc.draws {
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := hand
		if tc.draws {
			want++
		}
		if me.Hand.Size() != want {
			t.Errorf("%s: hand %d, want %d", tc.name, me.Hand.Size(), want)
		}
	}
}

// --- Colfenor's Plans ------------------------------------------------------------

func TestOldB546PlansExilesSevenFaceDownAndLetsYouCastOneSpellFromThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	var cards []uuid.UUID
	for i := 0; i < 9; i++ {
		cards = append(cards, effectstest.TopOfLibrary(me, "Exiled Spell", "Sorcery", "{1}"))
	}
	effectstest.CastCatalogSpell(t, g, "Colfenor's Plans", "Enchantment", oldB546Plans, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Size() != 7 {
		t.Fatalf("%d cards exiled, want 7", g.Exile.Size())
	}
	if me.Library.Size() != 2 {
		t.Fatalf("library has %d cards, want 2", me.Library.Size())
	}
	// Plans was this turn's first spell, so nothing more may be cast from
	// the exile until a later turn of mine, and its draw step is skipped.
	cast := func(id uuid.UUID) error {
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "exile"})
	}
	top := cards[len(cards)-1]
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := cast(top); err == nil {
		t.Fatal("cast a second spell in one turn")
	}
	library := me.Library.Size()
	for i := 0; i < 200 && g.Turn.ActiveSeat == 0; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for i := 0; i < 400 && !(g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepPrecombatMain); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepPrecombatMain {
		t.Fatalf("never reached my next main phase (seat %d, %s)", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if me.Library.Size() != library {
		t.Errorf("library %d, want %d: the draw step was not skipped", me.Library.Size(), library)
	}
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := cast(top); err != nil {
		t.Fatalf("cast from exile on a later turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := cast(cards[len(cards)-2]); err == nil {
		t.Fatal("cast a second spell in one turn")
	}
	if g.Exile.Contains(cards[0]) {
		t.Error("an eighth card (the second from the bottom) was exiled")
	}
}

// --- Sylvan Echoes ---------------------------------------------------------------

func TestOldB546EchoesDrawsWhenYouWinAClash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Sylvan Echoes", "Enchantment", oldB546Echoes, false)
	me.Library.Cards = nil
	opp.Library.Cards = nil
	effectstest.TopOfLibrary(me, "Padding", "Sorcery", "{1}")
	effectstest.TopOfLibrary(me, "Padding", "Sorcery", "{1}")
	effectstest.TopOfLibrary(me, "Padding", "Sorcery", "{1}")
	effectstest.TopOfLibrary(me, "Padding", "Sorcery", "{1}")
	effectstest.TopOfLibrary(me, "Padding", "Sorcery", "{1}")
	effectstest.TopOfLibrary(me, "Padding", "Sorcery", "{1}")
	effectstest.TopOfLibrary(me, "My One", "Sorcery", "{B}")
	effectstest.TopOfLibrary(me, "Draw Two", "Sorcery", "{1}")
	effectstest.TopOfLibrary(me, "Draw One", "Sorcery", "{1}")
	effectstest.TopOfLibrary(me, "My Five", "Sorcery", "{3}{B}{B}")
	effectstest.TopOfLibrary(me, "Draw Two", "Sorcery", "{1}")
	effectstest.TopOfLibrary(me, "Draw One", "Sorcery", "{1}")
	effectstest.TopOfLibrary(opp, "Their Two", "Sorcery", "{1}{U}")
	effectstest.TopOfLibrary(opp, "Their Two", "Sorcery", "{1}{U}")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Hoarder's Greed", "Sorcery", oldB546Greed, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// First clash: my 5 against their 2, both decide to keep on top; I win.
	oldB546Clash(t, g, me, opp, false, false)
	// The Greed repeats on a win; the second clash is my 1 against their 2: I lose.
	oldB546Clash(t, g, me, opp, false, false)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	// Greed drew four cards over its two passes; Echoes drew a fifth.
	if me.Hand.Size() != hand+5 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+5)
	}
}

func oldB546Clash(t *testing.T, g *game.Game, me, opp *game.Player, myBottom, theirBottom bool) {
	t.Helper()
	var open []*game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick {
			open = append(open, c)
		}
	}
	if len(open) != 1 {
		t.Fatalf("%d option picks open, want the opponent choice", len(open))
	}
	for i, o := range open[0].PickOptions {
		if o.Player == opp.ID {
			if err := g.ResolveOptionPick(open[0].ID, me.ID, i); err != nil {
				t.Fatalf("choose the opponent: %v", err)
			}
		}
	}
	for _, who := range []struct {
		p      *game.Player
		bottom bool
	}{{me, myBottom}, {opp, theirBottom}} {
		var pick *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == who.p.ID {
				pick = c
			}
		}
		if pick == nil {
			t.Fatalf("no clash decision for %s", who.p.Name)
		}
		idx := 0
		if who.bottom {
			idx = 1
		}
		if err := g.ResolveOptionPick(pick.ID, who.p.ID, idx); err != nil {
			t.Fatalf("decide: %v", err)
		}
	}
}

// --- Lammastide Weave ------------------------------------------------------------

func TestOldB546WeaveGainsLifeEqualToTheNamedCardsManaValueAndDraws(t *testing.T) {
	for _, guess := range []string{"Named Card", "Another Card"} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.TopOfLibrary(opp, "Named Card", "Sorcery", "{2}{G}")
		life, hand, oppHand := me.Life, me.Hand.Size(), opp.Hand.Size()
		libOpp := opp.Library.Size()
		effectstest.CastCatalogSpell(t, g, "Lammastide Weave", "Instant", oldB546Weave, oldB546Player(opp.ID))
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.ActNameFor(t, g, me.ID, guess)
		effectstest.PassPriorityAroundTable(t, g)
		wantLife := life
		if guess == "Named Card" {
			wantLife += 3
		}
		if me.Life != wantLife {
			t.Errorf("guess %q: life %d, want %d", guess, me.Life, wantLife)
		}
		if me.Hand.Size() != hand+1 {
			t.Errorf("guess %q: hand %d, want %d", guess, me.Hand.Size(), hand+1)
		}
		if opp.Library.Size() != libOpp-1 || opp.Hand.Size() != oppHand {
			t.Errorf("guess %q: the opponent did not mill exactly one card", guess)
		}
		if opp.Graveyard.Size() != 1 {
			t.Errorf("guess %q: %d cards in their graveyard, want 1", guess, opp.Graveyard.Size())
		}
	}
}

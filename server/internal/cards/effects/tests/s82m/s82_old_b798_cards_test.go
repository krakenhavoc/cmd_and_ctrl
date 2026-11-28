package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 798 (Mirage, #5633): Fallow Earth, Early Harvest, Painful
// Memories, Soulshriek, Flash, Lure of Prey, Sealed Fate and Natural Balance.

const (
	oldB798Fallow    = "05a4acfd-3bf3-4ec2-a28a-bc43ba7c6fbf"
	oldB798Sealed    = "0b9b5856-9396-4a55-8675-58cd171df708"
	oldB798Painful   = "27a3faa1-4f1a-4581-9e19-9265faa9c10b"
	oldB798Harvest   = "32001382-3079-4bbe-8d86-05da654c0d1a"
	oldB798Lure      = "3c83ea11-7141-41a2-ba77-1ed2722dfe74"
	oldB798Flash     = "3e057471-e415-4105-9064-8236a0a5047d"
	oldB798Balance   = "69acaa6f-3ba2-420c-9069-02958f844121"
	oldB798Soulshrk  = "77ae796e-fc56-47b9-b8a0-97e24fbd60f7"
	oldB798ForestTyp = "Basic Land — Forest"
)

func TestOldB798EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB798Fallow: "Fallow Earth", oldB798Sealed: "Sealed Fate", oldB798Painful: "Painful Memories",
		oldB798Harvest: "Early Harvest", oldB798Lure: "Lure of Prey", oldB798Flash: "Flash",
		oldB798Balance: "Natural Balance", oldB798Soulshrk: "Soulshriek",
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

func oldB798Main(t *testing.T, g *game.Game) *game.Player {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	return g.Seats[g.Turn.ActiveSeat]
}

func oldB798Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := effectstest.HandCardOf(g, me, name, typeLine, cost, oracle)
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func oldB798Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Tapped = true })
}

func TestOldB798FallowEarthPutsALandOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB798Main(t, g)
	opp := g.Seats[1]
	// A land I control but the opponent owns goes to the opponent's library.
	land := effectstest.AcPermanent(g, opp.ID, "Island", "Basic Land — Island", 0, 0)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, land).Controller = me.ID })
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Fallow Earth", "Sorcery", oldB798Fallow, effectstest.LtCardTarget(bear)); err == nil {
		t.Fatal("targeted a creature that is not a land")
	}
	effectstest.CastCatalogSpell(t, g, "Fallow Earth", "Sorcery", oldB798Fallow, effectstest.LtCardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Fatal("the land is still on the battlefield")
	}
	if top := effectstest.LibraryTopIDs(opp, 1); len(top) != 1 || top[0] != land {
		t.Errorf("the land is not on top of its owner's library: %v", top)
	}
}

func TestOldB798EarlyHarvestUntapsOnlyTheTargetsBasicLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB798Main(t, g)
	opp := g.Seats[1]
	forest := effectstest.AcPermanent(g, opp.ID, "Forest", oldB798ForestTyp, 0, 0)
	island := effectstest.AcPermanent(g, opp.ID, "Island", "Basic Land — Island", 0, 0)
	nonbasic := effectstest.AcPermanent(g, opp.ID, "Dryad Arbor", "Land Creature — Forest Dryad", 1, 1)
	mine := effectstest.AcPermanent(g, me.ID, "Forest", oldB798ForestTyp, 0, 0)
	for _, id := range []uuid.UUID{forest, island, nonbasic, mine} {
		oldB798Tap(g, id)
	}
	effectstest.CastCatalogSpell(t, g, "Early Harvest", "Instant", oldB798Harvest, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PermanentTapped(g, forest) || effectstest.PermanentTapped(g, island) {
		t.Error("a basic land of the target stayed tapped")
	}
	if !effectstest.PermanentTapped(g, nonbasic) {
		t.Error("a nonbasic land was untapped")
	}
	if !effectstest.PermanentTapped(g, mine) {
		t.Error("the caster's own land was untapped")
	}
}

func TestOldB798PainfulMemoriesPutsTheChosenCardOnTopOfTheirLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB798Main(t, g)
	opp := g.Seats[1]
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	keep := effectstest.HandCardOf(g, opp, "Keeper", "Creature — Bear", "{G}", "")
	take := effectstest.HandCardOf(g, opp, "Taken", "Sorcery", "{2}{R}", "")
	effectstest.CastCatalogSpell(t, g, "Painful Memories", "Sorcery", oldB798Painful, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(c.ID, me.ID, []uuid.UUID{take}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(take) || !opp.Hand.Contains(keep) {
		t.Errorf("hand after the pick: taken %v keeper %v", opp.Hand.Contains(take), opp.Hand.Contains(keep))
	}
	if top := effectstest.LibraryTopIDs(opp, 1); len(top) != 1 || top[0] != take {
		t.Errorf("the chosen card is not on top of their library: %v", top)
	}
}

func TestOldB798PainfulMemoriesDoesNothingToAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB798Main(t, g)
	opp := g.Seats[1]
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	before := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Painful Memories", "Sorcery", oldB798Painful, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.NoVariant(t, g)
	if opp.Library.Size() != before {
		t.Errorf("their library changed size %d → %d", before, opp.Library.Size())
	}
}

func TestOldB798SoulshriekBoostsByCreatureCardsThenSacrificesAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB798Main(t, g)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead One", TypeLine: "Creature — Elf", Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Two", TypeLine: "Creature — Elf", Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Dead", TypeLine: "Creature — Elf", Owner: opp.ID, Controller: opp.ID})
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Soulshriek", "Instant", oldB798Soulshrk, effectstest.LtCardTarget(theirs)); err == nil {
		t.Fatal("targeted a creature the caster does not control")
	}
	effectstest.CastCatalogSpell(t, g, "Soulshriek", "Instant", oldB798Soulshrk, effectstest.LtCardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	// X counts my graveyard's creature cards only (the Soulshriek in it is not one).
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4 (2 + two creature cards)", p)
	}
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature was sacrificed before the end step")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not sacrificed at the beginning of the end step")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("their creature was sacrificed")
	}
}

func TestOldB798FlashPaysItsReducedCostOrSacrifices(t *testing.T) {
	for _, tc := range []struct {
		name      string
		cost      string
		pay       string
		wantAsked bool
		answerPay bool
		wantStays bool
	}{
		{"paid", "{2}{G}{G}", "{G}{G}", true, true, true},
		{"declined", "{2}{G}{G}", "", true, false, false},
		{"free after the reduction", "{2}", "", false, false, true},
	} {
		g := effectstest.NewCatalogGame(t)
		me := oldB798Main(t, g)
		bear := effectstest.HandCardOf(g, me, "Test Beast", "Creature — Beast", tc.cost, "")
		oldB798Cast(t, g, me, "Flash", "Instant", "{1}{U}", oldB798Flash, game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(bear) {
			t.Fatalf("%s: the creature did not enter", tc.name)
		}
		asked := effectstest.HasPayUnlessFor(g, me.ID)
		if asked != tc.wantAsked {
			t.Fatalf("%s: asked %v, want %v", tc.name, asked, tc.wantAsked)
		}
		if asked {
			if tc.pay != "" {
				effectstest.DcMana(t, g, me.ID, tc.pay)
			}
			effectstest.AnswerPayUnless(t, g, me.ID, tc.answerPay)
			effectstest.PassPriorityAroundTable(t, g)
		}
		if g.Battlefield.Contains(bear) != tc.wantStays {
			t.Errorf("%s: creature on the battlefield %v, want %v", tc.name, g.Battlefield.Contains(bear), tc.wantStays)
		}
	}
}

func TestOldB798FlashLetsYouDeclineToPutAnything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB798Main(t, g)
	bear := effectstest.HandCardOf(g, me, "Test Beast", "Creature — Beast", "{2}{G}{G}", "")
	oldB798Cast(t, g, me, "Flash", "Instant", "{1}{U}", oldB798Flash, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("declining still put the creature onto the battlefield or asked for payment")
	}
}

func TestOldB798ReducedManaCostTakesTwoGenericOff(t *testing.T) {
	for in, want := range map[string]string{
		"{2}{G}{G}": "{G}{G}", "{4}{R}": "{2}{R}", "{1}{G}": "{G}", "{2}": "", "{1}": "", "{G}": "{G}", "{X}{G}": "{G}", "{3}": "{1}", "": "",
		"{1}{W/U}{W/U}": "{W/U}{W/U}",
	} {
		if got := effects.OldB798ReducedManaCost(in, 2); got != want {
			t.Errorf("reduce %q by 2 = %q, want %q", in, got, want)
		}
	}
}

func TestOldB798LureOfPreyIsCastOnlyAfterAnOpponentCastsACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB798Main(t, g)
	opp := g.Seats[1]
	green := effectstest.HandCardOf(g, me, "Green Beast", "Creature — Beast", "{3}{G}", "")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == green {
				me.Hand.Cards[i].Colors = []string{"G"}
			}
		}
	})
	red := effectstest.HandCardOf(g, me, "Red Beast", "Creature — Beast", "{3}{R}", "")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == red {
				me.Hand.Cards[i].Colors = []string{"R"}
			}
		}
	})
	lure := effectstest.HandCardOf(g, me, "Lure of Prey", "Instant", "{2}{G}{G}", oldB798Lure)
	if err := g.CastSpell(me.ID, lure, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no opponent having cast a creature spell")
	}
	// An opponent's noncreature spell is not enough.
	g.WithWriteLock(func() { g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{opp.ID: {Total: 1, Noncreature: 1}} })
	if err := g.CastSpell(me.ID, lure, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast after only a noncreature spell")
	}
	// My own creature spell does not count.
	g.WithWriteLock(func() { g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{me.ID: {Total: 1}} })
	if err := g.CastSpell(me.ID, lure, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast after only my own creature spell")
	}
	g.WithWriteLock(func() { g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{opp.ID: {Total: 2, Noncreature: 1}} })
	effectstest.DcMana(t, g, me.ID, "{G}{G}{G}{G}")
	if err := g.CastSpell(me.ID, lure, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast after an opponent's creature spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt for the green creature card")
	}
	for _, id := range c.ChooseCards {
		if id == red {
			t.Error("a red creature card was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, green)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(green) {
		t.Error("the green creature did not enter")
	}
}

func TestOldB798SealedFateExilesOneAndOrdersTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB798Main(t, g)
	opp := g.Seats[1]
	g.WithWriteLock(func() { opp.Library.Cards = nil })
	var ids []uuid.UUID
	for _, n := range []string{"One", "Two", "Three", "Four", "Five"} {
		ids = append(ids, effectstest.PushLibraryCardForTest(opp, game.Card{Name: n, TypeLine: "Sorcery"}))
	}
	oldB798Cast(t, g, me, "Sealed Fate", "Sorcery", "{X}{U}{B}", oldB798Sealed,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt to exile one of the looked-at cards")
	}
	for _, id := range c.ChooseCards {
		if id == ids[3] || id == ids[4] {
			t.Error("a card below the top three was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no prompt to order the rest")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{ids[2], ids[0]}); err != nil {
		t.Fatalf("order: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, ids[1]) {
		t.Error("the chosen card is not in exile")
	}
	want := []uuid.UUID{ids[2], ids[0], ids[3], ids[4]}
	if got := effectstest.LibraryTopIDs(opp, 4); !effectstest.SameIDs(got, want) {
		t.Errorf("their library from the top is %v, want %v", got, want)
	}
}

func TestOldB798SealedFateWithXZeroDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB798Main(t, g)
	opp := g.Seats[1]
	before := opp.Library.Size()
	oldB798Cast(t, g, me, "Sealed Fate", "Sorcery", "{X}{U}{B}", oldB798Sealed,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, XValue: 0})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != before || effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("an X of 0 looked at or exiled something")
	}
}

func oldB798Lands(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		out = append(out, effectstest.AcPermanent(g, owner, "Forest", oldB798ForestTyp, 0, 0))
	}
	return out
}

func TestOldB798NaturalBalanceTrimsToFiveAndFetchesUpToFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB798Main(t, g)
	s1, s2, s3 := g.Seats[1], g.Seats[2], g.Seats[3]
	mine := oldB798Lands(g, me.ID, 7)
	short := oldB798Lands(g, s1.ID, 2)
	five := oldB798Lands(g, s2.ID, 5)
	big := oldB798Lands(g, s3.ID, 6)
	var basics []uuid.UUID
	for i := 0; i < 4; i++ {
		basics = append(basics, effectstest.PushLibraryCardForTest(s1, game.Card{Name: "Forest", TypeLine: oldB798ForestTyp}))
	}
	oldB798Cast(t, g, me, "Natural Balance", "Sorcery", "{2}{G}{G}", oldB798Balance, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	// Each player with six or more lands is asked once, in turn order.
	if p := effectstest.ChooseCardsChoiceFor(g, s2.ID); p != nil {
		t.Fatal("a player with five lands was asked to choose")
	}
	if err := g.ResolveChooseCards(effectstest.ChooseCardsChoiceFor(g, me.ID).ID, me.ID, mine[:4]); err == nil {
		t.Fatal("kept four lands, not five")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, mine[:5]...)
	effectstest.AnswerChooseCards(t, g, s3.ID, big[1:6]...)
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range mine {
		if g.Battlefield.Contains(id) != (i < 5) {
			t.Errorf("my land %d on the battlefield: %v", i, g.Battlefield.Contains(id))
		}
	}
	if g.Battlefield.Contains(big[0]) {
		t.Error("the sixth land of the last player was kept")
	}
	for _, id := range five {
		if !g.Battlefield.Contains(id) {
			t.Error("a land of the five-land player was sacrificed")
		}
	}
	// The two-land player may fetch up to three; take two.
	effectstest.AnswerSearchByID(t, g, s1.ID, basics[0], basics[1])
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, id := range append(short, basics...) {
		if g.Battlefield.Contains(id) {
			n++
		}
	}
	if n != 4 {
		t.Errorf("the short player controls %d of their lands, want 4 (two plus two fetched)", n)
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil || effectstest.SearchChoiceFor(g, s3.ID) != nil {
		t.Error("a player who had five lands was offered a search")
	}
}

func TestOldB798NaturalBalanceSearchesWhenNobodyHasToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB798Main(t, g)
	oldB798Lands(g, me.ID, 4)
	for _, s := range g.Seats[1:] {
		oldB798Lands(g, s.ID, 5)
	}
	basic := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: oldB798ForestTyp})
	oldB798Cast(t, g, me, "Natural Balance", "Sorcery", "{2}{G}{G}", oldB798Balance, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, basic)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(basic) {
		t.Error("the fetched land did not enter the battlefield")
	}
}

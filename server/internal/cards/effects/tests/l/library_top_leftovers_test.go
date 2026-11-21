package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Jace, the Mind Sculptor --------------------------------------

func TestJaceTheMindSculptorPlusTwoLooksAtAnotherPlayersTop(t *testing.T) {
	for _, bury := range []bool{true, false} {
		name := "leave it on top"
		if bury {
			name = "bury it"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.ToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
			lib := effectstest.SeedLibrary(opp, "Their Top", "Their Next")
			jace := effectstest.PushCatalogWalker(g, me.ID, "Jace, the Mind Sculptor", effectstest.JaceTheMindSculptorOracle, 3)

			effectstest.B16Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{
				Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
			})
			if got := effectstest.LoyaltyCount(g, jace); got != 5 {
				t.Errorf("loyalty %d after +2 on 3", got)
			}
			ask := effectstest.PutInLibraryChoiceFor(g, me.ID)
			if ask == nil {
				t.Fatal("the +2 asks Jace's controller top or bottom")
			}
			if ask.LibraryPlacement != game.LibraryPlaceTopOrBottom || len(ask.ScryCards) != 1 || ask.ScryCards[0] != lib[0] {
				t.Fatalf("prompt %+v, want top-or-bottom over the TARGET's top card", ask)
			}
			if !effectstest.KnownTo(g, lib[0], me.ID) {
				t.Error("Jace's controller looked at the card")
			}
			if effectstest.KnownTo(g, lib[0], opp.ID) || effectstest.KnownTo(g, lib[0], other.ID) {
				t.Error("a look is not a reveal: the owner and the table learn nothing")
			}
			var top, bottom []uuid.UUID
			if bury {
				bottom = []uuid.UUID{lib[0]}
			} else {
				top = []uuid.UUID{lib[0]}
			}
			if err := g.ResolvePutInLibrary(ask.ID, me.ID, bottom, top); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}
			if bury {
				if opp.Library.Cards[0].InstanceID != lib[0] {
					t.Error("buried on the bottom of THAT player's library")
				}
				if got := effectstest.LibraryTopIDs(opp, 1); got[0] != lib[1] {
					t.Error("the next card is now on top")
				}
			} else if got := effectstest.LibraryTopIDs(opp, 1); got[0] != lib[0] {
				t.Error("left on top")
			}
			if me.Library.Contains(lib[0]) {
				t.Error("the card never leaves its owner's library")
			}
			if effectstest.KnownTo(g, lib[0], opp.ID) {
				t.Error("still not known to its owner after the move")
			}
		})
	}
}

func TestJaceTheMindSculptorZeroIsBrainstorm(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "D1", "D2", "D3")
	for _, n := range []string{"H1", "H2"} {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	before := me.Hand.Size()
	jace := effectstest.PushCatalogWalker(g, me.ID, "Jace, the Mind Sculptor", effectstest.JaceTheMindSculptorOracle, 3)
	effectstest.B16Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{})
	if me.Hand.Size() != before+3 {
		t.Fatalf("drew three: hand %d → %d", before, me.Hand.Size())
	}
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("then pick exactly two to put back: %+v", pick)
	}
	back := []uuid.UUID{pick.ChooseCards[0], pick.ChooseCards[1]}
	if err := g.ResolveChooseCards(pick.ID, me.ID, back); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceTop {
		t.Fatalf("then order them on top: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{back[1], back[0]}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := effectstest.LibraryTopIDs(me, 2); !effectstest.SameIDs(got, []uuid.UUID{back[1], back[0]}) {
		t.Errorf("top two %v, want the order chosen", got)
	}
	if effectstest.LoyaltyCount(g, jace) != 3 {
		t.Error("the 0 costs nothing")
	}
}

func TestJaceTheMindSculptorMinusOneBounces(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.SeedCreature(g, "Their Bear", opp.ID)
	jace := effectstest.PushCatalogWalker(g, me.ID, "Jace, the Mind Sculptor", effectstest.JaceTheMindSculptorOracle, 3)
	effectstest.B16Activate(t, g, me.ID, jace, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if !opp.Hand.Contains(bear) {
		t.Error("returned to its owner's hand")
	}
	if effectstest.LoyaltyCount(g, jace) != 2 {
		t.Error("the −1 costs one loyalty")
	}
}

func TestJaceTheMindSculptorUltimateExilesTheLibraryAndShufflesTheHandIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.SeedLibrary(opp, "L1", "L2", "L3")
	libraryBefore := make([]uuid.UUID, 0, opp.Library.Size())
	for _, c := range opp.Library.Cards {
		libraryBefore = append(libraryBefore, c.InstanceID)
	}
	h1, h2 := uuid.New(), uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: h1, Name: "Hand 1", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	opp.Hand.PushTop(game.Card{InstanceID: h2, Name: "Hand 2", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	handBefore := opp.Hand.Size()
	jace := effectstest.PushCatalogWalker(g, me.ID, "Jace, the Mind Sculptor", effectstest.JaceTheMindSculptorOracle, 12)
	effectstest.B16Activate(t, g, me.ID, jace, 3, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	for _, id := range libraryBefore {
		if !g.Exile.Contains(id) {
			t.Fatal("every card of the library is exiled")
		}
	}
	if opp.Hand.Size() != 0 {
		t.Errorf("the hand is shuffled in: %d cards left in hand", opp.Hand.Size())
	}
	if opp.Library.Size() != handBefore || !opp.Library.Contains(h1) || !opp.Library.Contains(h2) {
		t.Errorf("the library is exactly the old hand: %d cards", opp.Library.Size())
	}
}

// --- Portent: an ordered look at another player's library ----------

func TestPortentOrdersAnotherPlayersTopAndDrawsNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lib := effectstest.SeedLibrary(opp, "A", "B", "C", "D")
	effectstest.CastCatalogSpell(t, g, "Portent", "Sorcery", effectstest.PortentOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	ask := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if ask == nil || ask.LibraryPlacement != game.LibraryPlaceTop || !effectstest.SameIDs(ask.ScryCards, lib[:3]) {
		t.Fatalf("a top placement over the target's top three: %+v", ask)
	}
	for _, id := range lib[:3] {
		if !effectstest.KnownTo(g, id, me.ID) || effectstest.KnownTo(g, id, opp.ID) {
			t.Fatal("the caster looks; the library's owner does not")
		}
	}
	want := []uuid.UUID{lib[2], lib[0], lib[1]}
	if err := g.ResolvePutInLibrary(ask.ID, me.ID, nil, want); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := effectstest.LibraryTopIDs(opp, 4); !effectstest.SameIDs(got, append(append([]uuid.UUID(nil), want...), lib[3])) {
		t.Errorf("opponent's top four %v, want [C A B D]", got)
	}
	shuffle := effectstest.LatestConfirmFor(g, me.ID)
	if shuffle == nil {
		t.Fatal("\"you may have that player shuffle\" is asked after the order")
	}
	if err := g.ResolveConfirm(shuffle.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if got := effectstest.LibraryTopIDs(opp, 3); !effectstest.SameIDs(got, want) {
		t.Error("declining leaves the order as chosen")
	}
	for _, id := range want {
		if effectstest.KnownTo(g, id, opp.ID) {
			t.Error("CR 401.4: the owner does not learn the order")
		}
	}

	before := me.Hand.Size()
	next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	effectstest.AdvanceToUpkeepOf(t, g, next)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("Portent draws a card at the next turn's upkeep: hand %d → %d", before, me.Hand.Size())
	}
}

// --- Temporal Cleansing: the OWNER chooses second from the top or the bottom

func TestTemporalCleansingOwnerChoosesSecondFromTheTopOrTheBottom(t *testing.T) {
	for _, bury := range []bool{false, true} {
		name := "second from the top"
		if bury {
			name = "on the bottom"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			lib := effectstest.SeedLibrary(opp, "Top", "Next")
			perm := effectstest.SeedCreature(g, "Their Creature", opp.ID)
			effectstest.CastCatalogSpell(t, g, "Temporal Cleansing", "Sorcery", effectstest.TemporalCleansingOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: perm}})
			effectstest.PassPriorityAroundTable(t, g)

			if effectstest.PutInLibraryChoiceFor(g, me.ID) != nil {
				t.Fatal("the caster does not choose")
			}
			ask := effectstest.PutInLibraryChoiceFor(g, opp.ID)
			if ask == nil || ask.LibraryPlacement != game.LibraryPlaceTopOrBottom || ask.LibraryTopDepth != 2 {
				t.Fatalf("the OWNER is asked, with the top lane at depth 2: %+v", ask)
			}
			var top, bottom []uuid.UUID
			if bury {
				bottom = []uuid.UUID{perm}
			} else {
				top = []uuid.UUID{perm}
			}
			if err := g.ResolvePutInLibrary(ask.ID, opp.ID, bottom, top); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}
			if effects.OnBattlefield(g, perm) {
				t.Fatal("the permanent left the battlefield")
			}
			if bury {
				if opp.Library.Cards[0].InstanceID != perm {
					t.Error("on the bottom of its owner's library")
				}
			} else if got := effectstest.LibraryTopIDs(opp, 3); !effectstest.SameIDs(got, []uuid.UUID{lib[0], perm, lib[1]}) {
				t.Errorf("top three %v, want [Top Their-Creature Next]", got)
			}
		})
	}
}

// --- Cream of the Crop: exactly one on top ---------------------------

func TestCreamOfTheCropPutsExactlyOneOnTopAndTheRestUnder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := effectstest.SeedLibrary(me, "A", "B", "C", "Fourth")
	effectstest.PushCatalogPermanent(g, me.ID, "Cream of the Crop", "Enchantment", effectstest.CreamOfTheCropOracle, false)

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Beast", TypeLine: "Token Creature — Beast", Power: 3, Toughness: 3}, 1)
	})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	ask := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if ask == nil || ask.LibraryTopCount != 1 || !effectstest.SameIDs(ask.ScryCards, lib[:3]) {
		t.Fatalf("X = 3: the top three, exactly one to stay on top: %+v", ask)
	}
	if err := g.ResolvePutInLibrary(ask.ID, me.ID, []uuid.UUID{lib[2]}, []uuid.UUID{lib[0], lib[1]}); err == nil {
		t.Fatal("two on top is not \"one of those cards\"")
	}
	if err := g.ResolvePutInLibrary(ask.ID, me.ID, []uuid.UUID{lib[2], lib[0]}, []uuid.UUID{lib[1]}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := effectstest.LibraryTopIDs(me, 2); !effectstest.SameIDs(got, []uuid.UUID{lib[1], lib[3]}) {
		t.Errorf("top two %v, want [B Fourth]", got)
	}
	if got := effectstest.LibraryBottomIDs(me, 2); !effectstest.SameIDs(got, []uuid.UUID{lib[2], lib[0]}) {
		t.Errorf("bottom two %v, want [C A] in the order chosen", got)
	}
}

func TestCreamOfTheCropPowerOneHasNothingToChoose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := effectstest.SeedLibrary(me, "A", "B")
	effectstest.PushCatalogPermanent(g, me.ID, "Cream of the Crop", "Enchantment", effectstest.CreamOfTheCropOracle, false)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Squirrel", TypeLine: "Token Creature — Squirrel", Power: 1, Toughness: 1}, 1)
	})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PutInLibraryChoiceFor(g, me.ID) != nil {
		t.Error("one card, one on top: no prompt")
	}
	if got := effectstest.LibraryTopIDs(me, 1); got[0] != lib[0] {
		t.Error("the one card stays on top")
	}
}

func TestHinderPutsTheSpellOnTheCounterersChoiceOfEnd(t *testing.T) {
	for _, bury := range []bool{false, true} {
		name := "top"
		if bury {
			name = "bottom"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.SeedLibrary(opp, "Their Top")
			lifeBefore := me.Life
			shock := effectstest.CastShockAt(t, g, opp, me.ID)
			hinder := effectstest.CastCounterAt(t, g, me, "Hinder", effectstest.HinderOracle, shock)
			effectstest.PassPriorityAroundTable(t, g)

			if effectstest.PutInLibraryChoiceFor(g, opp.ID) != nil {
				t.Fatal("the choice is Hinder's controller's, not the spell's owner's")
			}
			ask := effectstest.PutInLibraryChoiceFor(g, me.ID)
			if ask == nil || ask.LibraryPlacement != game.LibraryPlaceTopOrBottom || !effectstest.SameIDs(ask.ScryCards, []uuid.UUID{shock}) {
				t.Fatalf("top or bottom over the countered spell: %+v", ask)
			}
			if !g.Stack.Contains(shock) {
				t.Fatal("the spell waits on the stack for the choice")
			}
			var top, bottom []uuid.UUID
			if bury {
				bottom = []uuid.UUID{shock}
			} else {
				top = []uuid.UUID{shock}
			}
			if err := g.ResolvePutInLibrary(ask.ID, me.ID, bottom, top); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if opp.Graveyard.Contains(shock) || g.Stack.Contains(shock) {
				t.Fatal("countered into the library instead of the graveyard")
			}
			if bury {
				if opp.Library.Cards[0].InstanceID != shock {
					t.Error("on the bottom of its OWNER's library")
				}
			} else if got := effectstest.LibraryTopIDs(opp, 1); got[0] != shock {
				t.Error("on top of its OWNER's library")
			}
			if me.Life != lifeBefore {
				t.Error("a countered Shock deals no damage")
			}
			if effectstest.CountEvents(g, game.EventCounterSpell) == 0 {
				t.Error("Hinder counters (CR 701.6a)")
			}
			if !me.Graveyard.Contains(hinder) {
				t.Error("Hinder itself goes to the graveyard")
			}
			if !effectstest.KnownTo(g, shock, me.ID) || !effectstest.KnownTo(g, shock, opp.ID) {
				t.Error("the whole table saw where the spell went")
			}
		})
	}
}

func TestSpellCrumpleBottomsTheSpellAndItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "My Top")
	effectstest.SeedLibrary(opp, "Their Top")
	shock := effectstest.CastShockAt(t, g, opp, me.ID)
	crumple := effectstest.CastCounterAt(t, g, me, "Spell Crumple", effectstest.SpellCrumpleOracle, shock)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.PutInLibraryChoiceFor(g, me.ID) != nil {
		t.Error("the bottom is not a choice: no prompt")
	}
	if opp.Library.Cards[0].InstanceID != shock {
		t.Error("the countered spell goes to the bottom of its owner's library")
	}
	if me.Library.Cards[0].InstanceID != crumple {
		t.Error("Spell Crumple goes to the bottom of its owner's library")
	}
	if me.Graveyard.Contains(crumple) || opp.Graveyard.Contains(shock) {
		t.Error("neither card reaches a graveyard")
	}
	if effectstest.CountEvents(g, game.EventCounterSpell) == 0 {
		t.Error("Spell Crumple counters (CR 701.6a)")
	}
}

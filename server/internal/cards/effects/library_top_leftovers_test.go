package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_top_leftovers_test.go — the proof cards for #1298, ADR 0088's
// 2026-09-23 amendment: Jace, the Mind Sculptor and Portent (a look at
// ANOTHER player's library), Temporal Cleansing (a top lane at a depth,
// chosen by the owner), Cream of the Crop (an exact top count), Hinder
// and Spell Crumple (a counter to a position, chosen by the counterer).

const (
	jaceTheMindSculptorOracle = "7f77a84e-5a4b-4834-aefa-3cecc175ae8e"
	portentOracle             = "e744dbb6-2d56-462a-8d49-d79c73944048"
	temporalCleansingOracle   = "5f67698a-28db-43cd-aaf1-52371b0b47eb"
	creamOfTheCropOracle      = "b61a87e3-dc98-4db9-abed-b47b677d81ab"
	hinderOracle              = "c9db6b94-a7b1-4b93-b454-4dead8f85e34"
	spellCrumpleOracle        = "d4ab7848-5c37-4c6b-be29-0bb703333e5b"
)

// libraryTopIDs is the top n cards of a library, top-first.
func libraryTopIDs(p *game.Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, p.Library.Cards[p.Library.Size()-1-i].InstanceID)
	}
	return out
}

func knownTo(g *game.Game, id, viewer uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && c.IsKnownTo(viewer)
}

// --- Jace, the Mind Sculptor --------------------------------------

func TestJaceTheMindSculptorPlusTwoLooksAtAnotherPlayersTop(t *testing.T) {
	for _, bury := range []bool{true, false} {
		name := "leave it on top"
		if bury {
			name = "bury it"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			toMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
			lib := seedLibrary(opp, "Their Top", "Their Next")
			jace := pushCatalogWalker(g, me.ID, "Jace, the Mind Sculptor", jaceTheMindSculptorOracle, 3)

			b16Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{
				Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
			})
			if got := loyaltyCount(g, jace); got != 5 {
				t.Errorf("loyalty %d after +2 on 3", got)
			}
			ask := putInLibraryChoiceFor(g, me.ID)
			if ask == nil {
				t.Fatal("the +2 asks Jace's controller top or bottom")
			}
			if ask.LibraryPlacement != game.LibraryPlaceTopOrBottom || len(ask.ScryCards) != 1 || ask.ScryCards[0] != lib[0] {
				t.Fatalf("prompt %+v, want top-or-bottom over the TARGET's top card", ask)
			}
			if !knownTo(g, lib[0], me.ID) {
				t.Error("Jace's controller looked at the card")
			}
			if knownTo(g, lib[0], opp.ID) || knownTo(g, lib[0], other.ID) {
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
				if got := libraryTopIDs(opp, 1); got[0] != lib[1] {
					t.Error("the next card is now on top")
				}
			} else if got := libraryTopIDs(opp, 1); got[0] != lib[0] {
				t.Error("left on top")
			}
			if me.Library.Contains(lib[0]) {
				t.Error("the card never leaves its owner's library")
			}
			if knownTo(g, lib[0], opp.ID) {
				t.Error("still not known to its owner after the move")
			}
		})
	}
}

func TestJaceTheMindSculptorZeroIsBrainstorm(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "D1", "D2", "D3")
	for _, n := range []string{"H1", "H2"} {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	before := me.Hand.Size()
	jace := pushCatalogWalker(g, me.ID, "Jace, the Mind Sculptor", jaceTheMindSculptorOracle, 3)
	b16Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{})
	if me.Hand.Size() != before+3 {
		t.Fatalf("drew three: hand %d → %d", before, me.Hand.Size())
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("then pick exactly two to put back: %+v", pick)
	}
	back := []uuid.UUID{pick.ChooseCards[0], pick.ChooseCards[1]}
	if err := g.ResolveChooseCards(pick.ID, me.ID, back); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceTop {
		t.Fatalf("then order them on top: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{back[1], back[0]}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := libraryTopIDs(me, 2); !sameIDs(got, []uuid.UUID{back[1], back[0]}) {
		t.Errorf("top two %v, want the order chosen", got)
	}
	if loyaltyCount(g, jace) != 3 {
		t.Error("the 0 costs nothing")
	}
}

func TestJaceTheMindSculptorMinusOneBounces(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := seedCreature(g, "Their Bear", opp.ID)
	jace := pushCatalogWalker(g, me.ID, "Jace, the Mind Sculptor", jaceTheMindSculptorOracle, 3)
	b16Activate(t, g, me.ID, jace, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if !opp.Hand.Contains(bear) {
		t.Error("returned to its owner's hand")
	}
	if loyaltyCount(g, jace) != 2 {
		t.Error("the −1 costs one loyalty")
	}
}

func TestJaceTheMindSculptorUltimateExilesTheLibraryAndShufflesTheHandIn(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seedLibrary(opp, "L1", "L2", "L3")
	libraryBefore := make([]uuid.UUID, 0, opp.Library.Size())
	for _, c := range opp.Library.Cards {
		libraryBefore = append(libraryBefore, c.InstanceID)
	}
	h1, h2 := uuid.New(), uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: h1, Name: "Hand 1", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	opp.Hand.PushTop(game.Card{InstanceID: h2, Name: "Hand 2", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	handBefore := opp.Hand.Size()
	jace := pushCatalogWalker(g, me.ID, "Jace, the Mind Sculptor", jaceTheMindSculptorOracle, 12)
	b16Activate(t, g, me.ID, jace, 3, game.ActivateAbilityParams{
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lib := seedLibrary(opp, "A", "B", "C", "D")
	castCatalogSpell(t, g, "Portent", "Sorcery", portentOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	ask := putInLibraryChoiceFor(g, me.ID)
	if ask == nil || ask.LibraryPlacement != game.LibraryPlaceTop || !sameIDs(ask.ScryCards, lib[:3]) {
		t.Fatalf("a top placement over the target's top three: %+v", ask)
	}
	for _, id := range lib[:3] {
		if !knownTo(g, id, me.ID) || knownTo(g, id, opp.ID) {
			t.Fatal("the caster looks; the library's owner does not")
		}
	}
	want := []uuid.UUID{lib[2], lib[0], lib[1]}
	if err := g.ResolvePutInLibrary(ask.ID, me.ID, nil, want); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := libraryTopIDs(opp, 4); !sameIDs(got, append(append([]uuid.UUID(nil), want...), lib[3])) {
		t.Errorf("opponent's top four %v, want [C A B D]", got)
	}
	shuffle := latestConfirmFor(g, me.ID)
	if shuffle == nil {
		t.Fatal("\"you may have that player shuffle\" is asked after the order")
	}
	if err := g.ResolveConfirm(shuffle.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if got := libraryTopIDs(opp, 3); !sameIDs(got, want) {
		t.Error("declining leaves the order as chosen")
	}
	for _, id := range want {
		if knownTo(g, id, opp.ID) {
			t.Error("CR 401.4: the owner does not learn the order")
		}
	}

	before := me.Hand.Size()
	next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	advanceToUpkeepOf(t, g, next)
	passPriorityAroundTable(t, g)
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
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			lib := seedLibrary(opp, "Top", "Next")
			perm := seedCreature(g, "Their Creature", opp.ID)
			castCatalogSpell(t, g, "Temporal Cleansing", "Sorcery", temporalCleansingOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: perm}})
			passPriorityAroundTable(t, g)

			if putInLibraryChoiceFor(g, me.ID) != nil {
				t.Fatal("the caster does not choose")
			}
			ask := putInLibraryChoiceFor(g, opp.ID)
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
			if onBattlefield(g, perm) {
				t.Fatal("the permanent left the battlefield")
			}
			if bury {
				if opp.Library.Cards[0].InstanceID != perm {
					t.Error("on the bottom of its owner's library")
				}
			} else if got := libraryTopIDs(opp, 3); !sameIDs(got, []uuid.UUID{lib[0], perm, lib[1]}) {
				t.Errorf("top three %v, want [Top Their-Creature Next]", got)
			}
		})
	}
}

// --- Cream of the Crop: exactly one on top ---------------------------

func TestCreamOfTheCropPutsExactlyOneOnTopAndTheRestUnder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := seedLibrary(me, "A", "B", "C", "Fourth")
	pushCatalogPermanent(g, me.ID, "Cream of the Crop", "Enchantment", creamOfTheCropOracle, false)

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Beast", TypeLine: "Token Creature — Beast", Power: 3, Toughness: 3}, 1)
	})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	ask := putInLibraryChoiceFor(g, me.ID)
	if ask == nil || ask.LibraryTopCount != 1 || !sameIDs(ask.ScryCards, lib[:3]) {
		t.Fatalf("X = 3: the top three, exactly one to stay on top: %+v", ask)
	}
	if err := g.ResolvePutInLibrary(ask.ID, me.ID, []uuid.UUID{lib[2]}, []uuid.UUID{lib[0], lib[1]}); err == nil {
		t.Fatal("two on top is not \"one of those cards\"")
	}
	if err := g.ResolvePutInLibrary(ask.ID, me.ID, []uuid.UUID{lib[2], lib[0]}, []uuid.UUID{lib[1]}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := libraryTopIDs(me, 2); !sameIDs(got, []uuid.UUID{lib[1], lib[3]}) {
		t.Errorf("top two %v, want [B Fourth]", got)
	}
	if got := libraryBottomIDs(me, 2); !sameIDs(got, []uuid.UUID{lib[2], lib[0]}) {
		t.Errorf("bottom two %v, want [C A] in the order chosen", got)
	}
}

func TestCreamOfTheCropPowerOneHasNothingToChoose(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := seedLibrary(me, "A", "B")
	pushCatalogPermanent(g, me.ID, "Cream of the Crop", "Enchantment", creamOfTheCropOracle, false)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Squirrel", TypeLine: "Token Creature — Squirrel", Power: 1, Toughness: 1}, 1)
	})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if putInLibraryChoiceFor(g, me.ID) != nil {
		t.Error("one card, one on top: no prompt")
	}
	if got := libraryTopIDs(me, 1); got[0] != lib[0] {
		t.Error("the one card stays on top")
	}
}

// --- Hinder and Spell Crumple: a counter to a position -------------

// castShockAt puts an opponent's Shock on the stack, aimed at `at`.
func castShockAt(t *testing.T, g *game.Game, caster *game.Player, at uuid.UUID) uuid.UUID {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Shock", TypeLine: "Instant",
		OracleID: b5ShockOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: at}},
	}); err != nil {
		t.Fatalf("CastSpell Shock: %v", err)
	}
	return id
}

// castCounterAt casts a catalog counterspell from `caster`'s hand at a
// spell on the stack.
func castCounterAt(t *testing.T, g *game.Game, caster *game.Player, name, oracle string, spell uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant",
		OracleID: oracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func TestHinderPutsTheSpellOnTheCounterersChoiceOfEnd(t *testing.T) {
	for _, bury := range []bool{false, true} {
		name := "top"
		if bury {
			name = "bottom"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			seedLibrary(opp, "Their Top")
			lifeBefore := me.Life
			shock := castShockAt(t, g, opp, me.ID)
			hinder := castCounterAt(t, g, me, "Hinder", hinderOracle, shock)
			passPriorityAroundTable(t, g)

			if putInLibraryChoiceFor(g, opp.ID) != nil {
				t.Fatal("the choice is Hinder's controller's, not the spell's owner's")
			}
			ask := putInLibraryChoiceFor(g, me.ID)
			if ask == nil || ask.LibraryPlacement != game.LibraryPlaceTopOrBottom || !sameIDs(ask.ScryCards, []uuid.UUID{shock}) {
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
			passPriorityAroundTable(t, g)
			if opp.Graveyard.Contains(shock) || g.Stack.Contains(shock) {
				t.Fatal("countered into the library instead of the graveyard")
			}
			if bury {
				if opp.Library.Cards[0].InstanceID != shock {
					t.Error("on the bottom of its OWNER's library")
				}
			} else if got := libraryTopIDs(opp, 1); got[0] != shock {
				t.Error("on top of its OWNER's library")
			}
			if me.Life != lifeBefore {
				t.Error("a countered Shock deals no damage")
			}
			if countEvents(g, game.EventCounterSpell) == 0 {
				t.Error("Hinder counters (CR 701.6a)")
			}
			if !me.Graveyard.Contains(hinder) {
				t.Error("Hinder itself goes to the graveyard")
			}
			if !knownTo(g, shock, me.ID) || !knownTo(g, shock, opp.ID) {
				t.Error("the whole table saw where the spell went")
			}
		})
	}
}

func TestSpellCrumpleBottomsTheSpellAndItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "My Top")
	seedLibrary(opp, "Their Top")
	shock := castShockAt(t, g, opp, me.ID)
	crumple := castCounterAt(t, g, me, "Spell Crumple", spellCrumpleOracle, shock)
	passPriorityAroundTable(t, g)

	if putInLibraryChoiceFor(g, me.ID) != nil {
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
	if countEvents(g, game.EventCounterSpell) == 0 {
		t.Error("Spell Crumple counters (CR 701.6a)")
	}
}

package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// slice_296a_test.go — slice 296-a (card draw spells, #296): the
// discard-and-draw / sacrifice-and-draw family, two evasion grants
// that cantrip, and the "cast as though flash" cantrip.

const (
	corruptedConvictionOracle = "b45e35df-9032-4482-89a6-c7c50c6d0a79"
	tormentingVoiceOracle     = "f307b5b4-e949-4f69-8dc7-856e33a45a16"
	seizeTheSpoilsOracle      = "58f83528-9110-4895-b5ea-51b90af30a8d"
	piratesPillageOracle      = "7a2e866b-9642-495e-b47f-5b13a24373cc"
	catharticReunionOracle    = "0f3c3e5f-6af3-4af2-8703-4ccc8ed8f675"
	expediteOracle            = "3501a839-eef5-44e4-8637-b5754780454e"
	enterTheEnigmaOracle      = "67554654-e751-4679-9e92-f3588525ae4f"
	borneUponAWindOracle      = "ce19962d-94f9-4b2b-b668-963c0acce308"
)

// --- Corrupted Conviction (sacrifice cost) -----------------------

func TestCorruptedConvictionEatsACreatureAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := seedCreature(g, "Doomed Traveler", me.ID)
	for i := 0; i < 4; i++ {
		pushLibraryCardForTest(me, game.Card{
			InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery",
			Owner: me.ID, Controller: me.ID,
		})
	}
	handBefore := me.Hand.Size()

	castWithSacrifice(t, g, "Corrupted Conviction", "Instant", corruptedConvictionOracle, victim)

	if _, ok := battlefieldCard(g, victim); ok {
		t.Error("victim still on the battlefield after the cast — the sacrifice is a cost, not an effect")
	}
	if !me.Graveyard.Contains(victim) {
		t.Error("victim not in the graveyard")
	}

	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

func TestCorruptedConvictionUncastableWithNoCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	id, err := tryCastWithSacrifice(g, "Corrupted Conviction", "Instant", corruptedConvictionOracle, nil)
	if err == nil {
		t.Fatal("Corrupted Conviction cast with no creature to sacrifice")
	}
	if !me.Hand.Contains(id) {
		t.Error("rejected spell left the caster's hand")
	}
	if g.Stack.Contains(id) {
		t.Error("rejected spell reached the stack")
	}
}

// --- Tormenting Voice (discard 1, draw 2) ------------------------

func TestTormentingVoiceDrawsTwoAfterTheDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	before := me.Library.Size()

	_, paid := castWithDiscard(t, g, "Tormenting Voice", tormentingVoiceOracle, "Sorcery", 1)
	if !me.Graveyard.Contains(paid[0]) {
		t.Fatalf("the discard should be paid at announce, not on resolution")
	}
	passPriorityAroundTable(t, g)

	if got := before - me.Library.Size(); got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

func TestTormentingVoiceUncastableWithNoCardToDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Tormenting Voice", TypeLine: "Sorcery", OracleID: tormentingVoiceOracle,
		Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err == nil {
		t.Fatal("Tormenting Voice cast with nothing to discard (the spell itself is not a legal payment)")
	}
	if !me.Hand.Contains(id) {
		t.Error("rejected spell left the caster's hand")
	}
}

// --- Seize the Spoils / Pirate's Pillage (discard 1, draw 2, Treasures) --

func TestSeizeTheSpoilsAndPiratesPillageMakeTreasures(t *testing.T) {
	for _, tc := range []struct {
		name      string
		oracle    string
		treasures int
	}{
		{"Seize the Spoils", seizeTheSpoilsOracle, 1},
		{"Pirate's Pillage", piratesPillageOracle, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			me.Hand.Cards = nil
			before := me.Library.Size()

			castWithDiscard(t, g, tc.name, tc.oracle, "Sorcery", 1)
			passPriorityAroundTable(t, g)

			if got := before - me.Library.Size(); got != 2 {
				t.Errorf("drew %d cards, want 2", got)
			}
			treasures := 0
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Treasure" && c.Controller == me.ID {
					treasures++
				}
			}
			if treasures != tc.treasures {
				t.Errorf("Treasures = %d, want %d", treasures, tc.treasures)
			}
		})
	}
}

// --- Cathartic Reunion (discard 2, draw 3) -----------------------

func TestCatharticReunionDiscardsTwoDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	before := me.Library.Size()

	_, paid := castWithDiscard(t, g, "Cathartic Reunion", catharticReunionOracle, "Sorcery", 2)
	if len(paid) != 2 {
		t.Fatalf("setup: paid %d cards, want 2", len(paid))
	}
	for _, id := range paid {
		if !me.Graveyard.Contains(id) {
			t.Error("the discards should be paid at announce, not on resolution")
		}
	}
	passPriorityAroundTable(t, g)

	if got := before - me.Library.Size(); got != 3 {
		t.Errorf("drew %d cards, want 3", got)
	}
}

func TestCatharticReunionUncastableWithOnlyOneCardToDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	fodder := handCard(me, "Fodder", "Sorcery")
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Cathartic Reunion", TypeLine: "Sorcery", OracleID: catharticReunionOracle,
		Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{DiscardIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("Cathartic Reunion cast while only one card could be discarded, want two")
	}
	if !me.Hand.Contains(id) {
		t.Error("rejected spell left the caster's hand")
	}
}

// --- Expedite (target creature: haste + draw) --------------------

func TestExpediteGrantsHasteAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := ctrlPushCreature(g, me.ID, "Bear")
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Expedite", "Instant", expediteOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	var abilities []string
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == bear {
				abilities = c.Effective().Abilities
			}
		}
	})
	if !eotHasAbility(abilities, "haste") {
		t.Errorf("effective abilities %v do not include haste", abilities)
	}
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("hand delta = %d, want 1 (the spell left, one card drawn)", got)
	}
}

func TestExpediteRefusesAnIllegalTarget(t *testing.T) {
	g := newCatalogGame(t)
	if err := castCatalogSpellErr(t, g, "Expedite", "Instant", expediteOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: uuid.New()}}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Expedite at a nonexistent creature: err = %v, want ErrIllegalTarget", err)
	}
}

// --- Enter the Enigma (target creature: unblockable + draw) ------

func TestEnterTheEnigmaMakesUnblockableAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := ctrlPushCreature(g, me.ID, "Bear")
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Enter the Enigma", "Sorcery", enterTheEnigmaOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	assertRestrictions(t, g, bear, game.CantBeBlocked)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("hand delta = %d, want 1 (the spell left, one card drawn)", got)
	}
}

func TestEnterTheEnigmaRefusesAnIllegalTarget(t *testing.T) {
	g := newCatalogGame(t)
	if err := castCatalogSpellErr(t, g, "Enter the Enigma", "Sorcery", enterTheEnigmaOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: uuid.New()}}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Enter the Enigma at a nonexistent creature: err = %v, want ErrIllegalTarget", err)
	}
}

// --- Borne Upon a Wind (grant flash this turn + draw) ------------

func TestBorneUponAWindGrantsFlashForTheTurnAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Borne Upon a Wind", "Instant", borneUponAWindOracle, nil)
	passPriorityAroundTable(t, g)

	if n := ctStoredTimings(me); n != 1 {
		t.Fatalf("the grant did not land on the player: %d statements", n)
	}
	if k := me.Statics[0].Duration.Kind; k != game.UntilEndOfTurn {
		t.Errorf("duration kind = %v, want UntilEndOfTurn — the clause is \"this turn\"", k)
	}
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("hand delta = %d, want 1 (the spell left, one card drawn)", got)
	}

	// The window is open in this turn's end step.
	advanceTo(t, g, game.StepEnd)
	bear := ctHandCard(me, "Proof Bear", "Creature — Bear")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("a creature spell in the end step under Borne Upon a Wind's grant: %v", err)
	}
	passPriorityAroundTable(t, g)

	// And it ends with the turn (CR 514.2).
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if n := ctStoredTimings(me); n != 0 {
		t.Errorf("the grant outlived its turn: %d statements left", n)
	}
	later := ctHandCard(me, "Proof Bear Two", "Creature — Bear")
	if err := g.CastSpell(me.ID, later, game.CastSpellParams{}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("a creature spell after the grant expired = %v, want ErrSorcerySpeedRequired", err)
	}
}

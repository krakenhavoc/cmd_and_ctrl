package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func castRescuer(t *testing.T, g *game.Game, params game.CastSpellParams) error {
	t.Helper()
	_, err := effectstest.CastWithTapParams(t, g, "The Wandering Rescuer",
		"Legendary Creature — Human Samurai Noble", effectstest.WanderingRescuerCost,
		effectstest.TheWanderingRescuerOracle, params)
	return err
}

// --- announce-time validation ------------------------------------

// The cost taps what the caster named, and the taps land on the
// board rather than being a bookkeeping fiction.
func TestConvokeTapsTheNamedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	ids := effectstest.PushTapCostSoldiers(g, g.Seats[0].ID, 2)
	if err := castRescuer(t, g, game.CastSpellParams{TapIDs: ids}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for _, id := range ids {
		if !effectstest.TapCostTapped(g, id) {
			t.Error("a convoked creature was not tapped")
		}
	}
}

// You may not tap more creatures than the spell costs (CR 702.51a).
// The Wandering Rescuer is {3}{W}{W} — five symbols, so a sixth
// creature is a rejection rather than a wasted tap, and a rejected
// cast leaves the board untouched.
func TestConvokeRejectsTappingBeyondTheCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	ids := effectstest.PushTapCostSoldiers(g, g.Seats[0].ID, 6)
	if err := castRescuer(t, g, game.CastSpellParams{TapIDs: ids}); err == nil {
		t.Fatal("convoking six creatures at a five-mana spell was accepted")
	}
	for _, id := range ids {
		if effectstest.TapCostTapped(g, id) {
			t.Fatal("a creature was tapped by a rejected cast")
		}
	}
}

// A land is not a creature, so convoke can't tap it — and waterbend,
// whose pool is wider, still can't.
func TestTapCostRejectsAPermanentOutsideTheClause(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	island := effectstest.PushTapCostPermanent(g, g.Seats[0].ID, "Island", "Basic Land — Island")

	if err := castRescuer(t, g, game.CastSpellParams{TapIDs: []uuid.UUID{island}}); err == nil {
		t.Error("convoke tapped a land")
	}
	if err := effectstest.CastRestoration(t, g, game.CastSpellParams{
		XValue: 1, TapIDs: []uuid.UUID{island},
	}); err == nil {
		t.Error("waterbend tapped a land")
	}
}

// Waterbend's pool is wider than convoke's: an artifact pays for {1}
// here and can't there.
func TestWaterbendAcceptsAnArtifactConvokeRefuses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.PushTapCostPermanent(g, me.ID, "Sol Ring", "Artifact")
	victim := effectstest.PushTapCostPermanent(g, me.ID, "Bear", "Creature — Bear", "G")

	if err := castRescuer(t, g, game.CastSpellParams{TapIDs: []uuid.UUID{rock}}); err == nil {
		t.Error("convoke tapped an artifact that isn't a creature")
	}
	if err := effectstest.CastRestoration(t, g, game.CastSpellParams{
		XValue:  1,
		TapIDs:  []uuid.UUID{rock},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("waterbend refused an artifact: %v", err)
	}
	if !effectstest.TapCostTapped(g, rock) {
		t.Error("the artifact paid the cost but was not tapped")
	}
}

// A tapped permanent has nothing left to give.
func TestTapCostRejectsAnAlreadyTappedPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	soldier := effectstest.PushTapCostSoldiers(g, g.Seats[0].ID, 1)[0]
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == soldier {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	if err := castRescuer(t, g, game.CastSpellParams{TapIDs: []uuid.UUID{soldier}}); err == nil {
		t.Error("a tapped creature paid a convoke cost")
	}
}

// You may only tap permanents YOU control.
func TestTapCostRejectsAnOpponentsPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	theirs := effectstest.PushTapCostSoldiers(g, g.Seats[1].ID, 1)[0]
	if err := castRescuer(t, g, game.CastSpellParams{TapIDs: []uuid.UUID{theirs}}); err == nil {
		t.Error("convoke tapped a creature the caster doesn't control")
	}
}

// A card that charges no such cost arriving WITH tap IDs is a client
// bug, and rejecting it keeps the wire honest — the same posture the
// additional-cost validator takes.
func TestTapIDsOnACardWithoutTheCostAreRejected(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	soldier := effectstest.PushTapCostSoldiers(g, g.Seats[0].ID, 1)[0]
	if _, err := effectstest.CastWithTapParams(t, g, "Filler", "Sorcery", "{1}",
		"00000000-0000-4000-8000-00000000face",
		game.CastSpellParams{TapIDs: []uuid.UUID{soldier}}); err == nil {
		t.Error("a card with no tap cost accepted tap IDs")
	}
}

// Tapping nothing is always a legal answer — "you MAY tap any
// number" — and the caster then owes the whole cost in mana.
func TestTapCostIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if err := castRescuer(t, g, game.CastSpellParams{}); err != nil {
		t.Errorf("casting a convoke card without convoking was rejected: %v", err)
	}
}

// --- the strict-mana gate ----------------------------------------

// The point of the whole component: convoking makes a spell the pool
// can't afford affordable. Five white creatures cast a {3}{W}{W} off
// an empty pool.
func TestConvokeMakesAnUnaffordableSpellPayable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	ids := effectstest.PushTapCostSoldiers(g, g.Seats[0].ID, 5)
	if err := castRescuer(t, g, game.CastSpellParams{Strict: true, TapIDs: ids}); err != nil {
		t.Fatalf("five convoked creatures could not pay {3}{W}{W}: %v", err)
	}
}

// Four of them cannot — the fifth symbol is still owed, and strict
// mode says so instead of letting it through.
func TestConvokeShortOfTheCostStillFailsTheStrictGate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	ids := effectstest.PushTapCostSoldiers(g, g.Seats[0].ID, 4)
	if err := castRescuer(t, g, game.CastSpellParams{Strict: true, TapIDs: ids}); err == nil {
		t.Fatal("four convoked creatures paid a five-mana spell")
	}
}

// Convoke's colour clause is real on a live board: {3}{W}{W} cast
// with three colourless creatures and two white ones is payable —
// the colourless pair can only cover the generic half, and the
// white ones have to land on the {W}{W}.
func TestConvokeColorClauseCoversTheColoredHalf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var ids []uuid.UUID
	for i := 0; i < 3; i++ {
		ids = append(ids, effectstest.PushTapCostPermanent(g, me.ID, "Scrap", "Artifact Creature — Construct"))
	}
	ids = append(ids, effectstest.PushTapCostSoldiers(g, me.ID, 2)...)
	if err := castRescuer(t, g, game.CastSpellParams{Strict: true, TapIDs: ids}); err != nil {
		t.Fatalf("three colourless + two white could not pay {3}{W}{W}: %v", err)
	}
}

// Swap the two white creatures for two more colourless ones and the
// {W}{W} is unpayable no matter how many bodies are tapped.
func TestConvokeColorlessCreaturesCannotPayTheColoredHalf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var ids []uuid.UUID
	for i := 0; i < 5; i++ {
		ids = append(ids, effectstest.PushTapCostPermanent(g, me.ID, "Scrap", "Artifact Creature — Construct"))
	}
	if err := castRescuer(t, g, game.CastSpellParams{Strict: true, TapIDs: ids}); err == nil {
		t.Fatal("five colourless creatures paid a {3}{W}{W}")
	}
}

// --- Waterbender's Restoration, now costed (issue #259) ----------

// The clause is "Exile X target creatures you control", and X is the
// waterbend cost that was paid. Announcing X=1 and naming two
// targets is a rejection; while the cost was uncharged, "any number
// of targets for free" was the card's actual behaviour.
func TestWaterbendersRestorationTargetCountIsTiedToX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushTapCostPermanent(g, me.ID, "Bear", "Creature — Bear", "G")
	b := effectstest.PushTapCostPermanent(g, me.ID, "Bear", "Creature — Bear", "G")
	both := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}

	if err := effectstest.CastRestoration(t, g, game.CastSpellParams{XValue: 1, Targets: both}); err == nil {
		t.Error("X=1 blinked two creatures")
	}
	if err := effectstest.CastRestoration(t, g, game.CastSpellParams{XValue: 2, Targets: both}); err != nil {
		t.Errorf("X=2 with two targets was rejected: %v", err)
	}
}

// X=0 takes no targets. Max 0 means "unbounded" everywhere else in
// TargetSpec, so this is the case the count resolution could most
// easily get wrong — and an unbounded free blink is precisely the
// bug #259 was filed about.
func TestWaterbendersRestorationAtXZeroTakesNoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.PushTapCostPermanent(g, g.Seats[0].ID, "Bear", "Creature — Bear", "G")
	if err := effectstest.CastRestoration(t, g, game.CastSpellParams{
		XValue:  0,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("X=0 blinked a creature for free")
	}
}

// End to end under the strict gate: {U}{U} plus waterbend {2}, the
// {2} paid by tapping two permanents. The pool still has to cover
// the {U}{U} — that is the half the tapping never touches, and the
// half an uncosted waterbend would have made free.
func TestWaterbendersRestorationChargesTheWaterbendCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushTapCostPermanent(g, me.ID, "Bear", "Creature — Bear", "G")
	b := effectstest.PushTapCostPermanent(g, me.ID, "Bear", "Creature — Bear", "G")
	helper := effectstest.PushTapCostPermanent(g, me.ID, "Sol Ring", "Artifact")
	both := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}

	// An empty pool can't pay the {U}{U}, tap all you like.
	if err := effectstest.CastRestoration(t, g, game.CastSpellParams{
		Strict: true, XValue: 2, TapIDs: []uuid.UUID{a, helper}, Targets: both,
	}); err == nil {
		t.Fatal("tapping paid the spell's own {U}{U}")
	}
	if effectstest.TapCostTapped(g, helper) {
		t.Fatal("a rejected cast still tapped the board")
	}

	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	if err := effectstest.CastRestoration(t, g, game.CastSpellParams{
		Strict: true, XValue: 2, TapIDs: []uuid.UUID{a, helper}, Targets: both,
	}); err != nil {
		t.Fatalf("waterbend {2} paid by two taps was rejected: %v", err)
	}
	if !effectstest.TapCostTapped(g, a) || !effectstest.TapCostTapped(g, helper) {
		t.Error("the permanents that paid the waterbend cost were not tapped")
	}
}

// Waterbend {2} with only one permanent tapped still owes {1} on top
// of the {U}{U}, so a two-blue pool is short.
func TestWaterbendersRestorationShortTapStillOwesTheRemainder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushTapCostPermanent(g, me.ID, "Bear", "Creature — Bear", "G")
	b := effectstest.PushTapCostPermanent(g, me.ID, "Bear", "Creature — Bear", "G")
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})

	if err := effectstest.CastRestoration(t, g, game.CastSpellParams{
		Strict: true, XValue: 2,
		TapIDs:  []uuid.UUID{a},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
	}); err == nil {
		t.Fatal("waterbend {2} was satisfied by a single tap and two blue mana")
	}
}

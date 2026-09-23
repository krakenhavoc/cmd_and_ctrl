package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_legend_sagas_test.go — issue #343: the four Avatar: The Last
// Airbender transforming Sagas, each a Saga front face that at
// chapter III exiles itself and returns transformed as a legendary
// Avatar creature (ADR 0079's second verb). Built on
// transform_cards_test.go's importAndCast / importToBattlefield /
// assertFaceInvariant fixtures for the reason that file gives: a
// transform card's whole point is Faces[1], which only reaches
// game.Card through deck.ToGameCard, so a fixture built with
// castCatalogSpell (no faces at all) would exercise a CanTransform
// that always answers false.
//
// Each card gets the same four checks the task asked for: the Saga
// enters with a lore counter and chapter I fires; chapter III exiles
// and returns the card transformed under its controller's control as
// the Avatar face; the SBA does not sacrifice a Saga that has left
// (folded into the chapter III assertion, exactly as
// TestFableChapterThreeExilesAndReturnsItTransformed does — the
// original object is gone but not in the graveyard); and one
// back-face ability works.

// avatarSagaRow builds the Scryfall-shaped fixture for one of the
// four Sagas. Not transformRow: that helper paints ONE Colors slice
// onto both faces, which would wrongly give the colourless Avatar
// creature the Saga's colour.
func avatarSagaRow(oracleID, sagaName, sagaCost, avatarName, power, toughness, sagaColor string) cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     sagaName + " // " + avatarName,
		Layout:   "transform",
		TypeLine: "Enchantment — Saga // Legendary Creature — Avatar",
		OracleID: uuid.MustParse(oracleID),
		CardFaces: []cards.CardFace{
			{Name: sagaName, TypeLine: "Enchantment — Saga", ManaCost: sagaCost, Colors: []string{sagaColor}},
			{Name: avatarName, TypeLine: "Legendary Creature — Avatar", Power: power, Toughness: toughness, Colors: []string{}},
		},
	}
}

func kurukRow() cards.Card {
	return avatarSagaRow(theLegendOfKurukOracleID, "The Legend of Kuruk", "{2}{U}{U}", "Avatar Kuruk", "4", "3", "U")
}

func kyoshiRow() cards.Card {
	return avatarSagaRow(theLegendOfKyoshiOracleID, "The Legend of Kyoshi", "{4}{G}{G}", "Avatar Kyoshi", "5", "4", "G")
}

func rokuRow() cards.Card {
	return avatarSagaRow(theLegendOfRokuOracleID, "The Legend of Roku", "{2}{R}{R}", "Avatar Roku", "4", "4", "R")
}

func yangchenRow() cards.Card {
	return avatarSagaRow(theLegendOfYangchenOracleID, "The Legend of Yangchen", "{3}{W}{W}", "Avatar Yangchen", "4", "5", "W")
}

// --- shared prompt helpers, one per pending-choice kind these cards
// --- raise that no existing test file already answers -------------

// answerScryKeepAll resolves a scry prompt keeping every card on top
// in the order it was shown — Kuruk's "scry 2, then draw" chapters
// don't care about order, only that the prompt is answered so the
// "then" runs.
func answerScryKeepAll(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	c := scryChoiceFor(g, chooser)
	if c == nil {
		t.Fatal("no scry prompt to answer")
	}
	if err := g.ResolveScry(c.ID, chooser, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
}

// resolveManaPickAnyColor answers a "add mana of any color" pipe
// prompt with its first offered colour and returns which one it
// picked.
func resolveManaPickAnyColor(t *testing.T, g *game.Game, chooser uuid.UUID) string {
	t.Helper()
	c := latestChoiceOfKind(g, game.PendingChoiceMana)
	if c == nil || c.Chooser != chooser {
		t.Fatal("no mana_pick prompt to answer")
	}
	if len(c.ColorOptions) == 0 {
		t.Fatal("mana_pick prompt has no color options")
	}
	color := c.ColorOptions[0]
	if err := g.ResolveManaChoice(c.ID, chooser, color); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	return color
}

// assertReturnedAsAvatar is chapter III's card-level proof, common to
// all four Sagas: the original object is gone (not merely
// transformed in place, and not sacrificed to CR 714.4 either — a
// Saga that has already left when the SBA next runs), a NEW object
// with the Avatar's name/type/P-T has arrived under the controller's
// control, with no lore counters, and it is summoning sick because it
// just entered (CR 400.7). Mirrors
// TestFableChapterThreeExilesAndReturnsItTransformed's assertion set.
func assertReturnedAsAvatar(t *testing.T, g *game.Game, saga, controller uuid.UUID, avatarName string, power, toughness int) *game.Card {
	t.Helper()
	if battlefieldCardFor(g, saga) != nil {
		t.Fatal("the original Saga object is still on the battlefield — chapter III is an exile and a return, not an in-place transform")
	}
	if inGraveyardOf(g, controller, saga) {
		t.Fatal("the Saga was sacrificed (CR 714.4) instead of coming back")
	}
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == avatarName {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatalf("%s never reached the battlefield", avatarName)
	}
	if back.InstanceID == saga {
		t.Error("it kept its InstanceID — CR 400.7 makes the returning permanent a new object")
	}
	if back.Controller != controller {
		t.Errorf("controller = %s, want %s (under your control)", back.Controller, controller)
	}
	if back.ActiveFace != 1 {
		t.Errorf("ActiveFace = %d, want the back face", back.ActiveFace)
	}
	if back.Counters[game.CounterLore] != 0 {
		t.Errorf("it came back with %d lore counters", back.Counters[game.CounterLore])
	}
	if !back.IsCreature() || back.Power != power || back.Toughness != toughness {
		t.Errorf("came back as %q %d/%d, want a %d/%d creature",
			back.TypeLine, back.Power, back.Toughness, power, toughness)
	}
	if !back.SummonedThisTurn {
		t.Error("it is not summoning sick — the returning permanent entered this turn")
	}
	return back
}

// putBackFaceOntoBattlefield puts a card straight onto the
// battlefield on its BACK face and clears summoning sickness — the
// back-face-ability tests care about the ability, not how the face
// got there (Reflection of Kiki-Jiki's own test takes the same
// shortcut).
func putBackFaceOntoBattlefield(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	id := importToHand(row, p)
	g.WithWriteLock(func() {
		for i := range p.Hand.Cards {
			if p.Hand.Cards[i].InstanceID == id {
				p.Hand.Cards[i].SetFace(1)
			}
		}
	})
	var entered uuid.UUID
	g.WithWriteLock(func() {
		var err error
		entered, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{})
		if err != nil {
			t.Fatalf("put the back face onto the battlefield: %v", err)
		}
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == entered {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	return entered
}

// --- The Legend of Kuruk // Avatar Kuruk ---------------------------

// TestKurukEntersAndChapterOneScriesThenDraws is the entry + chapter
// I proof: one lore counter from CR 714.3, then "scry 2, then draw a
// card" (Preordain's own body, chaptered).
func TestKurukEntersAndChapterOneScriesThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	handBefore := me.Hand.Size()

	saga := importAndCast(t, g, kurukRow(), me)
	passPriorityAroundTable(t, g)

	if got := loreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}
	answerScryKeepAll(t, g, me.ID)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Fatalf("hand delta after chapter I = %d, want 1 (scry 2, then draw a card)", got)
	}
}

// TestKurukChapterThreeReturnsAsAvatarKuruk walks all three chapters
// (answering the identical chapter II scry along the way) and checks
// the transform.
func TestKurukChapterThreeReturnsAsAvatarKuruk(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]

	saga := importAndCast(t, g, kurukRow(), me)
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)
	if got := loreCountersOn(g, saga); got != 2 {
		t.Fatalf("lore counters after chapter II = %d, want 2", got)
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)

	assertReturnedAsAvatar(t, g, saga, me.ID, "Avatar Kuruk", 4, 3)
	assertFaceInvariant(t, g)
}

// TestAvatarKurukCreatesASpiritWhenYouCastASpell is the back-face
// ability: "Whenever you cast a spell, create a 1/1 colorless
// Spirit." The Spirit's own "can't block or be blocked by non-Spirit
// creatures" clause is the declared caveat (#521) — not asserted
// here, because it does not exist to assert.
func TestAvatarKurukCreatesASpiritWhenYouCastASpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	putBackFaceOntoBattlefield(t, g, kurukRow(), me)

	before := countBattlefieldByName(g, "Spirit")
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)

	if got := countBattlefieldByName(g, "Spirit") - before; got != 1 {
		t.Fatalf("Spirits created = %d, want 1", got)
	}
}

// --- The Legend of Kyoshi // Avatar Kyoshi --------------------------

// TestKyoshiChapterOneDrawsEqualToGreatestPower is the entry +
// chapter I proof.
func TestKyoshiChapterOneDrawsEqualToGreatestPower(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Beast", TypeLine: "Creature — Beast",
		Owner: me.ID, Controller: me.ID, Power: 5, Toughness: 5,
	})
	handBefore := me.Hand.Size()

	saga := importAndCast(t, g, kyoshiRow(), me)
	passPriorityAroundTable(t, g)

	if got := loreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}
	if got := me.Hand.Size() - handBefore; got != 5 {
		t.Fatalf("hand delta after chapter I = %d, want 5 (the greatest power on the board)", got)
	}
}

// TestKyoshiChapterThreeReturnsAsAvatarKyoshi walks all three
// chapters. Chapter II fizzles (no land in play to target — CR
// 603.3d, zero legal targets), which is fine: this test is chapter
// III's, not chapter II's.
func TestKyoshiChapterThreeReturnsAsAvatarKyoshi(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]

	saga := importAndCast(t, g, kyoshiRow(), me)
	passPriorityAroundTable(t, g)
	if got := loreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if got := loreCountersOn(g, saga); got != 2 {
		t.Fatalf("lore counters after chapter II = %d, want 2", got)
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)

	assertReturnedAsAvatar(t, g, saga, me.ID, "Avatar Kyoshi", 5, 4)
	assertFaceInvariant(t, g)
}

// TestAvatarKyoshiManaAbilityAddsGreatestPowerInOneColor is the
// back-face ability: "{T}: Add X mana of any one color, X = the
// greatest power among creatures you control" — Selvala's shape with
// one colour instead of a combination.
func TestAvatarKyoshiManaAbilityAddsGreatestPowerInOneColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	entered := putBackFaceOntoBattlefield(t, g, kyoshiRow(), me)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Beast", TypeLine: "Creature — Beast",
		Owner: me.ID, Controller: me.ID, Power: 5, Toughness: 5,
	})

	if err := g.ActivateManaAbility(me.ID, entered, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	color := resolveManaPickAnyColor(t, g, me.ID)

	got := 0
	for _, tok := range me.ManaPool {
		if tok.Color == color {
			got++
		}
	}
	if got != 5 {
		t.Fatalf("pool has %d mana of %s, want 5 (the greatest power)", got, color)
	}
}

// --- The Legend of Roku // Avatar Roku ------------------------------

// TestRokuChapterOneExilesTopThreeWithPlayPermission is the entry +
// chapter I proof: Reckless Impulse's primitive with three cards
// instead of two.
func TestRokuChapterOneExilesTopThreeWithPlayPermission(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	top := make([]uuid.UUID, 3)
	for i := range top {
		top[i] = me.Library.Cards[me.Library.Size()-1-i].InstanceID
	}

	saga := importAndCast(t, g, rokuRow(), me)
	passPriorityAroundTable(t, g)

	if got := loreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}
	for _, id := range top {
		if !g.Exile.Contains(id) {
			t.Errorf("card %s never reached exile", id)
			continue
		}
		if perm := g.CastPermissionOnCardByIDForEffect(id); perm == nil {
			t.Errorf("exiled card %s has no play permission", id)
		}
	}
}

// TestRokuChapterThreeReturnsAsAvatarRoku walks all three chapters,
// answering chapter II's colour pick along the way.
func TestRokuChapterThreeReturnsAsAvatarRoku(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]

	saga := importAndCast(t, g, rokuRow(), me)
	passPriorityAroundTable(t, g)

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	resolveManaPickAnyColor(t, g, me.ID)
	if got := loreCountersOn(g, saga); got != 2 {
		t.Fatalf("lore counters after chapter II = %d, want 2", got)
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)

	assertReturnedAsAvatar(t, g, saga, me.ID, "Avatar Roku", 4, 4)
	assertFaceInvariant(t, g)
}

// TestAvatarRokuCreatesADragonToken is the back-face ability:
// "{8}: Create a 4/4 red Dragon creature token with flying and
// firebending 4." Permissive mode waives the unfunded mana cost
// (AGENTS.md §7's mana section — the human default), so the test does
// not need to float eight mana first.
func TestAvatarRokuCreatesADragonToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	entered := putBackFaceOntoBattlefield(t, g, rokuRow(), me)

	before := countBattlefieldByName(g, "Dragon")
	if err := g.ActivateCatalogAbility(me.ID, entered, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := countBattlefieldByName(g, "Dragon") - before; got != 1 {
		t.Fatalf("Dragon tokens created = %d, want 1", got)
	}
}

// --- The Legend of Yangchen // Avatar Yangchen ----------------------

// TestYangchenChapterOneExilesAnOpponentsExpensivePermanent is the
// entry + chapter I proof: the declared slice of the printed ability
// (the controller's own target, not every player's).
func TestYangchenChapterOneExilesAnOpponentsExpensivePermanent(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Big Rock", TypeLine: "Artifact",
		ManaCost: "{4}", Owner: opp.ID, Controller: opp.ID,
	})

	saga := importAndCast(t, g, yangchenRow(), me)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	if got := loreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("chapter I did not exile the qualifying permanent")
	}
}

// TestYangchenChapterThreeReturnsAsAvatarYangchen walks all three
// chapters, picking chapter II's target opponent and declining the
// "you may" so the assertion stays about chapter III.
func TestYangchenChapterThreeReturnsAsAvatarYangchen(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]

	saga := importAndCast(t, g, yangchenRow(), me)
	passPriorityAroundTable(t, g)
	if got := loreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	if got := loreCountersOn(g, saga); got != 2 {
		t.Fatalf("lore counters after chapter II = %d, want 2", got)
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)

	assertReturnedAsAvatar(t, g, saga, me.ID, "Avatar Yangchen", 4, 5)
	assertFaceInvariant(t, g)
}

// TestAvatarYangchenAirbendsOnSecondSpellEachTurn is the back-face
// ability: "Whenever you cast your second spell each turn, airbend up
// to one other target nonland permanent."
func TestAvatarYangchenAirbendsOnSecondSpellEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := pushCreatureToBattlefieldForTest(g, opp.ID, "Their Blocker")
	putBackFaceOntoBattlefield(t, g, yangchenRow(), me)

	castCatalogSpell(t, g, "First Spell", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("the first spell each turn does not trigger: %+v", g.PendingChoices)
	}

	castCatalogSpell(t, g, "Second Spell", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	assertAirbent(t, g, victim, opp.ID)
}

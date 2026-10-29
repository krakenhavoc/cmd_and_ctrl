package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tom_bombadil_bilbo_4505_test.go — Tom Bombadil and Bilbo, Thief in the
// Night (seam #4505, The Hobbit #4454).

// --- Tom Bombadil ---------------------------------------------------

// tomTable is a game with Tom on the active player's battlefield.
func tomTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	// In the main phase, so no turn-based lore counter lands mid-test.
	toMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tom := apaPush(g, me.ID, me.ID, game.Card{Name: "Tom Bombadil", OracleID: tomBombadilOracleID,
		TypeLine: "Legendary Creature — God Bard", ManaCost: "{W}{U}{B}{R}{G}", Power: 4, Toughness: 4})
	return g, me, opp, tom
}

// tomBenalia puts a History of Benalia (three chapters) onto the
// battlefield under `p` with `lore` lore counters already on it.
func tomBenalia(g *game.Game, p *game.Player, lore int) uuid.UUID {
	return apaPush(g, p.ID, p.ID, game.Card{Name: "History of Benalia", OracleID: historyOfBenaliaOracle,
		TypeLine: "Enchantment — Saga", ManaCost: "{1}{W}{W}", Counters: map[string]int{game.CounterLore: lore}})
}

// tomTestSaga is a Saga card with no chapters: it never triggers and is
// never sacrificed, so it sits still wherever a test puts it.
func tomTestSaga(p *game.Player, name string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Enchantment — Saga",
		Owner: p.ID, Controller: p.ID}
}

func tomPlainCard(p *game.Player, name string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Sorcery", ManaCost: "{1}",
		Owner: p.ID, Controller: p.ID}
}

// tomLibrary replaces p's library with `cards`, bottom first.
func tomLibrary(g *game.Game, p *game.Player, cards ...game.Card) {
	g.WithWriteLock(func() { p.Library.Cards = cards })
}

// tomAdvance puts one lore counter on the Saga and lets everything that
// triggers resolve.
func tomAdvance(t *testing.T, g *game.Game, saga uuid.UUID) {
	t.Helper()
	if err := g.AddCounter(saga, game.CounterLore, 1); err != nil {
		t.Fatalf("add a lore counter: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func tomTriggered(g *game.Game, tom uuid.UUID) int {
	return g.TriggeredThisTurn(tom, tomBombadilLabel)
}

// The final chapter resolving reveals down to the first Saga card, puts it
// onto the battlefield, and bottoms the cards above it; the card below it
// stays where it was, now on top.
func TestTomBombadilFinalChapterPutsTheFirstSagaOntoTheBattlefield(t *testing.T) {
	g, me, _, tom := tomTable(t)
	saga := tomBenalia(g, me, 2)
	below := tomPlainCard(me, "Below")
	hit := tomTestSaga(me, "Revealed Saga")
	a, b := tomPlainCard(me, "Top"), tomPlainCard(me, "Second")
	tomLibrary(g, me, below, hit, b, a)

	tomAdvance(t, g, saga)

	if got := tomTriggered(g, tom); got != 1 {
		t.Fatalf("Tom triggered %d times, want 1", got)
	}
	if findBattlefieldCardForTest(g, hit.InstanceID) == nil {
		t.Fatal("the revealed Saga card is not on the battlefield")
	}
	lib := libraryIDs(me)
	if len(lib) != 3 || lib[len(lib)-1] != below.InstanceID {
		t.Fatalf("library = %v, want the card below the Saga on top of the two bottomed ones", lib)
	}
	if !slices.Contains(lib[:2], a.InstanceID) || !slices.Contains(lib[:2], b.InstanceID) {
		t.Errorf("the revealed cards above the Saga are not on the bottom: %v", lib)
	}
}

// "This ability triggers only once each turn": a second final chapter the
// same turn does nothing, and the next turn it triggers again.
func TestTomBombadilTriggersOnlyOnceEachTurn(t *testing.T) {
	g, me, _, tom := tomTable(t)
	first, second := tomBenalia(g, me, 2), tomBenalia(g, me, 2)
	s1, s2, s3 := tomTestSaga(me, "Saga One"), tomTestSaga(me, "Saga Two"), tomTestSaga(me, "Saga Three")
	tomLibrary(g, me, s3, s2, s1)

	tomAdvance(t, g, first)
	tomAdvance(t, g, second)

	if got := tomTriggered(g, tom); got != 1 {
		t.Fatalf("Tom triggered %d times this turn, want 1", got)
	}
	if findBattlefieldCardForTest(g, s1.InstanceID) == nil || findBattlefieldCardForTest(g, s2.InstanceID) != nil {
		t.Fatal("want exactly one Saga put onto the battlefield this turn")
	}

	advanceToPrecombatMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	passPriorityAroundTable(t, g)
	third := tomBenalia(g, me, 2)
	tomAdvance(t, g, third)
	if got := tomTriggered(g, tom); got != 1 {
		t.Fatalf("Tom triggered %d times on the next turn, want 1", got)
	}
	if findBattlefieldCardForTest(g, s2.InstanceID) == nil {
		t.Error("the next turn's final chapter put no Saga onto the battlefield")
	}
}

// Only the FINAL chapter ability, and only of a Saga Tom's controller
// controls: chapter II of my Saga and chapter III of an opponent's do
// nothing.
func TestTomBombadilIgnoresOtherChaptersAndOtherPlayersSagas(t *testing.T) {
	g, me, opp, tom := tomTable(t)
	mine := tomBenalia(g, me, 1)
	theirs := tomBenalia(g, opp, 2)
	hit := tomTestSaga(me, "Library Saga")
	tomLibrary(g, me, hit)

	tomAdvance(t, g, mine)
	if countBattlefieldByName(g, "Knight") == 0 {
		t.Fatal("chapter II did not resolve")
	}
	if got := tomTriggered(g, tom); got != 0 {
		t.Fatalf("Tom triggered on chapter II (%d)", got)
	}

	tomAdvance(t, g, theirs)
	if inGraveyardOf(g, opp.ID, theirs) == false {
		t.Fatal("the opponent's final chapter did not resolve (its Saga was not sacrificed)")
	}
	if got := tomTriggered(g, tom); got != 0 {
		t.Fatalf("Tom triggered on an opponent's final chapter (%d)", got)
	}
	if findBattlefieldCardForTest(g, hit.InstanceID) != nil {
		t.Error("a Saga was put onto the battlefield")
	}
}

// No Saga card in the library: every card is revealed, nothing enters,
// and all of them go to the bottom.
func TestTomBombadilWithNoSagaInTheLibraryBottomsEverything(t *testing.T) {
	g, me, _, tom := tomTable(t)
	saga := tomBenalia(g, me, 2)
	x, y, z := tomPlainCard(me, "X"), tomPlainCard(me, "Y"), tomPlainCard(me, "Z")
	tomLibrary(g, me, x, y, z)
	before := len(g.Battlefield.Cards)

	tomAdvance(t, g, saga)

	if got := tomTriggered(g, tom); got != 1 {
		t.Fatalf("Tom triggered %d times, want 1", got)
	}
	// History of Benalia was sacrificed (CR 714.4); its two Knights stay.
	if got := len(g.Battlefield.Cards); got != before-1 {
		t.Errorf("battlefield went from %d to %d cards, want only the sacrificed Saga gone", before, got)
	}
	lib := libraryIDs(me)
	if len(lib) != 3 {
		t.Fatalf("library has %d cards, want all three back", len(lib))
	}
	for _, id := range []uuid.UUID{x.InstanceID, y.InstanceID, z.InstanceID} {
		if !slices.Contains(lib, id) {
			t.Errorf("card %s left the library", id)
		}
	}
	var revealed []uuid.UUID
	for _, ev := range g.Events {
		if ev.Kind == game.EventRevealCards && ev.Source == tom {
			revealed = append(revealed, ev.CardID)
		}
	}
	if len(revealed) != 3 {
		t.Fatalf("Tom revealed %d cards, want the whole library of 3", len(revealed))
	}
}

// Hexproof and indestructible at four or more lore counters among Sagas
// Tom's controller controls, added up across them; an opponent's do not
// count.
func TestTomBombadilHexproofAndIndestructibleAtFourLoreCounters(t *testing.T) {
	g, me, opp, tom := tomTable(t)
	one := tomTestSaga(me, "Saga A")
	one.Counters = map[string]int{game.CounterLore: 1}
	two := tomTestSaga(me, "Saga B")
	two.Counters = map[string]int{game.CounterLore: 2}
	apaPush(g, me.ID, me.ID, one)
	b := apaPush(g, me.ID, me.ID, two)
	theirs := tomTestSaga(opp, "Their Saga")
	theirs.Counters = map[string]int{game.CounterLore: 5}
	apaPush(g, opp.ID, opp.ID, theirs)

	live := apaLive(g, tom)
	if game.HasKeyword(live, "hexproof") || game.HasKeyword(live, "indestructible") {
		t.Fatal("three lore counters of mine already grant the keywords")
	}

	if err := g.AddCounter(b, game.CounterLore, 1); err != nil {
		t.Fatalf("add a lore counter: %v", err)
	}
	live = apaLive(g, tom)
	if !game.HasKeyword(live, "hexproof") || !game.HasKeyword(live, "indestructible") {
		t.Fatal("four lore counters among my Sagas do not grant hexproof and indestructible")
	}
}

// The chapter's triggering event survives a restore while the final
// chapter waits on the stack, so Tom still sees it resolve.
func TestTomBombadilFinalChapterOnTheStackSurvivesARestore(t *testing.T) {
	g, me, _, tom := tomTable(t)
	saga := tomBenalia(g, me, 2)
	hit := tomTestSaga(me, "Restored Saga")
	tomLibrary(g, me, hit)
	if err := g.AddCounter(saga, game.CounterLore, 1); err != nil {
		t.Fatalf("add a lore counter: %v", err)
	}
	if len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0 {
		t.Fatal("chapter III is not waiting")
	}

	r := restoreRoundTrip(t, g, true)
	passPriorityAroundTable(t, r)

	if got := r.TriggeredThisTurn(tom, tomBombadilLabel); got != 1 {
		t.Fatalf("after the restore Tom triggered %d times, want 1", got)
	}
	if findBattlefieldCardForTest(r, hit.InstanceID) == nil {
		t.Error("after the restore no Saga was put onto the battlefield")
	}
}

// --- Bilbo, Thief in the Night --------------------------------------

func bilboTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bilbo := pushDiesCreatureForTest(g, me.ID, bilboThiefName, bilboThiefOracleID,
		"Legendary Creature — Halfling Rogue", 2, 2)
	return g, me, opp, bilbo
}

func bilboGraveyardCast(g *game.Game, me *game.Player, id uuid.UUID) error {
	return g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true,
	})
}

// The discount applies to a cast from the graveyard and from exile, and not
// to a cast from the hand.
func TestBilboDiscountsCastsFromAnywhereButTheHand(t *testing.T) {
	g, me, _, _ := bilboTable(t)
	toMainForCost(t, g)

	hand := handCardFull(me, "Divination", "Sorcery", "{2}{U}", divinationOracle, []string{"U"})
	floatForTest(g, me, "UU")
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a hand cast of a {2}{U} spell was discounted")
	}
	floatForTest(g, me, "U")
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast from hand for the full {2}{U}: %v", err)
	}
	passPriorityAroundTable(t, g)

	exiled := game.Card{InstanceID: uuid.New(), Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}",
		OracleID: divinationOracle, Owner: me.ID, Controller: me.ID}
	gy := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)
	g.WithWriteLock(func() {
		g.Exile.PushTop(exiled)
		for _, id := range []uuid.UUID{exiled.InstanceID, gy} {
			c, _ := g.LookupCardForEffect(id)
			z := g.FindCardZoneForEffect(id)
			g.GrantCastPermissionToCardsForEffect(game.CastPermission{
				Player: me.ID, Zone: z.Kind, Duration: g.UntilEndOfTurnDuration(),
				AltCostKey: game.AltCostKeyGrantedCost, CastOnly: true, Label: "test grant",
			}, []game.Card{c})
		}
	})
	for _, tc := range []struct {
		id   uuid.UUID
		zone string
	}{{exiled.InstanceID, "exile"}, {gy, "graveyard"}} {
		floatForTest(g, me, "U")
		params := game.CastSpellParams{FromZone: tc.zone, AlternativeCost: game.AltCostKeyGrantedCost, Strict: true}
		if err := g.CastSpell(me.ID, tc.id, params); err == nil {
			t.Fatalf("a cast from %s for {U} alone went through", tc.zone)
		}
		floatForTest(g, me, "U")
		if err := g.CastSpell(me.ID, tc.id, params); err != nil {
			t.Fatalf("cast from %s for {1}{U}: %v", tc.zone, err)
		}
		passPriorityAroundTable(t, g)
	}
}

// The attack trigger offers the artifact, instant and sorcery cards in the
// graveyard, and nothing else.
func TestBilboAttackOffersOnlyArtifactsInstantsAndSorceries(t *testing.T) {
	g, me, opp, bilbo := bilboTable(t)
	div := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)
	inst := gyCard(me, "Test Instant", "Instant", "{1}{U}", "")
	relic := gyCard(me, "Test Relic", "Artifact", "{2}", "")
	bear := gyCard(me, "Bear", "Creature — Bear", "{1}{G}", "")
	land := gyCard(me, "Test Vault", "Artifact Land", "", "")

	declareAttack(t, g, opp.ID, bilbo)
	passPriorityAroundTable(t, g)

	c := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, div, inst, relic)
	if slices.Contains(c.DiscardOptions, bear) || slices.Contains(c.DiscardOptions, land) {
		t.Fatalf("a creature card or a land is offered: %v", c.DiscardOptions)
	}
	if len(c.DiscardOptions) != 3 {
		t.Fatalf("offered %d cards, want 3", len(c.DiscardOptions))
	}
}

// bilboOffer attacks with Bilbo and leaves the cast offer open.
func bilboOffer(t *testing.T, g *game.Game, me, opp *game.Player, bilbo uuid.UUID, want ...uuid.UUID) {
	t.Helper()
	declareAttack(t, g, opp.ID, bilbo)
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, want...)
}

// An instant or sorcery cast this way is exiled instead of going to the
// graveyard, at its mana cost less Bilbo's {1}; the one cast closes the
// offer.
func TestBilboInstantOrSorceryCastThisWayIsExiled(t *testing.T) {
	g, me, opp, bilbo := bilboTable(t)
	div := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)
	inst := gyCard(me, "Test Instant", "Instant", "{1}{U}", "")
	bilboOffer(t, g, me, opp, bilbo, div, inst)

	floatForTest(g, me, "UU")
	if err := bilboGraveyardCast(g, me, div); err != nil {
		t.Fatalf("cast Divination for {1}{U}: %v", err)
	}
	castInsideTheResolution(t, g, div)
	floatForTest(g, me, "U")
	if err := bilboGraveyardCast(g, me, inst); err == nil {
		t.Fatal("a second spell was cast off the one offer")
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(div) || !inExile(g, div) {
		t.Error("Divination cast this way went to the graveyard, want exile")
	}

	// Countered, too.
	g2, me2, opp2, bilbo2 := bilboTable(t)
	inst2 := gyCard(me2, "Test Instant", "Instant", "{1}{U}", "")
	bilboOffer(t, g2, me2, opp2, bilbo2, inst2)
	floatForTest(g2, me2, "U")
	if err := bilboGraveyardCast(g2, me2, inst2); err != nil {
		t.Fatalf("cast the instant for {U}: %v", err)
	}
	g2.WithWriteLock(func() {
		if err := g2.CounterTargetForEffect(inst2); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	if me2.Graveyard.Contains(inst2) || !inExile(g2, inst2) {
		t.Error("the countered instant went to the graveyard, want exile")
	}
}

// The exile clause names instants and sorceries only: an artifact spell
// cast this way that is countered goes to the graveyard.
func TestBilboCounteredArtifactSpellGoesToTheGraveyard(t *testing.T) {
	g, me, opp, bilbo := bilboTable(t)
	relic := gyCard(me, "Test Relic", "Artifact", "{2}", "")
	bilboOffer(t, g, me, opp, bilbo, relic)

	floatForTest(g, me, "C")
	if err := bilboGraveyardCast(g, me, relic); err != nil {
		t.Fatalf("cast the artifact for {1}: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(relic); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	if !me.Graveyard.Contains(relic) || inExile(g, relic) {
		t.Error("the countered artifact spell was exiled, want the graveyard")
	}
}

// Declining is the "you may": the cards stay in the graveyard and the
// offer is gone.
func TestBilboDecliningLeavesTheGraveyard(t *testing.T) {
	g, me, opp, bilbo := bilboTable(t)
	div := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)
	bilboOffer(t, g, me, opp, bilbo, div)

	declineResolutionCast(t, g, me.ID)
	if !me.Graveyard.Contains(div) {
		t.Fatal("the declined card left the graveyard")
	}
	floatForTest(g, me, "UU")
	if err := bilboGraveyardCast(g, me, div); err == nil {
		t.Fatal("the declined offer still allowed a cast")
	}
}

// The offer, and its instant-or-sorcery-only exile, survive a restore
// while it is open.
func TestBilboOfferSurvivesARestore(t *testing.T) {
	g, me, opp, bilbo := bilboTable(t)
	relic := gyCard(me, "Test Relic", "Artifact", "{2}", "")
	bilboOffer(t, g, me, opp, bilbo, relic)

	r := restoreRoundTrip(t, g, true)
	rme := r.Seats[r.Turn.ActiveSeat]
	mustResolutionCastOffer(t, r, rme.ID, game.MayCastKeywordPaidCast, relic)
	floatForTest(r, rme, "C")
	if err := bilboGraveyardCast(r, rme, relic); err != nil {
		t.Fatalf("cast the artifact after the restore: %v", err)
	}
	r.WithWriteLock(func() {
		if err := r.CounterTargetForEffect(relic); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	if !rme.Graveyard.Contains(relic) || inExile(r, relic) {
		t.Error("after a restore the countered artifact spell was exiled, want the graveyard")
	}
}

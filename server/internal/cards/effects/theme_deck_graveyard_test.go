package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// theme_deck_graveyard_test.go — S29's exit criterion (issue #94):
//
//	"Theme-deck smoke test (Muldrotha-style graveyard deck casts via
//	 flashback + escape)"
//
// A scripted engine test rather than a bot game. The bots cannot take
// any of these casts: legal/cast.go enumerates hand and command-zone
// casts only and never offers an alternative cost, so a bot-driven
// run would pass without ever touching the graveyard. The bot version
// waits on #89.
//
// The harness is the S27 smoke test's (theme_deck_smoke_test.go), for
// the same reasons:
//
//  1. THE CARDS COME OFF THE IMPORT ROAD. Every card is a cards.Card
//     row passed through deck.ToGameCard (themeDeckRow and
//     importThemeCard), so the flashback and escape offers are found
//     by the oracle ID the importer stamped, not by a field the test
//     set.
//
//  2. THE MANA IS REAL. Every cast, from hand or from the graveyard,
//     goes through Strict + AutoTap. A flashback or escape cost the
//     deck cannot pay fails here instead of resolving for free, and
//     the colours matter: Lingering Souls' flashback is black and its
//     printed cost is white, and there is no Plains on the table.
//
//  3. THE TURNS ARE REAL. Three of the theme seat's turns at a
//     four-player table, with the other three seats' turns in
//     between. The graveyard carries from turn to turn because
//     nothing resets it.
//
// The concessions: nine lands start on the battlefield rather than
// being played one per turn (a land drop per turn would make the
// test about the mana curve), and the opening hand of fillers goes
// back into the library so the hand holds exactly the four spells and
// no cleanup discard gets in the way. The graveyard is NOT seeded:
// every card in it got there by being milled, discarded or cast.
//
// WHAT IT ASSERTS
//
//	Flashback  cast from the graveyard for the flashback cost; the
//	           card is exiled as it leaves the stack (CR 702.34a), so
//	           it is gone from the graveyard and cannot be cast again
//	           on a later turn; casting from the graveyard does not
//	           lift a sorcery's timing (CR 117.1a).
//	Escape     cast from the graveyard for the escape cost, exiling
//	           exactly the N other cards named (CR 702.138a); an
//	           escaped sorcery goes back to the graveyard on
//	           resolution and can escape again next turn; cards
//	           already exiled to pay one escape cannot pay the next;
//	           a graveyard too thin to pay is refused and the card
//	           stays put.
//
// WHAT IS DELIBERATELY NOT HERE
//
// Granted flashback and escape (Snapcaster Mage, Past in Flames,
// Underworld Breach), "unless it escaped" (Uro, Kroxa, Phlage) and
// madness: none of them has an engine primitive yet, which is the
// tail #94 splits into its own issues. A flashback spell that is
// countered or fizzles is pinned in the game package
// (flashback_test.go there); this file only drives the resolution
// road.

// The oracle keys for the three cards no other test in this package
// names as a constant. Faithless Looting's is faithlessLootingOracle
// in pirates_test.go.
const (
	gyThemeLingeringSoulsOracle = "0b8c3337-04dd-4798-8203-6d8b8cfb936b"
	gyThemeFruitOfTizerusOracle = "8d6ad0a0-3b71-4fab-8874-470285c39299"
	gyThemeSweetOblivionOracle  = "c023538d-2feb-4af5-a44e-b355a190f081"
)

// seedGraveyardThemeDeck puts the lands on the battlefield and the
// four spells in hand, all through deck.ToGameCard, after moving the
// opening hand back into the library. Returns the spells' instance
// IDs by name.
func seedGraveyardThemeDeck(t *testing.T, g *game.Game, p *game.Player) map[string]uuid.UUID {
	t.Helper()

	for len(p.Hand.Cards) > 0 {
		if _, err := game.MoveCard(p.Hand, p.Library, p.Hand.Cards[0].InstanceID); err != nil {
			t.Fatalf("return the opening hand: %v", err)
		}
	}

	// Enough of each colour that no auto-tap order on one turn can
	// strand the next cast: turn 1 needs {U}{R}{R} plus three generic,
	// turn 2 needs {B}{B} plus four.
	lands := []struct {
		name, typeLine, color string
		n                     int
	}{
		{"Swamp", "Basic Land — Swamp", "B", 3},
		{"Mountain", "Basic Land — Mountain", "R", 3},
		{"Island", "Basic Land — Island", "U", 3},
	}
	for _, l := range lands {
		row := themeDeckRow(l.name, "", l.typeLine, "")
		row.ProducedMana = []string{l.color}
		for i := 0; i < l.n; i++ {
			land := importThemeCard(row, p.ID)
			land.InstanceID = uuid.New()
			pushBattlefieldCardWithTimestamp(g, land)
		}
	}

	spells := []cards.Card{
		themeDeckRow("Sweet Oblivion", gyThemeSweetOblivionOracle, "Sorcery", "{1}{U}"),
		themeDeckRow("Faithless Looting", faithlessLootingOracle, "Sorcery", "{R}"),
		themeDeckRow("Lingering Souls", gyThemeLingeringSoulsOracle, "Sorcery", "{2}{W}"),
		themeDeckRow("Fruit of Tizerus", gyThemeFruitOfTizerusOracle, "Sorcery", "{B}"),
	}
	ids := make(map[string]uuid.UUID, len(spells))
	for _, row := range spells {
		c := importThemeCard(row, p.ID)
		ids[c.Name] = c.InstanceID
		p.Hand.PushTop(c)
	}
	return ids
}

// castFromGraveyard claims an alternative cost on a graveyard cast
// with the strict gate and the auto-tapper engaged. It does not
// settle the stack, so a caller can look at the spell mid-flight.
func castFromGraveyard(g *game.Game, p *game.Player, id uuid.UUID, key string, pay []uuid.UUID, targets []game.TargetRef) error {
	return g.CastSpell(p.ID, id, game.CastSpellParams{
		FromZone:        "graveyard",
		AlternativeCost: key,
		AltCostIDs:      pay,
		Targets:         targets,
		Strict:          true,
		AutoTap:         true,
	})
}

// discardLoot answers Faithless Looting's "then discard two cards"
// with the named cards. Since #651 the loot leaves a real prompt
// behind and the table is gated until it is answered, so the script
// cannot walk on without paying it.
func discardLoot(t *testing.T, g *game.Game, p *game.Player, ids ...uuid.UUID) {
	t.Helper()
	if got := discardOwed(g, p.ID); got != len(ids) {
		t.Fatalf("discards owed after the loot = %d, want %d", got, len(ids))
	}
	answerDiscard(t, g, p.ID, ids...)
}

// otherGraveyardCards returns up to n cards in p's graveyard that are
// not in `except`, oldest first. Oldest-first is only so a failure
// message is reproducible; which fodder pays is not load-bearing.
func otherGraveyardCards(p *game.Player, n int, except ...uuid.UUID) []uuid.UUID {
	skip := make(map[uuid.UUID]bool, len(except))
	for _, id := range except {
		skip[id] = true
	}
	var out []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if len(out) == n {
			break
		}
		if !skip[c.InstanceID] {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// TestGraveyardThemeDeckCastsFromTheGraveyardAcrossTurns is S29's
// exit criterion. See the file comment for the harness.
func TestGraveyardThemeDeckCastsFromTheGraveyardAcrossTurns(t *testing.T) {
	g := newCatalogGame(t)
	themeSeat := g.Turn.ActiveSeat
	me := g.Seats[themeSeat]
	opponent := g.Seats[(themeSeat+1)%len(g.Seats)]
	hand := seedGraveyardThemeDeck(t, g, me)

	oblivion := hand["Sweet Oblivion"]
	looting := hand["Faithless Looting"]
	souls := hand["Lingering Souls"]
	fruit := hand["Fruit of Tizerus"]
	drain := []game.TargetRef{{Kind: game.TargetPlayer, ID: opponent.ID}}

	// --- turn 1: fill the graveyard, flash back the loot -------------
	advanceToMainOf(t, g, themeSeat)
	// Turn.Number counts rounds of the table, so it goes up by one
	// between each of the theme seat's turns.
	firstRound := g.Turn.Number
	if me.Graveyard.Size() != 0 {
		t.Fatalf("graveyard before the first cast = %d cards, want 0", me.Graveyard.Size())
	}

	// Sweet Oblivion on yourself: four cards milled, and the spell
	// joins them.
	if err := g.CastSpell(me.ID, oblivion, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
		Strict:  true, AutoTap: true,
	}); err != nil {
		t.Fatalf("cast Sweet Oblivion: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != 5 {
		t.Fatalf("graveyard after milling four = %d, want 5 (four milled + Sweet Oblivion)", got)
	}

	// Faithless Looting from hand for {R}, discarding the two cards
	// the rest of the script casts out of the graveyard.
	castThemeSpell(t, g, me, looting)
	discardLoot(t, g, me, souls, fruit)
	if !me.Graveyard.Contains(looting) {
		t.Fatal("Faithless Looting cast from hand did not go to the graveyard")
	}
	for name, id := range map[string]uuid.UUID{"Lingering Souls": souls, "Fruit of Tizerus": fruit} {
		if !me.Graveyard.Contains(id) {
			t.Fatalf("the discarded %s is not in the graveyard", name)
		}
	}

	// Flashback the loot for {2}{R}. The same OnResolve runs, and the
	// card goes to exile instead of back to the graveyard.
	handBefore := me.Hand.Size()
	if err := castFromGraveyard(g, me, looting, "flashback", nil, nil); err != nil {
		t.Fatalf("flashback Faithless Looting: %v", err)
	}
	if me.Graveyard.Contains(looting) {
		t.Error("Faithless Looting is still in the graveyard after being cast from it")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand after the flashback draw = %d, want %d", got, handBefore+2)
	}
	discardLoot(t, g, me, me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID)
	if !g.Exile.Contains(looting) {
		t.Error("a flashed-back sorcery did not go to exile (CR 702.34a)")
	}
	if me.Graveyard.Contains(looting) {
		t.Error("a flashed-back sorcery returned to the graveyard (CR 702.34a)")
	}
	// Four milled, Sweet Oblivion, Lingering Souls, Fruit of Tizerus,
	// and the two cards the flashback loot discarded.
	if got := me.Graveyard.Size(); got != 9 {
		t.Fatalf("graveyard at the end of turn 1 = %d, want 9", got)
	}

	// --- turn 2: flashback is sorcery-speed; escape the drain --------
	advanceToStepOf(t, g, themeSeat, game.StepUpkeep)
	err := castFromGraveyard(g, me, souls, "flashback", nil, nil)
	if !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("flashback of a sorcery in upkeep returned %v, want ErrSorcerySpeedRequired", err)
	}
	if !me.Graveyard.Contains(souls) {
		t.Fatal("a refused flashback moved Lingering Souls out of the graveyard")
	}

	advanceToMainOf(t, g, themeSeat)
	if g.Turn.Number != firstRound+1 {
		t.Fatalf("round at the second main phase = %d, want %d", g.Turn.Number, firstRound+1)
	}
	// Flashback is {1}{B}, off a table with no white source: the
	// printed {2}{W} could not have been paid.
	if err := castFromGraveyard(g, me, souls, "flashback", nil, nil); err != nil {
		t.Fatalf("flashback Lingering Souls: %v", err)
	}
	passPriorityAroundTable(t, g)
	spirits := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.Controller == me.ID {
			spirits++
		}
	}
	if spirits != 2 {
		t.Errorf("Spirit tokens after flashback = %d, want 2", spirits)
	}
	if !g.Exile.Contains(souls) || me.Graveyard.Contains(souls) {
		t.Error("the flashed-back Lingering Souls is not in exile (CR 702.34a)")
	}

	// Escape Fruit of Tizerus: {3}{B} and three OTHER cards.
	firstPayment := otherGraveyardCards(me, 3, fruit, oblivion)
	lifeBefore := opponent.Life
	if err := castFromGraveyard(g, me, fruit, "escape", firstPayment, drain); err != nil {
		t.Fatalf("escape Fruit of Tizerus: %v", err)
	}
	for _, id := range firstPayment {
		if !g.Exile.Contains(id) {
			t.Error("a card named to pay the escape cost is not in exile (CR 702.138a)")
		}
	}
	passPriorityAroundTable(t, g)
	if got := lifeBefore - opponent.Life; got != 2 {
		t.Errorf("life lost to the escaped Fruit = %d, want 2", got)
	}
	if !me.Graveyard.Contains(fruit) {
		t.Fatal("an escaped sorcery did not return to the graveyard — escape has no exile clause")
	}
	if g.Exile.Contains(fruit) {
		t.Error("an escaped sorcery went to exile")
	}
	// Nine, less Lingering Souls, less the three paid. The Fruit left
	// and came back.
	if got := me.Graveyard.Size(); got != 5 {
		t.Fatalf("graveyard at the end of turn 2 = %d, want 5", got)
	}

	// --- turn 3: escape again; flashback stays spent -----------------
	//
	// Via upkeep: advanceToMainOf returns at once when it is already
	// this seat's precombat main, which would replay turn 2.
	advanceToStepOf(t, g, themeSeat, game.StepUpkeep)
	advanceToMainOf(t, g, themeSeat)
	if g.Turn.Number != firstRound+2 {
		t.Fatalf("round at the third main phase = %d, want %d", g.Turn.Number, firstRound+2)
	}

	// The flashed-back cards are in exile, a zone neither declares
	// castable, and the graveyard no longer holds them.
	for name, id := range map[string]uuid.UUID{"Faithless Looting": looting, "Lingering Souls": souls} {
		if err := castFromGraveyard(g, me, id, "flashback", nil, nil); err == nil {
			t.Errorf("%s was flashed back a second time", name)
		}
		if !g.Exile.Contains(id) {
			t.Errorf("%s left exile", name)
		}
	}

	// The cards that paid the first escape are exiled and cannot pay
	// the second.
	lifeBefore = opponent.Life
	if err := castFromGraveyard(g, me, fruit, "escape", firstPayment, drain); err == nil {
		t.Fatal("an escape cost was paid with cards already in exile")
	}
	if !me.Graveyard.Contains(fruit) {
		t.Fatal("the refused escape moved Fruit of Tizerus out of the graveyard")
	}

	secondPayment := otherGraveyardCards(me, 3, fruit, oblivion)
	if err := castFromGraveyard(g, me, fruit, "escape", secondPayment, drain); err != nil {
		t.Fatalf("escape Fruit of Tizerus a second time: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeBefore - opponent.Life; got != 2 {
		t.Errorf("life lost to the second escape = %d, want 2", got)
	}
	if !me.Graveyard.Contains(fruit) {
		t.Fatal("the twice-escaped sorcery did not return to the graveyard")
	}
	for _, id := range append(append([]uuid.UUID{}, firstPayment...), secondPayment...) {
		if !g.Exile.Contains(id) {
			t.Error("a card paid to an escape cost is not in exile")
		}
	}

	// Sweet Oblivion and the Fruit are all that is left: the Fruit's
	// three-other-cards cost can no longer be paid, and the card stays
	// where it is.
	if got := me.Graveyard.Size(); got != 2 {
		t.Fatalf("graveyard after the second escape = %d, want 2", got)
	}
	if err := castFromGraveyard(g, me, fruit, "escape", []uuid.UUID{oblivion}, drain); err == nil {
		t.Error("escape was paid with one other card when the cost is three")
	}
	if !me.Graveyard.Contains(fruit) || !me.Graveyard.Contains(oblivion) {
		t.Error("the refused escape moved a card out of the graveyard")
	}
}

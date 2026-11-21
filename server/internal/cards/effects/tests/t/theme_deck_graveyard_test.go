package testst

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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
		row := effectstest.ThemeDeckRow(l.name, "", l.typeLine, "")
		row.ProducedMana = []string{l.color}
		for i := 0; i < l.n; i++ {
			land := effectstest.ImportThemeCard(row, p.ID)
			land.InstanceID = uuid.New()
			effectstest.PushBattlefieldCardWithTimestamp(g, land)
		}
	}

	spells := []cards.Card{
		effectstest.ThemeDeckRow("Sweet Oblivion", effectstest.GyThemeSweetOblivionOracle, "Sorcery", "{1}{U}"),
		effectstest.ThemeDeckRow("Faithless Looting", effectstest.FaithlessLootingOracle, "Sorcery", "{R}"),
		effectstest.ThemeDeckRow("Lingering Souls", effectstest.GyThemeLingeringSoulsOracle, "Sorcery", "{2}{W}"),
		effectstest.ThemeDeckRow("Fruit of Tizerus", effectstest.GyThemeFruitOfTizerusOracle, "Sorcery", "{B}"),
	}
	ids := make(map[string]uuid.UUID, len(spells))
	for _, row := range spells {
		c := effectstest.ImportThemeCard(row, p.ID)
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
	if got := effectstest.DiscardOwed(g, p.ID); got != len(ids) {
		t.Fatalf("discards owed after the loot = %d, want %d", got, len(ids))
	}
	effectstest.AnswerDiscard(t, g, p.ID, ids...)
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
	g := effectstest.NewCatalogGame(t)
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
	effectstest.AdvanceToMainOf(t, g, themeSeat)
	// Turn.Round counts rounds of the table, so it goes up by one
	// between each of the theme seat's turns.
	firstRound := g.Turn.Round
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
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != 5 {
		t.Fatalf("graveyard after milling four = %d, want 5 (four milled + Sweet Oblivion)", got)
	}

	// Faithless Looting from hand for {R}, discarding the two cards
	// the rest of the script casts out of the graveyard.
	effectstest.CastThemeSpell(t, g, me, looting)
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
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.AdvanceToStepOf(t, g, themeSeat, game.StepUpkeep)
	err := castFromGraveyard(g, me, souls, "flashback", nil, nil)
	if !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("flashback of a sorcery in upkeep returned %v, want ErrSorcerySpeedRequired", err)
	}
	if !me.Graveyard.Contains(souls) {
		t.Fatal("a refused flashback moved Lingering Souls out of the graveyard")
	}

	effectstest.AdvanceToMainOf(t, g, themeSeat)
	if g.Turn.Round != firstRound+1 {
		t.Fatalf("round at the second main phase = %d, want %d", g.Turn.Round, firstRound+1)
	}
	// Flashback is {1}{B}, off a table with no white source: the
	// printed {2}{W} could not have been paid.
	if err := castFromGraveyard(g, me, souls, "flashback", nil, nil); err != nil {
		t.Fatalf("flashback Lingering Souls: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.AdvanceToStepOf(t, g, themeSeat, game.StepUpkeep)
	effectstest.AdvanceToMainOf(t, g, themeSeat)
	if g.Turn.Round != firstRound+2 {
		t.Fatalf("round at the third main phase = %d, want %d", g.Turn.Round, firstRound+2)
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
	effectstest.PassPriorityAroundTable(t, g)
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

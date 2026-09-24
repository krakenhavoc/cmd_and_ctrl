package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// deep_gnome_terramancer_test.go — the proof card for #1326 (CR
// 305.4): "without being played" now reads a real engine marker
// (Event.Played) instead of nothing at all.

const deepGnomeTerramancerOracle = "d4e3440d-4e34-40d7-8a42-c673225c0332"

// A land RETURNED to the battlefield under an opponent's control by
// an effect — never played — is exactly what Mold Earth watches for.
func TestDeepGnomeTerramancerSearchesOnAnUnplayedOpponentLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Deep Gnome Terramancer", "Creature — Gnome Wizard", deepGnomeTerramancerOracle, 2, 2)
	ids := seedSearchLibrary(me, game.Card{Name: "Hallowed Fountain", TypeLine: "Land — Plains Island"})

	buried := b17GraveyardCard(opp, "Island", "Basic Land — Island", "")
	g.WithWriteLock(func() { _ = g.ReturnFromGraveyardUnderControlForEffect(buried, game.ZoneBattlefield, opp.ID) })
	passPriorityAroundTable(t, g)

	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Mold Earth should offer the search after a land was PUT under an opponent's control")
	}
	answerSearchNamed(t, g, me.ID, "Hallowed Fountain")
	if !g.Battlefield.Contains(ids[0]) {
		t.Fatal("the picked Plains card lands on the battlefield")
	}
	if !b16Tapped(t, g, ids[0]) {
		t.Error("the fetched land enters tapped")
	}
}

// CR 305.4's whole point: an opponent's ORDINARY land drop is a PLAY,
// and Mold Earth stays quiet for it.
func TestDeepGnomeTerramancerStaysQuietOnAnOrdinaryOpponentLandPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Deep Gnome Terramancer", "Creature — Gnome Wizard", deepGnomeTerramancerOracle, 2, 2)

	b13PlayAs(t, g, 1, "Mountain", "Basic Land — Mountain", "")
	passPriorityAroundTable(t, g)

	if c := searchChoiceFor(g, me.ID); c != nil {
		t.Error("a PLAYED land must not trigger Mold Earth (CR 305.4)")
	}
}

// "Under an opponent's control" excludes the Terramancer's own
// controller, played or put.
func TestDeepGnomeTerramancerStaysQuietOnYourOwnLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Deep Gnome Terramancer", "Creature — Gnome Wizard", deepGnomeTerramancerOracle, 2, 2)

	buried := b17GraveyardCard(me, "Island", "Basic Land — Island", "")
	g.WithWriteLock(func() { _ = g.ReturnFromGraveyardUnderControlForEffect(buried, game.ZoneBattlefield, me.ID) })
	passPriorityAroundTable(t, g)

	if c := searchChoiceFor(g, me.ID); c != nil {
		t.Error("Mold Earth only watches an opponent's lands, never your own")
	}
}

// "One or more lands" collapses a simultaneous entry into ONE
// trigger, and "do this only once each turn" gates out a second,
// later batch in the same turn.
func TestDeepGnomeTerramancerCollapsesOneOrMoreAndOnlyOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dgt := b12Push(g, me.ID, "Deep Gnome Terramancer", "Creature — Gnome Wizard", deepGnomeTerramancerOracle, 2, 2)

	island := b17GraveyardCard(opp, "Island", "Basic Land — Island", "")
	swamp := b17GraveyardCard(opp, "Swamp", "Basic Land — Swamp", "")
	g.WithWriteLock(func() {
		// Nothing resolves and no step advances between these two
		// calls, so both land entries share one event batch (#829) —
		// exactly the "one or more" the printed text describes.
		_ = g.ReturnFromGraveyardUnderControlForEffect(island, game.ZoneBattlefield, opp.ID)
		_ = g.ReturnFromGraveyardUnderControlForEffect(swamp, game.ZoneBattlefield, opp.ID)
	})
	// The harvest runs synchronously off each EmitEvent, so the
	// OncePerBatch dedup has already applied by the time both calls
	// above return: PendingTriggers holds the one trigger before
	// priority ever passes to promote it onto the stack.
	count := 0
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == dgt {
			count++
		}
	}
	if count != 1 {
		t.Errorf("two lands entering in one batch is ONE trigger, got %d", count)
	}
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		answerSearchFailToFind(t, g, me.ID)
	}

	// A later, SEPARATE batch this turn must not trigger again.
	forest := b17GraveyardCard(opp, "Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() { _ = g.ReturnFromGraveyardUnderControlForEffect(forest, game.ZoneBattlefield, opp.ID) })
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		t.Error("\"only once each turn\" should gate a second batch out")
	}
}

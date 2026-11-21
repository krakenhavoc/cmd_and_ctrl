package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const shareTheSpoilsOracle = "d39af66c-4997-48b8-a704-f21d93247d7d"

func TestShareTheSpoilsIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(shareTheSpoilsOracle)
	if !ok || spec.Name != "Share the Spoils" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// shareTheSpoilsTable casts Share the Spoils for the active player and
// resolves its enters trigger, after seeding the top card of each
// player's library with tops[i] (a sorcery unless it names a land).
// It returns the enchantment and, per seat, the card exiled from it.
func shareTheSpoilsTable(t *testing.T, tops []struct{ name, typeLine, cost string }) (*game.Game, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	exiled := make([]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats {
		_ = effectstest.TopOfLibrary(p, "Deeper Card", "Sorcery", "{B}")
		exiled[i] = effectstest.TopOfLibrary(p, tops[i].name, tops[i].typeLine, tops[i].cost)
	}
	spoils := effectstest.RfdHandCard(me, "Share the Spoils", "Enchantment", "{1}{R}", shareTheSpoilsOracle, 0, 0)
	if err := g.CastSpell(me.ID, spoils, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Share the Spoils: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g) // the enchantment resolves
	effectstest.PassPriorityAroundTable(t, g) // its enters trigger resolves
	return g, spoils, exiled
}

func shareTheSpoilsTops(g4 ...[3]string) []struct{ name, typeLine, cost string } {
	out := make([]struct{ name, typeLine, cost string }, len(g4))
	for i, x := range g4 {
		out[i] = struct{ name, typeLine, cost string }{x[0], x[1], x[2]}
	}
	return out
}

func shareTheSpoilsPlay(g *game.Game, p uuid.UUID, card uuid.UUID) error {
	return g.CastSpell(p, card, game.CastSpellParams{FromZone: "exile", Strict: true})
}

// The enters trigger exiles the top card of each player's library,
// linked to Share the Spoils. On their turn, the active player casts one
// of them (any player's) spending mana as though it were any colour;
// when they do, the top card of their own library is exiled with it.
// That is their one play for the turn: no second spell, and no land.
func TestShareTheSpoilsEntersExilesEachTopAndAPlayExilesTheirTop(t *testing.T) {
	tops := shareTheSpoilsTops(
		[3]string{"Seat0 Spell", "Sorcery", "{G}"},
		[3]string{"Seat1 Spell", "Sorcery", "{U}"},
		[3]string{"Seat2 Land", "Basic Land — Forest", ""},
		[3]string{"Seat3 Spell", "Sorcery", "{W}"},
	)
	g, spoils, exiled := shareTheSpoilsTable(t, tops)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	for i, id := range exiled {
		if got := effectstest.CardInExile(t, g, id).ExiledWith.ID; got != spoils {
			t.Fatalf("seat %d's top card linked to %s, want Share the Spoils", i, got)
		}
	}
	other := exiled[(seat+1)%len(g.Seats)]
	land := exiled[2]
	if seat == 2 {
		t.Fatal("test table assumes seat 2 is not active")
	}
	// {R} pays a {U} or {G} or {W} sorcery: any colour.
	nextTop := effectstest.TopOfLibrary(me, "My Next Top", "Sorcery", "{R}")
	effectstest.FloatMana(t, g, me, "{R}")
	if !effectstest.EnumeratedFor(g, me.ID, other) {
		t.Fatal("the bot is offered the exiled spell")
	}
	if !effectstest.ViewCastable(t, g, me.ID, other) {
		t.Fatal("the client is offered the exiled spell")
	}
	if err := shareTheSpoilsPlay(g, me.ID, other); err != nil {
		t.Fatalf("cast an exiled spell with off-colour mana: %v", err)
	}
	// The "when they do" trigger goes on the stack above the spell.
	if effectstest.E2Zone(g, nextTop) != game.ZoneLibrary {
		t.Fatal("the top card left before the trigger resolved")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CardInExile(t, g, nextTop).ExiledWith.ID; got != spoils {
		t.Fatalf("their new top card linked to %s, want Share the Spoils", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.E2Zone(g, other) != game.ZoneGraveyard {
		t.Fatalf("the cast spell is in %q, want its owner's graveyard", effectstest.E2Zone(g, other))
	}

	// One land or one spell, not one of each.
	effectstest.FloatMana(t, g, me, "{R}")
	if err := shareTheSpoilsPlay(g, me.ID, nextTop); err == nil {
		t.Error("a second spell from the pile in one turn")
	}
	if err := shareTheSpoilsPlay(g, me.ID, land); err == nil {
		t.Error("a land from the pile after a spell this turn")
	}
	if effectstest.EnumeratedFor(g, me.ID, nextTop) {
		t.Error("the spent use is still offered")
	}
}

// Another player may play from the pile only during their own turn,
// with normal timing: a land needs their main phase. Their land play
// spends their use for that turn and exiles the top card of THEIR
// library. Each player's use is their own: the active player's play
// does not spend the next player's.
func TestShareTheSpoilsEachPlayerOnTheirOwnTurn(t *testing.T) {
	tops := shareTheSpoilsTops(
		[3]string{"Seat0 Land", "Basic Land — Mountain", ""},
		[3]string{"Seat1 Land", "Basic Land — Island", ""},
		[3]string{"Seat2 Spell", "Instant", "{B}"},
		[3]string{"Seat3 Spell", "Sorcery", "{W}"},
	)
	g, spoils, exiled := shareTheSpoilsTable(t, tops)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	nextSeat := (seat + 1) % len(g.Seats)
	opp := g.Seats[nextSeat]
	instant := exiled[2]
	land := exiled[nextSeat]

	// Not on my turn, even an instant.
	effectstest.FloatMana(t, g, opp, "{B}")
	if err := shareTheSpoilsPlay(g, opp.ID, instant); err == nil {
		t.Fatal("an opponent cast from the pile during another player's turn")
	}
	// I spend my use on a spell; it is mine, not theirs.
	if seat == 2 {
		t.Fatal("test table assumes seat 2 is not active")
	}
	effectstest.FloatMana(t, g, me, "{B}")
	if err := shareTheSpoilsPlay(g, me.ID, instant); err != nil {
		t.Fatalf("cast the instant on my turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceToUpkeepOf(t, g, nextSeat)
	if err := shareTheSpoilsPlay(g, opp.ID, land); err == nil {
		t.Fatal("a land played from the pile in upkeep")
	}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if !effectstest.EnumeratedFor(g, opp.ID, land) {
		t.Fatal("the land is offered to its player on their turn")
	}
	theirTop := effectstest.TopOfLibrary(opp, "Their Next Top", "Sorcery", "{U}")
	if err := shareTheSpoilsPlay(g, opp.ID, land); err != nil {
		t.Fatalf("play the land from the pile on their turn: %v", err)
	}
	if effectstest.E2Zone(g, land) != game.ZoneBattlefield {
		t.Fatalf("the land is in %q, want the battlefield", effectstest.E2Zone(g, land))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CardInExile(t, g, theirTop).ExiledWith.ID; got != spoils {
		t.Fatalf("their top card linked to %s after the land play, want Share the Spoils", got)
	}
	effectstest.FloatMana(t, g, opp, "{U}")
	if err := shareTheSpoilsPlay(g, opp.ID, theirTop); err == nil {
		t.Error("a spell from the pile after a land this turn")
	}
}

// Whenever an opponent loses the game, the top card of each remaining
// player's library is exiled. Once Share the Spoils leaves, the exiled
// cards can't be played, and a new Share the Spoils has no memory of
// them.
func TestShareTheSpoilsOpponentLosesAndLeaving(t *testing.T) {
	tops := shareTheSpoilsTops(
		[3]string{"Seat0 Spell", "Instant", "{R}"},
		[3]string{"Seat1 Spell", "Instant", "{R}"},
		[3]string{"Seat2 Spell", "Instant", "{R}"},
		[3]string{"Seat3 Spell", "Instant", "{R}"},
	)
	g, spoils, exiled := shareTheSpoilsTable(t, tops)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	loser := g.Seats[(seat+2)%len(g.Seats)]
	survivor := g.Seats[(seat+1)%len(g.Seats)]
	myTop := effectstest.TopOfLibrary(me, "Mine After", "Instant", "{R}")
	theirTop := effectstest.TopOfLibrary(survivor, "Theirs After", "Instant", "{R}")
	if err := g.Concede(loser.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{myTop, theirTop} {
		if got := effectstest.CardInExile(t, g, id).ExiledWith.ID; got != spoils {
			t.Fatalf("after an opponent lost, a top card is linked to %s, want Share the Spoils", got)
		}
	}

	// Instants, so the checks below do not lean on the step: playing
	// from the pile is open whenever the card's own timing is.
	effectstest.FloatMana(t, g, me, "{R}")
	if !effectstest.EnumeratedFor(g, me.ID, myTop) {
		t.Fatal("non-vacuity: the card is offered while Share the Spoils stays")
	}
	effectstest.B27Kill(g, spoils)
	if effectstest.EnumeratedFor(g, me.ID, myTop) {
		t.Error("a card from the pile is offered after Share the Spoils left")
	}
	if err := shareTheSpoilsPlay(g, me.ID, exiled[(seat+1)%len(g.Seats)]); err == nil {
		t.Fatal("played a card from the pile after Share the Spoils left")
	}

	// A new Share the Spoils is a new object with no claim on them.
	effectstest.B12Push(g, me.ID, "Share the Spoils", "Enchantment", shareTheSpoilsOracle, 0, 0)
	if effectstest.EnumeratedFor(g, me.ID, myTop) {
		t.Error("a new Share the Spoils opened the old one's cards")
	}
}

// The once-each-turn use is kept per player and survives a save and
// restore, as does the "when they do" trigger waiting on the stack.
func TestShareTheSpoilsUseSurvivesARestore(t *testing.T) {
	tops := shareTheSpoilsTops(
		[3]string{"Seat0 Spell", "Sorcery", "{G}"},
		[3]string{"Seat1 Spell", "Sorcery", "{G}"},
		[3]string{"Seat2 Spell", "Sorcery", "{G}"},
		[3]string{"Seat3 Spell", "Sorcery", "{G}"},
	)
	g, spoils, exiled := shareTheSpoilsTable(t, tops)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	first, second := exiled[(seat+1)%len(g.Seats)], exiled[(seat+2)%len(g.Seats)]
	nextTop := effectstest.TopOfLibrary(me, "My Next Top", "Sorcery", "{G}")
	effectstest.FloatMana(t, g, me, "{R}")
	if err := shareTheSpoilsPlay(g, me.ID, first); err != nil {
		t.Fatalf("cast from the pile: %v", err)
	}
	restored := effectstest.RestoreRoundTrip(t, g, true)
	effectstest.PassPriorityAroundTable(t, restored)
	if got := effectstest.CardInExile(t, restored, nextTop).ExiledWith.ID; got != spoils {
		t.Fatalf("restored: the trigger linked their top card to %s, want Share the Spoils", got)
	}
	effectstest.PassPriorityAroundTable(t, restored)
	rme := restored.Seats[seat]
	effectstest.FloatMana(t, restored, rme, "{R}")
	if err := shareTheSpoilsPlay(restored, rme.ID, second); err == nil {
		t.Error("restored: a second play from the pile this turn")
	}
}

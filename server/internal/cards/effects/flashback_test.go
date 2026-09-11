package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flashback_test.go — S29. The game package pins the mechanic
// (cast_zones_test.go, flashback_test.go there); this file pins the
// three cards end to end, against the REAL catalog rather than a
// stubbed hook. That is the half a hook stub cannot check: that the
// card file actually declared the zone, that Register accepted the
// pair, and that the offer the client would be shown is the one the
// cast path accepts.

// seedGraveyardCard puts a card in the active seat's graveyard at a
// main phase and returns its instance ID. The mirror of
// castCatalogSpell's hand seeding.
func seedGraveyardCard(t *testing.T, g *game.Game, name, typeLine, oracleID string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id
}

// Faithless Looting — the simplest flashback card there is, and the
// one that proves the two declarations compose: the graveyard cast
// runs the same OnResolve the hand cast runs, and the card ends in
// exile rather than back where it started.
func TestFaithlessLootingFlashback(t *testing.T) {
	const oracle = "3d6fa57a-aa53-4b5c-b8af-a7612c823117"
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	id := seedGraveyardCard(t, g, "Faithless Looting", "Sorcery", oracle)
	handBefore := active.Hand.Size()

	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	// Draw two, then queue the discard of two against the post-draw
	// hand (the discard is a PendingChoice — see
	// TestFaithlessLootingDrawsThenQueuesTheDiscard). The point here
	// is only that the graveyard cast ran the same OnResolve a hand
	// cast runs.
	if got := active.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand size after the draw: got %d, want %d", got, handBefore+2)
	}
	if !g.Exile.Contains(id) {
		t.Errorf("the flashed-back sorcery is not in exile")
	}
	if active.Graveyard.Contains(id) {
		t.Errorf("the flashed-back sorcery returned to the graveyard")
	}
}

// The offer the client is shown is the offer the server accepts.
// This is the assertion that catches a card file that declared the
// cost but forgot the zone (or vice versa) — Register panics on one
// direction of that mistake, and this catches the other.
func TestFlashbackCardsDeclareBothHalves(t *testing.T) {
	cards := []struct {
		name   string
		oracle string
		cost   string
	}{
		{"Faithless Looting", "3d6fa57a-aa53-4b5c-b8af-a7612c823117", "{2}{R}"},
		{"Lingering Souls", "0b8c3337-04dd-4798-8203-6d8b8cfb936b", "{1}{B}"},
		{"Cackling Counterpart", "9e2adca5-f39c-4a09-bcce-8238ebac2c4a", "{5}{U}{U}"},
	}
	for _, c := range cards {
		t.Run(c.name, func(t *testing.T) {
			if !game.CardCastableFromZone(c.oracle, game.ZoneGraveyard) {
				t.Fatalf("%s does not declare the graveyard castable", c.name)
			}
			offers := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneGraveyard)
			if len(offers) != 1 || offers[0].Key != "flashback" {
				t.Fatalf("%s graveyard offers = %+v, want one flashback", c.name, offers)
			}
			if offers[0].ManaCost != c.cost {
				t.Errorf("%s flashback cost: got %q, want %q", c.name, offers[0].ManaCost, c.cost)
			}
			if !offers[0].ExileOnLeavingStack {
				t.Errorf("%s flashback does not exile on leaving the stack", c.name)
			}
			// And nothing is claimable from hand: flashback is the
			// only offer these cards make, and it is not one of them.
			if got := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneHand); len(got) != 0 {
				t.Errorf("%s offers %+v from hand", c.name, got)
			}
		})
	}
}

// Lingering Souls — two Spirits with flying, and the reason flashback
// is a cast path rather than a second mana cost: the printed cost is
// white and the flashback cost is black.
func TestLingeringSoulsFlashbackMakesTwoSpirits(t *testing.T) {
	const oracle = "0b8c3337-04dd-4798-8203-6d8b8cfb936b"
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	id := seedGraveyardCard(t, g, "Lingering Souls", "Sorcery", oracle)

	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	spirits := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.Controller == active.ID {
			spirits++
			if !game.HasKeyword(&c, "flying") {
				t.Errorf("Spirit token has no flying")
			}
		}
	}
	if spirits != 2 {
		t.Errorf("Spirit tokens: got %d, want 2", spirits)
	}
	if !g.Exile.Contains(id) {
		t.Errorf("the flashed-back sorcery is not in exile")
	}
}

// Cackling Counterpart — the first flashback card with a target
// clause. A graveyard cast has to validate its target exactly as a
// hand cast does, which is the whole reason the client routes the
// graveyard cast through the Board's prompt chain.
func TestCacklingCounterpartFlashbackCopiesTheTarget(t *testing.T) {
	const oracle = "9e2adca5-f39c-4a09-bcce-8238ebac2c4a"
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]

	bear := game.NewCard("Grizzly Bears", active.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Power, bear.Toughness = 2, 2
	bear.Controller = active.ID
	g.Battlefield.PushTop(bear)

	id := seedGraveyardCard(t, g, "Cackling Counterpart", "Instant", oracle)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: bear.InstanceID}}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Targets: targets,
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	copies := 0
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID != bear.InstanceID && c.Name == "Grizzly Bears" {
			copies++
			if !IsToken(c) {
				t.Errorf("the copy is not a token: %q", c.TypeLine)
			}
		}
	}
	if copies != 1 {
		t.Errorf("token copies: got %d, want 1", copies)
	}
	if !g.Exile.Contains(id) {
		t.Errorf("the flashed-back instant is not in exile")
	}
}

// The other direction, on a real card: the flashback cost is not
// claimable from hand, and a graveyard cast that claims nothing is
// refused rather than charged the printed cost.
func TestFaithlessLootingRefusesTheWrongCastPaths(t *testing.T) {
	const oracle = "3d6fa57a-aa53-4b5c-b8af-a7612c823117"

	t.Run("flashback from hand", func(t *testing.T) {
		g := newCatalogGame(t)
		active := g.Seats[g.Turn.ActiveSeat]
		id := uuid.New()
		active.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Faithless Looting", TypeLine: "Sorcery",
			OracleID: oracle, Owner: active.ID, Controller: active.ID,
		})
		for g.Turn.Step != game.StepPrecombatMain {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		err := g.CastSpell(active.ID, id, game.CastSpellParams{AlternativeCost: "flashback"})
		if err != game.ErrCastZoneNotAllowed {
			t.Fatalf("got %v, want ErrCastZoneNotAllowed", err)
		}
	})

	t.Run("graveyard cast at the printed cost", func(t *testing.T) {
		g := newCatalogGame(t)
		active := g.Seats[g.Turn.ActiveSeat]
		id := seedGraveyardCard(t, g, "Faithless Looting", "Sorcery", oracle)
		err := g.CastSpell(active.ID, id, game.CastSpellParams{FromZone: "graveyard"})
		if err != game.ErrCastCostRequired {
			t.Fatalf("got %v, want ErrCastCostRequired", err)
		}
	})
}

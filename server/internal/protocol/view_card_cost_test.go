package protocol

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// view_card_cost_test.go pins #2113: the per-card costs the view paid
// after #1498.
//
//   - A land's CR 305.6 intrinsic mana ability used to allocate a fresh
//     one-row slice (and an origin row) every time the view asked, for
//     every land in every zone. The one-type land now gets a shared,
//     read-only row (game/mutations.go, intrinsicLandRows).
//   - viewOfCard took game.Card (over a kilobyte) by value and called
//     about ten by-value helpers, each copying it again. It takes a
//     *game.Card now, and the helpers it calls have pointer twins.

// landTable returns a two-seat game whose every seat holds `bf` lands on
// the battlefield (a mix of one-type basics and a two-type land), and
// `off` cards in each of hand, graveyard and exile (a mix of lands and
// non-lands) — the shape of a mid-game commander table, where the lands
// outnumber everything else.
func landTable(t testing.TB, bf, off int) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := range 2 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := range 10 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	basics := []game.Card{
		basicLand("Forest", "Forest"),
		basicLand("Plains", "Plains"),
		basicLand("Island", "Island"),
		basicLand("Swamp", "Swamp"),
		basicLand("Mountain", "Mountain"),
		// Two land types: the multi-row path, which still allocates.
		{Name: "Sandsteppe Citadel", TypeLine: "Land — Plains Forest"},
	}
	others := []game.Card{
		{Name: "Elvish Visionary", TypeLine: "Creature — Elf Shaman", ManaCost: "{1}{G}", Power: 1, Toughness: 1},
		{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"},
	}
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			for i := 0; i < bf; i++ {
				c := basics[i%len(basics)]
				c.Name = fmt.Sprintf("%s %d", c.Name, i)
				put(g.Battlefield, p, c)
			}
			for i := 0; i < off; i++ {
				var c game.Card
				if i%3 == 2 {
					c = others[i%len(others)]
				} else {
					c = basics[i%len(basics)]
				}
				c.Name = fmt.Sprintf("%s %d", c.Name, i)
				put(p.Hand, p, c)
				put(p.Graveyard, p, c)
				put(g.Exile, p, c)
			}
		}
	})
	return g
}

// TestSharedLandRowsAreNotMutatedByAViewBuild pins the read-only half of
// #2113's shared intrinsic row: every one-type land in every zone of
// every game hands out the SAME backing array, so a consumer that wrote
// through an element would change every land's mana ability at once.
// The rows are read before and after a full view build for every seat
// (legal-move enumeration and per-viewer filtering included) and after
// a land's mana ability is activated through the engine, and must not
// change.
func TestSharedLandRowsAreNotMutatedByAViewBuild(t *testing.T) {
	subs := []string{"Plains", "Island", "Swamp", "Mountain", "Forest"}
	shared := make([][]game.ManaAbilityShape, len(subs))
	snapshot := make([][]game.ManaAbilityShape, len(subs))
	for i, sub := range subs {
		c := basicLand(sub, sub)
		rows := game.ManaAbilitiesForCard(c)
		if len(rows) != 1 {
			t.Fatalf("%s: %d rows, want 1", sub, len(rows))
		}
		again := game.ManaAbilitiesForCard(c)
		if &rows[0] != &again[0] {
			t.Fatalf("%s: two reads returned different backing arrays; the shared row is not shared, so this test proves nothing", sub)
		}
		shared[i] = rows
		snapshot[i] = append([]game.ManaAbilityShape(nil), rows...)
	}

	g := landTable(t, 24, 10)
	advanceTo(t, g, game.StepPrecombatMain)
	_ = ViewOfGame(g)
	for _, s := range g.Seats {
		v := ViewOfGameFor(g, s.ID.String())
		_ = FilterViewFor(v, s.ID.String())
	}
	_ = FilterViewFor(ViewOfGame(g), SpectatorViewerID)

	// A land actually tapped for mana, which reads the row for its cost.
	for _, c := range g.Battlefield.Cards {
		if c.Controller != g.Seats[g.Turn.ActiveSeat].ID || !c.IsLand() {
			continue
		}
		if err := g.ActivateManaAbility(c.Controller, c.InstanceID, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		break
	}
	_ = ViewOfGame(g)

	for i, sub := range subs {
		if !reflect.DeepEqual(shared[i], snapshot[i]) {
			t.Errorf("%s: the shared row changed from %+v to %+v", sub, snapshot[i], shared[i])
		}
	}
}

// BenchmarkViewOfGameLands measures one steady-state view of a mid-size
// board with lands in the battlefield, hand, graveyard and exile.
func BenchmarkViewOfGameLands(b *testing.B) {
	g := landTable(b, 24, 10)
	_ = ViewOfGame(g)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = ViewOfGame(g)
	}
}

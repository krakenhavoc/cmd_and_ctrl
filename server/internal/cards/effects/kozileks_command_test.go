package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const kozileksCommandOracle = "2f2c549d-0293-4b9d-b9a8-ff392800f3a5"

// TestKozileksCommandSpawnsTokensAndExilesACreature picks bullets 0
// and 2: X Eldrazi Spawn tokens for a target player, and exiling a
// creature with mana value X or less.
func TestKozileksCommandSpawnsTokensAndExilesACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	weenie := b12Creature(g, opp.ID, "Weenie", "Creature — Bear", 2, 2)

	b19CastXModal(t, g, "Kozilek's Command", "Kindred Instant — Eldrazi", kozileksCommandOracle, "{X}{C}{C}", 2,
		[]int{0, 2},
		[]game.TargetRef{
			modeRef(game.TargetPlayer, me.ID, 0, 0),
			modeRef(game.TargetCard, weenie, 1, 0),
		})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(weenie) {
		t.Error("the mana-value-2-or-less creature should have been exiled")
	}
	spawns := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Eldrazi Spawn" {
			spawns++
		}
	}
	if spawns != 2 {
		t.Errorf("X=2 should create two Eldrazi Spawn tokens, got %d", spawns)
	}
}

// TestKozileksCommandScryThenDraw picks bullet 1 alongside bullet 0.
func TestKozileksCommandScryThenDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Bottom Me", "Bottom Too")

	b19CastXModal(t, g, "Kozilek's Command", "Kindred Instant — Eldrazi", kozileksCommandOracle, "{X}{C}{C}", 2,
		[]int{0, 1},
		[]game.TargetRef{
			modeRef(game.TargetPlayer, me.ID, 0, 0),
			modeRef(game.TargetPlayer, me.ID, 1, 0),
		})
	passPriorityAroundTable(t, g)

	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no scry-X prompt: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("scry X=2 should offer two cards: %d", len(c.ScryCards))
	}
	hand := me.Hand.Size()
	if err := g.ResolveScry(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("scry X then draw a card: hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

// TestKozileksCommandExilesExactlyXFromGraveyards pins the declared
// caveat: the fourth bullet exiles exactly X cards, not up to X.
func TestKozileksCommandExilesExactlyXFromGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := graveCreature(me, "Mine A", "{1}")
	b := graveCreature(opp, "Theirs B", "{2}")

	b19CastXModal(t, g, "Kozilek's Command", "Kindred Instant — Eldrazi", kozileksCommandOracle, "{X}{C}{C}", 2,
		[]int{0, 3},
		[]game.TargetRef{
			modeRef(game.TargetPlayer, me.ID, 0, 0),
			modeRef(game.TargetCard, a, 1, 0),
			modeRef(game.TargetCard, b, 1, 0),
		})
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Error("both graveyard cards named should be exiled")
	}
}

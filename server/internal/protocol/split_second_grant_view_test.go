package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// split_second_grant_view_test.go — #3502: a spell that has split second
// only because a battlefield static grants it (Samut) still carries the
// `split_second` chip on the stack, and the table-wide
// `split_second_active` flag agrees; both go when the source leaves.

func TestViewShowsSplitSecondGrantedByAStatic(t *testing.T) {
	const samut = "test-3502-view-samut"
	prev := game.CatalogSplitSecondGrants
	game.CatalogSplitSecondGrants = func(key string) []game.SplitSecondGrant {
		if key != samut {
			return nil
		}
		return []game.SplitSecondGrant{{Spells: game.PermissionFilter{InstantOrSorceryOnly: true}, Label: "x"}}
	}
	t.Cleanup(func() { game.CatalogSplitSecondGrants = prev })

	g := buildActiveGame(t)
	castWindowOpen(t, g)
	me, them := g.Seats[0], g.Seats[1]
	seen := map[uuid.UUID]bool{me.ID: true, them.ID: true}

	src := game.NewCard("Test Samut", me.ID)
	src.TypeLine = "Legendary Creature — Human Wizard"
	src.OracleID = samut
	src.Controller = me.ID
	src.KnownBy = seen
	g.Battlefield.PushTop(src)

	spell := game.NewCard("Test Instant", me.ID)
	spell.TypeLine = "Instant"
	spell.KnownBy = seen
	me.Hand.PushTop(spell)
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}

	chip := func() (item, table bool) {
		v := ViewOfGameFor(g, them.ID.String())
		if len(v.StackItems) == 0 {
			t.Fatal("the stack view is empty")
		}
		return v.StackItems[len(v.StackItems)-1].SplitSecond, v.SplitSecondActive
	}
	if item, table := chip(); !item || !table {
		t.Fatalf("view with Samut out: chip=%v table=%v, want both", item, table)
	}

	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(src.InstanceID); err != nil {
			t.Fatalf("sacrifice the source: %v", err)
		}
	})
	if item, table := chip(); item || table {
		t.Fatalf("view after the source left: chip=%v table=%v, want neither", item, table)
	}
}

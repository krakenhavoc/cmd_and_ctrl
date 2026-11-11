package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const thievingVarmintOracle = "b9f3cf9f-24c7-43ed-9ba1-f4570df3d09e"

// The ability costs a tap and 1 life and adds two mana of one colour;
// that mana pays for a spell its controller does not own and for nothing
// else.
func TestThievingVarmintManaPaysOnlyForSpellsYouDontOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	varmint := b27Push(g, me.ID, "Thieving Varmint", "Creature — Varmint", thievingVarmintOracle, "{1}{B}", 2, 1, "B")
	advanceToMain(t, g)
	mine, theirs := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		known := map[uuid.UUID]bool{me.ID: true}
		me.Hand.PushTop(game.Card{InstanceID: mine, Name: "My Rock", TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID, KnownBy: known})
		me.Hand.PushTop(game.Card{InstanceID: theirs, Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{2}", Owner: opp.ID, Controller: me.ID, KnownBy: known})
	})
	life := me.Life

	if err := g.ActivateManaAbility(me.ID, varmint, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Fatalf("pool %v, want two red mana", got)
	}
	if me.Life != life-1 {
		t.Errorf("life = %d, want %d (pay 1 life)", me.Life, life-1)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == varmint && !c.Tapped {
			t.Errorf("the Varmint did not tap for its ability")
		}
	}

	// A spell I own: the mana is refused, and nothing is spent.
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the restricted mana paid for a spell its controller owns")
	}
	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool holds %d mana after a refused cast, want 2", me.ManaPool.Total())
	}
	// A spell I don't own: it pays.
	if err := g.CastSpell(me.ID, theirs, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast of a spell I don't own with the Varmint's mana: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool still holds %d mana", me.ManaPool.Total())
	}
}

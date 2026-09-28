package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// must_block_view_test.go — #1597: the wire half of block requirements.
// `CardView.must_block` marks the creatures a pending defender owes a
// block with, and clears once the declaration obeys the requirement.
func TestMustBlockIsStampedUntilTheRequirementIsObeyed(t *testing.T) {
	g := threeSeatsInDeclareAttackers(t)
	active := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%3]
	push := func(owner *game.Player, name string) uuid.UUID {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: owner.ID, Controller: owner.ID,
		})
		return id
	}
	atk := push(active, "Lured Bear")
	wall := push(def, "Wall")
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.New(), g.PinnedObjectsLocked(atk),
			[]game.Mod{game.AddBlockRequirementMod(game.BlockRequirementLure)}, game.IndefiniteDuration(), "test — Lure")
	})
	if err := g.DeclareAttacker(atk, def.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	stamp := func() map[string]bool {
		out := map[string]bool{}
		for _, c := range ViewOfGame(g).Battlefield.Cards {
			out[c.InstanceID] = c.MustBlock
		}
		return out
	}
	s := stamp()
	if !s[wall.String()] {
		t.Error("the creature the Lure asks for is not marked must_block")
	}
	if s[atk.String()] {
		t.Error("the attacker is marked must_block")
	}
	if err := g.DeclareBlocker(wall, atk); err != nil {
		t.Fatal(err)
	}
	if stamp()[wall.String()] {
		t.Error("must_block survived the block that obeys it")
	}
}

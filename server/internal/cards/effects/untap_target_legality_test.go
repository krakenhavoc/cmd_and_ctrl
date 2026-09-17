package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestFrostBreathSkipsATargetThatLeavesInResponse(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0].ID
	stays := pushCreatureToBattlefieldForTest(g, owner, "Stays")
	leaves := pushCreatureToBattlefieldForTest(g, owner, "Leaves")
	castCatalogSpell(t, g, "Frost Breath", "Instant", "382097b3-f753-493c-bde4-101c0538feb4", []game.TargetRef{
		{Kind: game.TargetCard, ID: stays},
		{Kind: game.TargetCard, ID: leaves},
	})
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(leaves); err != nil {
			t.Error(err)
		}
	})
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, stays)
	if !c.Tapped || len(c.NextUntapSkips) != 1 {
		t.Fatalf("remaining legal target was not tapped and marked: %+v", c)
	}
	for _, card := range g.Seats[0].Graveyard.Cards {
		if card.InstanceID == leaves && len(card.NextUntapSkips) != 0 {
			t.Fatal("departed target received a next-untap marker")
		}
	}
	for _, event := range g.Events {
		if event.Kind == game.EventEffectError {
			t.Fatalf("partial target resolution failed: %+v", event)
		}
	}
}

func TestCryogenRelicRechecksThatItsTargetIsTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	relic := pushPermanentForTest(g, me.ID, "Cryogen Relic", "106e9c0d-67a2-4db7-97d9-03e1ea1b40b5", "Artifact")
	target := pushCreatureToBattlefieldForTest(g, me.ID, "Tapped target")
	if err := g.TapCard(target, true); err != nil {
		t.Fatal(err)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "U"}, game.ManaToken{Color: "C"})
	handBefore := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, relic, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := g.TapCard(target, false); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, target)
	if c.Counters[game.CounterStun] != 0 {
		t.Fatal("Cryogen Relic put a stun counter on an illegal, untapped target")
	}
	if !me.Graveyard.Contains(relic) || me.Hand.Size() != handBefore+1 {
		t.Fatal("invalidated target undid the sacrifice cost or its separate leave trigger")
	}
}

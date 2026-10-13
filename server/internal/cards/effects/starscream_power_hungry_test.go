package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func starscreamRow() cards.Card {
	return transformerRow(starscreamOracleID,
		disturbFace{"Starscream, Power Hungry", "Legendary Artifact Creature — Robot", "{3}{B}", "2", "3"},
		disturbFace{"Starscream, Seeker Leader", "Legendary Artifact — Vehicle", "", "2", "3"},
		"More Than Meets the Eye {2}{B}\nFlying\nWhenever you draw a card, if you're the monarch, target opponent loses 2 life.\n"+
			"Whenever one or more creatures deal combat damage to you, convert Starscream.",
		"Living metal\nFlying, menace, haste\nWhenever Starscream deals combat damage to a player, if there is no monarch, that player becomes the monarch.\n"+
			"Whenever you become the monarch, convert Starscream.",
		"Living metal", "Convert", "Flying", "Menace", "Haste")
}

// "Whenever you draw a card, if you're the monarch, target opponent
// loses 2 life": nothing without the crown, 2 life with it.
func TestStarscreamDrainsOnDrawOnlyWhileMonarch(t *testing.T) {
	g, me, opp := flipMain(t)
	transformerIn(t, g, starscreamRow(), me)

	life := opp.Life
	_ = g.DrawCard(me.ID)
	tfSettle(t, g)
	if opp.Life != life || len(g.PendingChoices) != 0 {
		t.Fatalf("a draw with no monarch: life %d → %d, pending %d; want nothing", life, opp.Life, len(g.PendingChoices))
	}

	if err := g.SetMonarch(me.ID); err != nil {
		t.Fatal(err)
	}
	_ = g.DrawCard(me.ID)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	tfSettle(t, g)
	if opp.Life != life-2 {
		t.Errorf("a draw as the monarch: life %d → %d, want -2", life, opp.Life)
	}

	// An opponent's draw is not "you draw".
	life = opp.Life
	_ = g.DrawCard(opp.ID)
	tfSettle(t, g)
	if opp.Life != life {
		t.Errorf("the opponent's own draw cost them %d life", life-opp.Life)
	}
}

// CR 603.4: the monarchy is checked again as the trigger resolves.
func TestStarscreamDrainChecksTheCrownAgainOnResolution(t *testing.T) {
	g, me, opp := flipMain(t)
	transformerIn(t, g, starscreamRow(), me)
	if err := g.SetMonarch(me.ID); err != nil {
		t.Fatal(err)
	}
	life := opp.Life
	_ = g.DrawCard(me.ID)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	if err := g.SetMonarch(opp.ID); err != nil { // in response
		t.Fatal(err)
	}
	tfSettle(t, g)
	if opp.Life != life {
		t.Errorf("lost the crown in response, yet the opponent lost %d life", life-opp.Life)
	}
}

// "Whenever one or more creatures deal combat damage to you, convert":
// any creature, once per combat damage step.
func TestStarscreamConvertsWhenCreaturesHitYou(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, starscreamRow(), me)
	a := pushCatalogPermanent(g, opp.ID, "Attacker A", "Creature — Bear", "", false)
	b := pushCatalogPermanent(g, opp.ID, "Attacker B", "Creature — Bear", "", false)

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: a, Target: me.ID, Amount: 1, Combat: true})
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: b, Target: me.ID, Amount: 1, Combat: true})
	})
	tfSettle(t, g)
	c := wantFace(t, g, id, 1, "Starscream, Seeker Leader")
	if c.TurnOvers != 1 {
		t.Errorf("two creatures in one step turned it over %d times, want once", c.TurnOvers)
	}
}

// Noncombat damage and damage to someone else do not convert it.
func TestStarscreamIgnoresOtherDamage(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, starscreamRow(), me)
	a := pushCatalogPermanent(g, opp.ID, "Attacker", "Creature — Bear", "", false)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: a, Target: me.ID, Amount: 2})
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: a, Target: opp.ID, Amount: 2, Combat: true})
	})
	tfSettle(t, g)
	wantFace(t, g, id, 0, "Starscream, Power Hungry")
}

// Seeker Leader: a living-metal Vehicle that crowns the player it hits
// when there is no monarch, and converts back when you become the
// monarch.
func TestStarscreamSeekerLeaderTakesTheCrownThenConvertsBack(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerConverted(t, g, starscreamRow(), me)
	c := flipState(t, g, id)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Fatalf("on its controller's turn Seeker Leader is %q, want an artifact creature", c.Effective().Types)
	}
	for _, kw := range []string{"flying", "menace", "haste"} {
		if !game.HasKeyword(&c, kw) {
			t.Errorf("Seeker Leader lacks %s", kw)
		}
	}

	emitCombatDamage(g, id, opp.ID, 2)
	tfSettle(t, g)
	if g.Monarch != opp.ID {
		t.Fatalf("monarch = %v after Starscream hit the opponent, want them", g.Monarch)
	}
	// "Whenever you become the monarch" is MY crown, not theirs.
	wantFace(t, g, id, 1, "Starscream, Seeker Leader")

	// With a monarch already, a second hit crowns nobody.
	other := g.Seats[2]
	emitCombatDamage(g, id, other.ID, 2)
	tfSettle(t, g)
	if g.Monarch != opp.ID {
		t.Errorf("monarch moved to %v although there was one", g.Monarch)
	}

	if err := g.SetMonarch(me.ID); err != nil {
		t.Fatal(err)
	}
	tfSettle(t, g)
	c = wantFace(t, g, id, 0, "Starscream, Power Hungry")
	if c.TurnOvers != 2 || c.InstanceID != id {
		t.Errorf("turn-overs %d, want 2 on the same object", c.TurnOvers)
	}
}

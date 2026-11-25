package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const oracleDanceOfMany = "72e1a239-11a7-4dbe-b093-65d395e3ccc9"

// danceOfManyEnter casts Dance of Many, answers its enters trigger with
// `target` and resolves everything. It returns the enchantment and the
// token copy.
func danceOfManyEnter(t *testing.T, g *game.Game, target uuid.UUID) (uuid.UUID, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	dance := effectstest.CastCatalogSpell(t, g, "Dance of Many", "Enchantment", oracleDanceOfMany, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerTriggerTargets(t, g, me.ID, target)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dance) {
		t.Fatal("Dance of Many is not on the battlefield")
	}
	var token uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("no token copy was created")
	}
	if c := effectstest.DcCard(t, g, token); c.Name != "Hill Giant" || c.Controller != me.ID {
		t.Fatalf("the token is %q under %v, want my copy of Hill Giant", c.Name, c.Controller)
	}
	return dance, token
}

// TestDanceOfManyLeavingExilesTheToken — #6173's acceptance, first half:
// bouncing Dance of Many exiles the token it made (CR 607.2, CR 608.2h).
func TestDanceOfManyLeavingExilesTheToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	model := effectstest.SeedCopyableCreature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	dance, token := danceOfManyEnter(t, g, model)

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(dance); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the token is still on the battlefield after Dance of Many left")
	}
	if !g.Battlefield.Contains(model) {
		t.Error("the original creature left too")
	}
}

// TestDanceOfManyTokenLeavingSacrificesIt — the second half, across a
// restore: the link survives a snapshot, and destroying the token makes
// the controller sacrifice Dance of Many.
func TestDanceOfManyTokenLeavingSacrificesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	model := effectstest.SeedCopyableCreature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	dance, token := danceOfManyEnter(t, g, model)

	g = effectstest.RestoreThroughJSON(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(token); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dance) {
		t.Error("Dance of Many survived its token leaving the battlefield")
	}
	if !g.Battlefield.Contains(model) {
		t.Error("the original creature left too")
	}
}

// TestDanceOfManyOtherCreaturesLeavingDoNothing — only the linked token
// counts: another creature dying leaves Dance of Many alone.
func TestDanceOfManyOtherCreaturesLeavingDoNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	model := effectstest.SeedCopyableCreature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	dance, token := danceOfManyEnter(t, g, model)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(model); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dance) || !g.Battlefield.Contains(token) {
		t.Error("the original creature dying took Dance of Many or the token with it")
	}
}

// TestDanceOfManyCannotTargetAToken — "target nontoken creature".
func TestDanceOfManyCannotTargetAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var tok []uuid.UUID
	g.WithWriteLock(func() {
		var err error
		tok, err = g.CreateTokensForEffect(me.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier",
			Power: 1, Toughness: 1, PrintedPTKnown: true}, 1, game.TokenEntryOptions{})
		if err != nil {
			t.Fatalf("CreateTokensForEffect: %v", err)
		}
	})
	effectstest.SeedCopyableCreature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Dance of Many", "Enchantment", oracleDanceOfMany, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("entering asked for no target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: tok[0]}}); err == nil {
		t.Fatal("a token was a legal target")
	}
}

// TestDanceOfManyUpkeepCost — the upkeep pay-or-sacrifice is declared.
func TestDanceOfManyUpkeepCost(t *testing.T) {
	spec, ok := effects.Lookup(oracleDanceOfMany)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Triggered) != 4 {
		t.Fatalf("Dance of Many: registered %v, %d triggered abilities, want 4 and Full", ok, len(spec.Triggered))
	}
}

package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// krile_baldesion_test.go — Krile Baldesion (#3657, seam #3782): the
// trigger's target clause is cut to the cast spell's mana value, X
// included, and the "may" and the once-each-turn limit sit at resolution.

const krileOracle = "546a217a-ac24-4fc0-97a8-424cfe1aafc0"

func krileGame(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Krile Baldesion", "Legendary Creature — Dwarf Wizard", krileOracle, false)
	effectstest.AdvanceToMain(t, g)
	return g, me
}

// krileCast casts a sorcery with the given cost and X from me's hand and
// settles priority up to the trigger's target prompt or its absence.
func krileCast(t *testing.T, g *game.Game, me *game.Player, manaCost string, x int) {
	t.Helper()
	c := game.Card{InstanceID: uuid.New(), Name: "Ponder Deeply", TypeLine: "Sorcery", ManaCost: manaCost, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func krileInHand(g *game.Game, me *game.Player, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range me.Hand.Cards {
			if c.InstanceID == id {
				found = true
			}
		}
	})
	return found
}

func TestKrileIsRegistered(t *testing.T) {
	spec, ok := effects.Lookup(krileOracle)
	if !ok || spec.Name != "Krile Baldesion" || spec.Completeness != effects.CompletenessFull {
		t.Fatalf("Krile Baldesion registration: ok=%v spec=%+v", ok, spec)
	}
}

func TestKrileReturnsOnlyTheCreatureWithTheSpellsManaValue(t *testing.T) {
	g, me := krileGame(t)
	three := effectstest.PushGraveyardCreature(g, me.ID, "Three-drop", "{2}{G}")
	two := effectstest.PushGraveyardCreature(g, me.ID, "Two-drop", "{1}{G}")

	krileCast(t, g, me, "{2}{U}", 0)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for a mana value 3 spell with a three-drop in the graveyard")
	}
	if effectstest.HasID(pick.PickTargetCards, two) || !effectstest.HasID(pick.PickTargetCards, three) {
		t.Errorf("candidates = %v, want only the three-drop %v", pick.PickTargetCards, three)
	}
	effectstest.PickTriggerTarget(t, g, me.ID, three)
	effectstest.PassPriorityAroundTable(t, g)
	// The "may" is asked as the ability resolves, not before.
	if krileInHand(g, me, three) {
		t.Fatal("the card was returned before the player said yes")
	}
	effectstest.KrileConfirm(t, g, me, true)
	if !krileInHand(g, me, three) {
		t.Fatal("the three-drop did not come back to hand")
	}
	if krileInHand(g, me, two) {
		t.Error("the two-drop came back too")
	}
}

func TestKrileCountsTheXAnXSpellWasCastWith(t *testing.T) {
	g, me := krileGame(t)
	zero := effectstest.PushGraveyardCreature(g, me.ID, "One-drop", "{G}")
	four := effectstest.PushGraveyardCreature(g, me.ID, "Four-drop", "{2}{G}{G}")

	// {X}{R} for X=3 is mana value 4 on the stack (CR 202.3e).
	krileCast(t, g, me, "{X}{R}", 3)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for an X=3 spell with a four-drop in the graveyard")
	}
	if ids := pick.PickTargetCards; len(ids) != 1 || ids[0] != four {
		t.Errorf("candidates = %v, want only the four-drop %v (the one-drop %v is the X=0 reading)", ids, four, zero)
	}
}

func TestKrileUsesTheTurnOnlyWhenACardIsReturned(t *testing.T) {
	g, me := krileGame(t)
	a := effectstest.PushGraveyardCreature(g, me.ID, "Two-drop A", "{1}{G}")
	b := effectstest.PushGraveyardCreature(g, me.ID, "Two-drop B", "{1}{G}")

	// Declined: nothing returns and the turn's use is still there.
	krileCast(t, g, me, "{1}{U}", 0)
	effectstest.PickTriggerTarget(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KrileConfirm(t, g, me, false)
	if krileInHand(g, me, a) {
		t.Fatal("a declined trigger returned the card")
	}

	krileCast(t, g, me, "{1}{U}", 0)
	effectstest.PickTriggerTarget(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KrileConfirm(t, g, me, true)
	if !krileInHand(g, me, a) {
		t.Fatal("after a decline the next spell should still return a card")
	}

	// Done: a third spell this turn puts no trigger on the stack.
	krileCast(t, g, me, "{1}{U}", 0)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm) != nil {
		t.Fatal("Krile offered a second return this turn")
	}
	if krileInHand(g, me, b) {
		t.Error("the second two-drop came back on the same turn")
	}
}

func TestKrileIgnoresCreatureSpellsAndEmptyGraveyards(t *testing.T) {
	g, me := krileGame(t)
	bear := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	effectstest.PushGraveyardCreature(g, me.ID, "Two-drop", "{1}{G}")
	me.Hand.PushTop(bear)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a creature spell triggered Trace Aether")
	}

	// No creature card of mana value 5: CR 603.3d, nothing goes on the stack.
	krileCast(t, g, me, "{3}{U}{U}", 0)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm) != nil {
		t.Fatal("Trace Aether prompted with no creature card of that mana value")
	}
}

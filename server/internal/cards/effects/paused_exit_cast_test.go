package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// paused_exit_cast_test.go — #1474 with real cards. A commander an
// effect is exiling or destroying waits where it was while its owner
// answers CR 903.9, and in that window it can be neither cast nor the
// source of a counter cost: Bennie Bracks cast out of a hand an effect
// is exiling it from, Devoted Druid's "Put a -1/-1 counter on this
// creature", Jace, the Mind Sculptor's loyalty ability, and Heart of
// Kiran's "remove a loyalty counter from a planeswalker you control"
// naming a destroyed walker. Without the gate each of them spent the
// commander while its owner was still deciding. The engine half is
// game/paused_exit_cast_test.go; the move and tap costs are #1445's and
// #1427's files beside this one.

// assertPausedRefusal checks a refusal paid nothing and left the
// prompt alone.
func assertPausedRefusal(t *testing.T, g *game.Game, prompt *game.PendingChoice, err error) {
	t.Helper()
	if !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("using the paused commander: err = %v, want ErrChoicePending", err)
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("stack %d — the refusal put something on it", len(g.StackMeta))
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].ID != prompt.ID {
		t.Fatalf("the refusal disturbed the commander's prompt (%d pending)", len(g.PendingChoices))
	}
}

// destroyPaused destroys `id` by effect and returns its owner's open
// CR 903.9 prompt.
func destroyPaused(t *testing.T, g *game.Game, id uuid.UUID) *game.PendingChoice {
	t.Helper()
	if err := g.DestroyPermanentForEffect(id); err != nil {
		t.Fatalf("DestroyPermanentForEffect: %v", err)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceOptionalReplacement {
		t.Fatalf("want exactly one CR 903.9 prompt, have %d pending choices", len(g.PendingChoices))
	}
	if _, ok := battlefieldCard(g, id); !ok {
		t.Fatal("the destroyed commander left before its owner answered — the premise is gone")
	}
	return g.PendingChoices[0]
}

// markCommander flags a battlefield card as a commander.
func markCommander(g *game.Game, id uuid.UUID) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].IsCommander = true
		}
	}
}

func TestBennieBracksCannotBeCastWhileAnEffectIsExilingItFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	bennie := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: bennie, Name: "Bennie Bracks, Zoologist", TypeLine: "Legendary Creature — Elf Druid",
			ManaCost: "{3}{W}", OracleID: b16BennieBracksOracle, Owner: me.ID, Controller: me.ID,
			Power: 2, Toughness: 3, IsCommander: true,
		})
	})
	floatForTest(g, me, "WWWW")
	if err := g.ExileCardForEffect(bennie); err != nil {
		t.Fatalf("ExileCardForEffect: %v", err)
	}
	if len(g.PendingChoices) != 1 || !me.Hand.Contains(bennie) {
		t.Fatalf("premise: the exile did not pause in hand (%d pending)", len(g.PendingChoices))
	}
	prompt := g.PendingChoices[0]

	assertPausedRefusal(t, g, prompt, g.CastSpell(me.ID, bennie, game.CastSpellParams{Strict: true}))
	if !me.Hand.Contains(bennie) || len(me.ManaPool) != 4 {
		t.Fatalf("the refused cast moved Bennie or paid (pool %d)", len(me.ManaPool))
	}

	// The answer is in, and Bennie casts from where it put him.
	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if err := g.CastSpell(me.ID, bennie, game.CastSpellParams{FromZone: "command", Strict: true}); err != nil {
		t.Fatalf("casting Bennie from the command zone after the answer: %v", err)
	}
	if !g.Stack.Contains(bennie) {
		t.Error("Bennie is not on the stack")
	}
}

func TestDevotedDruidCommanderCannotPutACounterOnWhileAsked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	druid := b12Push(g, me.ID, "Devoted Druid", "Legendary Creature — Elf Druid", devotedDruidOracle, 0, 2)
	markCommander(g, druid)
	other := b12Push(g, me.ID, "Devoted Druid", "Creature — Elf Druid", devotedDruidOracle, 0, 2)
	prompt := destroyPaused(t, g, druid)

	assertPausedRefusal(t, g, prompt, g.ActivateCatalogAbility(me.ID, druid, 0, game.ActivateAbilityParams{}))
	if got := counterCount(g, druid, game.CounterMinusOne); got != 0 {
		t.Errorf("-1/-1 counters on the destroyed Druid = %d, want 0", got)
	}

	// An ordinary Druid pays with the prompt open.
	if err := g.ActivateCatalogAbility(me.ID, other, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("an ordinary Devoted Druid's untapper with the prompt open: %v", err)
	}
	if got := counterCount(g, other, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters on the ordinary Druid = %d, want 1", got)
	}
}

func TestJaceTheMindSculptorCommanderCannotActivateLoyaltyWhileAsked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	jace := pushCatalogWalker(g, me.ID, "Jace, the Mind Sculptor", jaceTheMindSculptorOracle, 3)
	markCommander(g, jace)
	prompt := destroyPaused(t, g, jace)

	// "0: Draw three cards, then put two cards back."
	assertPausedRefusal(t, g, prompt, g.ActivateCatalogAbility(me.ID, jace, 1, game.ActivateAbilityParams{}))
	if got := counterCount(g, jace, game.CounterLoyalty); got != 3 || g.LoyaltyActivatedThisTurn[jace] {
		t.Errorf("loyalty %d (activated %v), want 3 and unspent", got, g.LoyaltyActivatedThisTurn[jace])
	}

	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if !me.Command.Contains(jace) {
		t.Error("Jace did not reach the command zone")
	}
}

func TestHeartOfKiranCannotRemoveLoyaltyFromADestroyedCommanderWalker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	heart := pushVehicleForTest(g, me.ID, "Heart of Kiran", heartOfKiranOracle, 4, 4)
	walker := pushWalkerForCounterCost(g, me.ID, 3)
	markCommander(g, walker)
	other := pushWalkerForCounterCost(g, me.ID, 3)
	renameBattlefieldForCounterCost(g, other, "Other Walker")
	prompt := destroyPaused(t, g, walker)

	assertPausedRefusal(t, g, prompt,
		g.ActivateCatalogAbility(me.ID, heart, 1, game.ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{walker}}))
	if got := counterCount(g, walker, game.CounterLoyalty); got != 3 {
		t.Errorf("the destroyed walker's loyalty = %d, want 3", got)
	}

	// An ordinary walker pays with the prompt open.
	if err := g.ActivateCatalogAbility(me.ID, heart, 1, game.ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{other}}); err != nil {
		t.Fatalf("crewing with an ordinary walker's loyalty while the prompt is open: %v", err)
	}
	if got := counterCount(g, other, game.CounterLoyalty); got != 2 {
		t.Errorf("the ordinary walker's loyalty = %d, want 2", got)
	}
}

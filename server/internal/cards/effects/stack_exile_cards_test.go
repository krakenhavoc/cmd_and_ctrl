package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stack_exile_cards_test.go — the proof cards for #1318 (a spell exiled
// from the stack takes its stack record with it; "exile target spell";
// plot) and #1320 (a zone change records what caused it).

const (
	avenInterrupterOracle = "d31fd12f-b4dd-4bc3-ace4-703dabd0f607"
	ranarOracle           = "c73a9939-0742-4919-94d6-c3b537697f17"
)

// passUntilOnBattlefield passes priority until `id` has resolved onto
// the battlefield, leaving everything beneath it on the stack.
func passUntilOnBattlefield(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	for i := 0; i < 8 && !g.Battlefield.Contains(id); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Battlefield.Contains(id) {
		t.Fatal("the creature never resolved")
	}
}

// assertTableMovesOn passes priority around an empty stack and fails
// unless the step changes — the #1318 wedge, observed the way a player
// saw it.
func assertTableMovesOn(t *testing.T, g *game.Game) {
	t.Helper()
	if len(g.StackMeta) != 0 {
		t.Fatalf("%d stack records remain with nothing on the stack", len(g.StackMeta))
	}
	start := g.Turn.Step
	for i := 0; i < 2*len(g.Seats); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		if g.Turn.Step != start {
			return
		}
	}
	t.Fatalf("the table is wedged in %q", start)
}

// --- Aang, Swift Savior (#1318) ------------------------------------

// TestAangSwiftSaviorAirbendsASpell is the live bug on the deck's
// commander: Aang is flashed in over a spell, airbends it, and the
// table carries on. Before #1318 the spell's card went to exile and its
// stack record stayed, and no number of passes left the step.
func TestAangSwiftSaviorAirbendsASpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "test-1318-airbent-sorcery", nil)

	aang := aangSwiftSaviorCard(me.ID)
	me.Hand.PushTop(aang)
	if err := g.CastSpell(me.ID, aang.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("flash in Aang: %v", err)
	}
	passUntilOnBattlefield(t, g, aang.InstanceID)
	if !g.Stack.Contains(spell) {
		t.Fatal("the spell under Aang resolved before the airbend")
	}

	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, spell)
	passPriorityAroundTable(t, g)

	assertAirbent(t, g, spell, me.ID)
	assertTableMovesOn(t, g)
}

// --- Aven Interrupter (#1318) --------------------------------------

// TestAvenInterrupterExilesAndPlotsASpell: the spell leaves the stack
// without a trace, becomes plotted, and its owner casts it for free on
// a later turn — not before.
func TestAvenInterrupterExilesAndPlotsASpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := castCatalogSpell(t, g, "Expensive Sorcery", "Sorcery", "test-1318-plotted-sorcery", nil)
	g.WithWriteLock(func() {
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == spell {
				g.Stack.Cards[i].ManaCost = "{5}{U}{U}"
			}
		}
	})
	aven := castCatalogSpell(t, g, "Aven Interrupter", "Creature — Bird Rogue", avenInterrupterOracle, nil)
	passUntilOnBattlefield(t, g, aven)
	pickCard(t, g, me.ID, spell)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(spell) {
		t.Fatal("the spell was not exiled")
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventCounterSpell && ev.CardID == spell {
			t.Error("exiling a spell is not countering it")
		}
	}
	perm := g.CastPermissionOnCardByIDForEffect(spell)
	if !perm.Granted() || perm.Player != me.ID || perm.Timing != game.TimingPlot || perm.Cost != "{0}" {
		t.Fatalf("plot permission = %+v, want the owner's free TimingPlot grant", perm)
	}
	assertTableMovesOn(t, g)

	// Its owner's later turn, main phase: free.
	g.WithWriteLock(func() {
		g.Turn.Number++
		g.Turn.ActiveSeat = 0
		g.Turn.PriorityHolder = 0
		g.Turn.Step = game.StepPrecombatMain
		g.Turn.Phase = game.PhasePrecombatMain
	})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast the plotted {5}{U}{U} for nothing: %v", err)
	}
	if !g.Stack.Contains(spell) {
		t.Fatal("the plotted spell never reached the stack")
	}
}

// TestAvenInterrupterTaxesOpponentsCastsFromExile: "spells your
// opponents cast from graveyards or from exile cost {2} more" — an
// opponent's airbent card costs {2} + {2}, and the same opponent's
// cast from hand costs what it prints.
func TestAvenInterrupterTaxesOpponentsCastsFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Aven Interrupter", avenInterrupterOracle, "Creature — Bird Rogue")
	advanceTo(t, g, game.StepPrecombatMain)

	exiled := game.NewCard("Their Instant", opp.ID)
	exiled.TypeLine = "Instant"
	exiled.ManaCost = "{U}"
	exiled.Layout = "normal"
	exiled.Controller = opp.ID
	g.Exile.PushTop(exiled)
	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(exiled.InstanceID, game.CastPermission{
			Player: opp.ID, Zone: game.ZoneExile, Cost: AirbendCost,
			Duration: game.WhileInZoneDuration(), CastOnly: true,
		})
		g.Turn.PriorityHolder = 1
	})

	pool := func(n int) {
		opp.ManaPool = nil
		for i := 0; i < n; i++ {
			opp.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
	}
	pool(3)
	var im *game.InsufficientManaError
	if err := g.CastSpell(opp.ID, exiled.InstanceID, game.CastSpellParams{Strict: true, FromZone: "exile"}); !errors.As(err, &im) {
		t.Fatalf("three mana for an airbent card under Aven Interrupter: got %v, want *InsufficientManaError", err)
	}
	pool(4)
	if err := g.CastSpell(opp.ID, exiled.InstanceID, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("four mana ({2} airbend + {2} tax): %v", err)
	}

	// From hand: untaxed.
	fromHand := handCardForTest(opp, "Their Other Instant", "Instant", "")
	for i := range opp.Hand.Cards {
		if opp.Hand.Cards[i].InstanceID == fromHand {
			opp.Hand.Cards[i].ManaCost = "{U}"
			opp.Hand.Cards[i].Layout = "normal"
		}
	}
	opp.ManaPool = nil
	opp.ManaPool.AddMana(game.ManaToken{Color: "U"})
	if err := g.CastSpell(opp.ID, fromHand, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a {U} instant from hand under Aven Interrupter: %v", err)
	}
}

// --- Ranar the Ever-Watchful (#1320) -------------------------------

func spiritsOf(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == "Spirit" && IsToken(c) {
			n++
		}
	}
	return n
}

// TestRanarCountsMyAbilityExilingTheirCreature: Aang's airbend is MY
// ability, and it exiles an OPPONENT's creature — the permanent's
// controller is not the question.
func TestRanarCountsMyAbilityExilingTheirCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Ranar the Ever-Watchful", ranarOracle, "Legendary Creature — Spirit Warrior")
	victim := pushCreatureToBattlefieldForTest(g, opp.ID, "Their Blocker")
	advanceTo(t, g, game.StepPrecombatMain)

	aang := aangSwiftSaviorCard(me.ID)
	me.Hand.PushTop(aang)
	if err := g.CastSpell(me.ID, aang.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passUntilOnBattlefield(t, g, aang.InstanceID)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(victim) {
		t.Fatal("the airbend did not exile the creature")
	}
	if got := spiritsOf(g, me.ID); got != 1 {
		t.Errorf("Spirits = %d, want 1 — my ability exiled a permanent", got)
	}
}

// TestRanarIgnoresAnOpponentsExile: the active player's Swords to
// Plowshares exiles Ranar's controller's creature. Not "a spell or
// ability you control".
func TestRanarIgnoresAnOpponentsExile(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushPermanentForTest(g, me.ID, "Ranar the Ever-Watchful", ranarOracle, "Legendary Creature — Spirit Warrior")
	mine := pushCreatureToBattlefieldForTest(g, me.ID, "My Bear")

	castCatalogSpell(t, g, "Swords to Plowshares", "Instant", "b1544f21-7e98-461b-aed5-e748b0168c52",
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(mine) {
		t.Fatal("Swords did not exile the creature")
	}
	if got := spiritsOf(g, me.ID); got != 0 {
		t.Errorf("Spirits = %d, want 0 — nothing I control exiled it", got)
	}
	if got := spiritsOf(g, caster.ID); got != 0 {
		t.Errorf("the caster got %d Spirits from a Ranar they do not control", got)
	}
}

// TestRanarTriggersOnceForOneExileOfMany: "one or more" — Curse of the
// Swine exiling two creatures in one resolution is one occurrence
// (CR 603.2c), one Spirit.
func TestRanarTriggersOnceForOneExileOfMany(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Ranar the Ever-Watchful", ranarOracle, "Legendary Creature — Spirit Warrior")
	a := pushCreatureToBattlefieldForTest(g, opp.ID, "Their A")
	b := pushCreatureToBattlefieldForTest(g, opp.ID, "Their B")

	castXSpell(t, g, "Curse of the Swine", "Sorcery", b07CurseOfTheSwineOracle, "{X}{U}{U}", 2,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Fatal("Curse of the Swine did not exile both creatures")
	}
	if got := spiritsOf(g, me.ID); got != 1 {
		t.Errorf("Spirits = %d, want 1 — one exile of two permanents is one trigger", got)
	}
}

// TestRanarCountsACardForetoldFromMyHand: the hand clause does not care
// what exiled the card — foretell is a special action, not a spell or
// ability, and it still counts.
func TestRanarCountsACardForetoldFromMyHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushPermanentForTest(g, me.ID, "Ranar the Ever-Watchful", ranarOracle, "Legendary Creature — Spirit Warrior")
	advanceTo(t, g, game.StepPrecombatMain)
	card := handCardForTest(me, "Saw It Coming", "Instant", sawItComingOracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, card, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := spiritsOf(g, me.ID); got != 1 {
		t.Errorf("Spirits = %d, want 1 — a card went into exile from my hand", got)
	}
}

// TestRanarTwoForetellsInOneWindowMakeOneSpirit PINS a declared gap
// (the card's second caveat): two foretells with nothing resolving
// between them are one event batch, so OncePerBatch sees one
// occurrence. The rules say two. When special actions open their own
// batch this flips to 2 — update the caveat with it.
func TestRanarTwoForetellsInOneWindowMakeOneSpirit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushPermanentForTest(g, me.ID, "Ranar the Ever-Watchful", ranarOracle, "Legendary Creature — Spirit Warrior")
	advanceTo(t, g, game.StepPrecombatMain)
	first := handCardForTest(me, "Saw It Coming", "Instant", sawItComingOracle)
	second := handCardForTest(me, "Behold the Multiverse", "Instant", beholdTheMultiverseOracl)
	for _, card := range []uuid.UUID{first, second} {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
		if err := g.PerformSpecialAction(me.ID, card, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
			t.Fatalf("foretell: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if got := spiritsOf(g, me.ID); got != 1 {
		t.Errorf("Spirits = %d — the declared one-batch gap has changed; update Ranar's caveat", got)
	}
}

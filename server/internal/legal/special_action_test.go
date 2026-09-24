package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// special_action_test.go — the enumerator half of the CR 116.2 verb
// (#658, #659, ADR 0062 Decision 4).
//
// #544's invariant, one verb over: offer exactly the special actions
// the engine accepts, and nothing the engine would refuse. The row
// that earns this file is the split-second one — the two enumerators
// beside this one return early on SplitSecondActive, and copying them
// would silently make foretell illegal under a Trickbind.

// withSpecialActions stubs the catalog hook for one test.
func withSpecialActions(t *testing.T, oracle string, actions []game.SpecialAction) {
	t.Helper()
	prev := game.CatalogSpecialActions
	game.CatalogSpecialActions = func(id string) []game.SpecialAction {
		if id != oracle {
			return nil
		}
		return actions
	}
	t.Cleanup(func() { game.CatalogSpecialActions = prev })
}

// specialActionsOf picks the special-action moves for one source.
func specialActionsOf(moves []legal.Move, source uuid.UUID) []legal.Move {
	return movesOfKindFor(moves, legal.KindSpecialAction, source)
}

func foretellCard(oracle string) game.Card {
	return game.Card{
		Name:     "Saw It Coming",
		TypeLine: "Instant",
		ManaCost: "{1}{U}{U}",
		OracleID: oracle,
	}
}

// Foretell is enumerated on its owner's turn when the {2} is payable,
// and not when it is not — the #544 rule.
func TestForetellIsEnumeratedOnlyWhenPayable(t *testing.T) {
	const oracle = "legal-foretell-oracle"
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	withSpecialActions(t, oracle, []game.SpecialAction{{
		Kind: game.SpecialActionForetell, Cost: "{2}", CastCost: "{1}{U}", Label: "Foretell {2}",
	}})
	card := handCard(active, foretellCard(oracle))
	advanceTo(t, g, game.StepPrecombatMain)

	if acts := specialActionsOf(legal.EnumerateFor(g, active.ID), card); len(acts) != 0 {
		t.Fatalf("a seat with no mana: want no foretell move, got %v", labels(acts))
	}

	battlefieldCard(g, active, basic("Island", "Island"))
	battlefieldCard(g, active, basic("Island", "Island"))
	moves := legal.EnumerateFor(g, active.ID)
	acts := specialActionsOf(moves, card)
	if len(acts) != 1 {
		t.Fatalf("want exactly one foretell move, got %v", labels(acts))
	}
	if acts[0].Type != legal.TypeSpecialAction {
		t.Errorf("move type = %q, want %q", acts[0].Type, legal.TypeSpecialAction)
	}
	// Every enumerated move is one the dispatcher accepts.
	dispatchAll(t, g, active.ID, moves)
}

// CR 702.61b, and the row this whole file earns. Every other
// enumerator in this package opens with
// `if g.SplitSecondActive { return }`; this one must not, because a
// special action is neither a cast nor an activation and CR 702.61b
// stops only those. A Trickbind on the stack does not stop a player
// foretelling.
func TestSplitSecondDoesNotStopForetell(t *testing.T) {
	const oracle = "legal-split-second-foretell-oracle"
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	withSpecialActions(t, oracle, []game.SpecialAction{{
		Kind: game.SpecialActionForetell, Cost: "{2}", CastCost: "{1}{U}", Label: "Foretell {2}",
	}})
	card := handCard(active, foretellCard(oracle))
	battlefieldCard(g, active, basic("Island", "Island"))
	battlefieldCard(g, active, basic("Island", "Island"))
	advanceTo(t, g, game.StepPrecombatMain)

	if n := len(specialActionsOf(legal.EnumerateFor(g, active.ID), card)); n != 1 {
		t.Fatalf("with no split second: want the foretell move, got %d", n)
	}

	g.SplitSecondActive = true
	acts := specialActionsOf(legal.EnumerateFor(g, active.ID), card)
	if len(acts) != 1 {
		t.Fatalf("under split second: want foretell still offered, got %v", labels(acts))
	}
	// And the enumerator really is under split second: the ordinary
	// cast of the same card is gone.
	if n := len(movesOfKindFor(legal.EnumerateFor(g, active.ID), legal.KindCast, card)); n != 0 {
		t.Errorf("split second is not actually active: %d cast moves still offered", n)
	}
}

// A card in an OPPONENT's hand is never enumerated for this seat, and
// foretell is never offered on a turn that is not yours (CR 702.143a).
func TestForetellIsNotEnumeratedOnAnotherSeatsTurn(t *testing.T) {
	const oracle = "legal-offturn-foretell-oracle"
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	var other *game.Player
	for _, p := range g.Seats {
		if p.ID != active.ID {
			other = p
			break
		}
	}
	clearHand(other)
	withSpecialActions(t, oracle, []game.SpecialAction{{
		Kind: game.SpecialActionForetell, Cost: "{2}", CastCost: "{1}{U}", Label: "Foretell {2}",
	}})
	card := handCard(other, foretellCard(oracle))
	battlefieldCard(g, other, basic("Island", "Island"))
	battlefieldCard(g, other, basic("Island", "Island"))
	advanceTo(t, g, game.StepPrecombatMain)

	if acts := specialActionsOf(legal.EnumerateFor(g, other.ID), card); len(acts) != 0 {
		t.Errorf("foretell offered on another seat's turn: %v", labels(acts))
	}
}

// CR 702.62c, and the other half of the split-second row above: a
// special action's legality is asked PER KIND, so the one enumerator
// bars suspend under split second while still offering foretell.
func TestSplitSecondStopsSuspendButNotForetell(t *testing.T) {
	const oracle = "legal-both-kinds-oracle"
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	withSpecialActions(t, oracle, []game.SpecialAction{
		{Kind: game.SpecialActionForetell, Cost: "{2}", CastCost: "{1}{U}", Label: "Foretell {2}"},
		{Kind: game.SpecialActionSuspend, Cost: "{U}", Counters: 2, Label: "Suspend 2"},
	})
	card := handCard(active, foretellCard(oracle))
	battlefieldCard(g, active, basic("Island", "Island"))
	battlefieldCard(g, active, basic("Island", "Island"))
	advanceTo(t, g, game.StepPrecombatMain)

	if n := len(specialActionsOf(legal.EnumerateFor(g, active.ID), card)); n != 2 {
		t.Fatalf("with no split second: want both kinds offered, got %d", n)
	}

	g.SplitSecondActive = true
	acts := specialActionsOf(legal.EnumerateFor(g, active.ID), card)
	if len(acts) != 1 {
		t.Fatalf("under split second: want foretell only, got %v", labels(acts))
	}
	if acts[0].Label != "Foretell {2} Saw It Coming" {
		t.Errorf("the surviving move is %q, want the foretell one", acts[0].Label)
	}
}

// CR 702.62c imports the card's own casting window: a SORCERY may be
// suspended only at sorcery speed, so the enumerator offers it in a
// main phase and not in an upkeep.
func TestSuspendIsEnumeratedAtTheCardsCastingSpeed(t *testing.T) {
	const oracle = "legal-suspend-sorcery-oracle"
	seed := func(t *testing.T, step game.Step) (*game.Game, *game.Player, uuid.UUID) {
		t.Helper()
		g := newTable(t)
		active := g.Seats[g.Turn.ActiveSeat]
		clearHand(active)
		withSpecialActions(t, oracle, []game.SpecialAction{{
			Kind: game.SpecialActionSuspend, Cost: "{R}", Counters: 1, Label: "Suspend 1",
		}})
		card := handCard(active, game.Card{
			Name:     "Rift Bolt",
			TypeLine: "Sorcery",
			ManaCost: "{2}{R}",
			OracleID: oracle,
		})
		battlefieldCard(g, active, basic("Mountain", "Mountain"))
		advanceTo(t, g, step)
		return g, active, card
	}

	// A main phase: the sorcery could begin to be cast, so it can be
	// suspended, and the move the enumerator offers is one the
	// dispatcher accepts.
	g, active, card := seed(t, game.StepPrecombatMain)
	moves := legal.EnumerateFor(g, active.ID)
	if n := len(specialActionsOf(moves, card)); n != 1 {
		t.Fatalf("a sorcery in a main phase: want one suspend move, got %d", n)
	}
	dispatchAll(t, g, active.ID, moves)

	// A combat step on the same seat's own turn: sorcery speed is
	// shut, so suspend is shut with it (CR 702.62c).
	g, active, card = seed(t, game.StepBeginCombat)
	if acts := specialActionsOf(legal.EnumerateFor(g, active.ID), card); len(acts) != 0 {
		t.Errorf("a sorcery offered for suspend outside a main phase: %v", labels(acts))
	}
}

// The other bot-facing half of suspend: the FREE CAST the last time
// counter offers. It is an ordinary granted exile permission
// (ADR 0066), so `grantedCastMoves` already enumerates it — what this
// pins is that the two fields suspend sets make the move appear where
// it has to: `{0}` so a seat with no mana can still take it, and
// flash timing so a SORCERY is offered outside a main phase, which is
// where the last-counter trigger resolves (CR 608.2g).
func TestTheSuspendFreeCastIsEnumeratedOutsideAMainPhase(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	// A step on the seat's own turn where priority is held and
	// SORCERY SPEED IS SHUT, which is the whole point: without the
	// permission's flash timing a sorcery would not be offered here.
	advanceTo(t, g, game.StepBeginCombat)

	bolt := game.NewCard("Rift Bolt", active.ID)
	bolt.TypeLine = "Sorcery"
	bolt.ManaCost = "{2}{R}"
	g.Exile.PushTop(bolt)
	g.WithWriteLock(func() {
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:   active.ID,
			Zone:     game.ZoneExile,
			Cost:     "{0}",
			Timing:   game.TimingFlash,
			CastOnly: true,
			Duration: g.UntilEndOfTurnDuration(),
			Label:    game.SuspendFreeCastLabel,
		}, []game.Card{bolt})
	})

	moves := legal.EnumerateFor(g, active.ID)
	casts := movesOfKindFor(moves, legal.KindCast, bolt.InstanceID)
	if len(casts) == 0 {
		t.Fatalf("the suspend free cast is not enumerated outside a main phase; moves = %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

// withCostModifiers stubs the CR 601.2f catalog hook for one test.
func withCostModifiers(t *testing.T, oracle string, mods []game.CostModifier) {
	t.Helper()
	prev := game.CatalogCostModifiers
	game.CatalogCostModifiers = func(id string) []game.CostModifier {
		if id != oracle {
			return nil
		}
		return mods
	}
	t.Cleanup(func() { game.CatalogCostModifiers = prev })
}

// ranarStyleForetellDiscount is a CostModifier in Ranar the
// Ever-Watchful's shape: "the first card you foretell each turn costs
// {0} to foretell" (#1319).
func ranarStyleForetellDiscount() game.CostModifier {
	return game.CostModifier{
		Kind:           game.CostReduction,
		SpecialActions: true,
		Label:          "The first card you foretell each turn costs {0} to foretell.",
		AppliesTo: func(q game.CostQuery) bool {
			return q.SpecialAction != nil && q.SpecialAction.Kind == game.SpecialActionForetell &&
				q.Game.ForetoldCountThisTurn(q.Controller) == 0
		},
		Amount: func(game.CostQuery) int { return 2 },
	}
}

// #1319: the enumerator prices a special action through the same
// CR 601.2f pass the engine charges with, so a discount that empties
// the cost out entirely offers the move to a seat with no mana at
// all — #544's invariant, the special-action verb over.
func TestForetellIsEnumeratedWhenACostModifierMakesItFree(t *testing.T) {
	const cardOracle = "legal-foretell-discount-card"
	const sourceOracle = "legal-foretell-discount-source"
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	withSpecialActions(t, cardOracle, []game.SpecialAction{{
		Kind: game.SpecialActionForetell, Cost: "{2}", CastCost: "{1}{U}", Label: "Foretell {2}",
	}})
	withCostModifiers(t, sourceOracle, []game.CostModifier{ranarStyleForetellDiscount()})
	card := handCard(active, foretellCard(cardOracle))
	battlefieldCard(g, active, game.Card{Name: "Ranar the Ever-Watchful", TypeLine: "Legendary Creature — Spirit Warrior", OracleID: sourceOracle})
	advanceTo(t, g, game.StepPrecombatMain)

	// No land on the board at all: the {2} would refuse this move
	// without the discount.
	acts := specialActionsOf(legal.EnumerateFor(g, active.ID), card)
	if len(acts) != 1 {
		t.Fatalf("with the discount and no mana: want the foretell move offered, got %v",
			labels(legal.EnumerateFor(g, active.ID)))
	}
	dispatchAll(t, g, active.ID, legal.EnumerateFor(g, active.ID))
}

// #1319's partition: a modifier written as an ordinary spell
// reduction must never reach a special action, in either direction —
// the same rule TestACastModifierDoesNotPriceAnActivation pins for
// activations, one door over.
func TestASpellCostModifierDoesNotMakeForetellFree(t *testing.T) {
	const cardOracle = "legal-foretell-nodiscount-card"
	const sourceOracle = "legal-foretell-nodiscount-source"
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	withSpecialActions(t, cardOracle, []game.SpecialAction{{
		Kind: game.SpecialActionForetell, Cost: "{2}", CastCost: "{1}{U}", Label: "Foretell {2}",
	}})
	withCostModifiers(t, sourceOracle, []game.CostModifier{{
		Kind:   game.CostReduction,
		Label:  "Spells you cast cost {2} less to cast.",
		Amount: func(game.CostQuery) int { return 2 },
	}})
	card := handCard(active, foretellCard(cardOracle))
	battlefieldCard(g, active, game.Card{Name: "Goblin Electromancer", TypeLine: "Creature — Goblin Wizard", OracleID: sourceOracle})
	advanceTo(t, g, game.StepPrecombatMain)

	// No land on the board: a spell-shaped reduction must not reach
	// the foretell special action, so the {2} is still unaffordable.
	if acts := specialActionsOf(legal.EnumerateFor(g, active.ID), card); len(acts) != 0 {
		t.Fatalf("a spell cost modifier discounted a foretell: %v", labels(legal.EnumerateFor(g, active.ID)))
	}
}

// #1342 / CR 702.170a: plot is enumerated in its owner's main phase
// with the stack empty, when the plot cost is payable — and not in
// combat, not to another seat. The move the enumerator offers is one
// the dispatcher accepts, and the plotted card's free cast is
// enumerated on a later turn with no mana at all.
func TestPlotIsEnumeratedInYourMainPhaseAndItsFreeCastLater(t *testing.T) {
	const oracle = "legal-plot-oracle"
	seed := func(t *testing.T, step game.Step) (*game.Game, *game.Player, uuid.UUID) {
		t.Helper()
		g := newTable(t)
		active := g.Seats[g.Turn.ActiveSeat]
		clearHand(active)
		withSpecialActions(t, oracle, []game.SpecialAction{{
			Kind: game.SpecialActionPlot, Cost: "{U}", Label: "Plot {U}",
		}})
		card := handCard(active, game.Card{
			Name:     "Djinn of Fool's Fall",
			TypeLine: "Creature — Djinn",
			ManaCost: "{4}{U}",
			Power:    4,
			OracleID: oracle,
		})
		battlefieldCard(g, active, basic("Island", "Island"))
		advanceTo(t, g, step)
		return g, active, card
	}

	g, active, card := seed(t, game.StepBeginCombat)
	if acts := specialActionsOf(legal.EnumerateFor(g, active.ID), card); len(acts) != 0 {
		t.Errorf("plot offered in combat: %v", labels(acts))
	}

	g, active, card = seed(t, game.StepPrecombatMain)
	for _, p := range g.Seats {
		if p.ID == active.ID {
			continue
		}
		if acts := specialActionsOf(legal.EnumerateFor(g, p.ID), card); len(acts) != 0 {
			t.Errorf("plot offered to a seat that does not hold the card: %v", labels(acts))
		}
	}
	moves := legal.EnumerateFor(g, active.ID)
	acts := specialActionsOf(moves, card)
	if len(acts) != 1 {
		t.Fatalf("a main phase with {U} available: want one plot move, got %v", labels(acts))
	}
	if acts[0].Label != "Plot {U} Djinn of Fool's Fall" {
		t.Errorf("plot move label = %q", acts[0].Label)
	}
	if err := g.PerformSpecialAction(active.ID, card, game.SpecialActionPlot, game.SpecialActionParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("plot: %v", err)
	}
	if n := len(movesOfKindFor(legal.EnumerateFor(g, active.ID), legal.KindCast, card)); n != 0 {
		t.Errorf("the plotted card is offered for casting on the turn it was plotted (%d moves)", n)
	}

	// A later turn of the same seat. The land stays tapped, so the
	// only way the cast can be offered is for free.
	g.WithWriteLock(func() { g.Turn.Number++ })
	moves = legal.EnumerateFor(g, active.ID)
	if n := len(movesOfKindFor(moves, legal.KindCast, card)); n == 0 {
		t.Fatalf("the plotted card's free cast is not enumerated on a later turn; moves = %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

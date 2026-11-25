package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cumulative_upkeep_events_test.go — #6188: paying a cumulative upkeep
// and declining it each announce themselves, so a card can watch either
// (Balduvian Fallen, Hibernation's End, Heart of Bogardan, Thought Lash).
//
//	702.24a Cumulative upkeep is a triggered ability that imposes an
//	        increasing cost on a permanent. "Cumulative upkeep [cost]"
//	        means "At the beginning of your upkeep, if this permanent is
//	        on the battlefield, put an age counter on this permanent.
//	        Then you may pay [cost] for each age counter on it. If you
//	        don't, sacrifice it." ...
//
//	702.24b If a permanent has multiple instances of cumulative upkeep,
//	        each triggers separately. However, the age counters are not
//	        connected to any particular ability; each cumulative upkeep
//	        ability will count the total number of age counters on the
//	        permanent at the time that ability resolves.

const cueSustainingSpirit = "683c81ae-38cd-4db1-b7e7-a09d34b6431e"

// cueLastEvent is the last event of `kind` whose card is `id`.
func cueLastEvent(g *game.Game, kind game.EventKind, id uuid.UUID) *game.Event {
	for i := len(g.Events) - 1; i >= 0; i-- {
		if ev := g.Events[i]; ev.Kind == kind && ev.CardID == id {
			return &g.Events[i]
		}
	}
	return nil
}

// cueEventIndex is the index of the last event of `kind` for `id`, or -1.
func cueEventIndex(g *game.Game, kind game.EventKind, id uuid.UUID) int {
	for i := len(g.Events) - 1; i >= 0; i-- {
		if ev := g.Events[i]; ev.Kind == kind && ev.CardID == id {
			return i
		}
	}
	return -1
}

// Sustaining Spirit's "Cumulative upkeep {1}{W}": paying announces the
// payment with the age counters it was paid for and every mana spent on
// it, which is what Balduvian Fallen's "+1/+0 for each {B} or {R} spent
// this way" counts.
func TestPayingACumulativeUpkeepAnnouncesItWithTheManaSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	spirit := effectstest.B12Push(g, me.ID, "Sustaining Spirit", "Creature — Angel Spirit", cueSustainingSpirit, 0, 3)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if prompt.PayCost != "{1}{W}" {
		t.Fatalf("cost = %q, want {1}{W}", prompt.PayCost)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "C"})
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, spirit); !ok {
		t.Fatal("a paid cumulative upkeep sacrificed the permanent")
	}
	ev := cueLastEvent(g, game.EventCumulativeUpkeepPaid, spirit)
	if ev == nil {
		t.Fatal("paying announced nothing")
	}
	if ev.Actor != me.ID || ev.Source != spirit || ev.Amount != 1 {
		t.Errorf("event = actor %s, source %s, age %d; want the payer, the permanent and 1 age counter", ev.Actor, ev.Source, ev.Amount)
	}
	if got := game.SpentSymbolsOf(*ev, "W"); got != 1 {
		t.Errorf("{W} spent this way = %d, want 1", got)
	}
	if got := game.SpentSymbolsOf(*ev, "B", "R"); got != 0 {
		t.Errorf("{B} or {R} spent this way = %d, want 0", got)
	}
	if len(ev.SpentSymbols) != 2 {
		t.Errorf("mana spent = %v, want both symbols of {1}{W}", ev.SpentSymbols)
	}
	if cueEventIndex(g, game.EventCumulativeUpkeepNotPaid, spirit) >= 0 {
		t.Error("paying also announced a decline")
	}

	// Second upkeep: two age counters, so the event's age is 2.
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt = effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "W"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("pay the second upkeep: %v", err)
	}
	if ev := cueLastEvent(g, game.EventCumulativeUpkeepPaid, spirit); ev == nil || ev.Amount != 2 {
		t.Errorf("second upkeep's age = %+v, want 2 (CR 702.24b)", ev)
	}
}

// A declined cumulative upkeep announces itself BEFORE the sacrifice, so
// Heart of Bogardan's "When a player doesn't pay this enchantment's
// cumulative upkeep" triggers from a source still on the battlefield.
func TestDecliningACumulativeUpkeepAnnouncesItBeforeTheSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	spirit := effectstest.B12Push(g, me.ID, "Sustaining Spirit", "Creature — Angel Spirit", cueSustainingSpirit, 0, 3)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if err := g.ResolvePayUnless(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	ev := cueLastEvent(g, game.EventCumulativeUpkeepNotPaid, spirit)
	if ev == nil {
		t.Fatal("declining announced nothing")
	}
	if ev.Actor != me.ID || ev.Source != spirit || ev.Amount != 1 {
		t.Errorf("event = actor %s, source %s, age %d; want the player, the permanent and 1 age counter", ev.Actor, ev.Source, ev.Amount)
	}
	if cueEventIndex(g, game.EventCumulativeUpkeepPaid, spirit) >= 0 {
		t.Error("declining also announced a payment")
	}
	sacrificed := cueEventIndex(g, game.EventSacrifice, spirit)
	if sacrificed < 0 {
		t.Fatal("the permanent was not sacrificed")
	}
	if cueEventIndex(g, game.EventCumulativeUpkeepNotPaid, spirit) > sacrificed {
		t.Error("the announcement came after the sacrifice: a trigger would not see its source on the battlefield")
	}
	if _, ok := effectstest.BattlefieldCard(g, spirit); ok {
		t.Error("the permanent survived a declined cumulative upkeep")
	}
}

// A "yes" the player cannot fund is the decline, and announces one (CR
// 118.3): nothing was paid.
func TestAnUnfundableCumulativeUpkeepAnnouncesTheDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	spirit := effectstest.B12Push(g, me.ID, "Sustaining Spirit", "Creature — Angel Spirit", cueSustainingSpirit, 0, 3)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("an unfundable yes: %v", err)
	}
	if cueLastEvent(g, game.EventCumulativeUpkeepNotPaid, spirit) == nil {
		t.Error("an unfundable payment announced no decline")
	}
	if cueEventIndex(g, game.EventCumulativeUpkeepPaid, spirit) >= 0 {
		t.Error("an unfundable payment announced a payment")
	}
}

// A payment that is not mana (Polar Kraken's "Cumulative upkeep—Sacrifice
// a land") announces itself with no mana spent.
func TestANonManaCumulativeUpkeepPaymentAnnouncesNoManaSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	kraken := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Polar Kraken", OracleID: effectstest.OraclePolarKraken,
		TypeLine: "Creature — Kraken", Power: 11, Toughness: 11, Owner: me.ID, Controller: me.ID,
	})
	land := effectstest.PushLand(g, me.ID, "Island")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.EchoPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no cumulative upkeep prompt")
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{land}); err != nil {
		t.Fatalf("sacrifice a land: %v", err)
	}
	ev := cueLastEvent(g, game.EventCumulativeUpkeepPaid, kraken)
	if ev == nil || ev.Amount != 1 {
		t.Fatalf("event = %+v, want a payment for 1 age counter", ev)
	}
	if len(ev.SpentSymbols) != 0 {
		t.Errorf("mana spent = %v, want none", ev.SpentSymbols)
	}
}

// The two watchers read their own source's event only.
func TestCumulativeUpkeepWatchersReadTheirOwnSource(t *testing.T) {
	paid := effects.WhenCumulativeUpkeepIsPaid("test — paid", func(*game.Game, *game.StackItem) error { return nil })
	notPaid := effects.WhenCumulativeUpkeepIsNotPaid("test — not paid", func(*game.Game, *game.StackItem) error { return nil })
	mine, theirs := uuid.New(), uuid.New()
	source := &game.Card{InstanceID: mine}
	for name, tc := range map[string]struct {
		ability game.TriggeredAbility
		event   game.Event
		want    bool
	}{
		"paid, mine":          {paid, game.Event{Kind: game.EventCumulativeUpkeepPaid, CardID: mine}, true},
		"paid, another's":     {paid, game.Event{Kind: game.EventCumulativeUpkeepPaid, CardID: theirs}, false},
		"not paid, mine":      {notPaid, game.Event{Kind: game.EventCumulativeUpkeepNotPaid, CardID: mine}, true},
		"not paid, another's": {notPaid, game.Event{Kind: game.EventCumulativeUpkeepNotPaid, CardID: theirs}, false},
	} {
		if got := tc.ability.AppliesTo(tc.event, source, game.Characteristic{}, nil); got != tc.want {
			t.Errorf("%s: applies = %v, want %v", name, got, tc.want)
		}
	}
}

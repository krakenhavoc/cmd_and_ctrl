package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_trigger_activate_test.go — ADR 0119 §5, the fuller log. A trigger
// gets a `trigger` line when it is queued and an activation an
// `activate` line, both named and redacted like the ability's resolve
// line. The engine emits EventTrigger for an activation too (a
// breadcrumb that names its source in CardID, which no trigger emit
// sets), so the projection drops an EventTrigger with a CardID; these
// tests drive the real engine so that the discriminator is pinned on
// the shape it actually emits.

const (
	logWardenOracle       = "protocol-test-log-warden"
	logPingerOracle       = "protocol-test-log-pinger"
	logCyclerOracle       = "protocol-test-log-cycler"
	logCycleWatcherOracle = "protocol-test-log-cycle-watcher"

	logWardenLabel       = "Log Warden — you gain 1 life"
	logPingerLabel       = "{1}: you gain 1 life"
	logCycleWatcherLabel = "Log Cycle Watcher — you gain 1 life"
)

func init() {
	gain := effects.Do(effects.GainLife{Amount: 1})
	// "Whenever another creature enters, you gain 1 life." Watches
	// EventETB, which writes no log line of its own, so two entries
	// in a row queue two triggers in a row.
	effects.Register(effects.Spec{
		OracleID: logWardenOracle,
		Name:     "Log Warden",
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID != source.InstanceID
			}, logWardenLabel, gain),
		},
	})
	effects.Register(effects.Spec{
		OracleID: logPingerOracle,
		Name:     "Log Pinger",
		Activated: []effects.ActivatedAbility{{
			Label:  logPingerLabel,
			Cost:   effects.ManaCost("{1}"),
			Effect: gain,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:  logCyclerOracle,
		Name:      "Log Cycler",
		Activated: []effects.ActivatedAbility{effects.Cycling("{1}")},
	})
	effects.Register(effects.Spec{
		OracleID: logCycleWatcherOracle,
		Name:     "Log Cycle Watcher",
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCycle, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			}, logCycleWatcherLabel, gain),
		},
	})
}

// catalogPermanent puts a catalog card onto the battlefield, known to
// every seat, without emitting anything.
func catalogPermanent(g *game.Game, controller uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	known := map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		known[p.ID] = true
	}
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Power: 1, Toughness: 1, Owner: controller, Controller: controller, KnownBy: known,
		})
	})
	return id
}

// enterCreature puts a vanilla creature onto the battlefield and emits
// its EventETB, the event Log Warden watches.
func enterCreature(t *testing.T, g *game.Game, controller uuid.UUID, name string) {
	t.Helper()
	id := catalogPermanent(g, controller, name, "Creature — Bear", "")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: controller, CardID: id})
	})
}

func addColorless(t *testing.T, g *game.Game, player uuid.UUID, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(player, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
}

// A harvested trigger is a line when it is queued, named by its label
// and owned by its controller, and the event it is projected from has
// no CardID — the discriminator.
func TestAHarvestedTriggerIsLoggedWhenItTriggers(t *testing.T) {
	g := buildMainPhaseGame(t)
	me := g.Seats[0]
	warden := catalogPermanent(g, me.ID, "Log Warden", "Creature — Human Cleric", logWardenOracle)

	enterCreature(t, g, me.ID, "Grizzly Bears")

	entry := findLog(t, ViewOfGame(g).Log, LogTrigger)
	if want := "P1's trigger: " + logWardenLabel; entry.Text != want {
		t.Errorf("text = %q, want %q", entry.Text, want)
	}
	if entry.CardID != warden.String() || entry.Label != logWardenLabel || entry.Seat != 0 {
		t.Errorf("entry = card_id %q label %q seat %d, want the source %s, the label and seat 0",
			entry.CardID, entry.Label, entry.Seat, warden)
	}
	if entry.Amount != 0 {
		t.Errorf("amount = %d on a single trigger, want it absent", entry.Amount)
	}
	assertNoUUID(t, entry.Text)

	found := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventTrigger && ev.Source == warden {
			found = true
			if ev.CardID != uuid.Nil || ev.Label != logWardenLabel {
				t.Errorf("the trigger's EventTrigger = card_id %s label %q; want no card_id and the label", ev.CardID, ev.Label)
			}
		}
	}
	if !found {
		t.Fatal("no EventTrigger for the harvested trigger")
	}
}

// A trigger announced by hand is the same line.
func TestAnAnnouncedTriggerIsLogged(t *testing.T) {
	g := buildMainPhaseGame(t)
	me := g.Seats[0]
	rock := battlefieldCard(t, g, "Mind Stone", "Artifact", me.ID)

	const label = "Mind Stone — draw a card"
	if err := g.AnnounceTrigger(me.ID, rock, game.AbilityParams{Label: label}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	entry := findLog(t, ViewOfGame(g).Log, LogTrigger)
	if want := "P1's trigger: " + label; entry.Text != want {
		t.Errorf("text = %q, want %q", entry.Text, want)
	}
	if entry.CardID != rock.String() {
		t.Errorf("card_id = %q, want the source %s", entry.CardID, rock)
	}
}

// The same trigger twice in a row is one line with a count, as a run of
// draws is. A different source breaks the run.
func TestConsecutiveIdenticalTriggersCollapse(t *testing.T) {
	g := buildMainPhaseGame(t)
	me := g.Seats[0]
	catalogPermanent(g, me.ID, "Log Warden", "Creature — Human Cleric", logWardenOracle)

	enterCreature(t, g, me.ID, "Bear A")
	enterCreature(t, g, me.ID, "Bear B")

	triggers := logEntriesOf(ViewOfGame(g).Log, LogTrigger)
	if len(triggers) != 1 {
		t.Fatalf("%d trigger lines, want the two collapsed into one: %v", len(triggers), triggers)
	}
	if want := "P1's trigger: " + logWardenLabel + " ×2"; triggers[0].Text != want {
		t.Errorf("text = %q, want %q", triggers[0].Text, want)
	}
	if triggers[0].Amount != 2 {
		t.Errorf("amount = %d, want 2", triggers[0].Amount)
	}

	// A second Warden is another source: its trigger is a line of its
	// own, and the counts still add up to every trigger there was.
	second := catalogPermanent(g, me.ID, "Log Warden", "Creature — Human Cleric", logWardenOracle)
	enterCreature(t, g, me.ID, "Bear C")
	total, ownLine := 0, false
	for _, e := range logEntriesOf(ViewOfGame(g).Log, LogTrigger) {
		total += max(e.Amount, 1)
		if e.CardID == second.String() {
			ownLine = true
			if e.Amount != 0 {
				t.Errorf("the second Warden's line counts %d; it triggered once", e.Amount)
			}
		}
	}
	if !ownLine {
		t.Error("the second Warden's trigger was folded into the first Warden's line")
	}
	if total != 4 {
		t.Errorf("the trigger lines count %d triggers, want 4", total)
	}
}

// A face-down source's trigger is redacted with the card, exactly as
// its resolve line is: no label, no name, no id for a seat that may not
// identify it.
func TestAFaceDownSourcesTriggerIsRedacted(t *testing.T) {
	g := buildMainPhaseGame(t)
	me, other := g.Seats[0], g.Seats[1]
	secret := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: secret, Name: "Secret Sphinx", TypeLine: "Creature — Sphinx",
			Owner: me.ID, Controller: me.ID,
			FaceDown: true, FaceDownKind: game.FaceDownMorphed,
			KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	const label = "Secret Sphinx — draw a card"
	if err := g.AnnounceTrigger(me.ID, secret, game.AbilityParams{Label: label}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}

	mine := findLog(t, FilterViewFor(ViewOfGame(g), me.ID.String()).Log, LogTrigger)
	if want := "P1's trigger: " + label; mine.Text != want {
		t.Errorf("the controller's line = %q, want %q", mine.Text, want)
	}
	theirs := findLog(t, FilterViewFor(ViewOfGame(g), other.ID.String()).Log, LogTrigger)
	if want := "P1's ability triggered"; theirs.Text != want {
		t.Errorf("a non-knower's line = %q, want %q", theirs.Text, want)
	}
	if theirs.Label != "" || theirs.CardID != "" || strings.Contains(theirs.Text, "Sphinx") {
		t.Errorf("a non-knower's entry = card_id %q label %q text %q; want nothing that identifies the card",
			theirs.CardID, theirs.Label, theirs.Text)
	}
}

// An activation is one `activate` line, and never a `trigger` line too:
// the breadcrumb EventTrigger it emits names its source in CardID.
func TestAnActivationIsLoggedOnce(t *testing.T) {
	g := buildMainPhaseGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pinger := catalogPermanent(g, me.ID, "Log Pinger", "Artifact", logPingerOracle)
	addColorless(t, g, me.ID, "{C}")

	if err := g.ActivateCatalogAbility(me.ID, pinger, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	log := ViewOfGame(g).Log
	acts := logEntriesOf(log, LogActivate)
	if len(acts) != 1 {
		t.Fatalf("%d activate lines, want 1: %v", len(acts), logKinds(log))
	}
	if want := me.Name + " activated Log Pinger — " + logPingerLabel; acts[0].Text != want {
		t.Errorf("text = %q, want %q", acts[0].Text, want)
	}
	if acts[0].CardID != pinger.String() || acts[0].Label != logPingerLabel {
		t.Errorf("entry = card_id %q label %q, want the source and the label", acts[0].CardID, acts[0].Label)
	}
	if got := logEntriesOf(log, LogTrigger); len(got) != 0 {
		t.Errorf("the activation's breadcrumb was logged as a trigger: %v", got)
	}
	breadcrumb := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventTrigger && ev.Source == pinger {
			breadcrumb = true
			if ev.CardID != pinger {
				t.Errorf("the activation's EventTrigger has card_id %s; the log needs it to name the source", ev.CardID)
			}
		}
	}
	if !breadcrumb {
		t.Error("the activation emitted no EventTrigger breadcrumb; this test no longer pins the discriminator")
	}

	resolveAbilities(t, g)
	if got := len(logEntriesOf(ViewOfGame(g).Log, LogResolve)); got != 1 {
		t.Errorf("%d resolve lines after the ability resolved, want 1", got)
	}
}

// Cycling is an activation the cycle line already tells. The activation
// after it gets no second line, even with a "whenever you cycle"
// trigger's line between the two; the next activation of another card
// does.
func TestACyclingIsNotLoggedTwice(t *testing.T) {
	g := buildMainPhaseGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	catalogPermanent(g, me.ID, "Log Cycle Watcher", "Enchantment", logCycleWatcherOracle)
	pinger := catalogPermanent(g, me.ID, "Log Pinger", "Artifact", logPingerOracle)
	cycler := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: cycler, Name: "Log Cycler", TypeLine: "Instant", OracleID: logCyclerOracle,
			Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	addColorless(t, g, me.ID, "{C}{C}")

	if err := g.ActivateCatalogAbility(me.ID, cycler, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	log := ViewOfGame(g).Log
	findLog(t, log, LogCycle)
	if trig := findLog(t, log, LogTrigger); trig.Label != logCycleWatcherLabel {
		t.Errorf("trigger line %q, want the cycle watcher's", trig.Text)
	}
	if got := logEntriesOf(log, LogActivate); len(got) != 0 {
		t.Errorf("the cycling was told twice: %v", got)
	}

	if err := g.ActivateCatalogAbility(me.ID, pinger, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	acts := logEntriesOf(ViewOfGame(g).Log, LogActivate)
	if len(acts) != 1 || acts[0].CardID != pinger.String() {
		t.Errorf("activate lines after the cycling = %v, want the one for Log Pinger", acts)
	}
}

// A face-down source's activation is redacted like its trigger.
func TestAFaceDownSourcesActivationIsRedacted(t *testing.T) {
	g := buildMainPhaseGame(t)
	me, other := g.Seats[0], g.Seats[1]
	secret := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: secret, Name: "Secret Sphinx", TypeLine: "Creature — Sphinx",
			Owner: me.ID, Controller: me.ID,
			FaceDown: true, FaceDownKind: game.FaceDownMorphed,
			KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventActivateAbility, Actor: me.ID, Source: secret, CardID: secret,
			StackItemID: uuid.New(), Label: "{U}: Secret Sphinx gains flying",
		})
	})
	mine := findLog(t, FilterViewFor(ViewOfGame(g), me.ID.String()).Log, LogActivate)
	// A face-down permanent has no name for anyone (CR 708.2a), so the
	// controller's line is the label alone.
	if want := "P1 activated {U}: Secret Sphinx gains flying"; mine.Text != want {
		t.Errorf("the controller's line = %q, want %q", mine.Text, want)
	}
	theirs := findLog(t, FilterViewFor(ViewOfGame(g), other.ID.String()).Log, LogActivate)
	if want := "P1 activated an ability"; theirs.Text != want {
		t.Errorf("a non-knower's line = %q, want %q", theirs.Text, want)
	}
	if theirs.Label != "" || theirs.CardID != "" {
		t.Errorf("a non-knower's entry kept card_id %q / label %q", theirs.CardID, theirs.Label)
	}
}

// A legendary's label usually starts with its short name; the line
// does not say the name twice.
func TestAbilityNameKnowsALegendarysShortName(t *testing.T) {
	for _, tc := range []struct{ label, card, want string }{
		{"Tatyova — gain 1 life and draw a card", "Tatyova, Benthic Druid", "Tatyova — gain 1 life and draw a card"},
		{"Syr Konrad, the Grim — 1 damage to each opponent", "Syr Konrad, the Grim", "Syr Konrad, the Grim — 1 damage to each opponent"},
		{"{T}: deal 1 damage to any target", "Prodigal Pyromancer", "Prodigal Pyromancer — {T}: deal 1 damage to any target"},
		// A word that merely begins the label is not the name.
		{"Tatyovas everywhere", "Tatyova, Benthic Druid", "Tatyova, Benthic Druid — Tatyovas everywhere"},
	} {
		if got := abilityName(tc.label, tc.card); got != tc.want {
			t.Errorf("abilityName(%q, %q) = %q, want %q", tc.label, tc.card, got, tc.want)
		}
	}
}

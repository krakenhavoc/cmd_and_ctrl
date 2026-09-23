package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_ability_resolve_test.go — #1257. A triggered or activated ability
// resolved in the public log as "a card resolved": the engine's ability
// branch emits EventResolve with the item's source and label and no
// CardID (an ability has no card on the stack), and the projection
// read only CardID. Two deliberate silences — EventTrigger and
// EventActivateAbility in log_event_kind_gate_test.go — rest on "told
// by the LogResolve of the ability it becomes", so that line has to
// actually tell it.
//
// Every test here drives the real engine (announce, pass, resolve)
// rather than emitting a hand-built event, because the bug was the
// SHAPE the engine emits, and a hand-built event would pin the shape
// the test author assumed.

// buildMainPhaseGame is buildActiveGame with the mulligans kept and the
// cursor in the first main phase, where somebody holds priority.
func buildMainPhaseGame(t *testing.T) *game.Game {
	t.Helper()
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	for i := 0; i < 16 && g.Turn.Step != game.StepPrecombatMain; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g
}

// resolveAbilities passes priority until nothing is left on the stack
// or waiting to go on it.
func resolveAbilities(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		empty := true
		g.WithWriteLock(func() {
			empty = len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0 && g.Stack.Size() == 0
		})
		if empty {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
}

// onlyStackItem returns the id of the single item on the stack.
func onlyStackItem(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.WithWriteLock(func() {
		if len(g.StackMeta) != 1 {
			t.Fatalf("want one item on the stack, have %d", len(g.StackMeta))
		}
		for k := range g.StackMeta {
			id = k
		}
	})
	return id
}

// logEntriesOf returns every entry of the given kind.
func logEntriesOf(entries []LogEvent, kind LogKind) []LogEvent {
	var out []LogEvent
	for _, e := range entries {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

func TestATriggeredAbilityResolvesUnderItsLabel(t *testing.T) {
	g := buildMainPhaseGame(t)
	me := g.Seats[0]
	rock := battlefieldCard(t, g, "Mind Stone", "Artifact", me.ID)

	const label = "Mind Stone — draw a card"
	if err := g.AnnounceTrigger(me.ID, rock, game.AbilityParams{Label: label}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	resolveAbilities(t, g)

	entry := findLog(t, ViewOfGame(g).Log, LogResolve)
	// The label already starts with the source's name, so it is not
	// said twice.
	if want := label + " resolved"; entry.Text != want {
		t.Errorf("text = %q, want %q", entry.Text, want)
	}
	if entry.Label != label {
		t.Errorf("label = %q, want %q", entry.Label, label)
	}
	if entry.CardID != rock.String() {
		t.Errorf("card_id = %q, want the ability's source %s", entry.CardID, rock)
	}
	if entry.Seat != 0 {
		t.Errorf("seat = %d, want the controller's 0", entry.Seat)
	}
	assertNoUUID(t, entry.Text)
}

// An activation's label is its cost and effect, which does not say the
// card — these were the lines that were truly anonymous.
func TestAnActivatedAbilityNamesItsSource(t *testing.T) {
	g := buildMainPhaseGame(t)
	me := g.Seats[0]
	seer := battlefieldCard(t, g, "Viscera Seer", "Creature — Vampire Wizard", me.ID)

	const label = "Sacrifice a creature: scry 1"
	if err := g.ActivateAbility(me.ID, seer, game.AbilityParams{Label: label}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	resolveAbilities(t, g)

	entry := findLog(t, ViewOfGame(g).Log, LogResolve)
	if want := "Viscera Seer — Sacrifice a creature: scry 1 resolved"; entry.Text != want {
		t.Errorf("text = %q, want %q", entry.Text, want)
	}
	// And a seat that can see the card reads the same line.
	theirs := findLog(t, FilterViewFor(ViewOfGame(g), g.Seats[1].ID.String()).Log, LogResolve)
	if theirs.Text != entry.Text || theirs.Label != label {
		t.Errorf("a knower's line = %q / label %q, want the unredacted one", theirs.Text, theirs.Label)
	}
}

// A CR 707.10 copy of an ability shares its source and label, so the
// copy's resolution names the same thing — the table sees two
// resolutions of the ability, which is what happened.
func TestACopiedAbilityResolvesUnderTheSameName(t *testing.T) {
	g := buildMainPhaseGame(t)
	me := g.Seats[0]
	rock := battlefieldCard(t, g, "Mind Stone", "Artifact", me.ID)

	const label = "Mind Stone — draw a card"
	if err := g.AnnounceTrigger(me.ID, rock, game.AbilityParams{Label: label}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	item := onlyStackItem(t, g)
	g.WithWriteLock(func() {
		if err := g.CopyAbilityForEffect(item, me.ID, false); err != nil {
			t.Fatalf("CopyAbilityForEffect: %v", err)
		}
	})
	resolveAbilities(t, g)

	resolves := logEntriesOf(ViewOfGame(g).Log, LogResolve)
	if len(resolves) != 2 {
		t.Fatalf("%d resolve entries, want the copy's and the original's", len(resolves))
	}
	for _, e := range resolves {
		if e.Text != label+" resolved" {
			t.Errorf("resolve text = %q, want %q", e.Text, label+" resolved")
		}
	}
}

// A face-down source is the leak the redaction has to hold: the label
// names the card as loudly as its name does. The controller reads the
// line, a seat that may not identify the permanent reads that an
// ability resolved — no label, no name, and no id.
func TestAFaceDownSourcesAbilityIsRedactedWithTheCard(t *testing.T) {
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
	resolveAbilities(t, g)

	mine := findLog(t, FilterViewFor(ViewOfGame(g), me.ID.String()).Log, LogResolve)
	if want := label + " resolved"; mine.Text != want {
		t.Errorf("the controller's line = %q, want %q", mine.Text, want)
	}

	theirs := findLog(t, FilterViewFor(ViewOfGame(g), other.ID.String()).Log, LogResolve)
	if want := "an ability resolved"; theirs.Text != want {
		t.Errorf("a non-knower's line = %q, want %q", theirs.Text, want)
	}
	if theirs.Label != "" {
		t.Errorf("the label survived the redaction as %q", theirs.Label)
	}
	if theirs.CardID != "" {
		t.Errorf("the source's id survived the redaction as %q", theirs.CardID)
	}
	if strings.Contains(theirs.Text, "Sphinx") {
		t.Errorf("a non-knower's line names the face-down card: %q", theirs.Text)
	}
}

// A source in a HIDDEN zone: the instance ID is itself the leak there,
// because the zone filter never hands a non-knower the ids of an
// opponent's hand — so the id goes with the name.
func TestAHiddenSourcesAbilityShipsNoIDToANonKnower(t *testing.T) {
	g := buildMainPhaseGame(t)
	me, other := g.Seats[0], g.Seats[1]
	inHand := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: inHand, Name: "Hidden Card", TypeLine: "Instant",
			Owner: me.ID, Controller: me.ID,
			KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	if err := g.AnnounceTrigger(me.ID, inHand, game.AbilityParams{Label: "Hidden Card — reveal"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	resolveAbilities(t, g)

	theirs := findLog(t, FilterViewFor(ViewOfGame(g), other.ID.String()).Log, LogResolve)
	if theirs.CardID != "" || theirs.Label != "" || theirs.Text != "an ability resolved" {
		t.Errorf("a non-knower's entry = card_id %q label %q text %q; want nothing that identifies the card",
			theirs.CardID, theirs.Label, theirs.Text)
	}
	// Redaction is idempotent: a second pass over the filtered log
	// does not re-reveal what the first took away.
	again := redactLogForViewer([]LogEvent{theirs}, func(CardView) bool { return true })
	if again[0].Text != theirs.Text {
		t.Errorf("a second pass re-rendered the line as %q", again[0].Text)
	}
}

// An ability countered by game rules (CR 608.2b) is named the same way.
func TestAnAbilityThatFizzlesIsNamed(t *testing.T) {
	g := buildMainPhaseGame(t)
	me := g.Seats[0]
	seer := battlefieldCard(t, g, "Prodigal Pyromancer", "Creature — Human Wizard", me.ID)
	victim := battlefieldCard(t, g, "Grizzly Bears", "Creature — Bear", g.Seats[1].ID)

	const label = "{T}: deal 1 damage to any target"
	if err := g.ActivateAbility(me.ID, seer, game.AbilityParams{
		Label:   label,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	g.WithWriteLock(func() {
		if _, err := g.Battlefield.Remove(victim); err != nil {
			t.Fatalf("remove the target: %v", err)
		}
	})
	resolveAbilities(t, g)

	entry := findLog(t, ViewOfGame(g).Log, LogFizzle)
	if want := "Prodigal Pyromancer — " + label + " was countered by game rules (no legal targets)"; entry.Text != want {
		t.Errorf("text = %q, want %q", entry.Text, want)
	}
}

// A spell's resolve line is unchanged: it carries its own card and no
// label, and renders as it always has.
func TestASpellsResolveLineIsUnchanged(t *testing.T) {
	g := buildMainPhaseGame(t)
	bolt := battlefieldCard(t, g, "Lightning Bolt", "Instant", g.Seats[0].ID)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Source: bolt, CardID: bolt})
	})
	entry := findLog(t, ViewOfGame(g).Log, LogResolve)
	if entry.Text != "Lightning Bolt resolved" || entry.Label != "" {
		t.Errorf("spell resolve = %q / label %q, want the unchanged line and no label", entry.Text, entry.Label)
	}
}

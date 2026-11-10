package game

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

// format_test.go: a format on the table (ADR 0222 §1).

// buildSixtyCardDeck is a constructed deck: sixty cards, no commander.
func buildSixtyCardDeck() []Card {
	deck := make([]Card, 0, 60)
	for i := 0; i < 60; i++ {
		deck = append(deck, NewCard(fmt.Sprintf("Filler Card %d", i+1), uuid.Nil))
	}
	return deck
}

func setFormat(t *testing.T, g *Game, f Format) {
	t.Helper()
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{Format: &f}); err != nil {
		t.Fatalf("UpdateSettings(format %q): %v", f, err)
	}
}

// A new table is a Commander table, and its settings carry no format
// field at all, so its snapshot is byte for byte what it was.
func TestANewTableIsCommander(t *testing.T) {
	g := NewGame()
	if got := g.Settings.EffectiveFormat(); got != FormatCommander {
		t.Errorf("format = %q, want commander", got)
	}
	if g.Settings.Format != "" {
		t.Errorf("stored format = %q, want empty", g.Settings.Format)
	}
	if got := g.SeatCap(); got != MaxPlayers {
		t.Errorf("seat cap = %d, want %d", got, MaxPlayers)
	}
	if g.Settings != DefaultTableSettings() {
		t.Errorf("settings = %+v, want the defaults", g.Settings)
	}
	raw, err := json.Marshal(g.Settings)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte("format")) {
		t.Errorf("a Commander table's settings serialise a format: %s", raw)
	}
}

// Naming Commander on a Commander table changes nothing and logs
// nothing: a saved setup's complete patch always names the format.
func TestNamingCommanderIsANoOp(t *testing.T) {
	g := NewGame()
	before, events := g.Settings, len(g.Events)
	setFormat(t, g, FormatCommander)
	if g.Settings != before {
		t.Errorf("settings changed: %+v -> %+v", before, g.Settings)
	}
	if len(g.Events) != events {
		t.Errorf("%d events emitted for a no-op", len(g.Events)-events)
	}
	empty := Format("")
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{Format: &empty}); err != nil {
		t.Fatalf("an empty format is Commander, got %v", err)
	}
	if g.Settings != before {
		t.Errorf("settings changed by an empty format: %+v", g.Settings)
	}
}

// The format table is data (ADR 0222 §1): Commander's row is what the
// engine's constants always said, Standard's is CR 103.4's 20 life, two
// seats, and CR 100.2a / 100.4a's deck numbers.
func TestFormatRows(t *testing.T) {
	c := FormatCommander.Spec()
	if c.StartingLife != StartingLife || c.MaxSeats != MaxPlayers || !c.CommanderExpected {
		t.Errorf("commander row = %+v", c)
	}
	if c.Deck != (DeckRules{MinCards: 100, ExactCards: true, MaxCopies: 1, Legality: "commander"}) {
		t.Errorf("commander deck rules = %+v", c.Deck)
	}
	s := FormatStandard.Spec()
	if s.StartingLife != 20 || s.MaxSeats != 2 || s.CommanderExpected {
		t.Errorf("standard row = %+v", s)
	}
	if s.Deck != (DeckRules{MinCards: 60, MaxCopies: 4, SideboardSize: 15, Legality: "standard"}) {
		t.Errorf("standard deck rules = %+v", s.Deck)
	}
	for _, f := range Formats() {
		if f.MaxSeats < MinPlayers || f.MaxSeats > MaxPlayers {
			t.Errorf("%s: %d seats is outside %d..%d; MaxPlayers stays the global bound", f.Format, f.MaxSeats, MinPlayers, MaxPlayers)
		}
		if f.StartingLife < MinStartingLife || f.StartingLife > MaxStartingLife {
			t.Errorf("%s: starting life %d is out of range", f.Format, f.StartingLife)
		}
	}
	if Formats()[0].Format != FormatCommander {
		t.Error("Commander is the primary format and is listed first")
	}
	if _, ok := LookupFormat("vintage"); ok {
		t.Error("an unknown format has a row")
	}
	if got := Format("vintage").Spec().Format; got != FormatCommander {
		t.Errorf("an unknown format reads as %q, want commander", got)
	}
}

// Choosing Standard sets the format's starting life and seat cap, and
// says so in the log, the format's line first.
func TestStandardTableSettings(t *testing.T) {
	g := NewGame()
	setFormat(t, g, FormatStandard)
	if got := g.Settings.EffectiveFormat(); got != FormatStandard {
		t.Fatalf("format = %q", got)
	}
	if g.Settings.StartingLife != 20 {
		t.Errorf("starting life = %d, want 20", g.Settings.StartingLife)
	}
	if got := g.SeatCap(); got != 2 {
		t.Errorf("seat cap = %d, want 2", got)
	}
	var labels []string
	for _, e := range g.Events {
		if e.Kind == EventSettingsChanged {
			labels = append(labels, e.Label+"="+e.SettingNew)
		}
	}
	want := []string{SettingFormat + "=standard", SettingStartingLife + "=20"}
	if fmt.Sprint(labels) != fmt.Sprint(want) {
		t.Errorf("settings events = %v, want %v", labels, want)
	}

	// Back to Commander: its own starting life again.
	setFormat(t, g, FormatCommander)
	if g.Settings != DefaultTableSettings() {
		t.Errorf("back at Commander the settings are %+v, want the defaults", g.Settings)
	}

	// A patch naming both keeps the life it names.
	life, std := 25, FormatStandard
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{Format: &std, StartingLife: &life}); err != nil {
		t.Fatal(err)
	}
	if g.Settings.StartingLife != 25 {
		t.Errorf("starting life = %d, want the patch's 25", g.Settings.StartingLife)
	}
}

func TestUnknownFormatIsRefused(t *testing.T) {
	g := NewGame()
	bad := Format("vintage")
	err := g.UpdateSettings(uuid.Nil, SettingsPatch{Format: &bad})
	if !errors.Is(err, ErrInvalidSetting) {
		t.Fatalf("err = %v, want ErrInvalidSetting", err)
	}
	if g.Settings != DefaultTableSettings() {
		t.Errorf("a refused patch changed the settings: %+v", g.Settings)
	}
}

// A Standard table seats two. A Commander table still seats four.
func TestSeatCapIsTheFormats(t *testing.T) {
	g := NewGame()
	setFormat(t, g, FormatStandard)
	for i := 0; i < 2; i++ {
		p, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), buildSixtyCardDeck())
		if err != nil {
			t.Fatalf("seat %d: %v", i, err)
		}
		if p.Life != 20 {
			t.Errorf("seat %d sat down at %d life, want 20", i, p.Life)
		}
	}
	if _, err := g.AddPlayer("P3", buildSixtyCardDeck()); !errors.Is(err, ErrGameFull) {
		t.Errorf("a third seat at a Standard table: err = %v, want ErrGameFull", err)
	}

	c := NewGame()
	for i := 0; i < MaxPlayers; i++ {
		if _, err := c.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck("C")); err != nil {
			t.Fatalf("commander seat %d: %v", i, err)
		}
	}
	if _, err := c.AddPlayer("P5", buildTestDeck("C")); !errors.Is(err, ErrGameFull) {
		t.Errorf("a fifth seat at a Commander table: err = %v, want ErrGameFull", err)
	}
	// Three seated: no format with two seats.
	std := FormatStandard
	if err := c.UpdateSettings(uuid.Nil, SettingsPatch{Format: &std}); !errors.Is(err, ErrFormatSeats) {
		t.Errorf("Standard with four seated: err = %v, want ErrFormatSeats", err)
	}
	if c.Settings != DefaultTableSettings() {
		t.Errorf("a refused format change changed the settings: %+v", c.Settings)
	}
}

// The format is fixed once the game starts (ADR 0222 §1).
func TestFormatIsLockedAfterStart(t *testing.T) {
	g := NewGame()
	for i := 0; i < 2; i++ {
		_, _ = g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck("C"))
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatal(err)
	}
	before := g.Settings
	std := FormatStandard
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{Format: &std}); !errors.Is(err, ErrFormatLocked) {
		t.Fatalf("err = %v, want ErrFormatLocked", err)
	}
	if g.Settings != before {
		t.Errorf("a refused format change changed the settings")
	}
	// Naming the format it already has is not a change.
	cmd := FormatCommander
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{Format: &cmd}); err != nil {
		t.Errorf("naming the current format after Start: %v", err)
	}
}

// The slice's headline (ADR 0222 slice 1): a two-seat Standard game
// starts at 20 life, and the player who plays first draws no card on
// turn one. CR 103.8a: "In a two-player game, the player who plays
// first skips the draw step (see rule 504, "Draw Step") of their first
// turn." The rule keys on the number of players, not the format; the
// four-seat Commander case (CR 103.8c) is TestFourSeatGameSkipsNoDraw.
func TestTwoSeatStandardGameStartsAtTwentyAndSkipsTheFirstDraw(t *testing.T) {
	g := NewGame()
	setFormat(t, g, FormatStandard)
	for i := 0; i < 2; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), buildSixtyCardDeck()); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if p.Life != 20 {
			t.Errorf("%s starts at %d life, want 20", p.Name, p.Life)
		}
		if n := p.Command.Size(); n != 0 {
			t.Errorf("%s has %d cards in the command zone, want none", p.Name, n)
		}
		if got := p.Hand.Size() + p.Library.Size(); got != 60 {
			t.Errorf("%s has %d cards between hand and library, want 60", p.Name, got)
		}
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	if g.HasCommander() {
		t.Error("a table of sixty-card decks reports a commander")
	}

	first := g.Seats[g.StartingSeat]
	if g.Turn.ActiveSeat != g.StartingSeat || g.Turn.Round != 1 {
		t.Fatalf("turn = round %d seat %d, want round 1 seat %d", g.Turn.Round, g.Turn.ActiveSeat, g.StartingSeat)
	}
	hand := first.Hand.Size()
	for g.Turn.Step != StepDraw {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat != g.StartingSeat || g.Turn.Round != 1 {
		t.Fatalf("the walk left turn one")
	}
	if got := first.Hand.Size(); got != hand {
		t.Errorf("the first player drew on turn one: hand %d -> %d (CR 103.8a)", hand, got)
	}

	// The second player draws on their first turn.
	second := g.Seats[1-g.StartingSeat]
	hand = second.Hand.Size()
	for i := 0; i < 64 && !(g.Turn.ActiveSeat == second.Seat && g.Turn.Step == StepDraw); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat != second.Seat || g.Turn.Step != StepDraw {
		t.Fatalf("never reached the second player's draw step")
	}
	if got := second.Hand.Size(); got != hand+1 {
		t.Errorf("the second player's hand went %d -> %d on their first draw step, want +1", hand, got)
	}
}

// CR 103.8c: "In all other multiplayer games, no player skips the draw
// step of their first turn." A four-seat Commander table is unchanged.
func TestFourSeatGameSkipsNoDraw(t *testing.T) {
	g := NewGame()
	for i := 0; i < 4; i++ {
		_, _ = g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck("C"))
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatal(err)
	}
	for _, p := range g.Seats {
		if p.Life != StartingLife {
			t.Errorf("%s starts at %d life, want %d", p.Name, p.Life, StartingLife)
		}
		_ = g.KeepHand(p.ID)
	}
	if !g.HasCommander() {
		t.Error("a Commander table reports no commander")
	}
	first := g.Seats[g.StartingSeat]
	hand := first.Hand.Size()
	for g.Turn.Step != StepDraw {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if got := first.Hand.Size(); got != hand+1 {
		t.Errorf("the first player's hand went %d -> %d at a four-seat table, want +1 (CR 103.8c)", hand, got)
	}
}

// A Standard table's format survives a snapshot, and a Commander
// table's snapshot still has no format key.
func TestFormatRoundTripsThroughASnapshot(t *testing.T) {
	g := NewGame()
	setFormat(t, g, FormatStandard)
	for i := 0; i < 2; i++ {
		_, _ = g.AddPlayer(fmt.Sprintf("P%d", i+1), buildSixtyCardDeck())
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	back, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if got := back.Settings.EffectiveFormat(); got != FormatStandard {
		t.Errorf("restored format = %q, want standard", got)
	}
	if back.SeatCap() != 2 || back.Settings.StartingLife != 20 {
		t.Errorf("restored seat cap %d, starting life %d", back.SeatCap(), back.Settings.StartingLife)
	}

	c := NewGame()
	for i := 0; i < 2; i++ {
		_, _ = c.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck("C"))
	}
	if err := c.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(c.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var generic struct {
		Settings map[string]json.RawMessage `json:"settings"`
	}
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatal(err)
	}
	if _, has := generic.Settings["format"]; has {
		t.Error("a Commander table's snapshot carries a format key")
	}
}

// Every restore point in the fixture corpus was written before formats
// existed: none has a format key, and each restores as a Commander
// table with four seats (ADR 0222 §1). The fixtures are read, never
// edited.
func TestPreFormatFixturesRestoreAsCommander(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "snapshots", "*", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no snapshot fixtures: %v", err)
	}
	restored := 0
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		// A corpus file is a restore point, {seq, snapshot}; a file
		// under real/ may be the bare snapshot.
		var probe struct {
			Snapshot *struct {
				Settings map[string]json.RawMessage `json:"settings"`
			} `json:"snapshot"`
			Settings map[string]json.RawMessage `json:"settings"`
		}
		if err := json.Unmarshal(raw, &probe); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		settings := probe.Settings
		if probe.Snapshot != nil {
			settings = probe.Snapshot.Settings
		}
		if _, has := settings["format"]; has {
			// A fixture recorded after this change, at a table of
			// another format: not a pre-format file.
			continue
		}
		var file struct {
			Snapshot *GameSnapshot `json:"snapshot"`
		}
		var snap GameSnapshot
		if err := json.Unmarshal(raw, &file); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if file.Snapshot != nil {
			snap = *file.Snapshot
		} else if err := json.Unmarshal(raw, &snap); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if got := snap.Settings.EffectiveFormat(); got != FormatCommander {
			t.Errorf("%s decodes as %q, want commander", f, got)
		}
		g, err := snap.Restore()
		if err != nil {
			// Some boards need the real catalog to restore; the corpus
			// test in internal/cards/effects covers those.
			continue
		}
		restored++
		if got := g.Settings.EffectiveFormat(); got != FormatCommander {
			t.Errorf("%s restored as %q, want commander", f, got)
		}
		if g.Settings.Format != "" {
			t.Errorf("%s restored with a stored format %q", f, g.Settings.Format)
		}
		if got := g.SeatCap(); got != MaxPlayers {
			t.Errorf("%s restored with %d seats, want %d", f, got, MaxPlayers)
		}
		// Capturing it again adds no format key.
		again, err := json.Marshal(g.CaptureSnapshot().Settings)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(again, []byte(`"format"`)) {
			t.Errorf("%s recaptured with a format key: %s", f, again)
		}
	}
	if restored < 100 {
		t.Fatalf("only %d of %d fixtures restored; the test proves too little", restored, len(files))
	}
	t.Logf("%d of %d pre-format fixtures restored as Commander", restored, len(files))
}

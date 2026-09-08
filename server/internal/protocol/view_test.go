package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func buildActiveGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := range 2 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := range 10 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g
}

func TestViewOfGameStructuralFields(t *testing.T) {
	g := buildActiveGame(t)
	v := ViewOfGame(g)

	if v.ID != g.ID.String() {
		t.Errorf("ID: got %q, want %q", v.ID, g.ID.String())
	}
	if v.State != "active" {
		t.Errorf("State: got %q, want %q", v.State, "active")
	}
	if len(v.Seats) != 2 {
		t.Errorf("Seats: got %d, want 2", len(v.Seats))
	}
	if v.Turn.Step != "untap" {
		t.Errorf("Turn.Step: got %q, want %q", v.Turn.Step, "untap")
	}
	if v.Battlefield.Kind != "battlefield" {
		t.Errorf("Battlefield.Kind: got %q", v.Battlefield.Kind)
	}
	if v.Battlefield.Owner != "" {
		t.Errorf("shared zone should have empty owner string, got %q", v.Battlefield.Owner)
	}
}

// TestS135FilterViewRedactsUnknownCards verifies the per-card S13.5
// redaction: cards the viewer doesn't know have their printed
// characteristics zeroed; cards the viewer knows keep theirs;
// known_by_you reflects the lookup.
func TestS135FilterViewRedactsUnknownCards(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	other := g.Seats[1]

	// Park a known-to-owner card on the battlefield + an unknown card.
	knownID := uuid.New()
	unknownID := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: knownID,
			Name:       "Known Bear",
			TypeLine:   "Creature — Bear",
			Power:      2,
			Toughness:  2,
			Owner:      owner.ID,
			Controller: owner.ID,
			KnownBy:    map[uuid.UUID]bool{owner.ID: true},
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: unknownID,
			Name:       "Mystery Card",
			TypeLine:   "Sorcery",
			Owner:      other.ID,
			Controller: other.ID,
		})
	})

	v := ViewOfGame(g)
	filtered := FilterViewFor(v, owner.ID.String())

	var knownView, unknownView *CardView
	for i := range filtered.Battlefield.Cards {
		c := &filtered.Battlefield.Cards[i]
		switch c.InstanceID {
		case knownID.String():
			knownView = c
		case unknownID.String():
			unknownView = c
		}
	}
	if knownView == nil || unknownView == nil {
		t.Fatalf("setup: didn't find both cards in filtered view")
	}
	if knownView.Name != "Known Bear" || !knownView.KnownByYou {
		t.Errorf("known card redacted unexpectedly: %+v", knownView)
	}
	if unknownView.Name != "" || unknownView.TypeLine != "" || unknownView.KnownByYou {
		t.Errorf("unknown card not redacted: %+v", unknownView)
	}
	// Instance ID + controller stay intact even on redacted cards.
	if unknownView.Controller != other.ID.String() {
		t.Errorf("controller field stripped on redacted card: got %q", unknownView.Controller)
	}
}

// TestViewOfGameS131StackScaffold verifies that the new S13.1 stack
// metadata fields surface on the wire when populated. Builds a
// synthetic StackItem directly on the game (the cast/activate
// mutations land in sub-PR 2; this just verifies the projection),
// then asserts the wire view picks it up under stack_items and
// preserves announce-time targets / X / split-second.
func TestViewOfGameS131StackScaffold(t *testing.T) {
	g := buildActiveGame(t)
	caster := g.Seats[0]
	target := g.Seats[1]

	itemID := uuid.New()
	g.WithWriteLock(func() {
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*game.StackItem)
		}
		g.StackMeta[itemID] = &game.StackItem{
			ID:           itemID,
			Kind:         game.StackItemSpell,
			Controller:   caster.ID,
			Owner:        caster.ID,
			SourceCardID: itemID,
			Targets: []game.TargetRef{
				{Kind: game.TargetPlayer, ID: target.ID},
			},
			XValue:      3,
			SplitSecond: true,
		}
		g.SplitSecondActive = true
	})

	v := ViewOfGame(g)
	if !v.SplitSecondActive {
		t.Errorf("SplitSecondActive on the wire: got false, want true")
	}
	if len(v.StackItems) != 1 {
		t.Fatalf("StackItems on the wire: got %d, want 1", len(v.StackItems))
	}
	got := v.StackItems[0]
	if got.ID != itemID.String() {
		t.Errorf("StackItem.ID: got %q, want %q", got.ID, itemID.String())
	}
	if got.Kind != string(game.StackItemSpell) {
		t.Errorf("StackItem.Kind: got %q, want %q", got.Kind, game.StackItemSpell)
	}
	if got.Controller != caster.ID.String() {
		t.Errorf("StackItem.Controller: got %q, want %q", got.Controller, caster.ID.String())
	}
	if got.XValue != 3 {
		t.Errorf("StackItem.XValue: got %d, want 3", got.XValue)
	}
	if !got.SplitSecond {
		t.Errorf("StackItem.SplitSecond: got false, want true")
	}
	if len(got.Targets) != 1 || got.Targets[0].ID != target.ID.String() {
		t.Errorf("StackItem.Targets: got %+v, want one player target referencing seat 1", got.Targets)
	}
}

// TestViewOfGameS13Fields verifies StartingSeat surfaces on the wire
// and the PriorityHolder == NoPriority sentinel round-trips as -1.
// The buildActiveGame helper does not close mulligans, so the cursor
// sits at Untap with PriorityHolder=NoPriority — the post-Start
// state the wire view must serialise faithfully.
func TestViewOfGameS13Fields(t *testing.T) {
	g := buildActiveGame(t)
	v := ViewOfGame(g)
	if v.Turn.PriorityHolder != -1 {
		t.Errorf("PriorityHolder on Untap (mulligans open): got %d, want -1",
			v.Turn.PriorityHolder)
	}
	if v.StartingSeat != 0 {
		t.Errorf("StartingSeat: got %d, want 0", v.StartingSeat)
	}

	// JSON round-trip — both fields must survive marshal/unmarshal as
	// integers (not omitted, not stringified).
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !bytes.Contains(raw, []byte(`"priority_holder":-1`)) {
		t.Errorf("JSON missing priority_holder=-1: %s", raw)
	}
	if !bytes.Contains(raw, []byte(`"starting_seat":0`)) {
		t.Errorf("JSON missing starting_seat=0: %s", raw)
	}

	var decoded GameView
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Turn.PriorityHolder != -1 {
		t.Errorf("decoded PriorityHolder: got %d, want -1", decoded.Turn.PriorityHolder)
	}
	if decoded.StartingSeat != 0 {
		t.Errorf("decoded StartingSeat: got %d, want 0", decoded.StartingSeat)
	}
}

func TestViewOfGamePlayerContents(t *testing.T) {
	g := buildActiveGame(t)
	v := ViewOfGame(g)
	p := v.Seats[0]

	if p.Life != 40 {
		t.Errorf("Life: got %d, want 40", p.Life)
	}
	if p.Command.Count != 1 {
		t.Errorf("command zone count: got %d, want 1", p.Command.Count)
	}
	// Library was 10 fillers; Start deals 7 into the opening hand.
	if p.Library.Count != 3 {
		t.Errorf("library count: got %d, want 3 (10 - 7 opening hand)", p.Library.Count)
	}
	if p.Hand.Count != 7 {
		t.Errorf("hand count: got %d, want 7 (opening hand)", p.Hand.Count)
	}
	if p.Command.Owner == "" {
		t.Error("private zone should have non-empty owner")
	}
	if !p.Command.Cards[0].IsCommander {
		t.Error("commander view should have IsCommander set")
	}
}

func TestViewOfCardBattlefieldPosition(t *testing.T) {
	g := buildActiveGame(t)
	p := g.Seats[0]
	// Put a card on the battlefield and stamp a position, then verify
	// the projection carries BattleX / BattleY on the wire.
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)
	_ = g.SetBattlefieldPosition(card.InstanceID, 0.3, 0.7)

	v := ViewOfGame(g)
	var found bool
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID == card.InstanceID.String() {
			found = true
			if c.BattleX != 0.3 || c.BattleY != 0.7 {
				t.Errorf("BattleX/Y: got (%v, %v), want (0.3, 0.7)", c.BattleX, c.BattleY)
			}
		}
	}
	if !found {
		t.Fatalf("card not on battlefield view")
	}

	// JSON must round-trip the new fields under their snake_case tags.
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !bytes.Contains(raw, []byte(`"battle_x":0.3`)) {
		t.Errorf("battle_x not in JSON: %s", raw)
	}
	if !bytes.Contains(raw, []byte(`"battle_y":0.7`)) {
		t.Errorf("battle_y not in JSON: %s", raw)
	}
}

func TestViewOfGameJSONRoundTrip(t *testing.T) {
	g := buildActiveGame(t)
	v := ViewOfGame(g)

	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back GameView
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.ID != v.ID {
		t.Errorf("id: got %q, want %q", back.ID, v.ID)
	}
	if back.Turn != v.Turn {
		t.Errorf("turn: got %+v, want %+v", back.Turn, v.Turn)
	}
	if len(back.Seats) != len(v.Seats) {
		t.Errorf("seats: got %d, want %d", len(back.Seats), len(v.Seats))
	}
}

func TestViewOfCardOmitsEmptyCounters(t *testing.T) {
	owner := uuid.New()
	c := game.NewCard("Sol Ring", owner)
	view := viewOfCard(c)

	raw, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// counters is omitempty — should not appear in the output at all.
	if contains := string(raw); containsKey(contains, "counters") {
		t.Errorf("counters field should be omitted when empty, got %s", contains)
	}
}

func containsKey(s, key string) bool {
	needle := `"` + key + `"`
	for i := 0; i+len(needle) <= len(s); i++ {
		if s[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// buildViewWithCardsInHands creates a 2-seat active game and returns
// the resulting GameView. As of S08, Start deals an opening hand of
// game.OpeningHandSize per seat, so each hand is non-empty out of the
// gate without a manual draw.
func buildViewWithCardsInHands(t *testing.T) GameView {
	t.Helper()
	g := buildActiveGame(t)
	return ViewOfGame(g)
}

func TestFilterViewForOwnSeatSeesHand(t *testing.T) {
	v := buildViewWithCardsInHands(t)
	own := v.Seats[0].ID

	filtered := FilterViewFor(v, own)

	if filtered.Seats[0].Hand.Count != 7 {
		t.Errorf("own hand count: got %d, want 7", filtered.Seats[0].Hand.Count)
	}
	if len(filtered.Seats[0].Hand.Cards) != 7 {
		t.Errorf("own hand.cards len: got %d, want 7", len(filtered.Seats[0].Hand.Cards))
	}
	// Own library cards also visible.
	if len(filtered.Seats[0].Library.Cards) == 0 {
		t.Error("own library.cards should be visible, got empty slice")
	}
}

func TestFilterViewForOpponentHandHidden(t *testing.T) {
	v := buildViewWithCardsInHands(t)
	own := v.Seats[0].ID

	filtered := FilterViewFor(v, own)

	// Opponent (seat 1): count preserved, cards zeroed.
	if filtered.Seats[1].Hand.Count != 7 {
		t.Errorf("opponent hand count: got %d, want 7 (count must survive filter)", filtered.Seats[1].Hand.Count)
	}
	if len(filtered.Seats[1].Hand.Cards) != 0 {
		t.Errorf("opponent hand.cards len: got %d, want 0 (must be hidden)", len(filtered.Seats[1].Hand.Cards))
	}
	// Opponent library contents hidden too (count preserved).
	if filtered.Seats[1].Library.Count == 0 {
		t.Error("opponent library count should be non-zero (preserved)")
	}
	if len(filtered.Seats[1].Library.Cards) != 0 {
		t.Errorf("opponent library.cards should be hidden, got %d", len(filtered.Seats[1].Library.Cards))
	}
	// Graveyard and Command zones remain visible.
	if filtered.Seats[1].Command.Count == 0 {
		t.Error("opponent command zone should be visible (count>0 at game start)")
	}
}

func TestFilterViewForSpectatorHidesAllHands(t *testing.T) {
	v := buildViewWithCardsInHands(t)

	// viewerID = "" is the spectator case: no seat matches, so every
	// seat is treated as an opponent.
	filtered := FilterViewFor(v, "")

	for i, seat := range filtered.Seats {
		if len(seat.Hand.Cards) != 0 {
			t.Errorf("seat %d hand.cards should be hidden from spectator, got %d", i, len(seat.Hand.Cards))
		}
	}
}

func TestFilterViewForDoesNotMutateInput(t *testing.T) {
	v := buildViewWithCardsInHands(t)
	own := v.Seats[0].ID

	origSeat1HandLen := len(v.Seats[1].Hand.Cards)
	_ = FilterViewFor(v, own)
	if got := len(v.Seats[1].Hand.Cards); got != origSeat1HandLen {
		t.Errorf("FilterViewFor mutated input: seat 1 hand.cards was %d, now %d", origSeat1HandLen, got)
	}
}

func TestFilterViewForZeroedHandMarshalsAsEmptyArray(t *testing.T) {
	// The hidden-hand zone must marshal to `"cards": []`, never
	// `"cards": null`. Client code treats `null` as "field missing"
	// and can trip on it.
	v := buildViewWithCardsInHands(t)
	own := v.Seats[0].ID
	filtered := FilterViewFor(v, own)

	raw, err := json.Marshal(filtered.Seats[1].Hand)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !bytes.Contains(raw, []byte(`"cards":[]`)) {
		t.Errorf("hidden hand must serialise with empty cards array, got: %s", raw)
	}
}

func TestActionPayloadRoundTrip(t *testing.T) {
	params, _ := json.Marshal(map[string]any{"delta": -3})
	a := ActionPayload{
		Type:   "change_life",
		Player: uuid.New().String(),
		Params: params,
	}
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back ActionPayload
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Type != a.Type || back.Player != a.Player {
		t.Errorf("action payload lost fields: %+v", back)
	}
}

// TestLegalTargetsStampedForOwnerOnly — S20 sub-PR 1: a hand card
// with a TargetSpec carries legal_targets computed from its owner's
// point of view; the same card is stripped of them on the wire an
// opponent sees (even when revealed), and non-catalog cards never
// carry the field.
func TestLegalTargetsStampedForOwnerOnly(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-view-target-spec"
	prev := game.CatalogTargetSpec
	game.CatalogTargetSpec = func(id string) *game.TargetSpec {
		if id != oracle {
			return nil
		}
		return &game.TargetSpec{
			Mode:  "creature",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				return c.IsCreature()
			},
			Min: 1, Max: 1,
		}
	}
	t.Cleanup(func() { game.CatalogTargetSpec = prev })

	bear := game.NewCard("Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(bear)
	rock := game.NewCard("Rock", opp.ID)
	rock.TypeLine = "Artifact"
	g.Battlefield.PushTop(rock)
	spell := game.NewCard("Removal", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = oracle
	// Revealed to the opponent too, to prove the strip isn't just
	// the hand-hiding.
	spell.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	me.Hand.PushTop(spell)

	mine := ViewOfGameFor(g, me.ID.String())
	var found *CardView
	for i := range mine.Seats[0].Hand.Cards {
		if mine.Seats[0].Hand.Cards[i].InstanceID == spell.InstanceID.String() {
			found = &mine.Seats[0].Hand.Cards[i]
		}
	}
	if found == nil || found.LegalTargets == nil {
		t.Fatalf("owner's view: legal_targets missing on the targeted spell")
	}
	if len(found.LegalTargets.Cards) != 1 || found.LegalTargets.Cards[0] != bear.InstanceID.String() {
		t.Errorf("legal cards = %v, want just the bear", found.LegalTargets.Cards)
	}
	if len(found.LegalTargets.Players) != 0 {
		t.Errorf("creature spec must not list players")
	}
	for _, c := range mine.Seats[0].Hand.Cards {
		if c.InstanceID != spell.InstanceID.String() && c.LegalTargets != nil {
			t.Errorf("non-catalog hand card %q carries legal_targets", c.Name)
		}
	}

	theirs := ViewOfGameFor(g, opp.ID.String())
	for _, c := range theirs.Seats[0].Hand.Cards {
		if c.InstanceID == spell.InstanceID.String() && c.LegalTargets != nil {
			t.Errorf("opponent's view of a revealed hand card must not carry legal_targets")
		}
	}
}

// TestModesStampedForOwnerOnly — S20 sub-PR 4: a modal hand card
// carries its options with per-option legal sets for its owner, and
// nothing for an opponent.
func TestModesStampedForOwnerOnly(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-view-mode-spec"
	prev := game.CatalogModeSpec
	game.CatalogModeSpec = func(id string) *game.ModeSpec {
		if id != oracle {
			return nil
		}
		return &game.ModeSpec{
			Prompt: "Choose one",
			Options: []game.ModeOption{
				{Label: "Destroy target artifact.", Targets: &game.TargetSpec{
					Mode: "permanent", Zones: []game.ZoneKind{game.ZoneBattlefield},
					CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
						return c.IsArtifact()
					},
					Min: 1, Max: 1,
				}},
				{Label: "Do nothing."},
			},
			Min: 1, Max: 1,
		}
	}
	t.Cleanup(func() { game.CatalogModeSpec = prev })

	rock := game.NewCard("Rock", opp.ID)
	rock.TypeLine = "Artifact"
	g.Battlefield.PushTop(rock)
	charm := game.NewCard("Charm", me.ID)
	charm.TypeLine = "Instant"
	charm.OracleID = oracle
	charm.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	me.Hand.PushTop(charm)

	mine := ViewOfGameFor(g, me.ID.String())
	var found *CardView
	for i := range mine.Seats[0].Hand.Cards {
		if mine.Seats[0].Hand.Cards[i].InstanceID == charm.InstanceID.String() {
			found = &mine.Seats[0].Hand.Cards[i]
		}
	}
	if found == nil || found.Modes == nil {
		t.Fatalf("owner's view: modes missing on the modal card")
	}
	if found.Modes.Prompt != "Choose one" || found.Modes.Min != 1 || found.Modes.Max != 1 || len(found.Modes.Options) != 2 {
		t.Fatalf("modes = %+v", found.Modes)
	}
	o0, o1 := found.Modes.Options[0], found.Modes.Options[1]
	if o0.TargetMode != "permanent" || o0.LegalTargets == nil || len(o0.LegalTargets.Cards) != 1 || o0.LegalTargets.Cards[0] != rock.InstanceID.String() {
		t.Errorf("targeted option = %+v, want permanent mode with just the rock", o0)
	}
	if o1.TargetMode != "" || o1.LegalTargets != nil {
		t.Errorf("untargeted option must carry no target fields: %+v", o1)
	}
	if found.LegalTargets != nil {
		t.Errorf("modal card must not carry card-level legal_targets")
	}

	theirs := ViewOfGameFor(g, opp.ID.String())
	for _, c := range theirs.Seats[0].Hand.Cards {
		if c.InstanceID == charm.InstanceID.String() && c.Modes != nil {
			t.Errorf("opponent's view of a revealed modal card must not carry modes")
		}
	}
}

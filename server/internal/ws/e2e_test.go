package ws

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// updateGolden, when set, causes the e2e test to overwrite its golden
// file with the current output. Run with `go test -update ./...` to
// regenerate after an intentional schema change.
var updateGolden = flag.Bool("update", false, "update golden files instead of diffing")

// normalizeView walks a protocol.GameView in a fixed, field-by-field
// traversal order, collects every UUID it encounters, assigns
// deterministic <uuid-NNN> placeholders to each unique UUID, and
// returns a copy of the view with placeholders substituted.
//
// An earlier version of this function normalized the marshalled JSON
// text with a regex. That broke subtly for views containing
// UUID-keyed maps (PlayerView.CommanderDamage, CardView.Counters):
// encoding/json sorts map keys alphabetically, so the first-
// occurrence order of UUIDs in the emitted bytes depended on the
// alphabetical sort of the raw UUID strings — which are random per
// test run. Walking the struct directly gives us explicit control
// over traversal order, so placeholder assignment is stable.
//
// Traversal order: top-level ID, then each seat in seat order (player
// ID → private zones → CommanderDamage lookups), then shared zones
// (battlefield, stack, exile, phased_out), then the player-ID-only
// fields (monarch, initiative, promises, vote, discard_pending, the
// loop notice), then the card-ID-bearing lists (stack_items,
// pending_triggers, delayed_triggers, log). UUIDs that first appear as
// map KEYS (CommanderDamage keys are opponent player IDs, Promises'
// and Ballots' are player IDs) have ALREADY been assigned placeholders
// during the preceding seat walk, so the non-deterministic map
// iteration order in Go never allocates a new placeholder — it only
// looks up an existing one.
//
// Two GameView fields are deliberately NOT normalized here, and
// [normalizeViewNotYetSupported] is the one place that says so
// (#1264): `pending_choices` and `legal_moves`. Both carry UUIDs
// nested inside sub-shapes this function has no method for yet
// (PendingChoiceView's ReplacementOptions / DamageAssignment /
// PickTarget / PickOptions / DoubledBy; legal.Move's Player and
// Source are a real uuid.UUID, not the string every other ID field
// on the wire is, so they cannot even hold a "<uuid-NNN>" placeholder
// without a parallel scheme, and Params is per-move-kind opaque JSON
// that may embed further UUIDs no generic walk can find safely). A
// half-normalization that missed one of those nested UUIDs would look
// complete and still leak — worse than the honest gap this leaves.
func normalizeView(v protocol.GameView) protocol.GameView {
	n := &normalizer{ids: make(map[string]string)}
	return n.game(v)
}

// normalizeViewNotYetSupported names every exported GameView field
// normalizeView does not yet carry through placeholder-normalized,
// with the reason. TestNormalizeViewCarriesEveryTopLevelField reads
// this as its allowlist, so removing a field here without also
// implementing it fails that test rather than the golden silently
// covering for it.
func normalizeViewNotYetSupported() map[string]string {
	return map[string]string{
		"PendingChoices": "nested UUIDs in ReplacementOptions / DamageAssignment / PickTarget / PickOptions / DoubledBy have no normalizer method yet",
		"LegalMoves":     "legal.Move.Player/Source are uuid.UUID, not string — cannot hold a placeholder without a parallel scheme, and Params is per-move-kind opaque JSON",
	}
}

type normalizer struct {
	ids  map[string]string
	next int
}

// id returns the stable placeholder for s. Empty string passes
// through unchanged (omitempty on the wire).
func (n *normalizer) id(s string) string {
	if s == "" {
		return ""
	}
	if v, ok := n.ids[s]; ok {
		return v
	}
	n.next++
	v := fmt.Sprintf("<uuid-%03d>", n.next)
	n.ids[s] = v
	return v
}

func (n *normalizer) game(v protocol.GameView) protocol.GameView {
	out := protocol.GameView{
		ID:            n.id(v.ID),
		State:         v.State,
		Turn:          v.Turn,
		MulligansOpen: v.MulligansOpen,
		UndoLimit:     v.UndoLimit,
		Settings:      v.Settings,
	}
	out.Seats = make([]protocol.PlayerView, len(v.Seats))
	// First pass: assign placeholders to every player's ID so that
	// CommanderDamage lookups on the second pass resolve to existing
	// placeholders rather than allocating new (nondeterministic) ones.
	for i, p := range v.Seats {
		_ = n.id(p.ID)
		out.Seats[i].ID = n.ids[p.ID]
	}
	for i, p := range v.Seats {
		out.Seats[i] = n.player(p, out.Seats[i].ID)
	}
	out.Battlefield = n.zone(v.Battlefield)
	out.Stack = n.zone(v.Stack)
	out.Exile = n.zone(v.Exile)
	// #1199: the CR 702.26 holding zone. Normalised like the other
	// three, so the golden shows an empty `phased_out` rather than the
	// zero ZoneView a dropped field marshals to — which is what makes
	// the golden able to notice the day something phases out here.
	out.PhasedOut = n.zone(v.PhasedOut)

	// #1264: the player-ID-only fields. Safe here — every seat's
	// placeholder was already assigned above, so a monarch/initiative/
	// promise/ballot naming a seated player resolves to the existing
	// placeholder rather than allocating a new one.
	out.Monarch = n.id(v.Monarch)
	out.Initiative = n.id(v.Initiative)
	out.Promises = n.promises(v.Promises)
	out.Vote = n.vote(v.Vote)
	out.StartingSeat = v.StartingSeat
	out.SplitSecondActive = v.SplitSecondActive
	out.DiscardPending = n.idKeyedIntMap(v.DiscardPending)
	out.LoopNotice = n.loopNotice(v.LoopNotice)
	out.Outcome = n.outcome(v.Outcome)

	// #1264: the card-ID-bearing lists. A card named here that never
	// appeared in a zone above (a log entry for a card now in a
	// hidden zone with both endpoints redacted, say) still gets a
	// placeholder on first sight — see (*normalizer).id — so traversal
	// order only fixes the NUMBERING, never correctness.
	out.StackItems = n.stackItems(v.StackItems)
	out.PendingTriggers = n.stackItems(v.PendingTriggers)
	out.DelayedTriggers = n.delayedTriggers(v.DelayedTriggers)
	out.Log = n.logEvents(v.Log)
	// Reveals: RevealView carries no instance IDs at all by design
	// (reveal_frame.go) — Source is a card NAME, Cards are printed
	// identity with a scryfall_id (a printing, not this instance).
	// Nothing here for a normalizer to touch.
	out.Reveals = v.Reveals

	return out
}

// promises normalizes GameView.Promises' "{from}->{to}" composite
// keys (viewOfPromises in protocol/view.go). Both halves are player
// IDs, already assigned a placeholder during the seat walk in game().
func (n *normalizer) promises(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		from, to, ok := strings.Cut(k, "->")
		if !ok {
			out[k] = v
			continue
		}
		out[n.id(from)+"->"+n.id(to)] = v
	}
	return out
}

// idKeyedIntMap normalizes a map whose keys are player-ID strings and
// whose values need no normalization of their own — DiscardPending
// (count owed) and a VoteView's Ballots (the chosen option index).
func (n *normalizer) idKeyedIntMap(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[n.id(k)] = v
	}
	return out
}

func (n *normalizer) vote(v *protocol.VoteView) *protocol.VoteView {
	if v == nil {
		return nil
	}
	return &protocol.VoteView{
		ID:        v.ID,
		Topic:     v.Topic,
		Options:   v.Options,
		Initiator: n.id(v.Initiator),
		Ballots:   n.idKeyedIntMap(v.Ballots),
	}
}

// outcome normalizes ADR 0057's GameView.outcome: the winner is a
// seated player (already a placeholder) and the source a card.
func (n *normalizer) outcome(v *protocol.OutcomeView) *protocol.OutcomeView {
	if v == nil {
		return nil
	}
	out := *v
	out.Winner = n.id(v.Winner)
	out.Source = n.id(v.Source)
	return &out
}

func (n *normalizer) loopNotice(v *protocol.LoopNoticeView) *protocol.LoopNoticeView {
	if v == nil {
		return nil
	}
	return &protocol.LoopNoticeView{
		Source:     n.id(v.Source),
		Label:      v.Label,
		Controller: n.id(v.Controller),
		Count:      v.Count,
	}
}

// stackItems normalizes both StackItems and PendingTriggers — the
// same StackItemView shape (protocol/view.go), announce-time metadata
// for one item on the stack or in the pending-trigger queue.
func (n *normalizer) stackItems(in []protocol.StackItemView) []protocol.StackItemView {
	out := make([]protocol.StackItemView, len(in))
	for i, it := range in {
		out[i] = protocol.StackItemView{
			ID:               it.ID,
			Kind:             it.Kind,
			Controller:       n.id(it.Controller),
			Owner:            n.id(it.Owner),
			SourceCardID:     n.id(it.SourceCardID),
			Label:            it.Label,
			Targets:          n.targetRefs(it.Targets),
			Modes:            it.Modes,
			ModeLabels:       it.ModeLabels,
			XValue:           it.XValue,
			Distribution:     n.idKeyedIntMap(it.Distribution),
			HoldPriority:     it.HoldPriority,
			SplitSecond:      it.SplitSecond,
			AltCost:          it.AltCost,
			IsCopy:           it.IsCopy,
			DoubledBy:        n.id(it.DoubledBy),
			DoubledByName:    it.DoubledByName,
			ManaSpent:        it.ManaSpent,
			ColorsSpent:      it.ColorsSpent,
			ManaSpentUnknown: it.ManaSpentUnknown,
		}
	}
	return out
}

// targetRefs normalizes a StackItemView's Targets — the announced
// target list, where ID is a card or player instance ID (empty for a
// mode with no target slot).
func (n *normalizer) targetRefs(in []protocol.TargetRefView) []protocol.TargetRefView {
	out := make([]protocol.TargetRefView, len(in))
	for i, t := range in {
		out[i] = protocol.TargetRefView{Kind: t.Kind, ID: n.id(t.ID), Slot: t.Slot, Mode: t.Mode}
	}
	return out
}

func (n *normalizer) delayedTriggers(in []protocol.DelayedTriggerView) []protocol.DelayedTriggerView {
	out := make([]protocol.DelayedTriggerView, len(in))
	for i, d := range in {
		cards := make([]string, len(d.Cards))
		for j, c := range d.Cards {
			cards[j] = n.id(c)
		}
		out[i] = protocol.DelayedTriggerView{
			ID:         d.ID,
			Controller: n.id(d.Controller),
			Source:     n.id(d.Source),
			Label:      d.Label,
			At:         d.At,
			CreatedSeq: d.CreatedSeq,
			Cards:      cards,
			On:         d.On,
		}
	}
	return out
}

// logEvents normalizes the public game log. CardID and Target are the
// only UUID-bearing fields (log.go); Text is the already-rendered,
// already-redacted sentence and names cards by printed NAME, not ID,
// so it needs no rewriting.
func (n *normalizer) logEvents(in []protocol.LogEvent) []protocol.LogEvent {
	out := make([]protocol.LogEvent, len(in))
	for i, e := range in {
		out[i] = protocol.LogEvent{
			Seq:         e.Seq,
			Kind:        e.Kind,
			Turn:        e.Turn,
			Step:        e.Step,
			Seat:        e.Seat,
			TargetSeat:  e.TargetSeat,
			CardID:      n.id(e.CardID),
			Target:      n.id(e.Target),
			Amount:      e.Amount,
			LookedAt:    e.LookedAt,
			OldZone:     e.OldZone,
			NewZone:     e.NewZone,
			Combat:      e.Combat,
			CombatStep:  e.CombatStep,
			Sides:       e.Sides,
			Results:     e.Results,
			Faces:       e.Faces,
			Call:        e.Call,
			Wins:        e.Wins,
			Choice:      e.Choice,
			Label:       e.Label,
			ActorIsHost: e.ActorIsHost,
			Text:        e.Text,
		}
	}
	return out
}

func (n *normalizer) player(p protocol.PlayerView, id string) protocol.PlayerView {
	out := protocol.PlayerView{
		ID:             id,
		Name:           p.Name,
		Seat:           p.Seat,
		Life:           p.Life,
		Poison:         p.Poison,
		Energy:         p.Energy,
		Library:        n.zone(p.Library),
		Hand:           n.zone(p.Hand),
		Graveyard:      n.zone(p.Graveyard),
		Command:        n.zone(p.Command),
		Eliminated:     p.Eliminated,
		HandKept:       p.HandKept,
		MulligansTaken: p.MulligansTaken,
		MaxHandSize:    p.MaxHandSize,
		// #500: carried through so the golden pins the real per-seat
		// land-drop numbers rather than a normalizer zero.
		LandDropsPerTurn:    p.LandDropsPerTurn,
		LandsPlayedThisTurn: p.LandsPlayedThisTurn,
	}
	// Always initialise (possibly empty) — matches the wire shape
	// emitted by protocol.viewOfPlayer, which always allocates the
	// map. Leaving it nil here would make the golden file show
	// `null` while the real wire shows `{}`.
	out.CommanderDamage = make(map[string]int, len(p.CommanderDamage))
	for k, v := range p.CommanderDamage {
		out.CommanderDamage[n.id(k)] = v
	}
	// Same shape concern as CommanderDamage: viewOfPlayer always
	// allocates the slice. For determinism the timestamp is replaced
	// with a stable placeholder — wall-clock time would diff every
	// run.
	out.LifeHistory = make([]protocol.LifeChangeView, len(p.LifeHistory))
	for i, c := range p.LifeHistory {
		out.LifeHistory[i] = protocol.LifeChangeView{
			Delta:    c.Delta,
			NewTotal: c.NewTotal,
			At:       "<timestamp>",
			// Seq is deterministic (a per-player counter), unlike the
			// timestamp, so it goes into the golden file as-is.
			Seq: c.Seq,
		}
	}
	return out
}

func (n *normalizer) zone(z protocol.ZoneView) protocol.ZoneView {
	out := protocol.ZoneView{
		Kind:  z.Kind,
		Owner: n.id(z.Owner),
		Count: z.Count,
	}
	// Always allocate, even when empty, to match protocol.viewOfZone
	// which always does `make([]CardView, len(z.Cards))`. An empty
	// but non-nil slice marshals to `[]`; a nil slice marshals to
	// `null`. The wire always emits `[]`.
	out.Cards = make([]protocol.CardView, len(z.Cards))
	for i, c := range z.Cards {
		out.Cards[i] = n.card(c)
	}
	return out
}

func (n *normalizer) card(c protocol.CardView) protocol.CardView {
	return protocol.CardView{
		InstanceID:  n.id(c.InstanceID),
		Name:        c.Name,
		Owner:       n.id(c.Owner),
		Controller:  n.id(c.Controller),
		Tapped:      c.Tapped,
		Counters:    c.Counters,
		IsCommander: c.IsCommander,
	}
}

// everyFieldGameViewForNormalizer returns a GameView with every
// top-level exported field set to a non-zero value —
// TestNormalizeViewCarriesEveryTopLevelField's fixture. The shape
// mirrors protocol's own (unexported, different package)
// everyFieldGameView from #1250: a fixture a test also reflects over
// for completeness, so a GameView field added later and left at its
// zero value here is caught by THIS test rather than silently
// exempted from the one below it.
func everyFieldGameViewForNormalizer() protocol.GameView {
	ownerID, oppID := uuid.NewString(), uuid.NewString()
	cardID := uuid.NewString()
	zone := func(kind string) protocol.ZoneView {
		return protocol.ZoneView{
			Kind: kind, Owner: ownerID, Count: 1,
			Cards: []protocol.CardView{{InstanceID: uuid.NewString(), Name: "Card", Owner: ownerID, Controller: ownerID}},
		}
	}
	seat := func(id, name string, n int) protocol.PlayerView {
		return protocol.PlayerView{
			ID: id, Name: name, Seat: n, Life: 40,
			Library: zone("library"), Hand: zone("hand"),
			Graveyard: zone("graveyard"), Command: zone("command"),
		}
	}
	return protocol.GameView{
		ID:    "game-1",
		State: "active",
		Seats: []protocol.PlayerView{seat(ownerID, "Owner", 0), seat(oppID, "Opponent", 1)},

		Battlefield: zone("battlefield"),
		Stack:       zone("stack"),
		Exile:       zone("exile"),
		PhasedOut:   zone("phased_out"),

		Turn: protocol.TurnView{Number: 3, ActiveSeat: 0, PriorityHolder: 0, Phase: "main1", Step: "precombat_main"},

		MulligansOpen: true,
		Monarch:       ownerID,
		Initiative:    ownerID,
		Promises:      map[string]int{ownerID + "->" + oppID: 1},
		Vote: &protocol.VoteView{
			ID: "vote-1", Topic: "raise a toast", Options: []string{"yes"},
			Initiator: ownerID, Ballots: map[string]int{ownerID: 0},
		},
		UndoLimit:    3,
		Settings:     &protocol.TableSettingsView{UndoLimit: 3, UndoScope: "own", StartingLife: 40, BotPace: "normal"},
		StartingSeat: 1,

		StackItems:      []protocol.StackItemView{{ID: "item-1", Kind: "spell", Controller: ownerID, Owner: ownerID, SourceCardID: cardID}},
		PendingTriggers: []protocol.StackItemView{{ID: "item-2", Kind: "triggered", Controller: ownerID, Owner: ownerID, SourceCardID: cardID}},
		DelayedTriggers: []protocol.DelayedTriggerView{{ID: "delayed-1", Controller: ownerID, At: "end_step"}},

		SplitSecondActive: true,
		DiscardPending:    map[string]int{ownerID: 1},
		PendingChoices: []protocol.PendingChoiceView{{
			ID: "choice-1", Kind: "choose_cards", Chooser: ownerID, FromPlayer: ownerID, Count: 1,
			Options: []protocol.CardView{{InstanceID: uuid.NewString(), Name: "Choice Card", Owner: ownerID, Controller: ownerID}},
		}},
		LegalMoves: []protocol.LegalMoveView{{Type: "pass_priority", Player: uuid.MustParse(ownerID), Label: "Pass"}},

		Log: []protocol.LogEvent{{Seq: 1, Kind: protocol.LogCast, Turn: 3, Seat: 0, CardID: cardID, Text: "Owner cast Card"}},
		Reveals: []protocol.RevealView{{
			Seq: 1, Turn: 3, Seat: 0, Source: "Fact or Fiction", From: "library",
			Cards: []protocol.RevealedCardView{{Name: "Island"}}, Count: 1,
		}},

		LoopNotice: &protocol.LoopNoticeView{Source: ownerID, Label: "loop", Controller: ownerID, Count: 3},
		Outcome:    &protocol.OutcomeView{Kind: "win", Winner: ownerID, Cause: "effect", Source: cardID},
	}
}

// TestNormalizeViewCarriesEveryTopLevelField is the #1264 guard, the
// same shape #1250's TestFilterViewForCarriesEveryTopLevelField uses
// for FilterViewFor: build a GameView with every top-level exported
// field non-zero, run it through normalizeView, and fail on any field
// that came back zero and is not on normalizeViewNotYetSupported's
// allowlist.
//
// Before #1264, normalizeView's struct literal in (*normalizer).game
// simply never mentioned thirteen of GameView's fields — Monarch,
// Initiative, Promises, Vote, StartingSeat, StackItems,
// PendingTriggers, DelayedTriggers, SplitSecondActive,
// DiscardPending, Log, Reveals and LoopNotice — so every one of them
// silently zeroed out of every golden file, the same shape #1250
// found in FilterViewFor's own struct literal one layer over. This
// test is what makes the NEXT field added to GameView fail here
// instead of doing the same thing quietly.
func TestNormalizeViewCarriesEveryTopLevelField(t *testing.T) {
	fixture := everyFieldGameViewForNormalizer()

	rv := reflect.ValueOf(fixture)
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if !f.IsExported() {
			continue
		}
		if rv.Field(i).IsZero() {
			t.Fatalf("everyFieldGameViewForNormalizer leaves GameView.%s zero; set it so this test actually exercises normalizeView's handling of it", f.Name)
		}
	}

	out := normalizeView(fixture)
	allow := normalizeViewNotYetSupported()

	outv := reflect.ValueOf(out)
	outt := outv.Type()
	for i := 0; i < outt.NumField(); i++ {
		f := outt.Field(i)
		if !f.IsExported() {
			continue // legalBySeat: unexported, never on the wire.
		}
		zero := outv.Field(i).IsZero()
		if reason, ok := allow[f.Name]; ok {
			if !zero {
				t.Errorf("GameView.%s is on normalizeViewNotYetSupported's allowlist (%s) but normalizeView returned a non-zero value — the field is handled now, so drop it from the allowlist", f.Name, reason)
			}
			continue
		}
		if zero {
			t.Errorf("GameView.%s came back zero from normalizeView — it dropped the field (add handling in (*normalizer).game, or list it in normalizeViewNotYetSupported with a reason)", f.Name)
		}
	}
}

// newE2EServer stands up a fully-wired httptest server with a room
// whose game has been seeded from a deterministic RNG. Returns the
// ws URL (without query params), the game (for player ID lookups),
// and a cleanup func. Callers append `?game=<id>&player=<id>` as
// needed to bind a specific seat.
func newE2EServer(t *testing.T) (string, *game.Game, func()) {
	t.Helper()
	wsURL, g, _, cleanup := newE2EServerWithRoom(t)
	return wsURL, g, cleanup
}

// newE2EServerWithRoom is newE2EServer plus the Room, for the tests
// that need to reach past the wire — today the table-settings gates
// (ADR 0075 §2.3), which ask the room who hosts and therefore need a
// way to say so. The lobby designates a host in production; there is
// no lobby here.
func newE2EServerWithRoom(t *testing.T) (string, *game.Game, *Room, func()) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	g := game.NewGame()
	// Fixed-shape synthetic decks for reproducibility.
	for i := range 2 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Test Commander %d", i+1), uuid.Nil)}
		for j := range 20 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Test Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Player %d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(42, 42))); err != nil {
		t.Fatalf("Start: %v", err)
	}

	room := NewRoom(g, log, "")
	hub := NewHub(log)
	hub.SetRoom(room)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	return wsURL, g, room, srv.Close
}

// dialAs opens a WS connection bound to the given player ID. Passing
// uuid.Nil yields a spectator connection (no player query param),
// which has every opponent hand filtered away in the received
// snapshots.
func dialAs(t *testing.T, wsURL string, playerID uuid.UUID) *websocket.Conn {
	t.Helper()
	u := wsURL
	if playerID != uuid.Nil {
		u += "?player=" + playerID.String()
	}
	return dial(t, u)
}

// sendActionAndWait sends a scripted action and returns the resulting
// snapshot payload. Fails the test on any protocol error.
func sendActionAndWait(t *testing.T, conn *websocket.Conn, a protocol.ActionPayload) protocol.SnapshotPayload {
	t.Helper()
	payload, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal action: %v", err)
	}
	frame, err := json.Marshal(protocol.Frame{
		V:       protocol.Version,
		Kind:    protocol.KindAction,
		ID:      uuid.New().String(),
		Payload: payload,
	})
	if err != nil {
		t.Fatalf("marshal frame: %v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, frame); err != nil {
		t.Fatalf("write: %v", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	var f protocol.Frame
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("unmarshal frame: %v", err)
	}
	if f.Kind != protocol.KindSnapshot {
		// Probably an error frame — surface it.
		var ep protocol.ErrorPayload
		_ = json.Unmarshal(f.Payload, &ep)
		t.Fatalf("expected snapshot after %q, got %q (code=%s message=%s)", a.Type, f.Kind, ep.Code, ep.Message)
	}
	var snap protocol.SnapshotPayload
	if err := json.Unmarshal(f.Payload, &snap); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	return snap
}

// readChatFrame reads the next frame from conn and asserts it is a
// chat frame, returning the decoded payload. Used by the S07 chat
// broadcast test.
func readChatFrame(t *testing.T, conn *websocket.Conn) protocol.ChatPayload {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read chat frame: %v", err)
	}
	var f protocol.Frame
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("unmarshal frame: %v", err)
	}
	if f.Kind != protocol.KindChat {
		t.Fatalf("kind: got %q, want %q", f.Kind, protocol.KindChat)
	}
	var p protocol.ChatPayload
	if err := json.Unmarshal(f.Payload, &p); err != nil {
		t.Fatalf("unmarshal chat payload: %v", err)
	}
	return p
}

// TestChatBroadcast covers the S07 chat path: a client-sent text
// frame is server-stamped (author ID, author name, RFC3339 timestamp)
// and broadcast to every client bound to the same game, including the
// originator. Snapshot state is not touched.
func TestChatBroadcast(t *testing.T) {
	wsURL, g, cleanup := newE2EServer(t)
	defer cleanup()

	seat0 := g.Seats[0]
	seat1 := g.Seats[1]

	connA := dialAs(t, wsURL, seat0.ID)
	defer connA.Close()
	connB := dialAs(t, wsURL, seat1.ID)
	defer connB.Close()

	// Both clients consume their initial snapshots so subsequent reads
	// see only the chat broadcast.
	_ = readSnapshotFrame(t, connA)
	_ = readSnapshotFrame(t, connB)

	// Client A sends a chat frame. The author_id/name/timestamp values
	// it puts on the wire are intentionally bogus — the server must
	// overwrite all three.
	out := protocol.ChatPayload{
		AuthorID:   "client-supplied-bogus",
		AuthorName: "client-supplied-bogus",
		Text:       "hello table",
		Timestamp:  "client-supplied-bogus",
	}
	payload, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal chat payload: %v", err)
	}
	frame, err := json.Marshal(protocol.Frame{
		V:       protocol.Version,
		Kind:    protocol.KindChat,
		ID:      uuid.New().String(),
		Payload: payload,
	})
	if err != nil {
		t.Fatalf("marshal chat frame: %v", err)
	}
	if err := connA.WriteMessage(websocket.TextMessage, frame); err != nil {
		t.Fatalf("write: %v", err)
	}

	for label, conn := range map[string]*websocket.Conn{"A": connA, "B": connB} {
		got := readChatFrame(t, conn)
		if got.Text != "hello table" {
			t.Errorf("conn %s: text=%q, want %q", label, got.Text, "hello table")
		}
		if got.AuthorID != seat0.ID.String() {
			t.Errorf("conn %s: author_id=%q, want %q (server should re-stamp from connection's player)",
				label, got.AuthorID, seat0.ID)
		}
		if got.AuthorName != seat0.Name {
			t.Errorf("conn %s: author_name=%q, want %q", label, got.AuthorName, seat0.Name)
		}
		if got.Timestamp == "" || got.Timestamp == "client-supplied-bogus" {
			t.Errorf("conn %s: server did not stamp timestamp (got %q)", label, got.Timestamp)
		}
	}
}

// TestE2EScriptedTurn is the S03 exit-criteria test: drive a scripted
// turn (draw, play, tap, pass_turn) through the real hub via an
// actual WebSocket dial, then diff the final normalized snapshot
// against a golden file.
func TestE2EScriptedTurn(t *testing.T) {
	wsURL, g, cleanup := newE2EServer(t)
	defer cleanup()

	// Drive the scripted turn from seat 0's perspective. Binding the
	// viewer matters now that S04 filters opponent hand + library
	// cards out of every broadcast — a spectator connection would
	// never see the drawn card and the test's `Hand.Cards[0]` lookup
	// would panic on an empty slice.
	seat0ID := g.Seats[0].ID
	conn := dialAs(t, wsURL, seat0ID)
	defer conn.Close()

	// Consume initial snapshot. As of S08, Start deals an opening
	// hand of 7 to each seat — see game.OpeningHandSize.
	initial := readSnapshotFrame(t, conn)
	if initial.Game.Seats[0].Hand.Count != 7 {
		t.Fatalf("initial hand count: got %d, want 7 (opening hand)", initial.Game.Seats[0].Hand.Count)
	}
	if !initial.Game.MulligansOpen {
		t.Fatalf("initial mulligans_open: got false, want true (window opens at Start)")
	}

	seat0 := seat0ID.String()

	// 1) Draw a card. Hand: 7 → 8, library: 13 (20 - 7 dealt) → 12.
	afterDraw := sendActionAndWait(t, conn, protocol.ActionPayload{
		Type:   "draw_card",
		Player: seat0,
	})
	if afterDraw.Game.Seats[0].Hand.Count != 8 {
		t.Errorf("after draw: hand=%d, want 8", afterDraw.Game.Seats[0].Hand.Count)
	}
	if afterDraw.Game.Seats[0].Library.Count != 12 {
		t.Errorf("after draw: library=%d, want 12", afterDraw.Game.Seats[0].Library.Count)
	}

	// 2) Play the first card in hand onto the battlefield. (No longer
	// necessarily the just-drawn card now that the hand starts non-
	// empty — but the play mechanics are identical.)
	drawnCardID := afterDraw.Game.Seats[0].Hand.Cards[0].InstanceID
	params, _ := json.Marshal(map[string]string{"instance_id": drawnCardID})
	afterPlay := sendActionAndWait(t, conn, protocol.ActionPayload{
		Type:   "play_card",
		Player: seat0,
		Params: params,
	})
	if afterPlay.Game.Seats[0].Hand.Count != 7 {
		t.Errorf("after play: hand=%d, want 7", afterPlay.Game.Seats[0].Hand.Count)
	}
	if afterPlay.Game.Battlefield.Count != 1 {
		t.Errorf("after play: battlefield=%d, want 1", afterPlay.Game.Battlefield.Count)
	}

	// 3) Tap the played card ("attack").
	tapParams, _ := json.Marshal(map[string]string{"instance_id": drawnCardID})
	afterTap := sendActionAndWait(t, conn, protocol.ActionPayload{
		Type:   "tap",
		Params: tapParams,
	})
	if !afterTap.Game.Battlefield.Cards[0].Tapped {
		t.Error("after tap: card should be tapped")
	}

	// 4) Pass turn to seat 1.
	afterPass := sendActionAndWait(t, conn, protocol.ActionPayload{
		Type: "pass_turn",
	})
	if afterPass.Game.Turn.ActiveSeat != 1 {
		t.Errorf("after pass_turn: seat=%d, want 1", afterPass.Game.Turn.ActiveSeat)
	}
	if afterPass.Game.Turn.Step != "untap" {
		t.Errorf("after pass_turn: step=%q, want untap", afterPass.Game.Turn.Step)
	}

	// Final state must match the golden file after UUID normalization.
	// normalizeView walks the struct in a deterministic order and
	// replaces every UUID with a stable placeholder; encoding/json
	// then produces identical bytes across runs for the same inputs.
	normalizedPayload := protocol.SnapshotPayload{
		Seq:  afterPass.Seq,
		Game: normalizeView(afterPass.Game),
	}
	normalized, err := json.MarshalIndent(normalizedPayload, "", "  ")
	if err != nil {
		t.Fatalf("marshal normalized snapshot: %v", err)
	}

	goldenPath := filepath.Join("testdata", "e2e_scripted_turn.golden.json")
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(goldenPath, normalized, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("wrote golden file: %s", goldenPath)
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden (run with -update to create): %v", err)
	}
	if !bytes.Equal(normalized, want) {
		t.Errorf("snapshot diverged from golden\n\n--- want (%s) ---\n%s\n--- got ---\n%s",
			goldenPath, want, normalized)
	}
}

// readNextFrame reads either a snapshot or error frame and returns
// the raw protocol.Frame so the caller can assert on Kind. Used by
// the S11 undo authorization tests where the server's response can
// be either an error (rejected) or a snapshot (accepted).
func readNextFrame(t *testing.T, conn *websocket.Conn) protocol.Frame {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read frame: %v", err)
	}
	var f protocol.Frame
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("unmarshal frame: %v", err)
	}
	return f
}

// Game state read helpers — the hub's write goroutine mutates the
// shared *game.Game under its own write lock, so tests that assert on
// fields from the test goroutine must take the read lock too. Without
// these helpers `go test -race` flags a data race on every field
// read. Wrapping each assertion in a ReadSnapshot closure would work
// but is noisy; a pair of tiny accessors keeps the tests readable.
func handSize(g *game.Game, seat int) int {
	var n int
	g.ReadSnapshot(func() { n = g.Seats[seat].Hand.Size() })
	return n
}

func undosRemaining(g *game.Game, seat int) int {
	var n int
	g.ReadSnapshot(func() { n = g.Seats[seat].UndosRemaining })
	return n
}

func undoLimit(g *game.Game) int {
	var n int
	g.ReadSnapshot(func() { n = g.Settings.UndoLimit })
	return n
}

// TestUndoRejectsCrossPlayerCaller covers the S11 caller gate: a
// seated player may only undo their own most recent action. Player A
// draws; Player B tries to undo A's draw; B's request must error
// out with bad_request (and not affect game state).
func TestUndoRejectsCrossPlayerCaller(t *testing.T) {
	wsURL, g, cleanup := newE2EServer(t)
	defer cleanup()

	playerA := g.Seats[0].ID
	playerB := g.Seats[1].ID

	connA := dialAs(t, wsURL, playerA)
	defer connA.Close()
	readNextFrame(t, connA) // initial snapshot to A
	connB := dialAs(t, wsURL, playerB)
	defer connB.Close()
	readNextFrame(t, connB) // initial snapshot to B

	// A draws. Both A and B see the broadcast snapshot.
	sendActionAndWait(t, connA, protocol.ActionPayload{
		Type: "draw_card", Player: playerA.String(),
	})
	readNextFrame(t, connB) // B's broadcast copy
	if n := handSize(g, 0); n != 8 {
		t.Fatalf("after A draw: A.hand=%d, want 8", n)
	}

	// B tries to undo A's action. Should fail with bad_request and
	// leave game state untouched.
	sendActionFrame(t, connB, protocol.ActionPayload{Type: "undo"})
	frame := readNextFrame(t, connB)
	if frame.Kind != protocol.KindError {
		t.Fatalf("expected error frame for cross-player undo, got %q", frame.Kind)
	}
	var ep protocol.ErrorPayload
	_ = json.Unmarshal(frame.Payload, &ep)
	if ep.Code != protocol.CodeBadRequest {
		t.Errorf("error code: got %q, want %q", ep.Code, protocol.CodeBadRequest)
	}
	if n := handSize(g, 0); n != 8 {
		t.Errorf("rejected undo mutated state: A.hand=%d, want 8", n)
	}

	// A undoes their own action. Should succeed; broadcast snapshot
	// goes to both clients.
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "undo"})
	readNextFrame(t, connB) // B's broadcast copy
	if n := handSize(g, 0); n != 7 {
		t.Errorf("after A undo: A.hand=%d, want 7", n)
	}
}

// TestUndoBudgetExhausted covers the per-player per-turn budget. A
// draws twice, then undoes once (budget 1→0), then tries to undo
// again — should fail with bad_request even though the stack has
// another A-owned entry to pop.
func TestUndoBudgetExhausted(t *testing.T) {
	wsURL, g, cleanup := newE2EServer(t)
	defer cleanup()

	playerA := g.Seats[0].ID
	connA := dialAs(t, wsURL, playerA)
	defer connA.Close()
	readNextFrame(t, connA) // initial

	if n := undosRemaining(g, 0); n != game.DefaultUndoLimit {
		t.Fatalf("budget at start: got %d, want %d", n, game.DefaultUndoLimit)
	}

	// Two draws.
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "draw_card", Player: playerA.String()})
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "draw_card", Player: playerA.String()})
	if n := handSize(g, 0); n != 9 {
		t.Fatalf("after two draws: A.hand=%d, want 9", n)
	}

	// First undo: succeeds, budget 1→0.
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "undo"})
	if n := undosRemaining(g, 0); n != 0 {
		t.Fatalf("budget after first undo: got %d, want 0", n)
	}
	if n := handSize(g, 0); n != 8 {
		t.Fatalf("after first undo: A.hand=%d, want 8", n)
	}

	// Second undo: should fail with bad_request ("no undos remaining").
	sendActionFrame(t, connA, protocol.ActionPayload{Type: "undo"})
	frame := readNextFrame(t, connA)
	if frame.Kind != protocol.KindError {
		t.Fatalf("expected error frame for exhausted budget, got %q", frame.Kind)
	}
	var ep protocol.ErrorPayload
	_ = json.Unmarshal(frame.Payload, &ep)
	if ep.Code != protocol.CodeBadRequest {
		t.Errorf("error code: got %q, want %q", ep.Code, protocol.CodeBadRequest)
	}
	if n := handSize(g, 0); n != 8 {
		t.Errorf("rejected undo mutated state: A.hand=%d, want 8", n)
	}
}

// TestUndoBudgetRefreshesOnTurnRollover confirms the per-player undo
// budget refreshes when the cursor enters that player's untap step
// (via PassTurn here; the same hook fires on AdvanceStep and on
// PassPriority's wrap branch).
func TestUndoBudgetRefreshesOnTurnRollover(t *testing.T) {
	wsURL, g, cleanup := newE2EServer(t)
	defer cleanup()

	playerA := g.Seats[0].ID
	connA := dialAs(t, wsURL, playerA)
	defer connA.Close()
	readNextFrame(t, connA) // initial

	// Spend A's budget.
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "draw_card", Player: playerA.String()})
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "undo"})
	if n := undosRemaining(g, 0); n != 0 {
		t.Fatalf("budget post-undo: got %d, want 0", n)
	}

	// Pass turn (B becomes active) then back to A. A's budget refreshes
	// on entering A's untap step.
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "pass_turn"})
	if n := undosRemaining(g, 1); n != game.DefaultUndoLimit {
		t.Errorf("B budget after pass_turn: got %d, want %d", n, game.DefaultUndoLimit)
	}
}

// TestUnlimitedUndoBudgetNeverRefuses covers ADR 0075's UndoUnlimited
// through the room: the pre-restore budget check asks the game rather
// than reading UndosRemaining, so an unlimited table undoes as often
// as it has entries, and a settings change made between an action and
// its undo is not rolled back by that undo.
func TestUnlimitedUndoBudgetNeverRefuses(t *testing.T) {
	wsURL, g, cleanup := newE2EServer(t)
	defer cleanup()

	playerA := g.Seats[0].ID
	connA := dialAs(t, wsURL, playerA)
	defer connA.Close()
	readNextFrame(t, connA) // initial

	for i := 0; i < 3; i++ {
		sendActionAndWait(t, connA, protocol.ActionPayload{Type: "draw_card", Player: playerA.String()})
	}
	limit := game.UndoUnlimited
	if err := g.UpdateSettings(uuid.Nil, game.SettingsPatch{UndoLimit: &limit}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	for i := 0; i < 3; i++ {
		sendActionAndWait(t, connA, protocol.ActionPayload{Type: "undo"})
	}
	if n := handSize(g, 0); n != 7 {
		t.Errorf("after three undos: A.hand=%d, want 7", n)
	}
	if n := undoLimit(g); n != game.UndoUnlimited {
		t.Errorf("undo rolled the settings back: UndoLimit=%d, want %d", n, game.UndoUnlimited)
	}
}

// TestSetUndoLimitRefreshesAllSeats covers the in-game host path: the
// HOST dials the undo budget up via the deprecated set_undo_limit
// alias and every player's UndosRemaining updates to the new limit
// immediately.
//
// "Any seated player" until S35 — ADR 0075 §2.3 narrowed it, and
// TestSetUndoLimitRefusedForNonHost is the other half of the change.
func TestSetUndoLimitRefreshesAllSeats(t *testing.T) {
	wsURL, g, room, cleanup := newE2EServerWithRoom(t)
	defer cleanup()

	playerA := g.Seats[0].ID
	room.SetHost(playerA)
	connA := dialAs(t, wsURL, playerA)
	defer connA.Close()
	readNextFrame(t, connA) // initial

	// Drain A's budget so we can verify the refresh.
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "draw_card", Player: playerA.String()})
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "undo"})
	if n := undosRemaining(g, 0); n != 0 {
		t.Fatalf("setup: A budget got %d, want 0", n)
	}

	// Bump the limit to 3 — every seat should snap to UndosRemaining=3.
	sendActionAndWait(t, connA, protocol.ActionPayload{
		Type:   "set_undo_limit",
		Params: json.RawMessage(`{"limit":3}`),
	})
	if n := undoLimit(g); n != 3 {
		t.Errorf("game UndoLimit: got %d, want 3", n)
	}
	g.ReadSnapshot(func() {
		for i, p := range g.Seats {
			if p.UndosRemaining != 3 {
				t.Errorf("seat %d UndosRemaining: got %d, want 3", i, p.UndosRemaining)
			}
		}
	})
}

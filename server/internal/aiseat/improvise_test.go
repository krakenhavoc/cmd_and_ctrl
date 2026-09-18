package aiseat_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// --- fixtures ------------------------------------------------------

// improviser offers one improvisation and then declines every
// ordinary move, so the bundle stays the top of the undo stack for
// the whole test. That is not just convenience: Room.Undo pops the
// TOP entry only, so "a human undoes the bot's improvisation" is only
// a meaningful assertion when nothing has landed on top of it. The
// shallowness of that safety valve is called out in ADR 0033 §8.
type improviser struct {
	im aiseat.Improvisation

	mu    sync.Mutex
	asked int
	given bool
}

func (p *improviser) Name() string { return "improviser" }

func (p *improviser) Decide(context.Context, aiseat.Input) (aiseat.Decision, error) {
	return aiseat.Decision{Index: aiseat.Decline, Reason: "holding"}, nil
}

func (p *improviser) Improvise(aiseat.Input) (aiseat.Improvisation, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.asked++
	if p.given {
		return aiseat.Improvisation{}, false
	}
	p.given = true
	return p.im, true
}

// chatBroadcaster records both halves of what a hub would do, so a
// test can assert on the announcement as well as the state fan-out.
// Implements aiseat.Announcer, which *ws.Hub also satisfies.
type chatBroadcaster struct {
	recordingBroadcaster
	chatMu sync.Mutex
	chats  []protocol.ChatPayload
}

func (b *chatBroadcaster) BroadcastChat(_ uuid.UUID, msg protocol.ChatPayload) {
	b.chatMu.Lock()
	defer b.chatMu.Unlock()
	b.chats = append(b.chats, msg)
}

func (b *chatBroadcaster) count() int {
	b.chatMu.Lock()
	defer b.chatMu.Unlock()
	return len(b.chats)
}

func (b *chatBroadcaster) lines(kind string) []protocol.ChatPayload {
	b.chatMu.Lock()
	defer b.chatMu.Unlock()
	var out []protocol.ChatPayload
	for _, c := range b.chats {
		if c.Kind == kind {
			out = append(out, c)
		}
	}
	return out
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}
	return b
}

// newRoomWithDump is newRoom plus an on-disk replay log, which is
// where the improvisation tag lands.
func newRoomWithDump(t *testing.T, seats int, seed uint64) (*ws.Room, string) {
	t.Helper()
	dir := t.TempDir()
	room := newRoom(t, seats, seed)
	return ws.NewRoom(room.Game, testLogger(), dir), dir
}

// handCardOf returns one instance ID from the seat's hand.
func handCardOf(t *testing.T, g *game.Game, seat uuid.UUID) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.ReadSnapshot(func() {
		p := g.PlayerByIDForEffect(seat)
		if p == nil || len(p.Hand.Cards) == 0 {
			return
		}
		id = p.Hand.Cards[0].InstanceID
	})
	if id == uuid.Nil {
		t.Fatalf("seat %s has an empty hand", seat)
	}
	return id
}

func lifeOf(g *game.Game, seat uuid.UUID) int {
	var life int
	g.ReadSnapshot(func() {
		if p := g.PlayerByIDForEffect(seat); p != nil {
			life = p.Life
		}
	})
	return life
}

func zoneSizes(g *game.Game, seat uuid.UUID) (hand, graveyard int) {
	g.ReadSnapshot(func() {
		if p := g.PlayerByIDForEffect(seat); p != nil {
			hand, graveyard = len(p.Hand.Cards), len(p.Graveyard.Cards)
		}
	})
	return
}

// grimTutor is the shape of a real improvisation: a card the catalog
// cannot run, executed with three sandbox verbs — one of which
// reaches ANOTHER seat's life total, which is exactly what the
// seated-caller gates would refuse and what the uuid.Nil stamp
// permits.
func grimTutor(t *testing.T, bot, opponent, card uuid.UUID) aiseat.Improvisation {
	t.Helper()
	return aiseat.Improvisation{
		Card:   "Grim Tutor",
		Effect: "discard a card, lose 3 life, and each opponent loses 2 life",
		Reason: "no catalog spec for this card; the line needs it now",
		Steps: []aiseat.ImprovStep{
			{
				Type:   aiseat.VerbMoveCard,
				Params: mustJSON(t, map[string]any{"instance_id": card.String(), "src": map[string]string{"kind": "hand", "owner": bot.String()}, "dst": map[string]string{"kind": "graveyard", "owner": bot.String()}}),
			},
			{Type: aiseat.VerbChangeLife, Player: bot, Params: mustJSON(t, map[string]any{"delta": -3})},
			{Type: aiseat.VerbChangeLife, Player: opponent, Params: mustJSON(t, map[string]any{"delta": -2})},
		},
	}
}

// --- validation ----------------------------------------------------

func TestImprovisationValidate(t *testing.T) {
	ok := []aiseat.ImprovStep{{Type: aiseat.VerbChangeLife, Player: uuid.New(), Params: json.RawMessage(`{"delta":-3}`)}}

	cases := map[string]struct {
		im   aiseat.Improvisation
		want error
	}{
		"valid":            {aiseat.Improvisation{Card: "Grim Tutor", Effect: "lose 3 life", Steps: ok}, nil},
		"no card":          {aiseat.Improvisation{Effect: "lose 3 life", Steps: ok}, aiseat.ErrImprovNoCard},
		"blank card":       {aiseat.Improvisation{Card: "   ", Effect: "lose 3 life", Steps: ok}, aiseat.ErrImprovNoCard},
		"no effect":        {aiseat.Improvisation{Card: "Grim Tutor", Steps: ok}, aiseat.ErrImprovNoEffect},
		"blank effect":     {aiseat.Improvisation{Card: "Grim Tutor", Effect: "\t\n", Steps: ok}, aiseat.ErrImprovNoEffect},
		"no steps":         {aiseat.Improvisation{Card: "Grim Tutor", Effect: "lose 3 life"}, aiseat.ErrImprovNoSteps},
		"unlisted verb":    {aiseat.Improvisation{Card: "Grim Tutor", Effect: "draw", Steps: []aiseat.ImprovStep{{Type: "draw_card", Player: uuid.New(), Params: json.RawMessage(`{}`)}}}, aiseat.ErrImprovVerb},
		"concede smuggled": {aiseat.Improvisation{Card: "Grim Tutor", Effect: "scoop", Steps: []aiseat.ImprovStep{{Type: "concede", Player: uuid.New(), Params: json.RawMessage(`{}`)}}}, aiseat.ErrImprovVerb},
		"empty params":     {aiseat.Improvisation{Card: "Grim Tutor", Effect: "lose 3 life", Steps: []aiseat.ImprovStep{{Type: aiseat.VerbChangeLife, Player: uuid.New()}}}, aiseat.ErrImprovParams},
		"broken params":    {aiseat.Improvisation{Card: "Grim Tutor", Effect: "lose 3 life", Steps: []aiseat.ImprovStep{{Type: aiseat.VerbChangeLife, Player: uuid.New(), Params: json.RawMessage(`{`)}}}, aiseat.ErrImprovParams},
		"life needs player": {aiseat.Improvisation{Card: "Grim Tutor", Effect: "lose 3 life", Steps: []aiseat.ImprovStep{
			{Type: aiseat.VerbChangeLife, Params: json.RawMessage(`{"delta":-3}`)},
		}}, aiseat.ErrImprovPlayer},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.im.Validate()
			if tc.want == nil {
				if err != nil {
					t.Fatalf("Validate: %v", err)
				}
				return
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("Validate = %v, want %v", err, tc.want)
			}
		})
	}

	t.Run("too many steps", func(t *testing.T) {
		many := make([]aiseat.ImprovStep, aiseat.MaxImprovSteps+1)
		for i := range many {
			many[i] = aiseat.ImprovStep{Type: aiseat.VerbMarkDamage, Params: json.RawMessage(`{"instance_id":"x","delta":1}`)}
		}
		err := aiseat.Improvisation{Card: "c", Effect: "e", Steps: many}.Validate()
		if !errors.Is(err, aiseat.ErrImprovTooLong) {
			t.Fatalf("Validate = %v, want ErrImprovTooLong", err)
		}
	})
}

// TestImprovVerbsMatchActions pins the four verb strings to the
// dispatcher's own constants — the aiseat side is a string so a
// policy never needs the engine's packages, which means nothing but
// this test stops the two drifting apart.
func TestImprovVerbsMatchActions(t *testing.T) {
	for verb, want := range map[string]actions.Type{
		aiseat.VerbMoveCard:   actions.TypeMoveCard,
		aiseat.VerbChangeLife: actions.TypeChangeLife,
		aiseat.VerbAddCounter: actions.TypeAddCounter,
		aiseat.VerbMarkDamage: actions.TypeMarkDamage,
	} {
		if verb != string(want) {
			t.Errorf("verb %q != actions.%v", verb, want)
		}
	}
}

// TestAnnouncementNamesEverythingItMustName is the disclosure
// contract in one assertion: the card, the effect, and that it was
// improvised rather than executed by the rules.
func TestAnnouncementNamesEverythingItMustName(t *testing.T) {
	line := aiseat.Improvisation{Card: "Grim Tutor", Effect: "search my library and lose 3 life"}.Announcement()
	for _, want := range []string{"Grim Tutor", "search my library and lose 3 life", "improvised", "undo"} {
		if !strings.Contains(line, want) {
			t.Errorf("announcement %q is missing %q", line, want)
		}
	}

	long := aiseat.Improvisation{Card: "Verbose", Effect: strings.Repeat("x", 4000)}.Announcement()
	if len(long) > protocol.MaxChatTextLen {
		t.Errorf("announcement is %d bytes, over the %d cap", len(long), protocol.MaxChatTextLen)
	}
	if !strings.Contains(long, "improvised") {
		t.Errorf("truncation dropped the disclosure: %q", long)
	}
}

// --- end to end ----------------------------------------------------

// TestRunnerAppliesAnnouncesAndTagsAnImprovisation covers three of
// sub-PR 8's four checklist items at once: the bundle applies whole,
// the table is told in chat, and the replay line carries a greppable
// tag.
func TestRunnerAppliesAnnouncesAndTagsAnImprovisation(t *testing.T) {
	room, dumpDir := newRoomWithDump(t, 2, 11)
	g := room.Game
	bot, human := g.Seats[0].ID, g.Seats[1].ID
	card := handCardOf(t, g, bot)
	handBefore, graveBefore := zoneSizes(g, bot)
	botLife, humanLife := lifeOf(g, bot), lifeOf(g, human)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bc := &chatBroadcaster{}
	pol := &improviser{im: grimTutor(t, bot, human, card)}
	r := aiseat.Start(ctx, room, bot, pol, aiseat.Config{}, bc, testLogger())

	waitFor(t, "the improvisation to apply", func() bool {
		return r.Stats().Improvisations == 1
	})
	cancel()
	<-r.Done()

	// Every step landed.
	if got := lifeOf(g, bot); got != botLife-3 {
		t.Errorf("bot life = %d, want %d", got, botLife-3)
	}
	if got := lifeOf(g, human); got != humanLife-2 {
		t.Errorf("opponent life = %d, want %d — the uuid.Nil caller is what lets a bot reach another seat", got, humanLife-2)
	}
	hand, grave := zoneSizes(g, bot)
	if hand != handBefore-1 || grave != graveBefore+1 {
		t.Errorf("hand/graveyard = %d/%d, want %d/%d", hand, grave, handBefore-1, graveBefore+1)
	}

	// The table was told, in one line, naming the card and the effect.
	lines := bc.lines(protocol.ChatKindBotImprovisation)
	if len(lines) != 1 {
		t.Fatalf("got %d improvisation chat lines, want exactly 1", len(lines))
	}
	line := lines[0]
	if !strings.Contains(line.Text, "Grim Tutor") || !strings.Contains(line.Text, "improvised") {
		t.Errorf("chat line does not disclose: %q", line.Text)
	}
	if line.AuthorID != bot.String() {
		t.Errorf("chat author = %q, want the bot seat %q", line.AuthorID, bot)
	}
	if line.Reason == "" {
		t.Errorf("chat line should carry the policy's reason for the 'show bot reasoning' setting")
	}

	// The replay log is greppable.
	ann := replayAnnotations(t, dumpDir, g.ID)
	if len(ann) != 1 {
		t.Fatalf("got %d tagged replay lines, want 1", len(ann))
	}
	got := ann[0]
	if got.Tag != protocol.ReplayTagBotImprovisation {
		t.Errorf("tag = %q", got.Tag)
	}
	if got.Card != "Grim Tutor" || got.Seat != bot.String() || got.Text != line.Text {
		t.Errorf("annotation does not match what the table was told: %+v", got)
	}
	if len(got.Steps) != 3 {
		t.Errorf("annotation steps = %v, want the three verbs", got.Steps)
	}
}

// TestHumanUndoOfBotImprovisationRevertsTheWholeBundle is sub-PR 8's
// exit criterion, in full: a SEATED human — no admin token, caller is
// their own seat — undoes the improvisation, all three verbs revert
// together as ONE stack entry, and it costs them none of their own
// UndosRemaining.
func TestHumanUndoOfBotImprovisationRevertsTheWholeBundle(t *testing.T) {
	room := newRoom(t, 2, 12)
	g := room.Game
	bot, human := g.Seats[0].ID, g.Seats[1].ID
	card := handCardOf(t, g, bot)

	botLife, humanLife := lifeOf(g, bot), lifeOf(g, human)
	handBefore, graveBefore := zoneSizes(g, bot)
	undosBefore := undosRemaining(g, human)
	if undosBefore <= 0 {
		t.Fatalf("fixture problem: the human starts with %d undos", undosBefore)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pol := &improviser{im: grimTutor(t, bot, human, card)}
	r := aiseat.Start(ctx, room, bot, pol, aiseat.Config{}, nil, testLogger())
	waitFor(t, "the improvisation to apply", func() bool {
		return r.Stats().Improvisations == 1
	})
	// Stop the bot before undoing so it cannot act between the undo
	// and the assertions.
	cancel()
	<-r.Done()

	if lifeOf(g, human) != humanLife-2 {
		t.Fatalf("fixture problem: the improvisation did not apply")
	}

	// The undo, as a seated player. uuid.Nil here would be the admin
	// path, which is exactly what this must NOT need.
	if _, _, err := room.Undo(human); err != nil {
		t.Fatalf("seated human could not undo the bot's improvisation: %v", err)
	}

	// One entry, three verbs, all the way back.
	if got := lifeOf(g, bot); got != botLife {
		t.Errorf("bot life = %d after undo, want %d", got, botLife)
	}
	if got := lifeOf(g, human); got != humanLife {
		t.Errorf("opponent life = %d after undo, want %d", got, humanLife)
	}
	hand, grave := zoneSizes(g, bot)
	if hand != handBefore || grave != graveBefore {
		t.Errorf("hand/graveyard = %d/%d after undo, want %d/%d — the bundle did not revert as one entry", hand, grave, handBefore, graveBefore)
	}

	// And it did not cost the human their own take-back.
	if got := undosRemaining(g, human); got != undosBefore {
		t.Errorf("human has %d undos left, want %d — cleaning up after a bot must not spend a player's budget", got, undosBefore)
	}

	// The bundle was ONE entry: it was the only thing on the stack,
	// so there is nothing left to pop.
	if _, _, err := room.Undo(human); !errors.Is(err, ws.ErrNothingToUndo) {
		t.Errorf("second undo = %v, want ErrNothingToUndo — the bundle left more than one entry", err)
	}
}

// TestImprovisationBundleIsAtomic is the other half of "one bundle":
// a step that fails takes the whole thing with it. A half-applied
// improvisation is the state nobody at the table can reason about,
// and it must not be reachable.
func TestImprovisationBundleIsAtomic(t *testing.T) {
	room, dumpDir := newRoomWithDump(t, 2, 13)
	g := room.Game
	bot, human := g.Seats[0].ID, g.Seats[1].ID
	card := handCardOf(t, g, bot)

	im := grimTutor(t, bot, human, card)
	// Third step names a card that does not exist. The first two are
	// perfectly good, which is the point — they must not survive.
	im.Steps = append(im.Steps, aiseat.ImprovStep{
		Type:   aiseat.VerbAddCounter,
		Params: mustJSON(t, map[string]any{"instance_id": uuid.New().String(), "name": "+1/+1", "delta": 1}),
	})

	botLife, humanLife := lifeOf(g, bot), lifeOf(g, human)
	handBefore, graveBefore := zoneSizes(g, bot)
	seqBefore := room.Seq()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bc := &chatBroadcaster{}
	pol := &improviser{im: im}
	r := aiseat.Start(ctx, room, bot, pol, aiseat.Config{}, bc, testLogger())
	waitFor(t, "the improvisation to be refused", func() bool {
		return r.Stats().ImprovRefused == 1
	})
	cancel()
	<-r.Done()

	if lifeOf(g, bot) != botLife || lifeOf(g, human) != humanLife {
		t.Errorf("life totals moved: a failed bundle must roll back whole")
	}
	if hand, grave := zoneSizes(g, bot); hand != handBefore || grave != graveBefore {
		t.Errorf("hand/graveyard = %d/%d, want %d/%d — the move_card step survived a failed bundle", hand, grave, handBefore, graveBefore)
	}
	if r.Stats().Improvisations != 0 {
		t.Errorf("a rolled-back bundle must not count as an improvisation")
	}
	if n := len(bc.lines(protocol.ChatKindBotImprovisation)); n != 0 {
		t.Errorf("announced %d improvisations that never happened", n)
	}
	if n := len(replayAnnotations(t, dumpDir, g.ID)); n != 0 {
		t.Errorf("tagged %d replay lines for a bundle that rolled back", n)
	}
	// No commit at all: a rolled-back bundle allocates no sequence
	// number, so nobody ever saw the half-applied board.
	if got := room.Seq(); got != seqBefore {
		t.Errorf("seq advanced from %d to %d for a bundle that rolled back", seqBefore, got)
	}
}

// TestUnannouncedImprovisationIsRefused: an improvisation that will
// not say what it is doing is not applied. An unannounced
// improvisation is a bot cheating, so this is a correctness property,
// not input hygiene.
func TestUnannouncedImprovisationIsRefused(t *testing.T) {
	room := newRoom(t, 2, 14)
	g := room.Game
	bot, human := g.Seats[0].ID, g.Seats[1].ID
	card := handCardOf(t, g, bot)

	im := grimTutor(t, bot, human, card)
	im.Card, im.Effect = "", "" // the steps are still perfectly valid

	botLife := lifeOf(g, bot)
	handBefore, _ := zoneSizes(g, bot)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bc := &chatBroadcaster{}
	r := aiseat.Start(ctx, room, bot, &improviser{im: im}, aiseat.Config{}, bc, testLogger())
	waitFor(t, "the improvisation to be refused", func() bool {
		return r.Stats().ImprovRefused == 1
	})
	cancel()
	<-r.Done()

	if lifeOf(g, bot) != botLife {
		t.Errorf("an unannounceable improvisation was applied anyway")
	}
	if hand, _ := zoneSizes(g, bot); hand != handBefore {
		t.Errorf("an unannounceable improvisation moved a card")
	}
	if r.Stats().Improvisations != 0 {
		t.Errorf("stats counted a refused improvisation as applied")
	}
	if n := bc.count(); n != 0 {
		t.Errorf("a refused improvisation produced %d chat lines", n)
	}
}

// TestBotReasoningIsNarratedOnlyWhenAsked covers the setting's server
// half. The client half — hiding bot_reasoning lines unless the
// viewer opted in — is settings.gameplay.showBotReasoning.
func TestBotReasoningIsNarratedOnlyWhenAsked(t *testing.T) {
	for name, narrate := range map[string]bool{"on": true, "off": false} {
		t.Run(name, func(t *testing.T) {
			room := newRoom(t, 2, 15)
			g := room.Game
			bot := g.Seats[0].ID
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			bc := &chatBroadcaster{}
			pol := &scripted{prefer: []string{"Keep hand"}}
			r := aiseat.Start(ctx, room, bot, pol, aiseat.Config{Narrate: narrate}, bc, testLogger())
			waitFor(t, "the bot to act", func() bool { return r.Stats().Applied > 0 })
			cancel()
			<-r.Done()

			lines := bc.lines(protocol.ChatKindBotReasoning)
			if narrate && len(lines) == 0 {
				t.Fatalf("Narrate is on but nothing was narrated")
			}
			if !narrate && len(lines) != 0 {
				t.Fatalf("Narrate is off but %d lines were sent", len(lines))
			}
			for _, l := range lines {
				if l.Reason == "" {
					t.Errorf("a reasoning line with no reason: %+v", l)
				}
			}
		})
	}
}

// TestImprovisationVerbsWouldBeRefusedFromTheSeat is why the bundle
// is stamped uuid.Nil rather than with the bot's own seat: an
// improvised effect routinely reaches another player, and the
// seated-caller gate exists to stop exactly that.
func TestImprovisationVerbsWouldBeRefusedFromTheSeat(t *testing.T) {
	room := newRoom(t, 2, 16)
	g := room.Game
	bot, human := g.Seats[0].ID, g.Seats[1].ID

	drain := actions.Action{
		Type:   actions.TypeChangeLife,
		Player: human,
		Caller: bot,
		Params: json.RawMessage(`{"delta":-2}`),
	}
	if err := actions.Dispatch(g, drain); err == nil {
		t.Fatal("a seated bot could drain an opponent directly — the caller gate is not doing its job")
	}
	drain.Caller = uuid.Nil
	if err := actions.Dispatch(g, drain); err != nil {
		t.Fatalf("the uuid.Nil caller an improvisation uses was refused: %v", err)
	}
}

// --- helpers -------------------------------------------------------

// replayAnnotations streams the game's replay JSONL and returns every
// annotation on it. This is the grep the ADR asks for, spelled as a
// parse so the test can assert on the fields too.
func replayAnnotations(t *testing.T, dumpDir string, id uuid.UUID) []protocol.ReplayAnnotation {
	t.Helper()
	path := filepath.Join(dumpDir, "replays", id.String()+".jsonl")
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("read replay: %v", err)
	}
	var out []protocol.ReplayAnnotation
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if line == "" {
			continue
		}
		var p protocol.SnapshotPayload
		if err := json.Unmarshal([]byte(line), &p); err != nil {
			t.Fatalf("replay line is not a snapshot payload: %v", err)
		}
		if p.Annotation != nil {
			out = append(out, *p.Annotation)
		}
	}
	// Belt and braces: the tag has to survive as a literal string in
	// the file, because `grep bot_improvisation` is the interface.
	if n := strings.Count(string(raw), protocol.ReplayTagBotImprovisation); n != len(out) {
		t.Errorf("%d grep hits for %q but %d parsed annotations", n, protocol.ReplayTagBotImprovisation, len(out))
	}
	return out
}

func undosRemaining(g *game.Game, seat uuid.UUID) int {
	var n int
	g.ReadSnapshot(func() {
		if p := g.PlayerByIDForEffect(seat); p != nil {
			n = p.UndosRemaining
		}
	})
	return n
}

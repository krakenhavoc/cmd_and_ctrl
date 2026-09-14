package aiseat_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// visibility_test.go is the runtime half of ADR 0033 §3's
// hidden-information guarantee. The compile-time half already exists:
// TestPolicyPackagesDoNotImportGame in aiseat/heuristic walks every
// package under aiseat/ and fails the build on a direct import of
// internal/game, so a policy physically cannot hold the authoritative
// state.
//
// That gate proves a policy cannot REACH the engine. It says nothing
// about what it is HANDED. aiseat/policy.go's package doc makes a
// second, stronger claim — that Input.View is "byte-identical to what
// a human client in that seat receives" — and until this file nothing
// proved it. A runner that built its view with, say,
// protocol.ViewOfGame(g) instead of the per-viewer projection would
// pass the import test and hand the bot every hand at the table.
//
// The tests below hold three separate lines, and it is worth saying
// why equality alone is not enough: if the bot path and the human
// path were broken the SAME way, a byte-comparison of the two would
// pass while every hand at the table sat on both sides of it.
//
//  1. Equality. The runner's Input.View marshals to exactly the bytes
//     ws/hub.go produces for a seated human at the same seat.
//  2. The negative. No opponent hand or library card's name OR
//     instance ID appears anywhere in the bot's serialised Input —
//     view and move list both — and the bot's own library gives up no
//     card identity either.
//  3. The control. Those same secrets ARE present in the unfiltered
//     view, so the scan in (2) is matching something real rather than
//     passing because the strings were never there.
//
// The instance ID is treated as a secret in its own right wherever
// the card itself is secret, for the reason sub-PR 0's log work
// settled: a stable ID a policy can correlate across turns is the
// leak, not just the name. Where the ID is ALREADY public — a morph
// on the battlefield — it stays, and only the name has to be gone;
// withholding an ID the opponent can see anyway would be theatre.

// --- fixture -------------------------------------------------------

// visibilityDeck is a 40-card deck whose every card name is unique to
// its seat AND to itself ("Seat2Land07"). That uniqueness is the
// whole fixture: a substring scan for "did seat 2's hand leak" is
// only meaningful when no other seat, and no other card, could have
// put that string on the wire. monoRedDeck cannot be used here —
// every seat's hand is full of cards called "Mountain".
//
// Names are zero-padded so no card name is a substring of another
// ("Seat0Land1" would match inside "Seat0Land12").
func visibilityDeck(seat int, owner uuid.UUID) []game.Card {
	deck := make([]game.Card, 0, 40)
	cmdr := game.NewCommander(fmt.Sprintf("Seat%dCommander", seat), owner)
	cmdr.TypeLine = "Legendary Creature — Bear"
	cmdr.ManaCost = "{2}{R}"
	cmdr.Power, cmdr.Toughness = 3, 3
	deck = append(deck, cmdr)
	for i := 0; i < 24; i++ {
		c := game.NewCard(fmt.Sprintf("Seat%dLand%02d", seat, i), owner)
		c.TypeLine = "Basic Land — Mountain"
		deck = append(deck, c)
	}
	for i := 0; i < 15; i++ {
		c := game.NewCard(fmt.Sprintf("Seat%dBear%02d", seat, i), owner)
		c.TypeLine = "Creature — Bear"
		c.ManaCost = "{1}{R}"
		c.Power, c.Toughness = 2, 2
		deck = append(deck, c)
	}
	return deck
}

// newVisibilityGame seats four players on seat-unique decks and
// starts. keep closes the mulligan window; leaving it open is one of
// the scenarios.
func newVisibilityGame(t *testing.T, seed uint64, keep bool) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 4; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("Seat%d", i), visibilityDeck(i, uuid.Nil)); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(seed, seed+1))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if keep {
		for _, p := range g.Seats {
			if err := g.KeepHand(p.ID); err != nil {
				t.Fatalf("KeepHand: %v", err)
			}
		}
	}
	return g
}

// newVisibilityRoom is the ordinary fixture: mulligans closed, and the
// bot takes the seat that holds priority — it has to, or the runner
// would never be asked for a decision and there would be nothing to
// capture.
func newVisibilityRoom(t *testing.T, seed uint64) (*ws.Room, int) {
	t.Helper()
	g := newVisibilityGame(t, seed, true)
	return ws.NewRoom(g, testLogger(), ""), g.Turn.PriorityHolder
}

// scenario is one game state to prove the guarantee in. build returns
// the room, the seat the bot plays, and what was hidden where.
//
// One state is not enough, because FilterViewFor is not one rule. A
// zone is redacted card-by-card through the knower set; a pending
// choice is redacted through filterPendingChoices, which is a
// different function with a different rule (search_library drops its
// options wholesale for anyone but the chooser, because the LENGTH of
// the list is hidden information); legal_moves are picked per seat out
// of a map; and the mulligan window and combat each stamp fields the
// others do not. A single main-phase capture exercises the first of
// those and none of the rest.
type scenario struct {
	name  string
	build func(t *testing.T) (*ws.Room, uuid.UUID, material)
}

// scenarios are the states every assertion in this file runs against.
func scenarios() []scenario {
	return []scenario{
		{
			// The ordinary case: the bot holds priority in a main
			// phase, four opening hands and four libraries hidden.
			name: "main-phase",
			build: func(t *testing.T) (*ws.Room, uuid.UUID, material) {
				room, botIdx := newVisibilityRoom(t, 4242)
				return room, room.Game.Seats[botIdx].ID, seedHiddenMaterial(t, room.Game, botIdx)
			},
		},
		{
			// The mulligan window, where every seat is owed a
			// decision and the board is nothing but hidden zones.
			name: "mulligan-window",
			build: func(t *testing.T) (*ws.Room, uuid.UUID, material) {
				g := newVisibilityGame(t, 4242, false)
				if !g.MulligansOpen {
					t.Fatal("the mulligan window is not open")
				}
				return ws.NewRoom(g, testLogger(), ""), g.Seats[0].ID, seedHiddenMaterial(t, g, 0)
			},
		},
		{
			// Declare blockers, with the bot defending. Combat
			// stamps attack targets and block legality onto the view
			// through their own projection.
			name: "declare-blockers",
			build: buildBlockersScenario,
		},
		{
			// A Thoughtseize the bot cast: it owes a choice over an
			// opponent's hand, and is entitled to read it. Proves the
			// choice path is knower-driven — a blanket "hide every
			// opponent card" would pass every negative assertion in
			// this file and break the game.
			name: "bot-owes-a-thoughtseize-choice",
			build: func(t *testing.T) (*ws.Room, uuid.UUID, material) {
				room, botIdx := newVisibilityRoom(t, 4242)
				g := room.Game
				m := seedHiddenMaterial(t, g, botIdx)
				bot := g.Seats[botIdx]
				victim := g.Seats[(botIdx+1)%len(g.Seats)]
				g.WithWriteLock(func() {
					g.QueueDiscardFromRevealedHand(bot.ID, victim.ID, uuid.Nil, 1, "Thoughtseize")
				})
				return room, bot.ID, m
			},
		},
		{
			// The mirror, and the sharper of the two: an OPPONENT
			// owes a discard over their own hand, so that hand is
			// inlined into PendingChoiceView.Options on the
			// unfiltered view and has to come back to the bot as
			// backs. The bot owes one of its own as well, or it would
			// have no move to be asked about — the engine refuses
			// pass_priority while any choice is open.
			name: "opponent-owes-a-choice",
			build: func(t *testing.T) (*ws.Room, uuid.UUID, material) {
				room, botIdx := newVisibilityRoom(t, 4242)
				g := room.Game
				m := seedHiddenMaterial(t, g, botIdx)
				bot := g.Seats[botIdx]
				opp := g.Seats[(botIdx+1)%len(g.Seats)]
				g.WithWriteLock(func() {
					for _, p := range []*game.Player{opp, bot} {
						g.QueueChoiceForEffect(game.PendingChoice{
							Kind:       game.PendingChoiceDiscardFromHand,
							Chooser:    p.ID,
							FromPlayer: p.ID,
							Count:      1,
							Reason:     "Mind Rot",
						})
					}
				})
				return room, bot.ID, m
			},
		},
	}
}

// buildBlockersScenario puts the bot in the declare-blockers step as
// a defender with a legal block, which is the one window a seat is
// asked to act in without holding priority.
func buildBlockersScenario(t *testing.T) (*ws.Room, uuid.UUID, material) {
	t.Helper()
	g := newVisibilityGame(t, 4242, true)
	active := g.Turn.ActiveSeat
	botIdx := (active + 1) % len(g.Seats)
	m := seedHiddenMaterial(t, g, botIdx)

	// Battlefield cards are public, so every seat is a knower —
	// that is what the engine's zone move does, and a fixture that
	// skips it produces creatures nobody can see.
	everyone := make([]uuid.UUID, 0, len(g.Seats))
	for _, p := range g.Seats {
		everyone = append(everyone, p.ID)
	}
	public := func(name string, owner *game.Player) uuid.UUID {
		c := game.Card{
			InstanceID: uuid.New(),
			Name:       name,
			TypeLine:   "Creature — Bear",
			Power:      2, Toughness: 2,
			Owner: owner.ID, Controller: owner.ID,
		}
		c.AddKnowersAll(everyone)
		g.Battlefield.PushTop(c)
		return c.InstanceID
	}
	atk := public("PublicAttacker", g.Seats[active])
	public("PublicBlocker", g.Seats[botIdx])

	advanceToStep(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, g.Seats[botIdx].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceToStep(t, g, game.StepDeclareBlockers)
	return ws.NewRoom(g, testLogger(), ""), g.Seats[botIdx].ID, m
}

func advanceToStep(t *testing.T, g *game.Game, step game.Step) {
	t.Helper()
	for i := 0; i < 40; i++ {
		if g.Turn.Step == step {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep towards %s: %v", step, err)
		}
	}
	t.Fatalf("never reached %s (stuck at %s)", step, g.Turn.Step)
}

// capturingPolicy records the first Input it is handed and then parks
// inside Decide until its context is cancelled.
//
// Parking is load-bearing, not laziness. The comparison this file
// makes is between two projections of ONE game state, so the state
// must not move between them. A policy that returned a decision would
// let the runner dispatch it and advance the game, and the "human
// view" computed a moment later would be a view of a different board
// — a flaky test that proves nothing. Held inside Decide, with no
// other runner on the table, the game is frozen for as long as the
// assertions need it.
type capturingPolicy struct {
	once sync.Once
	got  chan aiseat.Input
}

func newCapturingPolicy() *capturingPolicy {
	return &capturingPolicy{got: make(chan aiseat.Input, 1)}
}

func (c *capturingPolicy) Name() string { return "capture" }

func (c *capturingPolicy) Decide(ctx context.Context, in aiseat.Input) (aiseat.Decision, error) {
	c.once.Do(func() { c.got <- in })
	<-ctx.Done()
	return aiseat.Decision{}, ctx.Err()
}

// captureBotInput starts one runner on the seat that holds priority
// and returns the exact Input the policy was handed, with the game
// still frozen at the state it was built from.
func captureBotInput(t *testing.T, room *ws.Room, seat uuid.UUID) aiseat.Input {
	t.Helper()
	pol := newCapturingPolicy()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	// MaxThink is an hour so the runner's own deadline cannot expire
	// mid-assertion, take the fallback pass, and move the board.
	r := aiseat.Start(ctx, room, seat, pol, aiseat.Config{MaxThink: time.Hour}, nil, testLogger())
	t.Cleanup(func() {
		cancel()
		<-r.Done()
	})
	select {
	case in := <-pol.got:
		return in
	case <-time.After(10 * time.Second):
		t.Fatal("the runner never asked the policy to decide")
		return aiseat.Input{}
	}
}

// secret is one thing the bot must not be able to read out of its
// own input, and what it is called in a failure message.
type secret struct {
	what  string
	token string
}

// hiddenFrom collects every string the bot is not entitled to: each
// opponent's hand and library cards (name AND instance ID), and the
// bot's own library cards (name and printed identity only — see the
// comment in TestBotInputLeaksNoHiddenInformation for why the ID is
// not on this list).
func hiddenFrom(g *game.Game, bot uuid.UUID) []secret {
	var out []secret
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			own := p.ID == bot
			for _, c := range p.Hand.Cards {
				if own || c.IsKnownTo(bot) {
					// The bot's own hand, and any card revealed to it
					// by a Thoughtseize-style effect, are its to see.
					continue
				}
				out = append(out,
					secret{fmt.Sprintf("%s's hand card %q", p.Name, c.Name), c.Name},
					secret{fmt.Sprintf("%s's hand card %q instance ID", p.Name, c.Name), c.InstanceID.String()},
				)
			}
			for _, c := range p.Library.Cards {
				if c.IsKnownTo(bot) {
					continue
				}
				out = append(out, secret{fmt.Sprintf("%s's library card %q", p.Name, c.Name), c.Name})
				if !own {
					out = append(out, secret{fmt.Sprintf("%s's library card %q instance ID", p.Name, c.Name), c.InstanceID.String()})
				}
			}
		}
	})
	return out
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

// --- 1. equality ----------------------------------------------------

// TestBotInputViewIsByteIdenticalToTheHumanView is aiseat/policy.go's
// package-doc claim, made executable: the view a policy is handed at
// a seat is the same bytes ws/hub.go sends to a human client at that
// seat.
//
// The human side of the comparison is deliberately built the way the
// hub builds it rather than the way the runner does — Room.Snapshot()
// for the full-fidelity view, then protocol.FilterViewFor with the
// seat's UUID string — because that is the expression at hub.go:381
// and hub.go:601 (viewerIDForFilter is the identity for a seated
// player; only a spectator's uuid.Nil becomes ""). If the two paths
// are ever "optimised" apart, this fails.
//
// Serialised comparison, not field-by-field, on purpose: a field
// added to GameView next sprint is covered by this test on the day it
// lands, without anyone remembering to extend a list.
func TestBotInputViewIsByteIdenticalToTheHumanView(t *testing.T) {
	for _, sc := range scenarios() {
		t.Run(sc.name, func(t *testing.T) {
			room, bot, _ := sc.build(t)
			in := captureBotInput(t, room, bot)

			full, _, err := room.Snapshot()
			if err != nil {
				t.Fatalf("snapshot: %v", err)
			}
			human := protocol.FilterViewFor(full, bot.String())

			got := mustMarshal(t, in.View)
			want := mustMarshal(t, human)
			if string(got) != string(want) {
				t.Errorf("the policy's view is not the human's view at that seat\n bot:   %s\n human: %s\n%s",
					truncate(string(got)), truncate(string(want)), firstDivergence(string(got), string(want)))
			}
			if in.Seat != bot {
				t.Errorf("Input.Seat = %s, want %s", in.Seat, bot)
			}
			if len(in.Moves) == 0 {
				t.Error("the policy was handed an empty move list; the capture is not exercising a real decision")
			}
		})
	}
}

// The same equality, for the OTHER three seats: a view built for seat
// N must not be the view for seat M. Without this the test above is
// satisfiable by a FilterViewFor that ignores its viewer argument and
// redacts everything for everyone.
func TestBotInputViewIsSeatSpecific(t *testing.T) {
	room, botIdx := newVisibilityRoom(t, 4242)
	g := room.Game
	bot := g.Seats[botIdx].ID
	seedHiddenMaterial(t, g, botIdx)

	in := captureBotInput(t, room, bot)
	full, _, err := room.Snapshot()
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	got := string(mustMarshal(t, in.View))
	for i, p := range g.Seats {
		if p.ID == bot {
			continue
		}
		other := string(mustMarshal(t, protocol.FilterViewFor(full, p.ID.String())))
		if got == other {
			t.Errorf("the bot's view at seat %d is identical to seat %d's view — FilterViewFor is not discriminating by viewer",
				botIdx, i)
		}
	}
}

// --- 2. the negative ------------------------------------------------

// TestBotInputLeaksNoHiddenInformation is the assertion that would
// actually catch a leak. Equality with the human view is necessary
// and not sufficient: two identically-broken projections compare
// equal. This one reads the bot's whole serialised Input — view AND
// move list — and looks for anything it should not know.
//
// Input.Moves matters more here than Input.View does, and it is the
// part a reader should be most suspicious of. The view goes through
// FilterViewFor. The move list does NOT: runner.go builds it with
// legal.EnumerateFor(r.room.Game, seat), straight off the
// authoritative state, and legal.Move.Label is free-form text built
// by calling cardName(g, id) on whatever the move touches. Nothing
// structural stops a future choice kind from naming a card the seat
// has not seen — legal/choices.go's discard_from_hand branch already
// labels every card in FromPlayer's hand, and is safe today only
// because the sole path that creates a chooser != discarder choice
// (QueueDiscardFromRevealedHand) reveals that hand to the chooser
// first. This test is what notices when that stops being true.
//
// Own library: the bot's library cards are not stripped from its view
// the way an opponent's are — the seat keeps a per-card entry so the
// client can render its own deck — so their INSTANCE IDs are on the
// wire, in library order. That is deliberate and matches what a human
// at the seat gets. It is also not a leak: an instance ID carries no
// identity, and MTG-legal knowledge of your own library (what you put
// on top, what you tutored) travels through KnownBy, which is exactly
// what makes those cards visible when they should be. What must never
// be recoverable is WHICH CARD is where, so the names, type lines and
// scryfall IDs are what this scans for, and a separate assertion below
// holds every printed characteristic to empty.
func TestBotInputLeaksNoHiddenInformation(t *testing.T) {
	for _, sc := range scenarios() {
		t.Run(sc.name, func(t *testing.T) {
			room, bot, mat := sc.build(t)
			g := room.Game
			assertNoLeak(t, g, bot, mat, captureBotInput(t, room, bot))
		})
	}
}

func assertNoLeak(t *testing.T, g *game.Game, bot uuid.UUID, mat material, in aiseat.Input) {
	t.Helper()

	view := string(mustMarshal(t, in.View))
	moves := string(mustMarshal(t, in.Moves))
	whole := string(mustMarshal(t, struct {
		View  protocol.GameView `json:"view"`
		Moves any               `json:"moves"`
		Seat  string            `json:"seat"`
	}{in.View, in.Moves, in.Seat.String()}))

	secrets := hiddenFrom(g, bot)
	if len(secrets) == 0 {
		t.Fatal("the fixture put nothing hidden on the table; the scan would pass vacuously")
	}
	for _, s := range secrets {
		if strings.Contains(view, s.token) {
			t.Errorf("Input.View leaked %s (%q)", s.what, s.token)
		}
		if strings.Contains(moves, s.token) {
			t.Errorf("Input.Moves leaked %s (%q) — the move list does not go through FilterViewFor", s.what, s.token)
		}
		if strings.Contains(whole, s.token) {
			t.Errorf("the bot's serialised Input leaked %s (%q)", s.what, s.token)
		}
	}

	// The morph and the face-down exiled card: on public zones, so
	// their instance IDs are legitimately on the wire and only the
	// NAME is secret. Same line sub-PR 0's log work took.
	for _, m := range []struct {
		what, name string
		id         uuid.UUID
	}{
		{"the opponent's morph", mat.morphName, mat.morphID},
		{"the face-down exiled card", mat.exileName, mat.exileID},
	} {
		if strings.Contains(whole, m.name) {
			t.Errorf("%s leaked its name %q to the bot", m.what, m.name)
		}
		if !strings.Contains(view, m.id.String()) {
			t.Errorf("%s lost its instance ID from the bot's view; it is in a public zone and the ID is not the secret", m.what)
		}
	}

	// Positive controls. A scan for absent strings passes trivially
	// against an empty input, so prove the input is real: what the
	// bot IS entitled to has to be there.
	for _, s := range visibleTo(g, bot) {
		if !strings.Contains(whole, s.token) {
			t.Errorf("the bot cannot see %s (%q) — the view is over-redacted, not just un-leaky", s.what, s.token)
		}
	}
	if !strings.Contains(whole, mat.revealedName) {
		t.Errorf("the Thoughtseize-revealed card %q is missing from the bot's input; redaction is scrubbing strings rather than honouring KnownBy",
			mat.revealedName)
	}

	// And the bot's own library gives up no identity, field by field
	// rather than by substring — a name is not the only way to
	// recognise a card.
	var lib protocol.ZoneView
	for _, s := range in.View.Seats {
		if s.ID == bot.String() {
			lib = s.Library
		}
	}
	if lib.Count == 0 || len(lib.Cards) == 0 {
		t.Fatalf("the bot's own library is empty in its view (count=%d cards=%d); nothing was checked", lib.Count, len(lib.Cards))
	}
	for i, c := range lib.Cards {
		if c.KnownByYou || c.Name != "" || c.TypeLine != "" || c.ScryfallID != "" || c.ManaCost != "" || c.Power != 0 || c.Toughness != 0 {
			t.Errorf("the bot can read its own library at position %d: %+v — the order of an unknown library is not the leak, its contents are", i, c)
		}
	}
}

// --- 3. the control -------------------------------------------------

// TestHiddenInformationScanIsLoadBearing is the control for the test
// above, and the reason it exists is the same reason
// TestImportBanIsLoadBearing exists next door: a scan for strings that
// are not there passes just as well when they were never there at
// all. Rename a fixture, change a projection so hands stop being
// serialised, and the leak test goes green for the wrong reason and
// stays green for a year.
//
// So: the identical secrets, scanned against the UNFILTERED view. All
// of them must be present. If this test fails, the leak test above is
// measuring nothing.
func TestHiddenInformationScanIsLoadBearing(t *testing.T) {
	for _, sc := range scenarios() {
		t.Run(sc.name, func(t *testing.T) {
			room, bot, mat := sc.build(t)
			assertScanIsLoadBearing(t, room.Game, bot, mat)
		})
	}
}

func assertScanIsLoadBearing(t *testing.T, g *game.Game, bot uuid.UUID, mat material) {
	t.Helper()
	unfiltered := string(mustMarshal(t, protocol.ViewOfGame(g)))
	secrets := hiddenFrom(g, bot)
	if len(secrets) == 0 {
		t.Fatal("no secrets in the fixture")
	}
	missing := 0
	for _, s := range secrets {
		if !strings.Contains(unfiltered, s.token) {
			missing++
			if missing <= 5 {
				t.Errorf("%s (%q) is absent from the UNFILTERED view — the leak scan is looking for a string the wire never carries", s.what, s.token)
			}
		}
	}
	if missing > 5 {
		t.Errorf("... and %d more", missing-5)
	}
	for _, name := range []string{mat.morphName, mat.exileName} {
		if !strings.Contains(unfiltered, name) {
			t.Errorf("%q is absent from the unfiltered view; the face-down fixture is not being projected at all", name)
		}
	}
}

// --- fixture material ----------------------------------------------

// material is what seedHiddenMaterial put on the table, so the
// assertions can name it.
type material struct {
	morphID              uuid.UUID
	morphName            string
	exileID              uuid.UUID
	exileName            string
	revealedName         string
	revealedID           uuid.UUID
	botGraveyardCardName string
}

// seedHiddenMaterial puts something in every hidden zone that has a
// distinct visibility rule, so the scan has more to find than four
// opening hands:
//
//   - a face-down morph on the BATTLEFIELD, known only to its
//     controller — a public zone holding a private identity, which is
//     the case where the instance ID is public and the name is not
//   - a face-down card in EXILE, same shape (Bolas's Citadel, the
//     foretell/cascade family)
//   - one card in an opponent's hand REVEALED to the bot, the
//     Thoughtseize case — proof that redaction honours KnownBy rather
//     than blanket-hiding every opponent card, which would make the
//     negative assertions pass for the wrong reason
//   - a card in the bot's own graveyard, which is public and must
//     survive
//
// Called before any runner starts, so no lock is needed — the same
// convention the rest of this package's fixtures use.
func seedHiddenMaterial(t *testing.T, g *game.Game, botIdx int) material {
	t.Helper()
	bot := g.Seats[botIdx]
	opp := g.Seats[(botIdx+1)%len(g.Seats)]
	third := g.Seats[(botIdx+2)%len(g.Seats)]

	var m material
	m.morphID, m.morphName = uuid.New(), "MorphSecretOfTheOpponent"
	morph := game.Card{
		InstanceID: m.morphID,
		Name:       m.morphName,
		TypeLine:   "Creature — Bear",
		Power:      2, Toughness: 2,
		Owner: opp.ID, Controller: opp.ID,
		FaceDown: true,
		KnownBy:  map[uuid.UUID]bool{opp.ID: true},
	}
	g.Battlefield.PushTop(morph)

	m.exileID, m.exileName = uuid.New(), "ExiledSecretOfTheThirdSeat"
	g.Exile.PushTop(game.Card{
		InstanceID: m.exileID,
		Name:       m.exileName,
		TypeLine:   "Sorcery",
		Owner:      third.ID, Controller: third.ID,
		FaceDown: true,
		KnownBy:  map[uuid.UUID]bool{third.ID: true},
	})

	// Thoughtseize: the bot saw one card in the opponent's hand.
	if len(opp.Hand.Cards) == 0 {
		t.Fatal("the opponent has no hand to reveal from")
	}
	opp.Hand.Cards[0].AddKnower(bot.ID)
	m.revealedName = opp.Hand.Cards[0].Name
	m.revealedID = opp.Hand.Cards[0].InstanceID

	// A graveyard is public, and in the engine that is expressed by
	// the zone move calling Card.AddKnowersAll — visibility is a
	// property of the KNOWER SET, not of the zone, so a fixture that
	// pushes a card straight onto a public zone with no knowers
	// produces a card nobody can read. Seat every viewer here, the
	// way MoveCard would.
	everyone := make([]uuid.UUID, 0, len(g.Seats))
	for _, p := range g.Seats {
		everyone = append(everyone, p.ID)
	}
	m.botGraveyardCardName = "BotGraveyardCard"
	gy := game.Card{
		InstanceID: uuid.New(),
		Name:       m.botGraveyardCardName,
		TypeLine:   "Instant",
		Owner:      bot.ID, Controller: bot.ID,
	}
	gy.AddKnowersAll(everyone)
	g.Seats[botIdx].Graveyard.PushTop(gy)
	return m
}

// visibleTo is hiddenFrom's mirror: what the bot IS entitled to see
// and must therefore find in its own input. Without it, every
// "does not contain" assertion in this file is satisfiable by handing
// the policy an empty struct.
func visibleTo(g *game.Game, bot uuid.UUID) []secret {
	var out []secret
	g.ReadSnapshot(func() {
		p := g.PlayerByID(bot)
		if p == nil {
			return
		}
		for _, c := range p.Hand.Cards {
			out = append(out,
				secret{"its own hand card " + c.Name, c.Name},
				secret{"its own hand card " + c.Name + " instance ID", c.InstanceID.String()},
			)
		}
		for _, c := range p.Graveyard.Cards {
			out = append(out, secret{"its own graveyard card " + c.Name, c.Name})
		}
	})
	return out
}

// --- failure-message helpers ---------------------------------------

func truncate(s string) string {
	const max = 600
	if len(s) <= max {
		return s
	}
	return s[:max] + fmt.Sprintf("… (%d bytes total)", len(s))
}

// firstDivergence points at where two JSON documents part company,
// because a byte-equality failure on a 40 KB view is unreadable
// otherwise.
func firstDivergence(a, b string) string {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			lo := i - 120
			if lo < 0 {
				lo = 0
			}
			hiA, hiB := i+120, i+120
			if hiA > len(a) {
				hiA = len(a)
			}
			if hiB > len(b) {
				hiB = len(b)
			}
			return fmt.Sprintf(" first difference at byte %d\n  bot:   …%s…\n  human: …%s…", i, a[lo:hiA], b[lo:hiB])
		}
	}
	return fmt.Sprintf(" identical for %d bytes, then one is longer (bot %d, human %d)", n, len(a), len(b))
}

package model

import (
	"fmt"
	"sort"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/boardtext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// prompt.go assembles the two halves of a Layer C call.
//
// # The split is the cost model
//
// Prompt caching is a PREFIX match: one changed byte anywhere in the
// prefix invalidates everything after it. So the split is not a
// tidiness preference, it is the whole reason the funnel is
// affordable:
//
//   - the STATIC half — rules primer, the bot's decklist with oracle
//     text, the archetype plan — is built once per seat, is
//     byte-identical for every decision in the game, and carries the
//     cache breakpoint. Nothing derived from the board may appear in
//     it, and nothing time-varying at all: a turn number or a
//     timestamp in here would silently cost the cache on every call
//     and nothing would report it.
//
//   - the DELTA — board state and the move list — is the user turn,
//     after the breakpoint, and is assembled from aiseat.Input and
//     nothing else.
//
// # Why the model is shown indices and not cards
//
// The move list is presented with its REAL indices into Input.Moves.
// No renumbering, no compaction of the index space: the answer maps
// straight back with no table in between, so there is no off-by-one
// to get wrong on a path that only ever runs in production. When the
// list is capped, the entries that are dropped simply do not appear,
// and their indices are absent from the sequence — which is fine,
// because the only contract is "an integer that indexes Moves".
//
// The reply also copies the chosen entry's label (#2196). A model
// that names the move it wants in words but writes a number that is
// not on the list has still said which move it meant, and
// ResolveAnswer takes the words when they name exactly one listed
// entry. The label is looked up in the shown list and nowhere else,
// so it can select a move and never describe one.

// DeckCard is one card in the bot's own list, as the static block
// describes it. Plain data: no engine types, nothing derived from a
// game in progress.
type DeckCard struct {
	Name   string
	Cost   string
	Type   string
	Oracle string
}

// DeckProfile is the bot's deck and plan — ADR 0033 §5's "decklist
// with oracle text, archetype plan". It is configuration handed to
// the seat when it is created, not something read off the board, so
// it stays byte-identical for the whole game and the cache holds.
type DeckProfile struct {
	// Name is the deck's name, e.g. "Izzet aggro (Mary Read and Anne
	// Bonny)".
	Name string
	// Archetype is the plan in prose: how this deck wins, what it
	// holds up, what it is happy to trade.
	Archetype string
	// Cards is the list. Order is normalised on use, so a caller
	// that shuffles it does not cost the cache.
	Cards []DeckCard
}

// rulesPrimer is the constant half of the static block. Short on
// purpose: the model knows Magic, it does not know THIS server's
// posture, and only the second is worth tokens.
const rulesPrimer = `You are playing one seat in a four-player game of Magic: the Gathering (Commander) on a private hobby server.

How you act:
- You are given a numbered list of the legal moves your seat may make right now. The list is closed and it is complete — it was produced by the server's rules engine, and every entry is a move the server will accept.
- You answer with ONE NUMBER from that list, and copy that entry's text beside it. You never describe an action of your own, and never propose a move that is not listed. An answer that names nothing on the list is discarded and a rule-based fallback plays instead.
- The list is everything you can do right now. A card in your hand that is not listed cannot be played in this window: it is the wrong step for it, or you cannot pay for it yet. Never answer with a number that is not listed.
- A line such as "Mountain: Add {R}" taps a land you already control for mana. It is not a land drop, and mana you do not spend empties at the end of the step, so tap for mana only when you are about to spend it.
- You see only what your seat is entitled to see: your own hand and command zone, and public information. Opponents' hands are counts. Do not reason about specific cards you have not been shown.

What this server is:
- It is a sandbox with rules grafted on. Only a few hundred cards have implemented rules text; a card marked "(unimplemented)" will physically do nothing when it resolves, whatever it says. Weigh those accordingly.
- Play to the standard of a good, ordinary Commander player: make the land drop, develop the board, spend removal on the biggest threat, attack the player who is winning, and hold up an instant when holding it is worth more than casting it.

How to answer:
- Reply with a single JSON object and nothing else, in this exact shape:
  {"index": <number>, "move": "<that entry's text, copied exactly as it is listed>", "why": "<at most twelve words>"}
- No preamble, no code fence, no explanation outside the JSON, and no internal or system XML tags.`

// staticBlocks builds the cached half. The breakpoint goes on the
// LAST block so tools (none) and system cache together.
func (d DeckProfile) staticBlocks() []Block {
	blocks := []Block{{Text: rulesPrimer}}
	var b strings.Builder
	if d.Name != "" {
		fmt.Fprintf(&b, "YOUR DECK: %s\n", d.Name)
	}
	if d.Archetype != "" {
		fmt.Fprintf(&b, "\nTHE PLAN\n%s\n", strings.TrimSpace(d.Archetype))
	}
	if len(d.Cards) > 0 {
		// Normalised order: the cache key is the exact bytes, so a
		// caller that built the list in a different order must not
		// pay for it.
		cards := append([]DeckCard(nil), d.Cards...)
		sort.Slice(cards, func(i, j int) bool { return cards[i].Name < cards[j].Name })
		b.WriteString("\nDECKLIST (oracle text)\n")
		for _, c := range cards {
			b.WriteString("- ")
			b.WriteString(c.Name)
			if c.Cost != "" {
				b.WriteString(" " + c.Cost)
			}
			if c.Type != "" {
				b.WriteString(" — " + c.Type)
			}
			if c.Oracle != "" {
				b.WriteString(": " + OneLine(c.Oracle))
			}
			b.WriteByte('\n')
		}
	}
	if b.Len() > 0 {
		blocks = append(blocks, Block{Text: strings.TrimRight(b.String(), "\n")})
	}
	blocks[len(blocks)-1].Cache = true
	return blocks
}

// --- the per-decision delta ----------------------------------------

// fallbackMarker ends the line of the move Layer B would take. It is a
// constant because three readers depend on its exact text: the model,
// fake.go's indexFromPrompt, and normLabel, which takes it off a label
// a model copied along with it.
const fallbackMarker = "<- the rule-based fallback would take this"

// shownMove is one entry of the move list as the model sees it.
type shownMove struct {
	index int
	label string
}

// buildDelta renders the board and the move list from Input alone.
// cands is the heuristic's ranking, or nil when the window is one the
// scorer does not price; fallback is the index Layer B would take.
func (p *Policy) buildDelta(in aiseat.Input, cands []heuristic.Candidate, fallback int) (string, []shownMove) {
	// The board half is shared with the MCP seat (ADR 0122 §5); the
	// bot leaves Options.NoteUnimplemented off, so its text is unchanged.
	var b strings.Builder
	b.WriteString(boardtext.Render(&in.View, in.Seat.String(), boardtext.Options{MaxZoneCards: p.cfg.MaxZoneCards}))

	shown := p.selectShown(in, cands, fallback)
	b.WriteString("\nYOUR LEGAL MOVES — answer with one of these numbers\n")
	for _, m := range shown {
		fmt.Fprintf(&b, "  %d: %s", m.index, m.label)
		if m.index == fallback {
			b.WriteString("   " + fallbackMarker)
		}
		b.WriteByte('\n')
	}
	if len(shown) < len(in.Moves) {
		fmt.Fprintf(&b, "(%d further legal moves the bot's scorer ranked below these are not listed.)\n",
			len(in.Moves)-len(shown))
	}
	b.WriteString("\nWhich number?")
	return b.String(), shown
}

// selectShown picks the moves the model is offered. Capped, because a
// full board can enumerate into the dozens and every entry is paid
// for on every call; ordered by the heuristic's own ranking when
// there is one, so the cap drops the moves the scorer liked least
// rather than whichever ones the enumerator happened to emit last.
//
// The pass and the fallback are always present. A model that cannot
// see the pass cannot choose to do nothing, and doing nothing is
// frequently right.
func (p *Policy) selectShown(in aiseat.Input, cands []heuristic.Candidate, fallback int) []shownMove {
	max := p.cfg.MaxCandidates
	if max <= 0 || max > len(in.Moves) {
		max = len(in.Moves)
	}
	picked := make([]int, 0, max)
	seen := make(map[int]bool, max)
	add := func(i int) {
		if i < 0 || i >= len(in.Moves) || seen[i] || len(picked) >= max {
			return
		}
		seen[i] = true
		picked = append(picked, i)
	}
	if pi := aiseat.PassIndex(in.Moves); pi >= 0 {
		add(pi)
	}
	add(fallback)
	for _, c := range cands {
		add(c.Index)
	}
	for i := range in.Moves {
		add(i)
	}
	out := make([]shownMove, 0, len(picked))
	for _, i := range picked {
		out = append(out, shownMove{index: i, label: in.Moves[i].Label})
	}
	// Ascending index, so the list reads as a list rather than as a
	// recommendation with the scorer's opinion baked into the order.
	sort.Slice(out, func(i, j int) bool { return out[i].index < out[j].index })
	return out
}

// choicesOf is the shown list as Request.Choices carries it: what a
// transport may constrain the reply to, and what ResolveAnswer looks a
// reply up in. The same entries the prompt rendered, in the same
// order, and nothing else — a move the cap dropped is not a choice.
func choicesOf(shown []shownMove) []Choice {
	out := make([]Choice, 0, len(shown))
	for _, m := range shown {
		out = append(out, Choice{Index: m.index, Label: m.label})
	}
	return out
}

// --- small renderers -----------------------------------------------
//
// The board renderers live in aiseat/boardtext (ADR 0122 §5). These are
// the bot-flavoured spellings the rest of this package (improvise.go and
// the tests) calls: no NoteUnimplemented, so the bytes are the old ones.

func permanentLabel(c *protocol.CardView) string {
	return boardtext.PermanentLabel(c, boardtext.Options{})
}

func emblemLine(e protocol.EmblemView) string { return boardtext.EmblemLine(e) }

func cardName(c *protocol.CardView) string { return boardtext.CardName(c) }

func stepName(step string) string { return boardtext.StepName(step) }

func seatLabel(s *protocol.PlayerView, me string) string { return boardtext.SeatLabel(s, me) }

func seatLabelByID(v *protocol.GameView, id, me string) string {
	return boardtext.SeatLabelByID(v, id, me)
}

func stackOwnership(it *protocol.StackItemView, v *protocol.GameView, me string) string {
	return boardtext.StackOwnership(it, v, me)
}

func stackLabel(it *protocol.StackItemView, v *protocol.GameView) string {
	return boardtext.StackLabel(it, v)
}

// OneLine collapses newlines and runs of whitespace into single
// spaces. Exported alongside Truncate, and for the same reason: a
// model's reply printed into a log line must not bring its own line
// breaks with it.
func OneLine(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "\n", " ")), " ")
}

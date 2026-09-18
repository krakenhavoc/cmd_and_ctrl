package model

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
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
- You answer with ONE NUMBER from that list. You never describe an action, never name a card as your answer, and never propose a move that is not listed. A number that is not in the list is discarded and a rule-based fallback plays instead.
- You see only what your seat is entitled to see: your own hand and command zone, and public information. Opponents' hands are counts. Do not reason about specific cards you have not been shown.

What this server is:
- It is a sandbox with rules grafted on. Only a few hundred cards have implemented rules text; a card marked "(unimplemented)" will physically do nothing when it resolves, whatever it says. Weigh those accordingly.
- Play to the standard of a good, ordinary Commander player: make the land drop, develop the board, spend removal on the biggest threat, attack the player who is winning, and hold up an instant when holding it is worth more than casting it.

How to answer:
- Reply with a single JSON object and nothing else, in this exact shape:
  {"index": <number>, "why": "<at most twelve words>"}
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

// shownMove is one entry of the move list as the model sees it.
type shownMove struct {
	index int
	label string
}

// buildDelta renders the board and the move list from Input alone.
// cands is the heuristic's ranking, or nil when the window is one the
// scorer does not price; fallback is the index Layer B would take.
func (p *Policy) buildDelta(in aiseat.Input, cands []heuristic.Candidate, fallback int) (string, []shownMove) {
	v := &in.View
	me := in.Seat.String()
	var b strings.Builder

	fmt.Fprintf(&b, "TURN %d — %s", v.Turn.Number, stepName(v.Turn.Step))
	if as := seatAt(v, v.Turn.ActiveSeat); as != nil {
		fmt.Fprintf(&b, " — active player: %s", seatLabel(as, me))
	}
	if ph := seatAt(v, v.Turn.PriorityHolder); ph != nil {
		fmt.Fprintf(&b, " — priority: %s", seatLabel(ph, me))
	}
	b.WriteByte('\n')

	// Seats, starting with this one.
	order := make([]*protocol.PlayerView, 0, len(v.Seats))
	for i := range v.Seats {
		if v.Seats[i].ID == me {
			order = append(order, &v.Seats[i])
		}
	}
	for i := range v.Seats {
		if v.Seats[i].ID != me {
			order = append(order, &v.Seats[i])
		}
	}
	for _, s := range order {
		b.WriteByte('\n')
		if s.Eliminated {
			fmt.Fprintf(&b, "%s — ELIMINATED\n", seatLabel(s, me))
			continue
		}
		fmt.Fprintf(&b, "%s — %d life, %d cards in hand, %d in library",
			seatLabel(s, me), s.Life, s.Hand.Count, s.Library.Count)
		if len(s.ManaPool) > 0 {
			fmt.Fprintf(&b, ", mana pool %s", strings.Join(s.ManaPool, ""))
		}
		b.WriteByte('\n')
		if bf := p.battlefieldOf(v, s.ID); bf != "" {
			fmt.Fprintf(&b, "  battlefield: %s\n", bf)
		}
		if gy := cardNames(s.Graveyard.Cards, p.cfg.MaxZoneCards); gy != "" {
			fmt.Fprintf(&b, "  graveyard: %s\n", gy)
		}
		if s.ID == me {
			if hand := describeCards(s.Hand.Cards, p.cfg.MaxZoneCards); hand != "" {
				fmt.Fprintf(&b, "  your hand: %s\n", hand)
			}
			if cmd := describeCards(s.Command.Cards, p.cfg.MaxZoneCards); cmd != "" {
				fmt.Fprintf(&b, "  your command zone: %s\n", cmd)
			}
		}
	}

	if len(v.StackItems) > 0 {
		b.WriteString("\nSTACK (bottom first — the last entry resolves next)\n")
		for i := range v.StackItems {
			it := &v.StackItems[i]
			fmt.Fprintf(&b, "  %s — cast by %s", stackLabel(it, v), seatLabelByID(v, it.Controller, me))
			if len(it.Targets) > 0 {
				fmt.Fprintf(&b, " — targets: %s", targetList(v, it.Targets, me))
			}
			b.WriteByte('\n')
		}
	}
	for i := range v.PendingChoices {
		ch := &v.PendingChoices[i]
		if ch.Chooser != me {
			continue
		}
		fmt.Fprintf(&b, "\nYOU OWE A CHOICE: %s", ch.Kind)
		if ch.Reason != "" {
			fmt.Fprintf(&b, " — %s", ch.Reason)
		}
		if ch.Count > 0 {
			fmt.Fprintf(&b, " (choose %d)", ch.Count)
		}
		b.WriteByte('\n')
	}

	shown := p.selectShown(in, cands, fallback)
	b.WriteString("\nYOUR LEGAL MOVES — answer with one of these numbers\n")
	for _, m := range shown {
		fmt.Fprintf(&b, "  %d: %s", m.index, m.label)
		if m.index == fallback {
			b.WriteString("   <- the rule-based fallback would take this")
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

// --- small renderers -----------------------------------------------

func (p *Policy) battlefieldOf(v *protocol.GameView, seat string) string {
	var parts []string
	for i := range v.Battlefield.Cards {
		c := &v.Battlefield.Cards[i]
		if c.Controller != seat {
			continue
		}
		if len(parts) >= p.cfg.MaxZoneCards {
			parts = append(parts, "…")
			break
		}
		parts = append(parts, permanentLabel(c))
	}
	return strings.Join(parts, ", ")
}

func permanentLabel(c *protocol.CardView) string {
	var b strings.Builder
	b.WriteString(cardName(c))
	if isCreatureLine(c.TypeLine) {
		fmt.Fprintf(&b, " %d/%d", c.Power, c.Toughness)
	}
	var flags []string
	if c.Tapped {
		flags = append(flags, "tapped")
	}
	if heuristic.WontUntap(c) {
		flags = append(flags, "won't untap")
	}
	if c.SummoningSick {
		flags = append(flags, "sick")
	}
	if c.DamageMarked > 0 {
		flags = append(flags, strconv.Itoa(c.DamageMarked)+" damage")
	}
	if c.AttackingTarget != "" {
		flags = append(flags, "attacking")
	}
	if c.BlockingTarget != "" {
		flags = append(flags, "blocking")
	}
	if c.IsCommander {
		flags = append(flags, "commander")
	}
	if c.Unimplemented {
		flags = append(flags, "unimplemented")
	}
	flags = append(flags, sortedCounters(c.Counters)...)
	if len(c.Abilities) > 0 {
		flags = append(flags, strings.Join(c.Abilities, " "))
	}
	if len(flags) > 0 {
		b.WriteString(" (" + strings.Join(flags, ", ") + ")")
	}
	return b.String()
}

// sortedCounters renders counters deterministically. Map iteration
// order is randomised in Go, and an unsorted render would change the
// prompt bytes between two identical boards — noise in the log, and a
// cache miss if any of this ever moved into the static block.
func sortedCounters(m map[string]int) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, fmt.Sprintf("%d %s", m[k], k))
	}
	return out
}

func describeCards(cards []protocol.CardView, max int) string {
	var parts []string
	for i := range cards {
		if len(parts) >= max {
			parts = append(parts, "…")
			break
		}
		c := &cards[i]
		s := cardName(c)
		if c.ManaCost != "" {
			s += " " + c.ManaCost
		}
		if c.TypeLine != "" {
			s += " — " + c.TypeLine
		}
		if c.Unimplemented {
			s += " (unimplemented)"
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, "; ")
}

func cardNames(cards []protocol.CardView, max int) string {
	var parts []string
	for i := range cards {
		if len(parts) >= max {
			parts = append(parts, "…")
			break
		}
		parts = append(parts, cardName(&cards[i]))
	}
	return strings.Join(parts, ", ")
}

// cardName respects the view's own redaction. A face-down card the
// seat may not look at arrives with no name, and the prompt says so
// rather than inventing one.
//
// ADR 0069: the question is `face_visible` — the rules permission to
// look at the face (CR 406.3a, CR 702.143d, CR 708.5) — not
// `known_by_you`. They agree on the wire today, but naming the
// permission is what stops a bot reading a hidden face if they ever
// come apart, and it is what makes a face-down PERMANENT read as the
// public CR 708.2 object it is: "a face-down 2/2" to the table, not
// the card under it even to its own controller, who sees the card in
// their client but is playing against an object with no name.
func cardName(c *protocol.CardView) string {
	if c.IsFaceDownPermanent() {
		return "a face-down 2/2 creature"
	}
	if c.FaceDown && !c.FaceVisible {
		return "a face-down card"
	}
	if c.Name == "" {
		return "an unknown card"
	}
	return c.Name
}

func isCreatureLine(t string) bool { return strings.Contains(strings.ToLower(t), "creature") }

func stepName(step string) string { return strings.ReplaceAll(step, "_", " ") }

func seatAt(v *protocol.GameView, i int) *protocol.PlayerView {
	if i < 0 || i >= len(v.Seats) {
		return nil
	}
	return &v.Seats[i]
}

func seatLabel(s *protocol.PlayerView, me string) string {
	if s == nil {
		return "?"
	}
	name := s.Name
	if name == "" {
		name = "seat " + strconv.Itoa(s.Seat)
	}
	if s.ID == me {
		return name + " (YOU)"
	}
	return name
}

func seatLabelByID(v *protocol.GameView, id, me string) string {
	for i := range v.Seats {
		if v.Seats[i].ID == id {
			return seatLabel(&v.Seats[i], me)
		}
	}
	return "someone"
}

func stackLabel(it *protocol.StackItemView, v *protocol.GameView) string {
	if it.Label != "" {
		return it.Label
	}
	for i := range v.Stack.Cards {
		if v.Stack.Cards[i].InstanceID == it.SourceCardID {
			return cardName(&v.Stack.Cards[i])
		}
	}
	return it.Kind
}

func targetList(v *protocol.GameView, targets []protocol.TargetRefView, me string) string {
	parts := make([]string, 0, len(targets))
	for _, t := range targets {
		switch t.Kind {
		case "player":
			parts = append(parts, seatLabelByID(v, t.ID, me))
		default:
			parts = append(parts, nameOfInstance(v, t.ID))
		}
	}
	return strings.Join(parts, ", ")
}

func nameOfInstance(v *protocol.GameView, id string) string {
	for i := range v.Battlefield.Cards {
		if v.Battlefield.Cards[i].InstanceID == id {
			return cardName(&v.Battlefield.Cards[i])
		}
	}
	for i := range v.Stack.Cards {
		if v.Stack.Cards[i].InstanceID == id {
			return cardName(&v.Stack.Cards[i])
		}
	}
	return "something"
}

// OneLine collapses newlines and runs of whitespace into single
// spaces. Exported alongside Truncate, and for the same reason: a
// model's reply printed into a log line must not bring its own line
// breaks with it.
func OneLine(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "\n", " ")), " ")
}

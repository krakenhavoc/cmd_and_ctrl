// Package boardtext renders a seat's filtered view of the table as the
// plain text the bot's prompt and (ADR 0122 §5) the MCP seat's compact
// view both carry: the turn line, each seat with its zones, the stack,
// and the seat's owed choices.
//
// It is the board half of what model.buildDelta used to do, moved here
// so the two readers share one renderer: a fix to one is a fix to the
// other. It is pure. Input is a protocol.GameView that has already been
// through FilterViewFor, so it can only say what the seat may see, and
// the package imports nothing from internal/game, ws, lobby or actions
// (board_imports_test.go holds that). The one engine-adjacent import is
// aiseat/heuristic, for WontUntap, a function of a CardView.
//
// The text is byte-for-byte what the bot's prompt carried before the
// move, and the model package's prompt tests hold it there.
package boardtext

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// DefaultMaxZoneCards is the zone cap an Options with no MaxZoneCards uses.
const DefaultMaxZoneCards = 24

// UnimplementedNote is the text an agent is given on a card whose
// text the engine does not run, when Options.NoteUnimplemented is set.
const UnimplementedNote = "unimplemented: the engine does not run this card's text"

// Options are the renderer's knobs.
type Options struct {
	// MaxZoneCards caps how much of a zone is listed. Zero or less
	// means DefaultMaxZoneCards.
	MaxZoneCards int
	// NoteUnimplemented spells out the unimplemented flag for a reader
	// that has no primer explaining it (the MCP seat, ADR 0122 §5). The
	// bot leaves it off and keeps its terse "(unimplemented)".
	NoteUnimplemented bool
}

func (o Options) maxZoneCards() int {
	if o.MaxZoneCards <= 0 {
		return DefaultMaxZoneCards
	}
	return o.MaxZoneCards
}

// Render writes the board for the seat named by seat (a player ID)
// from view alone. The result ends with a newline, so a caller can
// append its own section after a blank line.
func Render(v *protocol.GameView, seat string, opts Options) string {
	me := seat
	max := opts.maxZoneCards()
	var b strings.Builder

	// Number is the ROUND (every seat has had a turn); Seq counts turns,
	// the figure the game log's turn field carries. Print both, labelled,
	// so a model never has to reconcile "Turn 8" here with a log that
	// counts differently (#2279). The human client's "T4" is the round.
	if v.Turn.Seq > 0 {
		fmt.Fprintf(&b, "TURN %d (round %d) — %s", v.Turn.Seq, v.Turn.Number, StepName(v.Turn.Step))
	} else {
		fmt.Fprintf(&b, "ROUND %d — %s", v.Turn.Number, StepName(v.Turn.Step))
	}
	// ADR 0059 Decision 13: the round number repeats on an extra turn
	// (CR 500.7), so say so, or a model reads "TURN 3" twice as a
	// replay of the same turn.
	if v.Turn.Extra {
		b.WriteString(" (extra turn)")
	}
	if as := seatAt(v, v.Turn.ActiveSeat); as != nil {
		fmt.Fprintf(&b, " — active player: %s", SeatLabel(as, me))
	}
	if ph := seatAt(v, v.Turn.PriorityHolder); ph != nil {
		fmt.Fprintf(&b, " — priority: %s", SeatLabel(ph, me))
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
			fmt.Fprintf(&b, "%s — ELIMINATED\n", SeatLabel(s, me))
			continue
		}
		fmt.Fprintf(&b, "%s — %d life, %d cards in hand, %d in library",
			SeatLabel(s, me), s.Life, s.Hand.Count, s.Library.Count)
		if len(s.ManaPool) > 0 {
			fmt.Fprintf(&b, ", mana pool %s", strings.Join(s.ManaPool, ""))
		}
		// ADR 0129 §7: the seat's player counters ("4 energy, 2
		// poison"), so a model seat and the MCP seat see the energy they
		// can pay and the poison they are racing.
		if pc := PlayerCounters(s.Counters); pc != "" {
			fmt.Fprintf(&b, ", %s", pc)
		}
		// ADR 0136 §8: the seat's speed (CR 702.179), so a model seat
		// sees who is close to switching their max-speed abilities on.
		if sp := SpeedPhrase(s.Speed); sp != "" {
			fmt.Fprintf(&b, ", %s", sp)
		}
		// ADR 0057 Decision 6: a seat behind a "can't lose" or "can't
		// win" gate plays by different arithmetic, and the model is
		// told so on the seat line, with the sources.
		if gates := endGateNote(s); gates != "" {
			fmt.Fprintf(&b, ", %s", gates)
		}
		b.WriteByte('\n')
		// ADR 0114 §7: one line per emblem, the Ring with its count.
		for _, e := range s.Emblems {
			fmt.Fprintf(&b, "  emblem: %s\n", EmblemLine(e))
		}
		if bf := battlefieldOf(v, s.ID, opts); bf != "" {
			fmt.Fprintf(&b, "  battlefield: %s\n", bf)
		}
		if gy := cardNames(s.Graveyard.Cards, max); gy != "" {
			fmt.Fprintf(&b, "  graveyard: %s\n", gy)
		}
		if s.ID == me {
			if hand := describeCards(s.Hand.Cards, max, opts.NoteUnimplemented); hand != "" {
				fmt.Fprintf(&b, "  your hand: %s\n", hand)
			}
			if cmd := describeCards(s.Command.Cards, max, opts.NoteUnimplemented); cmd != "" {
				fmt.Fprintf(&b, "  your command zone: %s\n", cmd)
			}
		}
	}

	if len(v.StackItems) > 0 {
		b.WriteString("\nSTACK (bottom first — the last entry resolves next)\n")
		for i := range v.StackItems {
			it := &v.StackItems[i]
			fmt.Fprintf(&b, "  %s — %s", StackLabel(it, v), StackOwnership(it, v, me))
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
		// ADR 0115 decision 4: the one fact the CR 903.9a question
		// carries beyond its card.
		if ch.Kind == "commander_return" && ch.PlayableFromZone {
			b.WriteString(" (you could cast it from where it is now)")
		}
		// #2390: CR 903.9b's question about a commander headed for a
		// hand, which the Reason does not name.
		if ch.Kind == "optional_replacement" && ch.PlayableFromZone {
			b.WriteString(" (if you say no it goes to your hand, where you can cast it without the commander tax)")
		}
		// ADR 0129 §3: what an energy payment asks, beside the energy
		// the seat line already shows.
		if ch.PayEnergy != nil {
			fmt.Fprintf(&b, " (pay %d energy)", *ch.PayEnergy)
		}
		if pa := ch.PayAmount; pa != nil {
			lo := pa.Min
			if lo < 1 {
				lo = 1
			}
			fmt.Fprintf(&b, " (pay nothing, or %d to %d energy", lo, pa.Max)
			if pa.Goal > 0 {
				fmt.Fprintf(&b, "; %d reaches the card's threshold", pa.Goal)
			}
			fmt.Fprintf(&b, "; one energy is one point of %s)", pa.Unit)
		}
		if ch.Count > 0 && ch.PayAmount == nil {
			fmt.Fprintf(&b, " (choose %d)", ch.Count)
		}
		b.WriteByte('\n')
	}

	return b.String()
}

func battlefieldOf(v *protocol.GameView, seat string, opts Options) string {
	var parts []string
	for i := range v.Battlefield.Cards {
		c := &v.Battlefield.Cards[i]
		if c.Controller != seat {
			continue
		}
		if len(parts) >= opts.maxZoneCards() {
			parts = append(parts, "…")
			break
		}
		parts = append(parts, PermanentLabel(c, opts))
	}
	return strings.Join(parts, ", ")
}

func PermanentLabel(c *protocol.CardView, opts Options) string {
	var b strings.Builder
	b.WriteString(CardName(c))
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
	if c.RingBearer {
		// ADR 0114 §7: whose Ring-bearer it is, is its controller.
		flags = append(flags, "Ring-bearer")
	}
	if c.Unimplemented {
		if opts.NoteUnimplemented {
			flags = append(flags, UnimplementedNote)
		} else {
			flags = append(flags, "unimplemented")
		}
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

// EmblemLine renders one emblem for the board text: its label and its
// current text, and for the Ring how many times it has tempted ("The
// Ring (tempted 3 times): …"). The text's line breaks become spaces.
func EmblemLine(e protocol.EmblemView) string {
	label := e.Label
	if e.Level > 0 {
		times := "times"
		if e.Level == 1 {
			times = "time"
		}
		label = fmt.Sprintf("%s (tempted %d %s)", label, e.Level, times)
	}
	if e.Text == "" {
		return label
	}
	return label + ": " + strings.ReplaceAll(e.Text, "\n", " ")
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

func describeCards(cards []protocol.CardView, max int, noteUnimplemented bool) string {
	var parts []string
	for i := range cards {
		if len(parts) >= max {
			parts = append(parts, "…")
			break
		}
		c := &cards[i]
		s := CardName(c)
		if c.ManaCost != "" {
			s += " " + c.ManaCost
		}
		if c.TypeLine != "" {
			s += " — " + c.TypeLine
		}
		if c.Unimplemented {
			if noteUnimplemented {
				s += " (" + UnimplementedNote + ")"
			} else {
				s += " (unimplemented)"
			}
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
		parts = append(parts, CardName(&cards[i]))
	}
	return strings.Join(parts, ", ")
}

// CardName respects the view's own redaction. A face-down card the
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
func CardName(c *protocol.CardView) string {
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

func StepName(step string) string { return strings.ReplaceAll(step, "_", " ") }

func seatAt(v *protocol.GameView, i int) *protocol.PlayerView {
	if i < 0 || i >= len(v.Seats) {
		return nil
	}
	return &v.Seats[i]
}

func SeatLabel(s *protocol.PlayerView, me string) string {
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

func SeatLabelByID(v *protocol.GameView, id, me string) string {
	for i := range v.Seats {
		if v.Seats[i].ID == id {
			return SeatLabel(&v.Seats[i], me)
		}
	}
	return "someone"
}

// StackOwnership is who a stack item belongs to, as the prompt says it.
// A spell is "cast by X" (CR 601); an activated ability is "activated by
// X" (CR 602) and a triggered one "triggered, controlled by X" (CR 603),
// since nobody casts an ability (#2279). A spell another player took on
// the stack (ADR 0104) is "controlled by X (cast by Y)" — the model has to
// know that it now acts for X, and that Y is who lost it.
func StackOwnership(it *protocol.StackItemView, v *protocol.GameView, me string) string {
	ctl := SeatLabelByID(v, it.Controller, me)
	switch it.Kind {
	case "activated":
		return "activated by " + ctl
	case "triggered":
		return "triggered, controlled by " + ctl
	}
	if it.DefaultController == "" {
		return "cast by " + ctl
	}
	return "controlled by " + ctl +
		" (cast by " + SeatLabelByID(v, it.DefaultController, me) + ")"
}

func StackLabel(it *protocol.StackItemView, v *protocol.GameView) string {
	if it.Label != "" {
		return it.Label
	}
	for i := range v.Stack.Cards {
		if v.Stack.Cards[i].InstanceID == it.SourceCardID {
			return CardName(&v.Stack.Cards[i])
		}
	}
	return it.Kind
}

func targetList(v *protocol.GameView, targets []protocol.TargetRefView, me string) string {
	parts := make([]string, 0, len(targets))
	for _, t := range targets {
		switch t.Kind {
		case "player":
			parts = append(parts, SeatLabelByID(v, t.ID, me))
		default:
			parts = append(parts, nameOfInstance(v, t.ID))
		}
	}
	return strings.Join(parts, ", ")
}

func nameOfInstance(v *protocol.GameView, id string) string {
	for i := range v.Battlefield.Cards {
		if v.Battlefield.Cards[i].InstanceID == id {
			return CardName(&v.Battlefield.Cards[i])
		}
	}
	for i := range v.Stack.Cards {
		if v.Stack.Cards[i].InstanceID == id {
			return CardName(&v.Stack.Cards[i])
		}
	}
	return "something"
}

// endGateNote is a seat's "can't lose the game" / "can't win the game"
// state for the seat line (ADR 0057 Decision 6), or "".
func endGateNote(s *protocol.PlayerView) string {
	var parts []string
	if len(s.CantLose) > 0 {
		parts = append(parts, "CAN'T LOSE the game ("+strings.Join(s.CantLose, ", ")+")")
	}
	if s.CantWin {
		parts = append(parts, "CAN'T WIN the game")
	}
	if len(parts) == 0 {
		return ""
	}
	var names []string
	for _, g := range s.EndGates {
		names = append(names, g.SourceName)
	}
	out := strings.Join(parts, ", ")
	if len(names) > 0 {
		out += " because of " + strings.Join(names, ", ")
	}
	return out
}

// SpeedPhrase is a seat's speed (CR 702.179, ADR 0136) as a phrase:
// "speed 3", "speed 4 (max speed)", or empty for a seat with none.
func SpeedPhrase(speed int) string {
	if speed <= 0 {
		return ""
	}
	if speed >= 4 {
		return "speed 4 (max speed)"
	}
	return "speed " + strconv.Itoa(speed)
}

// PlayerCounters is a seat's non-zero player counters as one phrase,
// in name order: "4 energy, 2 poison". Empty when it has none.
func PlayerCounters(counters map[string]int) string {
	names := make([]string, 0, len(counters))
	for name, n := range counters {
		if n > 0 {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	parts := make([]string, len(names))
	for i, name := range names {
		parts[i] = strconv.Itoa(counters[name]) + " " + name
	}
	return strings.Join(parts, ", ")
}

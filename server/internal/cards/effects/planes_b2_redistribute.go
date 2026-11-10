package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b2_redistribute.go is "redistribute any number of players' life
// totals" (The Doctor's Tomb, #4870): the instruction's controller chooses
// any number of players, then each of those players gets one of the chosen
// players' life totals, one total each (the rulings: a total can't be split).
//
// The choosing is two keyed pick chains, so a table waiting on either is a
// restore point: first one yes/no per player in turn order from the
// controller, then, per chosen player in that order, which of the totals still
// unassigned they get (the last one is forced and not asked). Every total is
// read when the second chain starts, before any is written, and each player
// then gains or loses the difference (the exchange's own rule, CR 119.5 and
// effects_life_exchange.go), so every life payoff at the table sees it.

// planesB2RedistributeLife starts the instruction for the effect's controller.
type planesB2RedistributeLife struct{}

func (planesB2RedistributeLife) Apply(ctx *Context) error {
	players := apnapPlayersFrom(ctx.Game, ctx.Controller())
	return planesB2AskInclude(ctx, nil, players)
}

// apnapPlayersFrom lists the live players in turn order starting with `first`.
func apnapPlayersFrom(g *game.Game, first uuid.UUID) []uuid.UUID {
	var seats []uuid.UUID
	at := -1
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if p.ID == first {
			at = len(seats)
		}
		seats = append(seats, p.ID)
	}
	if at <= 0 {
		return seats
	}
	return append(append([]uuid.UUID(nil), seats[at:]...), seats[:at]...)
}

// planesB2Carry packs two lists into one carry, divided by a nil ID.
func planesB2Carry(a, b []uuid.UUID) []uuid.UUID {
	out := append([]uuid.UUID(nil), a...)
	out = append(out, uuid.Nil)
	return append(out, b...)
}

// planesB2Unpack is planesB2Carry's inverse.
func planesB2Unpack(carry []uuid.UUID) (a, b []uuid.UUID) {
	for i, id := range carry {
		if id == uuid.Nil {
			return append([]uuid.UUID(nil), carry[:i]...), append([]uuid.UUID(nil), carry[i+1:]...)
		}
	}
	return append([]uuid.UUID(nil), carry...), nil
}

func planesB2PlayerName(g *game.Game, id uuid.UUID) string {
	if p := g.PlayerByIDForEffect(id); p != nil {
		return p.Name
	}
	return "a player"
}

// planesB2AskInclude asks about the first of `rest`, or, with none left, starts
// the assignment of the chosen players' totals.
func planesB2AskInclude(ctx *Context, chosen, rest []uuid.UUID) error {
	for len(rest) > 0 {
		p := ctx.Game.PlayerByIDForEffect(rest[0])
		if p == nil {
			rest = rest[1:]
			continue
		}
		name := p.Name
		return PickOption{
			Question: fmt.Sprintf("Redistribute life totals — include %s (%d life)?", name, p.Life),
			Options: []game.ChoiceOption{
				{Label: fmt.Sprintf("Leave %s's life total as it is", name)},
				{Label: fmt.Sprintf("Redistribute %s's life total (%d)", name, p.Life)},
			},
			ThenKey:  planesB2RedistributeIncluded,
			Carry:    planesB2Carry(chosen, rest),
			WithItem: true,
		}.Apply(ctx)
	}
	return planesB2AskAssign(ctx, chosen, nil, chosen, nil)
}

// planesB2RedistributeIncluded is one include answer. An on-disk key: never
// renamed, never reused.
var planesB2RedistributeIncluded game.OptionPickThen

// planesB2RedistributeAssigned is one assignment answer. An on-disk key.
var planesB2RedistributeAssigned game.OptionPickThen

func init() {
	planesB2RedistributeIncluded = OptionPickThen("option-pick/redistribute-life-include", func(ctx *Context, r game.OptionPicked) error {
		chosen, rest := planesB2Unpack(r.Carry)
		if len(rest) == 0 {
			return nil
		}
		if r.Option != nil && r.Index == 1 {
			chosen = append(chosen, rest[0])
		}
		return planesB2AskInclude(ctx, chosen, rest[1:])
	})
	planesB2RedistributeAssigned = OptionPickThen("option-pick/redistribute-life-assign", func(ctx *Context, r game.OptionPicked) error {
		// Carry: [receivers left | totals left | pairs so far], two nil
		// dividers; a pair is receiver, donor.
		lists := planesB2Split(r.Carry, 3)
		receivers, donors, pairs := lists[0], lists[1], lists[2]
		if len(receivers) == 0 || len(donors) == 0 || r.Option == nil || r.Index < 0 || r.Index >= len(donors) {
			return nil
		}
		pairs = append(pairs, receivers[0], donors[r.Index])
		donors = append(donors[:r.Index:r.Index], donors[r.Index+1:]...)
		return planesB2AskAssign(ctx, receivers[1:], donors, nil, pairs)
	})
}

// planesB2Split divides a carry into n lists at nil IDs.
func planesB2Split(carry []uuid.UUID, n int) [][]uuid.UUID {
	out := make([][]uuid.UUID, 0, n)
	cur := []uuid.UUID{}
	for _, id := range carry {
		if id == uuid.Nil && len(out) < n-1 {
			out = append(out, cur)
			cur = []uuid.UUID{}
			continue
		}
		cur = append(cur, id)
	}
	out = append(out, cur)
	for len(out) < n {
		out = append(out, nil)
	}
	return out
}

// planesB2AskAssign asks which unassigned total the next receiver gets. When
// `donors` is nil on entry it starts from `start`, the chosen players, whose
// totals are all to be handed out.
func planesB2AskAssign(ctx *Context, receivers, donors, start, pairs []uuid.UUID) error {
	if donors == nil {
		donors = append([]uuid.UUID(nil), start...)
	}
	// With one total left, the last receiver takes it without a question.
	for len(receivers) == 1 && len(donors) == 1 {
		pairs = append(pairs, receivers[0], donors[0])
		receivers, donors = nil, nil
	}
	if len(receivers) == 0 {
		return planesB2ApplyRedistribution(ctx, pairs)
	}
	opts := make([]game.ChoiceOption, len(donors))
	for i, d := range donors {
		life := 0
		if p := ctx.Game.PlayerByIDForEffect(d); p != nil {
			life = p.Life
		}
		opts[i] = game.ChoiceOption{Label: fmt.Sprintf("%s's life total (%d)", planesB2PlayerName(ctx.Game, d), life)}
	}
	carry := append(planesB2Carry(receivers, donors), uuid.Nil)
	carry = append(carry, pairs...)
	return PickOption{
		Question: fmt.Sprintf("Redistribute life totals — which life total does %s get?", planesB2PlayerName(ctx.Game, receivers[0])),
		Options:  opts,
		ThenKey:  planesB2RedistributeAssigned,
		Carry:    carry,
		WithItem: true,
	}.Apply(ctx)
}

// planesB2ApplyRedistribution gives each receiver the donor's total as it was
// when the assignments were made: every total is read first, then each player
// gains or loses the difference.
func planesB2ApplyRedistribution(ctx *Context, pairs []uuid.UUID) error {
	totals := map[uuid.UUID]int{}
	for i := 0; i+1 < len(pairs); i += 2 {
		if p := ctx.Game.PlayerByIDForEffect(pairs[i+1]); p != nil {
			totals[pairs[i+1]] = p.Life
		}
	}
	for i := 0; i+1 < len(pairs); i += 2 {
		receiver, donor := pairs[i], pairs[i+1]
		want, ok := totals[donor]
		p := ctx.Game.PlayerByIDForEffect(receiver)
		if !ok || p == nil || p.Life == want || receiver == donor {
			continue
		}
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), receiver, want-p.Life); err != nil {
			return err
		}
	}
	return nil
}

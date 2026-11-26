package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b456_helpers.go — shared seams of Commander 2011, batch 456 (#5574):
// join forces and Malfegor's per-card edict.

// S82OldB456JoinForces is the join forces process (Alliance of Arms, Shared
// Trauma, Mana-Charged Dragon): "starting with you, each player may pay any
// amount of mana", then something happens with X, the total paid. Each living
// seat is asked in turn order starting with the controller (MayPayX over {1},
// each paying from their own mana, CR 608.2d); the seats still to ask and the
// running total ride the prompt as data: the seats first, one nil marker, then
// one placeholder per mana paid. When the last seat has answered, Finish runs
// with the total (zero when nobody paid).
type S82OldB456JoinForces struct {
	label, buys string
	finish      func(ctx *Context, total int) error
	then        game.PayUnlessElse
}

// NewS82OldB456JoinForces builds one card's join forces. key is the on-disk
// identity of its payment step (never renamed, never reused); label names the
// card in the prompts and buys says what X buys ("each player creates X 1/1
// white Soldier creature tokens"). Call it once, from a package-level var.
func NewS82OldB456JoinForces(key, label, buys string, finish func(ctx *Context, total int) error) *S82OldB456JoinForces {
	j := &S82OldB456JoinForces{label: label, buys: buys, finish: finish}
	j.then = PayUnlessBranches(key,
		func(ctx *Context, r game.PayUnlessAnswer) error { return j.next(ctx, r, r.Number) },
		func(ctx *Context, r game.PayUnlessAnswer) error { return j.next(ctx, r, 0) })
	return j
}

// Start begins the process for the resolving effect's controller.
func (j *S82OldB456JoinForces) Start(ctx *Context) error {
	return j.ask(ctx, SeatsFromController(ctx.Game, ctx.Controller()), 0)
}

// ask asks seats[0] how much to pay, then carries on with the rest and the new
// total. With nobody left to ask, the finishing step runs.
func (j *S82OldB456JoinForces) ask(ctx *Context, seats []uuid.UUID, total int) error {
	if len(seats) == 0 {
		return j.finish(ctx, total)
	}
	carry := append([]uuid.UUID(nil), seats[1:]...)
	carry = append(carry, uuid.Nil) // the marker between the seats and the total
	for i := 0; i < total; i++ {
		carry = append(carry, uuid.Nil)
	}
	return MayPayX{
		Chooser: seats[0],
		Symbol:  "{1}",
		Label:   j.label,
		Buys:    j.buys,
		Unit:    game.PayAmountOther,
		Then:    j.then,
		Carry:   carry,
	}.Apply(ctx)
}

// next reads the seats still to ask and the total so far off the carry and asks
// the next seat with `paid` added.
func (j *S82OldB456JoinForces) next(ctx *Context, r game.PayUnlessAnswer, paid int) error {
	var seats []uuid.UUID
	total := 0
	pastMarker := false
	for _, id := range r.Carry {
		switch {
		case id != uuid.Nil && !pastMarker:
			seats = append(seats, id)
		case id == uuid.Nil && !pastMarker:
			pastMarker = true
		default:
			total++
		}
	}
	return j.ask(ctx, seats, total+max(paid, 0))
}

// S82OldB456EachOpponentSacrificesN is "each opponent sacrifices N creatures of
// their choice" (Malfegor, N the cards discarded): every living opponent in
// APNAP order makes ONE choice of N creatures and each choice is sacrificed
// together (CR 701.21a). An opponent with fewer than N creatures sacrifices
// what they have; one with none is not asked. The choice is not a target.
func S82OldB456EachOpponentSacrificesN(ctx *Context, n int, question string) error {
	if n <= 0 {
		return nil
	}
	g := ctx.Game
	var legs []game.ChooseCardsPrompt
	spec := SacrificeSpec("a creature", Creature())
	for _, seat := range ApnapLivingSeats(g) {
		if seat == ctx.Controller() || !g.CanEffectCauseSacrificeForEffect(seat) {
			continue
		}
		candidates := g.EffectSacrificeCandidatesForEffect(seat, spec)
		if len(candidates) == 0 {
			continue
		}
		k := min(n, len(candidates))
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser: seat, FromPlayer: seat, Source: ctx.Source(),
			Question: question,
			Cards:    candidates, Min: k, Max: k, Zone: game.ZoneBattlefield,
			Destination: game.ChooseSacrifice,
		})
	}
	return g.ChoosePermanentsInOrderSacrificeForEffect(ctx.Source(), legs)
}

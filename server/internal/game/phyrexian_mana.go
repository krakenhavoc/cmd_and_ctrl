package game

import (
	"fmt"

	"github.com/google/uuid"
)

// phyrexian_mana.go — the life half of a Phyrexian mana symbol
// (CR 107.4, #787).
//
// A Phyrexian symbol can be paid with one mana of its colour, or with
// 2 life (CR 107.4c). Since CR 107.4 the family also has ten HYBRID
// Phyrexian symbols — {W/U/P} and the rest of the wheel — which can
// be paid with one mana of EITHER of two colours, or with 2 life
// (CR 107.4f). ParseCost reads every one of them into the same
// ColorRequirement: a set of colour options plus the Phyrexian flag
// (see mana_cost.go). There is no third symbol kind, so there is no
// third payment path either — what follows is the one life path, and
// it reads the flag, never a colour count.
//
// # Who chooses
//
// CR 601.2b: the caster announces, as part of announcing the spell,
// how they intend to pay each hybrid and Phyrexian symbol. That is an
// announce-time PARAMETER — CastSpellParams.PhyrexianLife, the number
// of the cost's Phyrexian symbols being paid with life — for exactly
// the reasons CastSpellParams.Face is one: every other announce
// decision rides the cast_spell action, and the PendingChoice
// machinery has no frame for a half-validated cast.
//
// CR 602.2b asks the same question of an ACTIVATED ability, in the
// same indivisible announcement, so the answer has the same shape:
// ActivateAbilityParams.PhyrexianLife, wire `phyrexian_life` — one
// field name for both paths (#917). What follows is ONE strike-and-pay
// helper with two callers: strikePhyrexianLifeLocked reduces the cost
// and prices the claim, payPhyrexianLifeLocked hands the life to
// PayLifeForEffect. applyCastCostLocked and payAbilityManaCostLocked
// call them in that order, with the pool spend between, and nothing
// about a Phyrexian symbol is decided anywhere else.
//
// The engine, not the caster, decides WHICH symbols those are, because
// the count is the only part a player cares about: the symbols a
// life payment can save go first (see PhyrexianLifePlan). Choosing
// between two symbols the pool can both pay changes nothing but which
// colour of mana is left floating.
//
// # What it costs
//
// 2 life per symbol, and it is paid through PayLifeForEffect — the
// one cost-shaped life path (#806): a real life LOSS that runs the
// CR 614 window and cannot pause on a CR 616 prompt, because CR
// 601.2h pays a spell's costs as one indivisible step, and CR 602.2b
// says the same of an activation. Nothing here touches a life total
// directly.

// PhyrexianLifePerSymbol is CR 107.4c's price for one Phyrexian
// symbol, hybrid Phyrexian included (CR 107.4f).
const PhyrexianLifePerSymbol = 2

// PhyrexianSymbols counts the symbols in the cost that carry the
// Phyrexian "or 2 life" option — {W/P} and {W/U/P} alike. It is the
// ceiling on CastSpellParams.PhyrexianLife and on
// ActivateAbilityParams.PhyrexianLife, and the count the view ships
// to the client as `phyrexian_symbols` so no client parses a mana
// string to find out (#916).
func (c ParsedCost) PhyrexianSymbols() int {
	n := 0
	for _, req := range c.Required {
		if req.Phyrexian {
			n++
		}
	}
	return n
}

// PhyrexianLifePlan returns `cost` with `n` of its Phyrexian symbols
// struck out — the ones the caster announced they are paying with
// life — and the life that costs (CR 107.4f: 2 each).
//
// Which n: the symbols the pool holds no mana for go FIRST, in
// printed order, then the rest in printed order. A caster who says
// "one symbol, by life" therefore never has the engine spend the life
// on a pip they could have paid and leaves one they cannot — and a
// caster with a full pool who wants the free cast anyway (Gitaxian
// Probe with an Island untapped) still gets it, because the second
// pass takes a payable symbol once the first has nothing to offer.
//
// Pure: `cost` is not mutated, and neither is the pool, and it takes
// no lock — which is why the read-only auto-tap preview can call it
// too (#916). An n outside [0, PhyrexianSymbols()] is the caller's
// error, not this function's — strikePhyrexianLifeLocked rejects it
// before anything is paid.
func PhyrexianLifePlan(cost ParsedCost, pool ManaPool, ctx ManaSpendContext, n int) (ParsedCost, int) {
	if n <= 0 {
		return cost, 0
	}
	spendable := spendOrder(pool, ctx)
	// Rank the Phyrexian symbols: pass 0 is the ones no token in the
	// pool could pay, pass 1 is the rest.
	struck := make(map[int]bool, n)
	for pass := 0; pass < 2 && len(struck) < n; pass++ {
		for i, req := range cost.Required {
			if len(struck) == n {
				break
			}
			if !req.Phyrexian || struck[i] {
				continue
			}
			payable := false
			for _, tok := range spendable {
				if matchColor(pool[tok].Color, req.Options) {
					payable = true
					break
				}
			}
			if payable == (pass == 0) {
				continue
			}
			struck[i] = true
		}
	}
	out := cost
	out.Required = make([]ColorRequirement, 0, len(cost.Required))
	for i, req := range cost.Required {
		if struck[i] {
			continue
		}
		out.Required = append(out.Required, req)
	}
	return out, len(struck) * PhyrexianLifePerSymbol
}

// strikePhyrexianLifeLocked is the announce half of the one
// strike-and-pay helper, and it has exactly two callers: a cast
// (CR 601.2b, applyCastCostLocked) and an activation (CR 602.2b,
// payAbilityManaCostLocked). The announcer may only claim a life
// payment for a symbol the cost actually prints, and only down to a
// life total of 0 (CR 119.4).
//
// `sourceName` is the card the cost is printed on — the spell being
// cast, or the permanent whose ability is being activated — and is
// read for the error message only. `spendCtx` is the context the
// remaining mana will be spent under, because which symbols the
// strike saves depends on what the pool can legally pay for THIS
// announcement (#352).
//
// Returns the reduced cost and the life owed, and adds that life to
// `paid` (ADR 0020's #958 addendum: PaidCost.LifePaid is every point
// this announcement paid, so an ability that ALSO prints a life
// component sums the two). `paid` may be nil for a caller that is
// only pricing — the auto-tap planner, which pays nothing.
//
// All of it is zero-work when the announcer claimed nothing, which is
// every cast and every activation of everything without a Phyrexian
// symbol in its cost.
//
// Caller must hold g.mu.
func (g *Game) strikePhyrexianLifeLocked(p *Player, sourceName string, cost ParsedCost, spendCtx ManaSpendContext, claimed int, paid *PaidCost) (ParsedCost, int, error) {
	if claimed == 0 {
		return cost, 0, nil
	}
	if claimed < 0 {
		return cost, 0, fmt.Errorf("%w: phyrexian_life must not be negative", ErrInvalidParam)
	}
	if have := cost.PhyrexianSymbols(); claimed > have {
		return cost, 0, fmt.Errorf("%w: phyrexian_life %d, but %s prints %d Phyrexian symbol(s)", ErrInvalidParam, claimed, sourceName, have)
	}
	reduced, life := PhyrexianLifePlan(cost, p.ManaPool, spendCtx, claimed)
	// CR 119.4: a player may pay life only down to 0. Checked before
	// anything is paid, so an over-claim rejects the announcement
	// without costing the announcer a single point. CR 119.8 rides
	// the same predicate (#1200): a player whose life total can't
	// change may not claim a Phyrexian symbol at all.
	if !g.CanPayLifeLocked(p, life) {
		return cost, 0, fmt.Errorf("%w: %s cannot pay the %d life for %d Phyrexian symbol(s) (CR 119.4, CR 119.8)", ErrInvalidParam, p.Name, life, claimed)
	}
	if paid != nil {
		paid.LifePaid += life
	}
	return reduced, life, nil
}

// payPhyrexianLifeLocked hands the announced life to the one
// cost-shaped life path (#806, ADR 0013 §5b): a real loss that runs
// the CR 614 window and settles it without a prompt, because CR
// 601.2h pays a spell's costs as one indivisible step. Nothing here
// writes a life total.
//
// The pay half of the one strike-and-pay helper, with the same two
// callers strikePhyrexianLifeLocked has. Both run it after the mana
// half is known to be payable and before the pool is spent, because
// the fallible payment goes first.
//
// A no-op for the zero the overwhelming majority of announcements
// make. Caller must hold g.mu.
func (g *Game) payPhyrexianLifeLocked(cardID, playerID uuid.UUID, life int) error {
	if life <= 0 {
		return nil
	}
	return g.PayLifeForEffect(cardID, playerID, life)
}

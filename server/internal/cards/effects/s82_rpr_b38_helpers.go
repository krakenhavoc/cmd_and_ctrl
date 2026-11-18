package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// plagueOfVerminOrder is "starting with you" (Plague of Vermin): the
// spell's controller, then the living opponents in turn order.
func plagueOfVerminOrder(g *game.Game, controller uuid.UUID) []uuid.UUID {
	return append([]uuid.UUID{controller}, opponentsInTurnOrder(g, controller)...)
}

// plagueOfVerminAsk asks the player at `pos` of `order` how much life they
// pay. nums is [position, life paid in this pass, then the total paid by
// each player in `order`]: plain data, so a table waiting on the question
// is a restore point (#4294).
func plagueOfVerminAsk(ctx *Context, order []uuid.UUID, nums []int) error {
	return PayLifeAmount{
		Chooser:  order[nums[0]],
		Question: "Plague of Vermin — pay any amount of life (each 1 life paid makes you a 1/1 Rat)",
		Unit:     game.PayAmountOther,
		Goal:     func(*Context) int { return 0 },
		Then:     plagueOfVerminPaid,
		Carry:    order,
		Nums:     nums,
	}.Apply(ctx)
}

// plagueOfVerminPaid records one player's payment and moves on. After the
// last player of a pass, the process repeats if anyone paid in it, and
// otherwise every player creates a Rat for each 1 life they paid.
var plagueOfVerminPaid game.NumberThen

// Registered in init: the branch asks the next question, so a package-level
// initialiser would be a cycle.
func init() {
	plagueOfVerminPaid = NumberThen("number/plague-of-vermin-paid", plagueOfVerminPaidBody)
}

func plagueOfVerminPaidBody(ctx *Context, r game.NumberAnswer) error {
	order := r.Carry
	nums := append([]int(nil), r.Nums...)
	pos := nums[0]
	nums[1] += r.Number
	nums[2+pos] += r.Number
	pos++
	if pos < len(order) {
		nums[0] = pos
		return plagueOfVerminAsk(ctx, order, nums)
	}
	if nums[1] > 0 {
		nums[0], nums[1] = 0, 0
		return plagueOfVerminAsk(ctx, order, nums)
	}
	for i, p := range order {
		if n := nums[2+i]; n > 0 {
			if err := (CreateToken{Controller: p, Template: TokenCard("1/1 black Rat"), N: n}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

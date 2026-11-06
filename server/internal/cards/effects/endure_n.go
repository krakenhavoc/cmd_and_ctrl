package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// endure_n.go — the keyword action endure (CR 701.63), Tarkir:
// Dragonstorm (#4463):
//
//	"Anafenza endures 2. (Put two +1/+1 counters on it or create a 2/2
//	 white Spirit creature token.)"
//
// CR 701.63a: "that permanent's controller creates an N/N white Spirit
// creature token unless they put N +1/+1 counters on that permanent."
// The token is the default and the counters are the alternative, so the
// option pick lists the token FIRST (the option that can always be
// taken, PickOption's contract). A permanent that has left the
// battlefield, or come back as a new object (CR 400.7), can't have
// counters put on it, so the token is created with no question. Endure 0
// does nothing (CR 701.63b).
//
// Not the card named Endure (endure.go). Append-only: a new endure card
// calls EndureN.

// EndureN is "<permanent> endures n", asked of the resolving effect's
// controller. `name` heads the prompt.
func EndureN(ctx *Context, permanent uuid.UUID, n int, name string) error {
	if n <= 0 {
		return nil
	}
	if !endureCanTakeCounters(ctx, permanent) {
		return endureSpiritToken(ctx, n)
	}
	count := strconv.Itoa(n)
	counters := "Put " + count + " +1/+1 counters on it"
	if n == 1 {
		counters = "Put a +1/+1 counter on it"
	}
	return PickOption{
		Question: name + " — endure " + count + ": put " + count + " +1/+1 counters on it, or create a " +
			count + "/" + count + " white Spirit creature token?",
		Options: []game.ChoiceOption{
			{Label: "Create a " + count + "/" + count + " white Spirit creature token"},
			{Label: counters},
		},
		ThenKey:  endureAnswered,
		Carry:    []uuid.UUID{permanent},
		Nums:     []int{n},
		WithItem: true,
	}.Apply(ctx)
}

// ThisEndures is a trigger body "<this> endures n": the resolving
// ability's own source endures (Sandskitter Outrider's enters trigger,
// Sinkhole Surveyor's attack trigger). `name` heads the prompt.
//
//	WhenThisEnters("Sandskitter Outrider — it endures 2", ThisEndures("Sandskitter Outrider", 2)),
func ThisEndures(name string, n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return EndureN(NewContext(g, item), item.SourceCardID, n, name)
	}
}

// endureCanTakeCounters is whether `permanent` is still the object the
// effect names and on the battlefield.
func endureCanTakeCounters(ctx *Context, permanent uuid.UUID) bool {
	return onBattlefield(ctx.Game, permanent) && !ctx.isNewSourceObject(permanent)
}

// endureAnswered is the choice, keyed: Carry is the permanent and Nums the
// N. Index 1 is the counters; anything else (the token, or nobody chose)
// is the token. An on-disk key: never renamed, never reused.
var endureAnswered = OptionPickThen("option-pick/endure-counters-or-spirit", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Nums) == 0 || len(r.Carry) == 0 {
		return nil
	}
	n, permanent := r.Nums[0], r.Carry[0]
	if r.Index == 1 && endureCanTakeCounters(ctx, permanent) {
		return AddCounter{Target: permanent, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
	}
	return endureSpiritToken(ctx, n)
})

// endureSpiritToken creates the N/N white Spirit creature token.
func endureSpiritToken(ctx *Context, n int) error {
	return CreateToken{Template: game.Card{
		Name:      "Spirit",
		TypeLine:  "Token Creature — Spirit",
		Colors:    []string{"W"},
		Power:     n,
		Toughness: n,
	}, N: 1}.Apply(ctx)
}

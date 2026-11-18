package effects

import (
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b12_helpers.go — the Adventures in the Forgotten Realms
// "Endeavor" sorceries and instants (Valiant, Grave, Wild, Arcane): "Roll
// two dN and choose one result. <first clause using that result>. Then
// <second clause using the other result>."
//
// afrEndeavorTail registers the rest of one such card. After the dice (a
// replaced roll waits for its chooser first, #4307) the controller
// chooses which result is "that result": a yes is the first roll as the
// chosen result and a no the second. Equal rolls are no choice, so
// nothing is asked. `apply` is told the chosen result and the other.
//
// key is the card's on-disk identity prefix ("dice/<card>"): the roll's
// tail is key+"-rolled" and the answer is "confirm/"+the key's last
// segment+"-result"; never renamed, never reused. Call it once, from a
// package-level var.
func afrEndeavorTail(key, name string, describe func(chosen, other int) string, apply func(ctx *Context, chosen, other int) error) game.TailRef {
	answer := MayAnswer("confirm/"+strings.TrimPrefix(key, "dice/")+"-result", func(ctx *Context, r game.ConfirmAnswer) error {
		first, second, ok := strings.Cut(r.Word, " ")
		a, errA := strconv.Atoi(first)
		b, errB := strconv.Atoi(second)
		if !ok || errA != nil || errB != nil {
			return nil
		}
		if !r.Accepted {
			a, b = b, a
		}
		return apply(ctx, a, b)
	})
	return DiceTail(key+"-rolled", func(ctx *Context, rolls []int) error {
		if len(rolls) < 2 {
			return nil
		}
		if rolls[0] == rolls[1] {
			return apply(ctx, rolls[0], rolls[1])
		}
		return MayChoice{
			Question: name + " — choose one result",
			YesLabel: describe(rolls[0], rolls[1]),
			NoLabel:  describe(rolls[1], rolls[0]),
			Then:     answer,
			Word:     strconv.Itoa(rolls[0]) + " " + strconv.Itoa(rolls[1]),
		}.Apply(ctx)
	})
}

// afrWheneverYouRollDiceDo is WheneverYouRollDice for a one-sentence effect
// that does not read the results, written as effects.Do primitives so the
// row declares its footprint (Brazen Dwarf's "deals 1 damage to each
// opponent"). The trigger is the same: once per instruction to roll.
func afrWheneverYouRollDiceDo(label string, effect Effect) game.TriggeredAbility {
	t := WheneverYouRollDice(label, func(*game.Game, *game.StackItem, int) error { return nil })
	t.Effect = effect
	return t
}

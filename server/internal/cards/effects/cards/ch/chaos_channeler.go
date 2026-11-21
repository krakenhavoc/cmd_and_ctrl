package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chaos Channeler — Creature — Human Shaman {2}{R}{R}, 4/3 (Adventures in
// the Forgotten Realms, S82):
//
//	"Wild Magic Surge — Whenever this creature attacks, roll a d20.
//	 1—9 | Exile the top card of your library. You may play it this turn.
//	 10—19 | Exile the top two cards of your library. You may play them
//	 this turn.
//	 20 | Exile the top three cards of your library. You may play them
//	 this turn."
//
// The d20 goes through the one dice door; the results table is the tail,
// which exiles that many cards with the usual until-end-of-turn play
// permission (ExileTopWithPermission). "Wild Magic Surge" is a flavor word.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea9031de-6169-4b1c-9390-7fdaf876c7cd",
		Name:         "Chaos Channeler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Chaos Channeler — roll a d20, exile that many cards from the top of your library", effects.Do(effects.RollDice{Sides: 20, N: 1, Then: chaosChannelerExile})),
		},
	})
}

// chaosChannelerExile is the results table, after the d20. An on-disk key.
var chaosChannelerExile = effects.DiceTail("dice/chaos-channeler-exile", func(ctx *effects.Context, rolls []int) error {
	n := 1
	if rolls[0] >= 10 {
		n = 2
	}
	if rolls[0] == 20 {
		n = 3
	}
	me := ctx.Controller()
	return effects.ExileTopWithPermission{From: me, GrantTo: me, N: n}.Apply(ctx)
})

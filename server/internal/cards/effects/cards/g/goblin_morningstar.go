package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Morningstar — Artifact — Equipment {1}{R} (Adventures in the
// Forgotten Realms, #5499):
//
//	"Equipped creature gets +1/+0 and has trample.
//	 Equip {1}{R}
//	 When this Equipment enters, roll a d20.
//	 1—9 | Create a 1/1 red Goblin creature token.
//	 10—20 | Create a 1/1 red Goblin creature token, then attach this
//	 Equipment to it."
//
// The attach rides the token's creation (createTokenAndAttachSource); if
// the Equipment has left, the Goblin is made and nothing is attached
// (CR 701.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61ff247e-a119-4f35-af62-2645d6762bba",
		Name:         "Goblin Morningstar",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 0), effects.GrantToAttached("trample")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Goblin Morningstar — roll a d20", effects.Do(effects.RollDice{Sides: 20, N: 1, Then: goblinMorningstarTable})),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}{R}")},
	})
}

// goblinMorningstarTable is the results table, after the d20.
var goblinMorningstarTable = effects.DiceTail("dice/goblin-morningstar-table", func(ctx *effects.Context, rolls []int) error {
	if rolls[0] >= 10 {
		return effects.CreateTokenAndAttachSource(ctx.Game, ctx.Item, "1/1 red Goblin", true)
	}
	return effects.CreateToken{Template: effects.RedGoblinToken(), N: 1}.Apply(ctx)
})

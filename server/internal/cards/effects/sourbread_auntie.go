package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sourbread Auntie — Creature — Goblin Warrior {2}{R}{R}, 4/3:
//
//	"When this creature enters, you may blight 2. If you do, create two
//	 1/1 black and red Goblin creature tokens. (To blight 2, put two -1/-1
//	 counters on a creature you control.)"
//
// The blight is asked as the trigger resolves (Blight, #4683, CR 701.68a):
// any creature you control, the Auntie herself included, hexproof or not.
// The counters are an effect's, so Doubling Season doubles them. "If you
// do" is a creature chosen; the tokens follow once the counters have
// landed. With no creature that can take the counters there is nothing to
// choose and no tokens (CR 701.68b).
func init() {
	Register(Spec{
		OracleID:     "6b087ae1-1412-4dd9-9b96-3b45be52d386",
		Name:         "Sourbread Auntie",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sourbread Auntie — you may blight 2; if you do, create two Goblin tokens", func(g *game.Game, item *game.StackItem) error {
				return Blight{N: 2, May: true, Then: sourbreadAuntieTokens}.Apply(NewContext(g, item))
			}),
		},
	})
}

// sourbreadAuntieTokens is "if you do, create two 1/1 black and red
// Goblin creature tokens". An on-disk key.
var sourbreadAuntieTokens = RegisterBlighted("blight/sourbread-auntie-tokens", func(ctx *Context, d BlightDone) error {
	if !d.Did() {
		return nil
	}
	return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 black and red Goblin"), N: 2}.Apply(ctx)
})

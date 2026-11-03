package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scuzzback Scrounger — Creature — Goblin Warrior {1}{R}, 3/2:
//
//	"At the beginning of your first main phase, you may blight 1. If you
//	 do, create a Treasure token. (To blight 1, put a -1/-1 counter on a
//	 creature you control. A Treasure token is an artifact with "{T},
//	 Sacrifice this token: Add one mana of any color.")"
//
// "Your first main phase" is the precombat main (CR 505.1a). The blight
// is asked as the trigger resolves (Blight, #4683), and the Treasure
// follows a creature chosen.
func init() {
	Register(Spec{
		OracleID:     "af38d683-77a3-4599-ad2a-1c293ae01d45",
		Name:         "Scuzzback Scrounger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Scuzzback Scrounger — you may blight 1; if you do, create a Treasure", func(g *game.Game, item *game.StackItem) error {
				return Blight{N: 1, May: true, Then: scuzzbackScroungerTreasure}.Apply(NewContext(g, item))
			}),
		},
	})
}

// scuzzbackScroungerTreasure is "if you do, create a Treasure token". An
// on-disk key.
var scuzzbackScroungerTreasure = RegisterBlighted("blight/scuzzback-scrounger-treasure", func(ctx *Context, d BlightDone) error {
	if !d.Did() {
		return nil
	}
	return CreateToken{Controller: ctx.Controller(), Template: TreasureToken(), N: 1}.Apply(ctx)
})

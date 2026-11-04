package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boggart Mischief — Kindred Enchantment — Goblin {2}{B}:
//
//	"When this enchantment enters, you may blight 1. If you do, create two
//	 1/1 black and red Goblin creature tokens. (To blight 1, put a -1/-1
//	 counter on a creature you control.)
//	 Whenever a Goblin creature you control dies, each opponent loses 1
//	 life and you gain 1 life."
//
// The blight is asked as the entry trigger resolves (Blight, #4683), and
// the two Goblins follow a creature chosen. The drain watches a Goblin
// that died under your control, read from what it was as it died (a
// token blighted to death is one). "Goblin" is a word of this card's
// text (CR 612.2).
func init() {
	Register(Spec{
		OracleID:     "ac65ed8c-abe8-4044-9ce5-70b1370d0e19",
		Name:         "Boggart Mischief",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Boggart Mischief — you may blight 1; if you do, create two Goblin tokens", func(g *game.Game, item *game.StackItem) error {
				return Blight{N: 1, May: true, Then: boggartMischiefTokens}.Apply(NewContext(g, item))
			}),
			On(game.EventLTB, aGoblinCreatureYouControlDied,
				"Boggart Mischief — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
	})
}

// boggartMischiefTokens is "if you do, create two 1/1 black and red Goblin
// creature tokens". An on-disk key.
var boggartMischiefTokens = RegisterBlighted("blight/boggart-mischief-tokens", func(ctx *Context, d BlightDone) error {
	if !d.Did() {
		return nil
	}
	return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 black and red Goblin"), N: 2}.Apply(ctx)
})

// aGoblinCreatureYouControlDied is "whenever a Goblin creature you control
// dies".
func aGoblinCreatureYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && leftUnderControlOf(ev, dead) == source.Controller && leftAsSubtype(ev, dead, source.TextSubtype("Goblin"))
}

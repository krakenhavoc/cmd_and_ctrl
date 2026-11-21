package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ac65ed8c-abe8-4044-9ce5-70b1370d0e19",
		Name:         "Boggart Mischief",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Boggart Mischief — you may blight 1; if you do, create two Goblin tokens", func(g *game.Game, item *game.StackItem) error {
				return effects.Blight{N: 1, May: true, Then: boggartMischiefTokens}.Apply(effects.NewContext(g, item))
			}),
			effects.On(game.EventLTB, aGoblinCreatureYouControlDied,
				"Boggart Mischief — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}

// boggartMischiefTokens is "if you do, create two 1/1 black and red Goblin
// creature tokens". An on-disk key.
var boggartMischiefTokens = effects.RegisterBlighted("blight/boggart-mischief-tokens", func(ctx *effects.Context, d effects.BlightDone) error {
	if !d.Did() {
		return nil
	}
	return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 black and red Goblin"), N: 2}.Apply(ctx)
})

// aGoblinCreatureYouControlDied is "whenever a Goblin creature you control
// dies".
func aGoblinCreatureYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := effects.DiedCreature(ev, g)
	return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && effects.LeftAsSubtype(ev, dead, source.TextSubtype("Goblin"))
}

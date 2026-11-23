package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rammas Echor, Ancient Shield — Legendary Artifact {3}{W} (Tales of
// Middle-earth Commander, #5948):
//
//	"Whenever you cast your second spell each turn, draw a card, then create
//	 a 0/3 white Wall creature token with defender.
//	 At the beginning of combat on your turn, creatures you control with
//	 defender gain exalted until end of turn. (Whenever a creature you control
//	 attacks alone, it gets +1/+1 until end of turn for each instance of
//	 exalted among permanents you control.)"
//
// "Your second spell each turn" is Avatar Yangchen's clock (the cast tally,
// bumped before the cast event). The draw and the Wall are two steps of one
// instruction, in printed order. The combat trigger grants exalted until end
// of turn to the creatures you control with defender as it resolves (CR
// 611.2c), so a Wall made later that turn gets none; exalted itself is the
// engine's own derived trigger, one per instance.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53d58401-c318-47e0-aab0-60001c071dbb",
		Name:         "Rammas Echor, Ancient Shield",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Rammas Echor, Ancient Shield — draw a card, then create a 0/3 white Wall creature token with defender",
				effects.Do(effects.DrawCards{N: 1}, effects.CreateToken{Template: effects.TokenCard("0/3 white Wall with defender"), N: 1})),
			effects.AtBeginningOfYourCombat("Rammas Echor, Ancient Shield — creatures you control with defender gain exalted until end of turn",
				effects.Do(effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl(), effects.HasKeyword("defender")),
					Keywords: []string{"exalted"},
					Label:    "Rammas Echor, Ancient Shield — exalted",
				})),
		},
	})
}

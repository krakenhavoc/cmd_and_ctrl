package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon Egg — Creature — Dragon Egg {2}{R}, 0/2 (Magic 2014, #5563):
//
//	"Defender
//	 When this creature dies, create a 2/2 red Dragon creature token with flying and "{R}: This token gets +1/+0 until end of turn.""
//
// The card Nesting Dragon's egg token copies: defender, and a dies trigger
// making the same 2/2 flying Dragon with "{R}: This token gets +1/+0 until
// end of turn" (nestingDragonDragonToken).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8f561498-66e7-4cbf-8cb9-e6fd4717993d",
		Name:            "Dragon Egg",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Dragon Egg — create a 2/2 red Dragon",
				effects.Do(effects.CreateToken{Template: effects.NestingDragonDragonToken(), N: 1})),
		},
	})
}

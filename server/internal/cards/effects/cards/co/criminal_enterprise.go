package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Criminal Enterprise — Enchantment {2}{B} (Marvel Super Heroes
// Commander, #4455):
//
//	"When this enchantment enters, create a 2/1 black Villain creature
//	 token with menace.
//	 Whenever a Villain you control dies, this enchantment deals 1 damage
//	 to each opponent and you gain 1 life."
//
// The token is the shared "2/1 black Villain with menace" row. The dies
// trigger reads the Villain as it last existed on the battlefield (CR
// 603.10a, leftAsSubtype), under the controller it died under, so a token
// Villain counts and a Villain an effect stopped making one does not. The
// body is Boomerang's: one damage to each opponent, then one life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a98f48d-a5e6-497f-ba9a-92af7b489871",
		Name:         "Criminal Enterprise",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Criminal Enterprise — create a 2/1 black Villain creature token with menace",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/1 black Villain with menace"), N: 1})),
			effects.On(game.EventLTB, msc5AVillainYouControlDied,
				"Criminal Enterprise — 1 damage to each opponent and you gain 1 life",
				effects.Do(effects.Instruction(effects.BoomerangPingAndGain))),
		},
	})
}

// msc5AVillainYouControlDied is "whenever a Villain you control dies".
func msc5AVillainYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := effects.DiedCreature(ev, g)
	return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && effects.LeftAsSubtype(ev, dead, source.TextSubtype("Villain"))
}

package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rampage of the Valkyries — Enchantment {3}{W}{B}:
//
//	"When this enchantment enters, create a 4/4 white Angel creature token
//	 with flying and vigilance.
//	 Whenever an Angel you control dies, each other player sacrifices a
//	 creature of their choice."
//
// The Angel token is the printed 4/4. The death trigger reads the creature as
// it last existed (CR 603.10a), so an Angel by an effect counts and a
// changeling is one; each other player chooses their own creature (Grave
// Pact's edict). Not targeted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49039519-62db-4605-b600-cef6f34c893f",
		Name:         "Rampage of the Valkyries",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rampage of the Valkyries — create a 4/4 white Angel creature token with flying and vigilance",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 white Angel with flying and vigilance"), N: 1})),
			effects.TriggerWithPurpose(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller &&
					effects.LeftAsSubtype(ev, dead, source.TextSubtype("Angel"))
			}, "Rampage of the Valkyries — each other player sacrifices a creature",
				effects.Do(effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Creature(), Label: "a creature"})),
				game.Purpose{DeathPayoff: true, Answers: game.AnswerRemove}),
		},
	})
}

package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goro-Goro and Satoru — Legendary Creature — Goblin Human {U}{B}{R}, 3/4
// (March of the Machine Commander):
//
//	"Whenever one or more creatures you control that entered this turn deal
//	 combat damage to a player, create a 5/5 red Dragon Spirit creature token
//	 with flying.
//	 {1}{R}: Creatures you control gain haste until end of turn."
//
// The trigger is one per damaged player in a combat damage step (CR 603.2c).
// A creature "entered this turn" when it came onto the battlefield since the
// turn began, read off the object that dealt the damage, so a creature that
// died in the same damage step still counts. The haste grant is decided at
// resolution for the creatures then on the battlefield (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "854bdd4a-d549-41d5-af43-0e2154ed017e",
		Name:         "Goro-Goro and Satoru",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(
				func(g *game.Game, _ uuid.UUID, c game.Card) bool { return g.EnteredThisTurn(c.InstanceID) },
				"Goro-Goro and Satoru — create a 5/5 red Dragon Spirit creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("5/5 red Dragon Spirit with flying"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: Creatures you control gain haste until end of turn.",
			Cost:    effects.ManaCost("{1}{R}"),
			Purpose: game.Purpose{TeamPump: &game.Pump{Keywords: []string{"haste"}}, Answers: game.AnswerCombatGrant},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"haste"},
					Label: "Goro-Goro and Satoru — haste"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

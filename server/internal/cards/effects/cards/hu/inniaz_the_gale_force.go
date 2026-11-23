package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inniaz, the Gale Force — Legendary Creature — Djinn {3}{U}{U}, 4/4
// (Jumpstart):
//
//	"Flying
//	 {2}{W/U}: Attacking creatures with flying get +1/+1 until end of turn.
//	 Whenever three or more creatures you control with flying attack, each
//	 player gains control of a nonland permanent of your choice controlled
//	 by the player to their right."
//
// The pump is every attacking creature with flying as the ability
// resolves, whoever controls it, locked in for the turn (CR 611.2c): one
// that gains flying later gets nothing.
//
// The attack trigger is one trigger per declaration (OncePerBatch: the
// engine emits EventAttack per creature), counted off the board once every
// attacker is declared, Inniaz herself among the three when she attacks.
// As it resolves, Inniaz's controller chooses, for each player in turn
// order starting with themselves, a nonland permanent controlled by the
// player to that player's right (the previous player in turn order still
// in the game); a player whose neighbour controls none gains nothing. The
// choices are not targets. Every gain of control is registered after the
// last choice, so they happen together, and each lasts indefinitely
// (jmpB7InniazAsk). With two players left it is an exchange.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b06ebf74-4042-4c8f-a70b-98ba9e74e6ed",
		Name:            "Inniaz, the Gale Force",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W/U}: Attacking creatures with flying get +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{W/U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{Match: effects.And(effects.AttackingCreature(), effects.HasKeyword("flying")), Power: 1, Toughness: 1,
					Label: "Inniaz, the Gale Force — attacking creatures with flying get +1/+1"}.Apply(effects.NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && effects.JmpB7AttackingFlyersYouControl(g, source.Controller) >= 3
			}, effects.JmpB7InniazLabel, func(g *game.Game, item *game.StackItem) error {
				return effects.JmpB7InniazAsk(g, item.SourceCardID, item.Controller, 0, nil)
			})),
		},
	})
}

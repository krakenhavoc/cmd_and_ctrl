package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drownyard Amalgam — Creature — Zombie Horror {4}{U}, 3/6 (Innistrad:
// Midnight Hunt, S82):
//
//	"When this creature enters, target player mills three cards.
//	 {2}{U}: This creature can't be blocked this turn."
//
// The player is the enters trigger's target, checked again as it resolves
// (CR 608.2b); you may name yourself. The unblockable grant is Daring
// Saboteur's: the until-end-of-turn restriction pinned to this permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2378d6fc-b6ba-4a44-b951-5698ed507e4c",
		Name:         "Drownyard Amalgam",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Drownyard Amalgam — target player mills three cards", effects.TargetPlayerMillsEffect(3)), effects.TargetPlayer("target player")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{2}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Drownyard Amalgam — can't be blocked this turn"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

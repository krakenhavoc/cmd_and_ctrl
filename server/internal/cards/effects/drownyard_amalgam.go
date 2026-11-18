package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2378d6fc-b6ba-4a44-b951-5698ed507e4c",
		Name:         "Drownyard Amalgam",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Drownyard Amalgam — target player mills three cards", targetPlayerMills(3)), TargetPlayer("target player")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{2}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Drownyard Amalgam — can't be blocked this turn"}.Apply(NewContext(g, item))
			},
		}},
	})
}

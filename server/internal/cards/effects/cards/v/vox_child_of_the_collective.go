package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Võx, Child of the Collective — {4}{U} Legendary Artifact Creature —
// Borg 3/3 (Star Trek Commander, #4974):
//
//	"When Võx enters and at the beginning of your first main phase, the
//	 next artifact creature spell and the next Spacecraft spell you cast
//	 this turn have undaunted. (They cost {1} less to cast for each
//	 opponent.)"
//
// One printed ability with two trigger conditions, so one row watching
// both events (Grell Philosopher's shape). It makes two next-spell
// promises (game.NextSpellPromise with the Undaunted rider), each spent
// by the first spell that matches its own filter, so an artifact
// creature spell cast first does not spend the Spacecraft promise. A
// spell that is both is the next of each and gets undaunted twice
// (CR 702.125c). The opponents are counted as the spell is priced
// (CR 702.125a), leaving out players who have left the game
// (CR 702.125b), and both promises end with the turn. Your first main
// phase is the precombat main phase.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e58ea13-4041-4ef4-bc09-1e4bf79a842a",
		Name:         "Võx, Child of the Collective",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventBeginPrecombatMain}, effects.ThisEntersOrYourFirstMain,
				"Võx, Child of the Collective — the next artifact creature spell and the next Spacecraft spell you cast this turn have undaunted",
				effects.Do(
					effects.GrantNextSpellPromise{From: "Võx, Child of the Collective", Promise: game.NextSpellPromise{
						Filter:    game.PermissionFilter{CardType: "artifact", CreatureOnly: true},
						Undaunted: true,
						Text:      "The next artifact creature spell you cast this turn has undaunted.",
					}},
					effects.GrantNextSpellPromise{From: "Võx, Child of the Collective", Promise: game.NextSpellPromise{
						Filter:    game.PermissionFilter{Subtype: "Spacecraft"},
						Undaunted: true,
						Text:      "The next Spacecraft spell you cast this turn has undaunted.",
					}},
				)),
		},
	})
}

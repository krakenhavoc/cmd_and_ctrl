package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boggart Mob — Creature — Goblin Warrior {3}{B}, 5/5 (Lorwyn, S82 OLD, #5593):
//
//	"Champion a Goblin (When this enters, sacrifice it unless you exile another
//	 Goblin you control. When this leaves the battlefield, that card returns to
//	 the battlefield.)
//	 Whenever a Goblin you control deals combat damage to a player, you may create
//	 a 1/1 black Goblin Rogue creature token."
//
// The shared champion helper (CR 702.72) over Goblins, and Bident of Thassa's
// combat-damage watch read for any Goblin of yours (the Mob included): one
// optional token for each Goblin that dealt damage (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5e2da6c-35fa-4f46-bfae-8ac8907c8300",
		Name:         "Boggart Mob",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			boggartMobChampion.Trigger("Boggart Mob"),
			effects.Optional(effects.On(game.EventDealDamage, effects.S82OldB556AGoblinYouControlDealtCombatDamageToAPlayer,
				"Boggart Mob — you may create a 1/1 black Goblin Rogue creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 black Goblin Rogue"), N: 1})),
				"Boggart Mob — create a 1/1 black Goblin Rogue?"),
		},
	})
}

// boggartMobChampion is the champion pick, keyed. An on-disk key: never renamed,
// never reused.
var boggartMobChampion = effects.NewChampion("Boggart Mob", "Goblin", "choose-cards/s82-old-b556-boggart-mob-champion", game.TailRef{})

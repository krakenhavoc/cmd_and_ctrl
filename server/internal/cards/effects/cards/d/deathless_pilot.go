package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deathless Pilot — Creature — Zombie Pilot {1}{B}, 2/2:
//
//	"This creature saddles Mounts and crews Vehicles as though its power
//	 were 2 greater.
//	 {3}{B}: Return this card from your graveyard to your hand."
//
// The first line is CrewsAsThoughPowerGreater (ADR 0186 Decision 4): the
// Pilot counts as power 4 for crew and saddle, and its real power stays
// 2. The second is Blessed Ghoul's graveyard ability, shared with it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31284a0b-9bd5-4186-bd27-6b16a3597e27",
		Name:         "Deathless Pilot",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.CrewsAsThoughPowerGreater(2, game.CostPowerCrew|game.CostPowerSaddle),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}

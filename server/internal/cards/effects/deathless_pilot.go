package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "31284a0b-9bd5-4186-bd27-6b16a3597e27",
		Name:         "Deathless Pilot",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CrewsAsThoughPowerGreater(2, game.CostPowerCrew|game.CostPowerSaddle),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{B}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{3}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  returnThisCardFromYourGraveyardToYourHand,
		}},
	})
}

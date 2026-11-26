package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Anafenza, the Foremost — Khans of Tarkir (S82 OLD b346, #5554):
//
//	Legendary Creature — Human Soldier {W}{B}{G}, 4/4: Whenever Anafenza attacks, put
//	a +1/+1 counter on another target tapped creature you control.
//	If a nontoken creature an opponent owns would die or a creature card not on
//	the battlefield would be put into an opponent's graveyard, exile that card
//	instead.
//
// One replacement covers both clauses: a nontoken creature card headed for an
// opponent's graveyard from any zone (the battlefield included) is exiled. The
// graveyard is its owner's, which is the "opponent owns" the first clause says.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebfafabc-5255-4024-be59-7403bdb16ee4",
		Name:         "Anafenza, the Foremost",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventAttack, effects.Self, "Anafenza, the Foremost — put a +1/+1 counter on another target tapped creature you control",
					effects.PutPlusOneCounterOnEachLegalTarget),
				effects.Another(effects.TargetCreature("another target tapped creature you control", effects.YouControl(), effects.RfCreatureDTapped()))),
		},
		Replacements: []game.ReplacementEffect{effects.GraveyardBecomesExile{
			OpponentsOnly: true,
			CardsOnly:     true,
			OnlyTypes:     []string{"creature"},
			Label:         "Anafenza, the Foremost: exile an opponent's creature card instead of putting it into a graveyard",
		}.Build()},
	})
}

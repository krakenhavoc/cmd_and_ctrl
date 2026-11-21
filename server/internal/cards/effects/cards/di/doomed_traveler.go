package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Doomed Traveler — 1/1 Creature — Human Soldier for {W}:
//
//	"When Doomed Traveler dies, create a 1/1 white Spirit creature
//	token with flying."
//
// S19 sub-PR 4: a mandatory LTB trigger that creates one token via
// the CreateToken primitive when it resolves. cardDied gates the
// trigger to graveyard-only.
//
// Flying on the Spirit token is real: SpiritToken carries it on
// Card.Keywords, which printedCharacteristic folds into the layer
// engine (S21 sub-PR 1). The "cosmetic" note here outlived the fix;
// corrected in the #338 sweep.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a30907c0-fbde-4fd3-a8c7-f304305fcea7",
		Name:         "Doomed Traveler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Doomed Traveler — create a 1/1 Spirit", effects.Do(effects.CreateToken{
				Template: effects.TokenCard("1/1 white Spirit with flying"),
				N:        1,
			})),
		},
	})
}

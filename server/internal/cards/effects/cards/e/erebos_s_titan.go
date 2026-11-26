package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Erebos's Titan — Creature — Giant {1}{B}{B}{B}, 5/5 (Magic Origins, S82
// older sets):
//
//	"As long as your opponents control no creatures, this creature has
//	 indestructible.
//	 Whenever a creature card leaves an opponent's graveyard, you may discard
//	 a card. If you do, return this card from your graveyard to your hand."
//
// The first ability is a layer 6 grant on the source alone, read live, so it
// ends the moment an opponent controls a creature. The second works only while
// the card is in your graveyard (CR 113.6k) and triggers once for every
// creature card that leaves any opponent's graveyard, by exile, return or a
// cast from it, asked of the card as it was there (CR 603.10a). The discard is
// optional and chosen as the ability resolves; only a card really discarded
// returns the Titan, and only if it is still in the graveyard (CR 608.2b).
//
// No simplification.
func init() {
	grant := effects.KeywordGrant(effects.OldB321SelfWhileOpponentsControlNoCreatures, "indestructible")
	grant.Label = "As long as your opponents control no creatures, this creature has indestructible."
	effects.Register(effects.Spec{
		OracleID:     "1341b64e-1f8f-4613-abea-a337cd0169de",
		Name:         "Erebos's Titan",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{grant},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.OnAny(effects.LeftGraveyardEvents, effects.OldB321CreatureCardLeftAnOpponentsGraveyard,
				"Erebos's Titan — you may discard a card; if you do, return this card from your graveyard to your hand",
				effects.OldB321ErebosDiscardReturn)),
		},
	})
}

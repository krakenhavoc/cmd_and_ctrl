package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Madame Vastra — Legendary Creature — Lizard Detective {2}{G}{U}, 3/3 (Doctor Who):
//
//	"Partner with Jenny Flint (When this creature enters, target player may put
//	 Jenny Flint into their hand from their library, then shuffle.)
//	 Madame Vastra must be blocked if able.
//	 Whenever a creature dealt damage by Madame Vastra this turn dies, create a
//	 Clue token and a Food token."
//
// Partner with is the engine's entry trigger (CR 702.124j). The block
// requirement is Canopy Stalker's. The death trigger is Markov Enforcer's: the
// engine records who Vastra damaged this turn (combat damage and any other), so a
// creature she hurt makes a Clue and a Food when it dies, whoever controls it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4789357c-3ab7-4327-b5ee-b1110ddb876f",
		Name:         "Madame Vastra",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		Static:       []game.StaticAbility{MustBeBlocked()},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Madame Vastra", "Jenny Flint"),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := diedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && dealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Madame Vastra — create a Clue token and a Food token",
				Do(CreateToken{Template: ClueToken(), N: 1}, CreateToken{Template: FoodToken(), N: 1})),
		},
	})
}

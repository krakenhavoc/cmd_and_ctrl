package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// King T'Challa // Black Panther, Hope Enduring — a double-faced
// Legendary Creature (Marvel Super Heroes, #4455):
//
//	King T'Challa {1}{W}{U}, Legendary Creature — Human Noble Hero, 3/2
//	  "Flash
//	   Whenever a player draws their second card each turn, you draw a
//	   card.
//	   {4}{W}{U}: Transform King T'Challa. Activate only as a sorcery."
//
//	Black Panther, Hope Enduring {4}{W}{U}, Legendary Creature — Human
//	Warrior Hero, 3/3
//	  "Flash
//	   Double strike
//	   Prevent all damage that would be dealt to Black Panther.
//	   Whenever Black Panther deals combat damage to a player, draw a
//	   card."
//
// The draw trigger watches every player's draws and reads the drawer's
// own turn tally after the draw (Atlantean Cavalry's reading with any
// player), so it fires once per player per turn, on that player's second
// card, and the controller draws. The transform is the permanent turning
// over in place (ADR 0079). The prevention is the standing self
// prevention the Glittering creatures use, with no way to turn it off.
//
// No simplification.
func init() {
	const oracle = "fe00278d-4385-4e3f-919c-12a794fd86f5"
	Register(Spec{
		OracleID:        oracle,
		Name:            "King T'Challa",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor != uuid.Nil && g.TurnTallyFor(ev.Actor).CardsDrawn == 2
			}, "King T'Challa — you draw a card (a player drew their second card)",
				Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{4}{W}{U}: Transform King T'Challa. Activate only as a sorcery.",
			Cost:         ManaCost("{4}{W}{U}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Black Panther, Hope Enduring",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "double strike"},
		Replacements: []game.ReplacementEffect{
			PreventAllDamageDealtTo(PreventionStatic{
				To:    ToThisCreature,
				Label: "Black Panther — prevent all damage that would be dealt to him",
			}),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Black Panther — draw a card", Do(DrawCards{N: 1})),
		},
	})
}

package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hildibrand Manderville // Gentleman's Rise — an Adventure card (CR 715):
//
//	Hildibrand Manderville — Legendary Creature — Human Detective {1}{W},
//	2/2
//	  "Creature tokens you control get +1/+1.
//	   When Hildibrand Manderville dies, you may cast it from your
//	   graveyard as an Adventure until the end of your next turn."
//	Gentleman's Rise — Instant — Adventure {2}{B}
//	  "Create a 2/2 black Zombie creature token. (Then exile this card.
//	   You may cast the creature later from exile.)"
//
// The anthem is Intangible Virtue's. The dies trigger is
// castItFromYourGraveyardAsAnAdventure (shared with Mosswood
// Dreadknight): a cast permission over the card in its owner's
// graveyard, narrowed to the Adventure face (the permission's Faces, as
// a defeated Siege's transformed cast is) and good until the end of
// your next turn. The grant is the "may": nothing is asked, and an
// uncast card simply stays. It names the object that died (CR 400.7),
// so a Hildibrand that left the graveyard first is not castable. Cast
// that way the Adventure resolves into exile like any other (CR 715.3d)
// and the creature can be cast from there later.
//
// No simplification.
const hildibrandOracleID = "6df03db9-e4b3-4931-94e3-c582d2b04110"

func init() {
	effects.Register(effects.Spec{
		OracleID:     hildibrandOracleID,
		Name:         "Hildibrand Manderville",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Creature tokens you control get +1/+1.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.B05CreatureTokenYouControl,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Hildibrand Manderville — you may cast it from your graveyard as an Adventure until the end of your next turn",
				effects.CastItFromYourGraveyardAsAnAdventure("Hildibrand Manderville", "Gentleman's Rise")),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     hildibrandOracleID + "#1",
		Name:         "Gentleman's Rise",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1}.Apply(ctx)
		},
	})
}

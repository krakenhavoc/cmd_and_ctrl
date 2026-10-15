package effects

import (
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
// The anthem is Intangible Virtue's. The dies trigger grants a cast
// permission over the card in its owner's graveyard, narrowed to the
// Adventure face (the permission's Faces, as a defeated Siege's
// transformed cast is) and good until the end of your next turn
// (UntilEndOfYourNextTurnDuration). The grant is the "may": nothing is
// asked, and an uncast card simply stays. It names the object that died
// (CR 400.7), so a Hildibrand that left the graveyard first is not
// castable. Cast that way the Adventure resolves into exile like any
// other (CR 715.3d) and the creature can be cast from there later.
//
// No simplification.
const hildibrandOracleID = "6df03db9-e4b3-4931-94e3-c582d2b04110"

func init() {
	Register(Spec{
		OracleID:     hildibrandOracleID,
		Name:         "Hildibrand Manderville",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Creature tokens you control get +1/+1.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: b05CreatureTokenYouControl,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Hildibrand Manderville — you may cast it from your graveyard as an Adventure until the end of your next turn",
				func(g *game.Game, item *game.StackItem) error {
					z := g.FindCardZoneForEffect(item.SourceCardID)
					if z == nil || z.Kind != game.ZoneGraveyard || !sameGraveyardObject(z, item) {
						return nil
					}
					c, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok {
						return nil
					}
					g.GrantCastPermissionToCardsForEffect(game.CastPermission{
						Player:     item.Controller,
						Zone:       game.ZoneGraveyard,
						Duration:   g.UntilEndOfYourNextTurnDuration(item.Controller),
						CastOnly:   true,
						Faces:      []int{1},
						Source:     item.SourceCardID,
						SourceName: "Hildibrand Manderville",
						Label:      "Hildibrand Manderville — cast Gentleman's Rise from your graveyard",
					}, []game.Card{c})
					return nil
				}),
		},
	})
	Register(Spec{
		OracleID:     hildibrandOracleID + "#1",
		Name:         "Gentleman's Rise",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1}.Apply(ctx)
		},
	})
}

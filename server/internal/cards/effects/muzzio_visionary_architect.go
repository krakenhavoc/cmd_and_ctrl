package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Muzzio, Visionary Architect — Legendary Creature — Human Artificer
// {1}{U}{U}, 1/3:
//
//	"{3}{U}, {T}: Look at the top X cards of your library, where X is the
//	 greatest mana value among artifacts you control. You may put an
//	 artifact card from among them onto the battlefield. Put the rest on
//	 the bottom of your library in any order."
//
// X is read as the ability resolves (post-layer artifacts, zero with
// none). The dig is Industrial Advancement's: an optional pick of one
// artifact card among the cards looked at, the rest to the bottom with the
// player's chosen order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc2768a5-621b-48de-9160-4dbaeddb3de1",
		Name:         "Muzzio, Visionary Architect",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}, {T}: Look at the top X cards of your library, where X is the greatest mana value among artifacts you control. You may put an artifact card from among them onto the battlefield. Put the rest on the bottom of your library in any order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}{U}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := b27GreatestManaValueAmongArtifactsControlled(g, ctx.Controller())
				if x <= 0 {
					return nil
				}
				cards := g.LookAtTopOfLibraryForEffect(ctx.Controller(), x)
				if len(cards) == 0 {
					return nil
				}
				return PutFromLibraryOntoBattlefield{
					Player:   ctx.Controller(),
					Cards:    cards,
					Match:    Artifact(),
					Max:      1,
					Optional: true,
					Label:    "Muzzio, Visionary Architect — you may put an artifact card onto the battlefield",
					ThenTail: rprB15RestOnBottomAnyOrderTail,
				}.Apply(ctx)
			},
		}},
	})
}

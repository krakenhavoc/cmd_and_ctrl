package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boneclad Necromancer — Creature — Human Wizard {3}{B}{B}, 3/3
// (Foundations Jumpstart):
//
//	"When this creature enters, you may exile target creature card from
//	 a graveyard. If you do, create a 2/2 black Zombie creature token."
//
// Same shape as Summoner's Sending: the trigger is optional, and the
// token follows only an exile that happened.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6494e3c4-5489-44a5-94b6-4ee2d7830896",
		Name:         "Boneclad Necromancer",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(WhenThisEnters("Boneclad Necromancer — exile target creature card from a graveyard; if you do, create a 2/2 black Zombie",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					return ExileTarget{Target: ts[0].ID, ThenTail: ctx.ItemTail(bonecladNecromancerZombie)}.Apply(ctx)
				}),
				TargetCardInGraveyard("target creature card from a graveyard", Creature())),
				"Boneclad Necromancer — exile target creature card from a graveyard?"),
		},
	})
}

var bonecladNecromancerZombie = CardTail("boneclad-necromancer/zombie-if-exiled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1}.Apply(ctx)
})

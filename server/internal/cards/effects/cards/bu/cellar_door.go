package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cellar Door — Artifact {2} (Jumpstart 2022, #5948):
//
//	"{3}, {T}: Target player puts the bottom card of their library into
//	 their graveyard. If it's a creature card, you create a 2/2 black Zombie
//	 creature token."
//
// The card is judged a creature card as it goes to the graveyard (the text
// asks what the card is, not where it ended up). A player with an empty
// library puts nothing and you get nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a2f79037-86c0-430c-ba91-62ac807a44ce",
		Name:         "Cellar Door",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target player puts the bottom card of their library into their graveyard. If it's a creature card, you create a 2/2 black Zombie creature token.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 || ts[0].Kind != game.TargetPlayer {
					return nil
				}
				p := g.PlayerByIDForEffect(ts[0].ID)
				if p == nil || p.Library == nil || len(p.Library.Cards) == 0 {
					return nil
				}
				// Index 0 is the bottom (Zone.Bottom).
				bottom := p.Library.Cards[0]
				isCreature := bottom.IsCreature()
				if err := g.PutIntoGraveyardForEffect(bottom.InstanceID); err != nil {
					return err
				}
				if !isCreature {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1}.Apply(ctx)
			},
		}},
	})
}

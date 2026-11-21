package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scouring Swarm — Creature — Insect {1}{B}{G}, 1/1 (Foundations
// Commander):
//
//	"Flying
//	 Whenever you sacrifice a land, create a tapped token that's a copy of
//	 this creature if seven or more land cards are in your graveyard.
//	 Otherwise, create a tapped 1/1 black Insect creature token with
//	 flying."
//
// One trigger per sacrificed land (Baloth Prime's watcher). The land
// cards are counted as the trigger resolves, so the sacrificed land is
// usually among them. The copy is the Swarm's copiable values, read last
// known if it has left (TokenCopyTemplate, CR 707.2 and the ruling), and
// enters tapped through the creation's own entry clause; the otherwise
// token is the plain Insect, tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aa64f830-b9b6-4d5e-9a5d-f9e120089335",
		Name:            "Scouring Swarm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, effects.CmdrB9YouSacrificedLand,
				"Scouring Swarm — create a tapped copy of this creature if seven or more land cards are in your graveyard, otherwise a tapped 1/1 black Insect with flying",
				scouringSwarmToken),
		},
	})
}

// scouringSwarmToken is the trigger's body.
func scouringSwarmToken(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if effects.B17LandCardsInGraveyard(g, item.Controller) < 7 {
		return effects.CreateTokenAdvanced{Spec: effects.Token(effects.TokenCard("1/1 black Insect with flying")).EntersTapped(), N: 1}.Apply(ctx)
	}
	tmpl, ok := effects.TokenCopyTemplate(g, item.SourceCardID)
	if !ok {
		return nil
	}
	_, err := g.CreateTokensForEffect(item.Controller, effects.CopiedToken(tmpl), 1, game.TokenEntryOptions{Tapped: true})
	return err
}

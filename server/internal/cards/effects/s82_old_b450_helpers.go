package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b450_helpers.go — shared pieces of OLD batch 450 (#5572,
// Innistrad).

// gutterGrimeOozeText is the ability text of the Ooze Gutter Grime makes.
const gutterGrimeOozeText = "This token's power and toughness are each equal to the number of slime counters on Gutter Grime."

// GutterGrimeOozeToken is the green Ooze creature token Gutter Grime
// makes. It is linked to the Grime object that made it (Card.CreatedWith,
// CR 607.1d) and sized by a layer 7a characteristic-defining ability
// (CR 604.3, 613.4a) reading that object's slime counters while it is on
// the battlefield as the same object; otherwise it is 0/0 (CR 400.7).
// Saproling Burst's token with slime counters in place of fade counters.
func GutterGrimeOozeToken() game.Card { return TokenFromCatalog(printedGutterGrimeOozeToken) }

func printedGutterGrimeOozeToken() TokenTemplate {
	return TokenTemplate{
		Slug: "gutter-grime-ooze",
		Card: game.Card{
			Name:     "Ooze",
			TypeLine: "Token Creature — Ooze",
			Colors:   []string{"G"},
		},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			Label:     gutterGrimeOozeText,
			AppliesTo: SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				if link := source.CreatedWith; link.ID != uuid.Nil {
					for _, p := range g.BattlefieldCardsForEffect() {
						if p.InstanceID == link.ID && p.ObjectEpoch == link.Epoch {
							n = p.Counters["slime"]
							break
						}
					}
				}
				c.Power, c.Toughness = n, n
			},
		}},
		Text: gutterGrimeOozeText,
	}
}

// GutterGrimeCreate is Gutter Grime's resolution: a slime counter on the
// Grime while it is still on the battlefield, then an Ooze linked to that
// object. "Then" keeps the token even when the Grime has gone (it is a
// 0/0 and dies to state-based actions).
func GutterGrimeCreate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (CounterOnThis{Kind: "slime", N: 1}).Apply(ctx); err != nil {
		return err
	}
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{{
			Template:    ctx.TextToken(GutterGrimeOozeToken()),
			Count:       1,
			CreatedWith: game.PermissionCardRef(ref),
		}},
	}, game.KeyedTail{})
}

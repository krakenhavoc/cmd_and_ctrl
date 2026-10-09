package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// explores.go — the card-facing half of the CR 701.44 keyword action
// "<permanent> explores" (#2720). The engine half is game/explore.go.
// (explore.go in this package is the card named Explore.)
//
//	"Target creature you control explores."  // Map token, Guidestone Compass
//
// The Map token (CR 111.10s) lives here too, beside the action it
// exists to perform.

// Explores has the permanent `Explorer` explore under the effect's
// controller. Then, when set, is the rest of the effect, run once the
// permanent has explored.
type Explores struct {
	Explorer uuid.UUID
	Then     func(g *game.Game) error
}

// Apply reads the explorer's object identity now, so a creature that
// leaves before the counter is placed gets none (CR 400.7), and the
// reveal still happens as it last existed (CR 701.44c).
func (e Explores) Apply(ctx *Context) error {
	ref := game.ObjectRef{ID: e.Explorer}
	if c, ok := ctx.Game.LookupCardForEffect(e.Explorer); ok {
		ref.Epoch = c.ObjectEpoch
	}
	return ctx.Game.ExploreForEffect(ctx.Source(), ref, ctx.Controller(), e.Then)
}

// targetCreatureYouControlExplores is the resolution of "Target creature
// you control explores": the chosen target, if it is still legal
// (CR 608.2b).
func targetCreatureYouControlExplores(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	return Explores{Explorer: t.ID}.Apply(ctx)
}

// targetCreatureYouControl is the clause every "target creature you
// control explores" ability targets with.
func targetCreatureYouControl() *game.TargetSpec {
	return TargetPermanent("target creature you control", Creature(), YouControl())
}

// MapToken — CR 111.10s: "{1}, {T}, Sacrifice this token: Target
// creature you control explores. Activate only as a sorcery."
func MapToken() game.Card { return tokenFromCatalog(printedMapToken) }

// printedMapToken is the Map as PRINTED — the abilities included. It is
// the catalog's entry for this token (token_catalog.go).
func printedMapToken() tokenTemplate {
	return tokenTemplate{
		Slug: "map",
		Card: game.Card{
			Name:     "Map",
			TypeLine: "Token Artifact — Map",
		},
		Text: "{1}, {T}, Sacrifice this token: Target creature you control explores. Activate only as a sorcery.",
		Activated: []game.ActivatedAbilityShape{{
			Label: "{1}, {T}, Sacrifice this artifact: Target creature you control explores",
			Cost: game.AbilityCost{
				Tap:           true,
				SacrificeSelf: true,
				Mana:          "{1}",
			},
			Targets:      targetCreatureYouControl(),
			SorcerySpeed: true,
			Effect:       targetCreatureYouControlExplores,
		}},
	}
}

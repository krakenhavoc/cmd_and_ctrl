package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silent Gravestone — Artifact {1}:
//
//	"Cards in graveyards can't be the targets of spells or abilities.
//	 {4}, {T}: Exile this artifact and all cards from all graveyards.
//	 Draw a card."
//
// THE STATIC is ADR 0109 §6's TargetingRestrictions (#1885), read at the
// engine's two targeting choke points: while the Gravestone is on the
// battlefield no spell or ability may target a card in any graveyard,
// and one already aimed at a graveyard card loses that target at
// resolution (CR 601.2c, 608.2b).
//
// THE ACTIVATION exiles the Gravestone and every card in every graveyard
// as ONE simultaneous exile, then draws. The Gravestone is exiled only if
// it is still on the battlefield: one bounced in response is a new object
// (CR 400.7) and stays where it went. A commander in a graveyard may take
// CR 903.9's offer instead, which pauses the exile, so the draw runs in
// the exile's continuation and happens after it, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "803815c5-12be-48b3-a101-f416c1ef9b7b",
		Name:         "Silent Gravestone",
		Completeness: CompletenessFull,
		TargetingRestrictions: []game.TargetingRestriction{
			CardsInGraveyardsCantBeTargeted("Cards in graveyards can't be the targets of spells or abilities."),
		},
		Activated: []ActivatedAbility{{
			Label:  "{4}, {T}: Exile this artifact and all cards from all graveyards. Draw a card.",
			Cost:   Plus(ManaCost("{4}"), TapCost()),
			Effect: silentGravestoneExile,
		}},
	})
}

// silentGravestoneExile is the activation's resolution.
func silentGravestoneExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := allGraveyardsCardIDs(ctx, func(game.Card) bool { return true })
	if z := g.FindCardZoneForEffect(item.SourceCardID); z != nil && z.Kind == game.ZoneBattlefield {
		ids = append([]uuid.UUID{item.SourceCardID}, ids...)
	}
	you := item.Controller
	return g.ExileCardsThenForEffect(ids, func(g *game.Game, _ []uuid.UUID) error {
		return g.DrawNForEffect(you, 1)
	})
}

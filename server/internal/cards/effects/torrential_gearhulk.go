package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Torrential Gearhulk — Artifact Creature — Construct {4}{U}{U}, 5/6:
//
//	"Flash
//	 When this creature enters, you may cast target instant card from
//	 your graveyard without paying its mana cost. If that spell would
//	 be put into your graveyard, exile it instead."
//
// The cast is a grant, the way every "you may cast it" a resolution
// offers is (ADR 0099 §4): a free ({0}, so X is locked at 0 and
// additional costs are still paid), any-timing (CR 608.2g), cast-only,
// pass-closed permission over the target card in the graveyard. The
// holder's next priority pass closes it and the card simply stays where
// it is. "Exile it instead" is the permission's ExileOnResolution, the
// replacement flashback uses, so the spell is exiled whether it
// resolves, is countered or fizzles. The target is chosen as the
// trigger goes on the stack and re-checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	enters := WhenThisEnters("Torrential Gearhulk — you may cast target instant card from your graveyard without paying its mana cost",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				z := g.FindCardZoneForEffect(t.ID)
				c, ok := g.LookupCardForEffect(t.ID)
				if !ok || z == nil || z.Kind != game.ZoneGraveyard {
					return nil
				}
				g.GrantCastPermissionToCardsForEffect(game.CastPermission{
					Player:            item.Controller,
					Zone:              game.ZoneGraveyard,
					Duration:          g.UntilEndOfTurnDuration(),
					Cost:              "{0}",
					AltCostKey:        game.AltCostKeyGrantedCost,
					Timing:            game.TimingFlash,
					CastOnly:          true,
					ExileOnResolution: true,
					// A pass closes the window; nothing moves, the card
					// stays in the graveyard.
					LapseOnPass: game.LapseStaysInHand,
					Source:      item.SourceCardID,
					SourceName:  "Torrential Gearhulk",
					Label:       "Torrential Gearhulk — cast " + c.Name + " without paying its mana cost",
				}, []game.Card{c})
				return nil
			}
			return nil
		})
	enters.Targets = TargetCardInGraveyard("target instant card from your graveyard", Instant(), YouOwn())
	Register(Spec{
		OracleID:        "a6395447-677d-4c39-8eda-2d57e527c94e",
		Name:            "Torrential Gearhulk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}

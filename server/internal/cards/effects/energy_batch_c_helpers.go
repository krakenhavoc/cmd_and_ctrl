package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy_batch_c_helpers.go — small bodies shared by the ADR 0129 PR 1
// energy cards (batch C).

// pumpAndGrantFirstTarget is "Target creature gets +P/+T and gains
// <keywords> until end of turn" as an ability's whole body: the still-
// legal target (CR 608.2b) gets the boost and the keywords, nothing if
// it left in response.
func pumpAndGrantFirstTarget(power, toughness int, label string, keywords ...string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		if err := (BoostUntilEOT{Target: target, Power: power, Toughness: toughness, Label: label}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{Target: target, Keywords: keywords, Label: label}.Apply(ctx)
	}
}

// returnFirstGraveyardTargetTapped puts the ability's still-legal
// graveyard target onto the battlefield tapped: under the activator's
// control when underYourControl is set, its owner's otherwise (CR
// 608.2b: nothing if it left).
func returnFirstGraveyardTargetTapped(underYourControl bool) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			r := ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Tapped: true}
			if underYourControl {
				r.Controller = ctx.Controller()
			}
			return r.Apply(ctx)
		}
		return nil
	}
}

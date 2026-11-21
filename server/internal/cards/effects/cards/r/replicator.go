package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Replicator — Artifact {2} (Star Trek, #4973):
//
//	"When this artifact enters, create a Food token.
//	 {2}, {T}: Create a tapped token that's a copy of target token you
//	 control."
//
// The Food is the printed Food token. The copy is Caretaker's Talent's
// CreateTokenCopy over a token you control, with the token entering
// tapped (Hashaton's exception sets Tapped on the template). The copy
// takes the token's copiable values (CR 707.2), so counters and tapped
// status are not copied, and a target that is gone by resolution makes
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a91b8eb4-d4f7-4e24-ae49-1f51e610c7ca",
		Name:         "Replicator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Replicator — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Create a tapped token that's a copy of target token you control",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target token you control", effects.IsTokenPredicate(), effects.YouControl()),
			Effect:  replicatorCopy,
		}},
	})
}

// replicatorCopy is the activated ability's resolution.
func replicatorCopy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return effects.CreateTokenCopy{
			Controller: item.Controller,
			Copy:       t.ID,
			N:          1,
			Except:     func(tok *game.Card) { tok.Tapped = true },
		}.Apply(ctx)
	}
	return nil
}

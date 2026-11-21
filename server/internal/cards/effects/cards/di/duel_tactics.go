package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duel Tactics — Sorcery {R} (Secrets of Strixhaven, #4456):
//
//	"Duel Tactics deals 1 damage to target creature. It can't block this
//	 turn.
//	 Flashback {1}{R} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// The damage is dealt first and the restriction is put on the same
// creature afterwards, whether or not the damage killed it (a creature that
// is gone takes no restriction). Flashback is the engine's alternative
// cost from the graveyard (Artful Dodge's shape), which exiles the spell
// after it resolves, is countered or otherwise leaves the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "262c5ac9-af41-42d1-a53a-8a671eedd5a6",
		Name:             "Duel Tactics",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{1}{R}")},
		Purpose: effects.ForTargets(game.TargetPurpose{
			Slot: 0, Damage: 1, UntilEndOfTurn: game.TargetHarm, Pump: &game.Pump{CantBlock: true},
		}),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
			return effects.RestrictUntilEOT{
				Target:       id,
				Restrictions: game.CantBlock,
				Label:        "Duel Tactics — can't block this turn",
			}.Apply(ctx)
		},
	})
}

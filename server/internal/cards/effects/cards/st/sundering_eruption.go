package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sundering Eruption // Volcanic Fissure — modal double-faced card.
// This file is the FRONT face, Sorcery {2}{R}:
//
//	"Destroy target land. Its controller may search their library for
//	 a basic land card, put it onto the battlefield tapped, then
//	 shuffle. Creatures without flying can't block this turn."
//
// The back face, Volcanic Fissure, is registered with the MDFC land
// cycle in mdfc_lands.go under "<oracle>#1".
//
// Path to Exile's search on Ghost Quarter's destruction: the land's
// controller is read BEFORE it dies, the search is an optional prompt
// addressed to them and happens whether or not the land actually died
// (indestructible, regeneration), and the basic enters tapped. The
// can't-block rider is Falter's rule and runs after the search (printed
// order), whether or not the controller took it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c95309e9-5c2f-4518-b2fd-825d3d0a4ae0",
		Name:         "Sundering Eruption",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			victim, ok := effects.ControllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return effects.SearchLibrary{
				Player:        victim,
				Predicate:     effects.IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Optional:      true,
				Reason:        "Sundering Eruption — you may search for a basic land",
				// #4253 slice 4: a registered tail.
				ThenTail: ctx.ItemTail(sunderingEruptionFalter),
			}.Apply(ctx)
		},
	})
}

// sunderingEruptionFalter is "creatures without flying can't block this
// turn". An on-disk key.
var sunderingEruptionFalter = effects.CardTail("sundering-eruption/creatures-cant-block", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.Faltering("Sundering Eruption")(ctx.Item, ctx)
})

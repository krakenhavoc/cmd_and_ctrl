package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Petrified Field — Land:
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Return target land card from your graveyard
//	 to your hand."
//
// Turntimber Sower's graveyard-land return behind a tap-and-sacrifice
// cost. The land is sacrificed as a cost, so it is never its own target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4bc5bc4-e589-42c5-91fa-2ebc96448e85",
		Name:         "Petrified Field",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this land: Return target land card from your graveyard to your hand.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCardInGraveyard("target land card in your graveyard", effects.YouOwn(), effects.Land()),
			Effect: effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
			})),
		}},
	})
}

package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "c4bc5bc4-e589-42c5-91fa-2ebc96448e85",
		Name:         "Petrified Field",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this land: Return target land card from your graveyard to your hand.",
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetCardInGraveyard("target land card in your graveyard", YouOwn(), Land()),
			Effect: Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
			})),
		}},
	})
}

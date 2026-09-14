package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Volatile Fault — Land — Cave (EDHREC rank 2640):
//
//	"{T}: Add {C}.
//	 {1}, {T}, Sacrifice this land: Destroy target nonbasic land an
//	 opponent controls. That player may search their library for a
//	 basic land card, put it onto the battlefield, then shuffle. You
//	 create a Treasure token."
//
// Demolition Field that pays you back in a Treasure instead of a
// basic. Same three-part cost, same three-predicate target clause
// (land, nonbasic, an opponent's), and the victim's search is a real
// "may": an Optional SearchLibrary addressed to the destroyed land's
// controller, who can decline the basic and the shuffle. The
// Treasure is created in the search's continuation, so it arrives
// after the victim has answered — printed order — and it arrives
// whether they found a basic, declined, or had none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "95c44f28-f7fa-4785-83b9-0d81be0db0c8",
		Name:         "Volatile Fault",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this land: Destroy target nonbasic land an opponent controls. Its controller may search for a basic land. You create a Treasure.",
			Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Targets: TargetPermanent("target nonbasic land an opponent controls", Land(), b03Nonbasic(), OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				target := item.Targets[0].ID
				victim, ok := controllerOfTarget(ctx, target)
				if !ok {
					return nil
				}
				if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
					return err
				}
				controller := item.Controller
				return SearchLibrary{
					Player:    victim,
					Predicate: IsBasicLand,
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Volatile Fault — you may search for a basic land",
					Then: func(g *game.Game, _ []uuid.UUID) error {
						return CreateToken{Controller: controller, Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
					},
				}.Apply(ctx)
			},
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Furnace Reins — Sorcery {2}{R} (March of the Machine):
//
//	"Gain control of target creature until end of turn. Untap that
//	 creature. Until end of turn, it gains haste and "Whenever this
//	 creature deals combat damage to a player or battle, create a
//	 Treasure token.""
//
// Act of Treason's control change, untap and haste, plus a granted
// trigger (Trash the Town's grant shape) on the player-or-battle combat
// damage predicate Bilbo uses. The Treasure is made under the trigger's
// controller, the one who stole the creature.
//
// No simplification.
const furnaceReinsGrant = "furnace-reins/combat-damage-treasure"

func init() {
	Register(Spec{
		OracleID:     "b623e61a-cdb9-46b0-a443-547be05db877",
		Name:         "Furnace Reins",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: furnaceReinsGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventDealDamage, thisDealtCombatDamageToAPlayerOrBattle,
					"Furnace Reins — create a Treasure token",
					Do(CreateToken{Template: TreasureToken(), N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player or battle, create a Treasure token.",
		}},
		Targets: TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (GainControl{
				Target:   target,
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Furnace Reins — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Furnace Reins — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantAbilitiesFor{
				Target: target,
				Keys:   []string{furnaceReinsGrant},
				Label:  "Furnace Reins — creates a Treasure when it deals combat damage to a player or battle",
			}.Apply(ctx)
		},
	})
}

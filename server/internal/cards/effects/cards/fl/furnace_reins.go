package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b623e61a-cdb9-46b0-a443-547be05db877",
		Name:         "Furnace Reins",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: furnaceReinsGrant,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayerOrBattle,
					"Furnace Reins — create a Treasure token",
					effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player or battle, create a Treasure token.",
		}},
		Targets: effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (effects.GainControl{
				Target:   target,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Furnace Reins — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Furnace Reins — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantAbilitiesFor{
				Target: target,
				Keys:   []string{furnaceReinsGrant},
				Label:  "Furnace Reins — creates a Treasure when it deals combat damage to a player or battle",
			}.Apply(ctx)
		},
	})
}

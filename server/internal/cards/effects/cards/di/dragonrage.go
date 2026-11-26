package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragonrage — {2}{R} Instant (Fate Reforged):
//
//	"Add {R} for each attacking creature you control. Until end of turn,
//	 attacking creatures you control gain "{R}: This creature gets +1/+0 until
//	 end of turn.""
//
// The mana is added as the spell resolves, one {R} per creature you control that
// is attacking then (CR 106.4: it empties with the step). The grant goes to the
// creatures attacking as it resolves (CR 611.2c: an effect that changes
// characteristics from a resolving spell affects only the objects at that time),
// so a creature that attacks or is put onto the battlefield attacking later does
// not get it; it lasts until the cleanup step (CR 514.2).
func init() {
	const key = "dragonrage/firebreathing"
	const text = "{R}: This creature gets +1/+0 until end of turn."
	effects.Register(effects.Spec{
		OracleID:     "fb16c4b0-9ac4-4664-9b57-97bfc22e5de9",
		Name:         "Dragonrage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
		Grants: []effects.AbilityGrant{{
			Key: key,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Dragonrage — +1/+0 until end of turn"),
			}},
			Text: text,
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var attackers []game.Card
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && c.IsCreature() && c.AttackingTarget != uuid.Nil {
					attackers = append(attackers, c)
				}
			}
			if err := (effects.AddMana{Amount: effects.ManaN(len(attackers), "R")}).Apply(ctx); err != nil {
				return err
			}
			for _, c := range attackers {
				if err := (effects.GrantAbilitiesFor{Target: c.InstanceID, Keys: []string{key}, Label: "Dragonrage — {R}: +1/+0 until end of turn"}).Apply(ctx.AsGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mana Vortex — Enchantment {1}{U}{U}:
//
//	"When you cast this spell, counter it unless you sacrifice a land.
//	 At the beginning of each player's upkeep, that player sacrifices a
//	 land of their choice.
//	 When there are no lands on the battlefield, sacrifice this
//	 enchantment."
//
// ADR 0107 PR 1 (#1858). Three existing shapes and a state trigger:
//
//   - The cast trigger fires from the stack (FromStack, cascade's shape)
//     and is controlled by the caster. As it resolves, a caster with a
//     land chooses between sacrificing one and letting the Vortex be
//     countered (The Gitrog Monster's "unless you sacrifice a land"); a
//     caster with none has nothing to pay, and the spell is countered. A
//     Vortex that has already left the stack is not countered.
//   - The upkeep trigger is "each player's": the player whose upkeep it
//     is sacrifices a land they choose. A player with no land sacrifices
//     nothing.
//   - The last line is a CR 603.8 state trigger over every land on the
//     battlefield, anyone's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "98b88734-09e2-4209-a787-9b5b345391f9",
		Name:         "Mana Vortex",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				FromStack: true,
				Watches:   []game.EventKind{game.EventCast},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				Key: manaVortexCastLabel,
				// A fill-in Build (ADR 0041 P9): the caster controls the
				// trigger (CR 603.3a).
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, manaVortexCastLabel)
					item.Controller, item.Owner = ev.Actor, ev.Actor
					return item
				},
				Effect: manaVortexCounterUnlessSacrifice,
			},
			AtEachUpkeep("Mana Vortex — that player sacrifices a land", func(g *game.Game, item *game.StackItem) error {
				player := NewContext(g, item).Trigger().Event.Actor
				if player == uuid.Nil {
					return nil
				}
				g.PlayerSacrificesForEffect(item.SourceCardID, player, sacrificeSpec("a land", Land()),
					"Mana Vortex — sacrifice a land")
				return nil
			}),
			WhenThereAreNo(QueryType("land"), "Mana Vortex — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}

const manaVortexCastLabel = "Mana Vortex — counter it unless you sacrifice a land"

// manaVortexCounterUnlessSacrifice is the cast trigger's resolution. The
// spell is the trigger's source, so its stack ID is the source's card ID.
func manaVortexCounterUnlessSacrifice(g *game.Game, item *game.StackItem) error {
	spell := item.SourceCardID
	counter := func(ctx *Context) error {
		if ctx.Game.StackItemForEffect(spell) == nil {
			return nil
		}
		return CounterTarget{StackID: spell}.Apply(ctx)
	}
	ctx := NewContext(g, item)
	if g.StackItemForEffect(spell) == nil {
		return nil
	}
	lands := permanentsControlledByMatching(g, item.Controller, Land())
	if len(lands) == 0 {
		return counter(ctx)
	}
	return PickOption{
		Player:   item.Controller,
		Question: "Mana Vortex — sacrifice a land, or Mana Vortex is countered",
		Options: []game.ChoiceOption{
			{Label: "Sacrifice a land"},
			{Label: "Let Mana Vortex be countered"},
		},
		Then: func(ctx *Context, index int) error {
			if index == 0 {
				return SacrificeChoice{
					Player:     item.Controller,
					Candidates: lands,
					Question:   "Mana Vortex — sacrifice a land",
				}.Apply(ctx)
			}
			return counter(ctx)
		},
	}.Apply(ctx)
}

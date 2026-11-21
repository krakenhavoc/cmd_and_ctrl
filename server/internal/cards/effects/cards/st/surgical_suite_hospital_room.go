package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surgical Suite // Hospital Room — Enchantment — Room (ADR 0103):
//
//	Surgical Suite {1}{W}: "When you unlock this door, return target
//	 creature card with mana value 3 or less from your graveyard to the
//	 battlefield."
//	Hospital Room {3}{W}: "Whenever you attack, put a +1/+1 counter on
//	 target attacking creature."
//
// Surgical Suite is Too Evil to Stay Dead's target clause on the
// reanimation helper. Hospital Room is one trigger per attack
// declaration (the OncePerBatch Adeline and Legion Warboss use) whose
// target, an attacking creature, is chosen as it goes on the stack.
func init() {
	effects.Register(effects.Room(effects.RoomSpec{
		OracleID:     "68bb76af-f933-430a-a66a-fe4a7d6be4e0",
		Name:         "Surgical Suite // Hospital Room",
		Completeness: effects.CompletenessFull,
		Left: effects.Door{Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenYouUnlockThisDoor(game.DoorLeft, "Surgical Suite — return a creature card with mana value 3 or less from your graveyard to the battlefield",
					effects.ReturnFirstLegalGraveyardTargetToBattlefield),
				effects.TargetCardInGraveyard("target creature card with mana value 3 or less in your graveyard", effects.YouOwn(), effects.Creature(), effects.ManaValueLE(3))),
		}},
		Right: effects.Door{Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Hospital Room — put a +1/+1 counter on target attacking creature", effects.B36CounterOnChosenAnimal),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature()))),
		}},
	}))
}

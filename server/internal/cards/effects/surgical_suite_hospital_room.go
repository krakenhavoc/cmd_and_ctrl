package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Room(RoomSpec{
		OracleID:     "68bb76af-f933-430a-a66a-fe4a7d6be4e0",
		Name:         "Surgical Suite // Hospital Room",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			Targeting(
				WhenYouUnlockThisDoor(game.DoorLeft, "Surgical Suite — return a creature card with mana value 3 or less from your graveyard to the battlefield",
					returnFirstLegalGraveyardTargetToBattlefield),
				TargetCardInGraveyard("target creature card with mana value 3 or less in your graveyard", YouOwn(), Creature(), ManaValueLE(3))),
		}},
		Right: Door{Triggered: []game.TriggeredAbility{
			OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Hospital Room — put a +1/+1 counter on target attacking creature", b36CounterOnChosenAnimal),
				TargetCreature("target attacking creature", AttackingCreature()))),
		}},
	}))
}

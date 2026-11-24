package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Halvar, God of Battle // Sword of the Realms — Kaldheim modal double-faced
// card (#5505), the front face under the bare oracle ID and the back face
// under "<oracle_id>#1":
//
//	Halvar, God of Battle {2}{W}{W}, Legendary Creature — God, 4/4
//	 "Creatures you control that are enchanted or equipped have double strike.
//	  At the beginning of each combat, you may attach target Aura or Equipment
//	  attached to a creature you control to target creature you control."
//
//	Sword of the Realms {1}{W}, Legendary Artifact — Equipment
//	 "Equipped creature gets +2/+0 and has vigilance.
//	  Whenever equipped creature dies, return it to its owner's hand.
//	  Equip {1}{W}"
//
// Halvar's grant reads the layered board on every recompute (an Aura or
// Equipment attached to the creature, whoever controls it). Its trigger is
// each combat on every player's turn, with two target clauses chosen as it
// goes on the stack; the move is AttachForEffect, which keeps an attachment
// that can't legally be attached to the new host where it is (CR 701.3b).
// The Sword's trigger reads the creature that died off the event and
// returns it only while it is still in the graveyard (CR 400.7).
func init() {
	double := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsCreature() && effects.CreatureIsEnchantedOrEquipped(g, target)
	}, "double strike")
	double.Label = "Creatures you control that are enchanted or equipped have double strike."
	move := effects.Targeting(effects.AtBeginningOfEachCombat(
		"Halvar — you may attach target Aura or Equipment attached to a creature you control to target creature you control",
		effects.AttachClauseTargets), effects.Clauses(
		effects.TargetPermanent("target Aura or Equipment attached to a creature you control", effects.AuraOrEquipmentOnACreatureYouControl()),
		effects.TargetCreature("target creature you control", effects.YouControl())))
	effects.Register(effects.Spec{
		OracleID:     "acfc3e07-757f-4ccf-a65e-bee07ce87f40",
		Name:         "Halvar, God of Battle",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{double},
		Triggered: []game.TriggeredAbility{
			effects.Optional(move, "Halvar, God of Battle — attach the Aura or Equipment to the target creature?"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "acfc3e07-757f-4ccf-a65e-bee07ce87f40#1",
		Name:         "Sword of the Realms",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			effects.GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Sword of the Realms — return equipped creature to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					dead := item.Trigger.Event.CardID
					if !effects.DiedCardStillInGraveyard(ctx, dead) {
						return nil
					}
					return effects.ReturnFromGraveyard{Target: dead, Dest: game.ZoneHand}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}{W}")},
	})
}

package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Satsuki, the Living Lore — Legendary Creature — Human Druid {G}{W}, 1/3
// (Kamigawa: Neon Dynasty, S82):
//
//	"{T}: Put a lore counter on each Saga you control. Activate only as a
//	 sorcery.
//	 When Satsuki dies, choose up to one —
//	 • Return target Saga or enchantment creature you control to its
//	   owner's hand.
//	 • Return target Saga card from your graveyard to your hand."
//
// Each lore counter goes on through the ordinary counter pipeline, so a
// Saga that crosses a chapter triggers it (CR 714.2b) and one on its final
// chapter is sacrificed once that ability is done (CR 714.4). The dies
// trigger is a real "choose up to one" with a floor of zero (CR 603.3c),
// each bullet's target announced as the trigger goes on the stack; a target
// that has left does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e41be9ad-3717-4f25-a8e0-c9947a2a0227",
		Name:         "Satsuki, the Living Lore",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: Put a lore counter on each Saga you control. Activate only as a sorcery.",
			Cost:         effects.TapCost(),
			SorcerySpeed: true,
			Effect:       satsukiLoreCounters,
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventLTB},
			AppliesTo: effects.ThisDied,
			Key:       "Satsuki, the Living Lore — choose up to one: return target Saga or enchantment creature you control to its owner's hand; or return target Saga card from your graveyard to your hand",
			Effect:    func(*game.Game, *game.StackItem) error { return nil },
			Modes: effects.ChooseN("Choose up to one", 0, 1,
				effects.ModeDoing("Return target Saga or enchantment creature you control to its owner's hand.",
					effects.TargetPermanent("target Saga or enchantment creature you control", effects.YouControl(),
						effects.Or(effects.HasSubtype("Saga"), effects.And(effects.Enchantment(), effects.Creature()))),
					effects.BounceTheModesTarget),
				effects.ModeDoing("Return target Saga card from your graveyard to your hand.",
					effects.TargetCardInGraveyard("target Saga card from your graveyard", effects.YouOwn(), effects.HasSubtype("Saga")),
					effects.ReturnTheModesGraveyardTargetToHand),
			),
		}},
	})
}

// satsukiLoreCounters puts a lore counter on each Saga the controller
// controls, one at a time in battlefield order.
func satsukiLoreCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var sagas []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.HasSubtype("Saga") {
			sagas = append(sagas, c)
		}
	}
	for _, c := range sagas {
		if err := (effects.AddCounter{Target: c.InstanceID, Kind: game.CounterLore, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const avB30aLostInMemoriesGrant = "avatar-b30/lost-in-memories-flashback"

// Lost in Memories — Enchantment — Aura {1}{R} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Flash
//	 Enchant creature you control
//	 Enchanted creature gets +1/+1 and has "Whenever this creature deals
//	 combat damage to a player, target instant or sorcery card in your
//	 graveyard gains flashback until end of turn. The flashback cost is
//	 equal to its mana cost.""
//
// Curious Obsession's shape: the pump is PumpAttached, and the trigger is
// GRANTED to the creature (ADR 0093), so it is the creature's ability and
// "your graveyard" is its controller's. The grant is Archmage's Newt's
// trigger: the target is chosen as it goes on the stack, and
// GrantFlashbackToCard prices the flashback at the card's own mana cost and
// exiles the card as it leaves the stack (CR 702.34a). Timing permissions
// still apply, so a sorcery is cast at sorcery speed (the ruling).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "69ac3757-66f8-47d4-af7b-dda66c96c254",
		Name:            "Lost in Memories",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(effects.YouControl()),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: avB30aLostInMemoriesGrant,
			Triggered: []game.TriggeredAbility{effects.Targeting(
				effects.WheneverThisDealsCombatDamageToAPlayer("Lost in Memories — target instant or sorcery card in your graveyard gains flashback",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						ts := ctx.LegalTargets()
						if len(ts) == 0 {
							return nil
						}
						return effects.GrantFlashbackToCard{Target: ts[0].ID, Label: "Flashback — its mana cost (Lost in Memories)"}.Apply(ctx)
					}),
				effects.TargetCardInGraveyard("target instant or sorcery card in your graveyard",
					effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
			)},
			Text: "Whenever this creature deals combat damage to a player, target instant or sorcery card in your graveyard gains flashback until end of turn. The flashback cost is equal to its mana cost.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantAbilitiesToAttached(avB30aLostInMemoriesGrant),
		},
	})
}

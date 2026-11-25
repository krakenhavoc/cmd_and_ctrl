package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demonic Embrace — Enchantment — Aura {1}{B}{B} (Core Set 2021, #5511):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+1, has flying, and is a Demon in addition to
//	 its other types.
//	 You may cast this card from your graveyard by paying 3 life and
//	 discarding a card in addition to paying its other costs."
//
// The aura is Dub's three layers. The graveyard cast is a path that pays the
// printed {1}{B}{B}, 3 life and a discarded card (a cost: refused when the hand
// has no card or the life can't be paid); the Aura goes to the graveyard
// normally when it leaves, so it can be cast again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "d185c829-ef12-403a-a071-7520a0293d4a",
		Name:             "Demonic Embrace",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.EnchantCreature(),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.S82OldB108CastFromGraveyardByLifeAndDiscard("{1}{B}{B}", 3)},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 1),
			effects.GrantToAttached("flying"),
			effects.AttachedIsAlso("Demon"),
		},
	})
}

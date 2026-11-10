package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cracked Skull — Enchantment — Aura {2}{B} (Duskmourn):
//
//	"Enchant creature
//	 When this Aura enters, look at target player's hand. You may choose a
//	 nonland card from it. That player discards that card.
//	 When enchanted creature is dealt damage, destroy it."
//
// The enters trigger is Salt Vampire's look (only you learn the hand,
// ChooseFromRevealedHand.Look), optional, over nonland cards, ending in a
// plain discard. The damage trigger is Sleep Magic's watch on the creature
// the Aura enchants, from any source, combat or not (damage of zero is not
// dealt); it destroys THAT creature, read off the event, so an Aura that has
// left by the time it resolves still destroys it (CR 608.2h). Destroying is
// not preventable by this Aura's absence, and a regeneration shield works.
func init() {
	Register(Spec{
		OracleID:     "4415b65d-5473-4b19-94f7-e5cf4ce64d5e",
		Name:         "Cracked Skull",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Cracked Skull — look at target player's hand; you may choose a nonland card from it, and that player discards it",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						return ChooseFromRevealedHand{
							Player:   TargetedPlayer(ctx),
							Filter:   Nonland(),
							Label:    "nonland card",
							Look:     true,
							Optional: true,
						}.Apply(ctx)
					}),
				TargetPlayer("target player")),
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && source.AttachedTo.Kind == game.TargetCard && source.AttachedTo.ID == ev.Target
			}, "Cracked Skull — destroy enchanted creature", crackedSkullDestroy),
		},
	})
}

// crackedSkullDestroy destroys the creature that was dealt damage.
func crackedSkullDestroy(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return DestroyTarget{Target: item.Trigger.Event.Target}.Apply(NewContext(g, item))
}

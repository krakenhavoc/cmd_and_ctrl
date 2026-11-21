package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starfield of Nyx — Enchantment {4}{W} (Commander Masters, #5948):
//
//	"At the beginning of your upkeep, you may return target enchantment
//	 card from your graveyard to the battlefield.
//	 As long as you control five or more enchantments, each other non-Aura
//	 enchantment you control is a creature in addition to its other types and
//	 has base power and base toughness each equal to its mana value."
//
// The animation is two layers over the same set (CR 613.1d, CR 613.4b): the
// creature type (layer 4) and the base size equal to each permanent's mana
// value (layer 7b). The count of five is read as the layers run and includes
// Starfield itself and Auras. A token enchantment has mana value zero.
//
// No simplification.
func init() {
	upkeep := Targeting(
		AtYourUpkeep("Starfield of Nyx — you may return target enchantment card from your graveyard to the battlefield",
			func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, id := range legalTargetCardIDs(ctx) {
					return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
				}
				return nil
			}),
		TargetCardInGraveyard("target enchantment card in your graveyard", Enchantment(), YouOwn()))
	upkeep.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Starfield of Nyx — return target enchantment card from your graveyard to the battlefield?"}
	Register(Spec{
		OracleID:     "922d603e-6623-4455-b91f-fad76f94045a",
		Name:         "Starfield of Nyx",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{upkeep},
		Static: []game.StaticAbility{
			{
				Label:     "As long as you control five or more enchantments, each other non-Aura enchantment you control is a creature in addition to its other types.",
				Layer:     game.Layer4Type,
				AppliesTo: starfieldOfNyxApplies,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !eotHasType(c.Types, "Creature") {
						c.Types = append(append([]string(nil), c.Types...), "Creature")
					}
				},
			},
			{
				Label:     "Each such enchantment has base power and base toughness each equal to its mana value.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: starfieldOfNyxApplies,
				Apply: func(c *game.Characteristic, target *game.Card, _ *game.Game, _ *game.Card) {
					mv := target.ManaValue()
					c.Power, c.Toughness = mv, mv
				},
			},
		},
	})
}

// starfieldOfNyxApplies is the set: each other non-Aura enchantment the
// source's controller controls, while they control five or more enchantments.
func starfieldOfNyxApplies(target *game.Card, g *game.Game, source *game.Card) bool {
	if target.Controller != source.Controller || target.InstanceID == source.InstanceID ||
		!target.IsEnchantment() || target.IsAura() {
		return false
	}
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == source.Controller && c.IsEnchantment() {
			n++
		}
	}
	return n >= 5
}

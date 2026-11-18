package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cobbled Lancer — Creature — Zombie Horse {U}, 3/3 (Innistrad: Crimson Vow):
//
//	"As an additional cost to cast this spell, exile a creature card from your
//	 graveyard.
//	 {3}{U}, Exile this card from your graveyard: Draw a card."
//
// The additional cost is Corpse Explosion's: the caster names the card at
// cast time and it is exiled with the spell on the stack (CR 601.2b, 601.2h),
// so with no creature card in the graveyard the spell cannot be cast. The
// graveyard ability exiles the card as a cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "200bbf1d-835d-4d61-a6ab-a0b2abfbed79",
		Name:         "Cobbled Lancer",
		Completeness: CompletenessFull,
		AdditionalCost: ExileFromGraveyardCost(1, "a creature card", func(c game.Card) bool {
			return c.IsCreature()
		}),
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}, Exile this card from your graveyard: Draw a card.",
			Purpose: game.Purpose{Draws: 1},
			Cost:    Plus(ManaCost("{3}{U}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}

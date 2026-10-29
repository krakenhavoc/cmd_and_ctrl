package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quinjet Technician — {1}{R} Creature — Human Artificer 2/1:
//
//	"{T}: Add {R}.
//	 {T}: Add {R}{R}. Spend this mana only to activate power-up
//	 abilities."
//
// Two mana abilities, one plain and one restricted. The restriction is
// the purpose tag ManaRestrictPowerUp (#4571, CR 106.6): the {R}{R} pays
// the cost of an activated ability that is a power-up ability
// (game.ManaSpendContext.PowerUp, set from ActivatedAbilityShape.PowerUp
// by the activation, the legal-move enumerator and the cost preview) and
// nothing else. It can't pay a spell, a loyalty or equip cost, another
// activated ability's cost, or a mana ability's own mana cost; a spend
// context that does not say what it pays for refuses it, so the tag can
// only make the mana weaker than printed, never stronger.
//
// No simplification. Marvel Super Heroes (#4455, seam #4571).
func init() {
	Register(Spec{
		OracleID:     "e8025c63-80b4-4cce-bf2d-df294450070d",
		Name:         "Quinjet Technician",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{R}",
				Label:    "{T}: Add {R}.",
			},
			{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     "{R}{R}",
				Label:        "{T}: Add {R}{R}. Spend this mana only to activate power-up abilities.",
				Restrictions: []string{game.ManaRestrictPowerUp},
			},
		},
	})
}

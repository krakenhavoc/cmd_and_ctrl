package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Corrupted Crossroads — Land (Modern Horizons 3 Commander, #5948):
//
//	"{T}: Add {C}. ({C} represents colorless mana.)
//	 {T}, Pay 1 life: Add one mana of any color. Spend this mana only to
//	 cast a spell with devoid."
//
// The second ability's life is a cost (CR 118.3, Horizon Canopy's shape).
// "Only to cast a spell with devoid" is a spend restriction on the mana
// itself (game.ManaRestrictDevoid, CR 106.6): it can pay for a spell that
// has devoid and nothing else, an activated ability and a colourless spell
// that merely prints no colour included. It does not auto-pay (the engine
// never reaches for a restricted source), so a player who wants it clicks
// for it.
//
// Seam: game.ManaRestrictDevoid.
func init() {
	Register(Spec{
		OracleID:     "6276a985-7630-476d-a94f-6c6adc88f6c4",
		Name:         "Corrupted Crossroads",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:         ManaAbilityCost{Tap: true, Life: 1},
				Produced:     "{W|U|B|R|G}",
				Label:        "{T}, Pay 1 life: Add one mana of any color. Spend this mana only to cast a spell with devoid.",
				Restrictions: []string{game.ManaRestrictDevoid},
			},
		},
	})
}

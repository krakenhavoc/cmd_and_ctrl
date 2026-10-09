package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tomik, Orzhov Lawmage — Legendary Creature — Human Advisor {1}{W}, 2/1
// (Reality Fracture, tracker #2795):
//
//	"Flying
//	 Planeswalkers you control have "No more than one creature can attack
//	 this planeswalker each combat."
//	 {T}: Target creature with a +1/+1 counter on it gains flying until end
//	 of turn."
//
// Flying and the tap ability are implemented. The granted attack limit is
// not: the engine reads an attack limit (game.AttackLimit) only off the
// permanent that prints it, and no AbilityGrant slot carries one, so
// "planeswalkers you control have ..." has nowhere to land (The Eternal
// Wanderer prints it on itself). Left out, the card is weaker than printed
// and says so. Waits on a granted-attack-limit slot.
func init() {
	Register(Spec{
		OracleID:     "7a501f7e-eec8-45f7-9ac3-483fd8f5ca5e",
		Name:         "Tomik, Orzhov Lawmage",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Your planeswalkers don't gain \"No more than one creature can attack this planeswalker each combat\" — that part isn't implemented.",
		},
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature with a +1/+1 counter on it gains flying until end of turn.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature with a +1/+1 counter on it", tomikHasPlusCounter),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return GrantKeywordUntilEOT{
						Target:   t.ID,
						Keywords: []string{"flying"},
						Label:    "Tomik, Orzhov Lawmage — gains flying until end of turn",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

// tomikHasPlusCounter is "with a +1/+1 counter on it".
func tomikHasPlusCounter(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.Counters[game.CounterPlusOne] > 0
}

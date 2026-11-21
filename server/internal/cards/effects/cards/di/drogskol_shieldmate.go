package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drogskol Shieldmate — Creature — Spirit Soldier {2}{W}, 2/3 (Innistrad
// Remastered):
//
//	"Flash
//	 When this creature enters, other creatures you control get +0/+1 until
//	 end of turn."
//
// The set of creatures is fixed as the ability resolves (CR 611.2c); this
// creature is excluded by instance.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "470028cf-0b99-4dd1-bb9d-aba91c52ee28",
		Name:            "Drogskol Shieldmate",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Drogskol Shieldmate — other creatures you control get +0/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID)),
						Toughness: 1,
						Label:     "Drogskol Shieldmate — +0/+1",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

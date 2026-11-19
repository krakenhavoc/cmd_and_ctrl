package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "470028cf-0b99-4dd1-bb9d-aba91c52ee28",
		Name:            "Drogskol Shieldmate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Drogskol Shieldmate — other creatures you control get +0/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match:     And(Creature(), YouControl(), OtherThan(item.SourceCardID)),
						Toughness: 1,
						Label:     "Drogskol Shieldmate — +0/+1",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

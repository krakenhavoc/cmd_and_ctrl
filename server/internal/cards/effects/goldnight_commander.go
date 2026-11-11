package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goldnight Commander — Creature — Human Cleric Soldier {3}{W}, 2/2
// (Foundations Commander):
//
//	"Whenever another creature you control enters, creatures you control get
//	 +1/+1 until end of turn."
//
// The trigger is the shared "another creature enters under your control" and
// the pump is Dwarven Lightsmith's: the set of creatures is fixed as the
// ability resolves (CR 611.2c), so the newcomer is included if it is still
// there (the ruling) and a creature that arrives later is not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3dc51dfc-7a60-41bf-b944-6afad6b06c22",
		Name:         "Goldnight Commander",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Goldnight Commander — creatures you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match:     And(Creature(), YouControl()),
						Power:     1,
						Toughness: 1,
						Label:     "Goldnight Commander — +1/+1",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

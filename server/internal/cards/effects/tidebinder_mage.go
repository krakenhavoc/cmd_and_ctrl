package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tidebinder Mage — Creature — Merfolk Wizard {U}{U}, 2/2:
//
//	"When this creature enters, tap target red or green creature an
//	 opponent controls. That creature doesn't untap during its
//	 controller's untap step for as long as you control this creature."
//
// #1313: Dungeon Geists' untap hold with a colour-narrowed target.
func init() {
	Register(Spec{
		OracleID:     "f881378b-b539-4ea8-981d-e01ee82af105",
		Name:         "Tidebinder Mage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("Tidebinder Mage — tap target red or green creature; it doesn't untap while you control Tidebinder Mage",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return TapAndHoldWhileYouControlThis(ctx, holdTargetIDs(ctx))
				}),
			TargetCreature("target red or green creature an opponent controls",
				Or(OfColor("R"), OfColor("G")), OpponentControls()),
		)},
	})
}

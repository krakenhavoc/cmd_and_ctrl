package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Longhorn Sharpshooter — {2}{R} Creature — Minotaur Rogue 3/3:
//
//	"Reach
//	 When this card becomes plotted, it deals 2 damage to any target.
//	 Plot {3}{R} (You may pay {3}{R} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// Waited on #1382: nothing emitted an event when a card became
// plotted, so the trigger — most of the reason to plot this rather than
// cast it — had nothing to watch. game.EventBecomesPlotted is now
// emitted by PlotExiledCardForEffect, the one place both the plot
// special action and "it becomes plotted" effects end, and the trigger
// watches from EXILE, where a plotted card is (WhenThisBecomesPlotted).
// So it also fires when an opponent's Aven Interrupter plots this card
// off the stack — the printed interaction: "this card" becomes
// plotted, whoever did it, and its owner controls the trigger.
//
// "It deals 2 damage": the card in exile is the source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6a0a7b02-10e6-4dbf-8356-659095519480",
		Name:            "Longhorn Sharpshooter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		SpecialActions:  []game.SpecialAction{Plot("{3}{R}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisBecomesPlotted("Longhorn Sharpshooter — 2 damage to any target", longhornSharpshooterDamage),
				TargetAny(),
			),
		},
	})
}

// longhornSharpshooterDamage deals 2 from the plotted card to the
// still-legal target (CR 608.2b — a creature that left in response is
// not hit, and the ability does nothing).
func longhornSharpshooterDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: targets[0].ID, Amount: 2}.Apply(ctx)
}

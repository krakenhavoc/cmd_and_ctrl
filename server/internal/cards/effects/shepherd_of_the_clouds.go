package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shepherd of the Clouds — Creature — Pegasus {4}{W}, 4/3 (Outlaws of
// Thunder Junction, #4468):
//
//	"Flying, vigilance
//	 When this creature enters, return target permanent card with mana
//	 value 3 or less from your graveyard to your hand. Return that card
//	 to the battlefield instead if you control a Mount."
//
// Annie Flash's graveyard target, sent to your hand, or to the
// battlefield under its owner's (your) control when you control a Mount
// as the ability resolves (CR 608.2h). The Mount check is Mobile
// Homestead's, with "Mount" read through the Shepherd's text (CR 612.2).
// The Shepherd is a Pegasus, so it never counts itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c0535854-e4e7-43e4-bc5c-7cc95fb17929",
		Name:            "Shepherd of the Clouds",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("Shepherd of the Clouds — return target permanent card with mana value 3 or less from your graveyard to your hand (to the battlefield if you control a Mount)",
				shepherdOfTheCloudsReturn),
			TargetCardInGraveyard("target permanent card with mana value 3 or less from your graveyard",
				YouOwn(), Permanent(), ManaValueLE(3)),
		)},
	})
}

// shepherdOfTheCloudsReturn returns the still-legal target to its
// owner's hand, or to the battlefield if you control a Mount now.
func shepherdOfTheCloudsReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 || ts[0].Kind != game.TargetCard {
		return nil
	}
	dest := game.ZoneHand
	if b40ControlsA(g, item.Controller, HasSubtype(ctx.TextSubtype("Mount"))) {
		dest = game.ZoneBattlefield
	}
	return ReturnFromGraveyard{Target: ts[0].ID, Dest: dest}.Apply(ctx)
}

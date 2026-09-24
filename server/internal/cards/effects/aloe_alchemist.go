package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aloe Alchemist — {1}{G} Creature — Plant Warlock 3/2:
//
//	"Trample
//	 When this card becomes plotted, target creature gets +3/+2 and
//	 gains trample until end of turn.
//	 Plot {1}{G} (You may pay {1}{G} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// Waited on #1382, like Longhorn Sharpshooter: the trigger watches
// game.EventBecomesPlotted from exile (WhenThisBecomesPlotted), which
// fires for the plot special action and for any "it becomes plotted"
// effect alike, and never for a plain exile.
//
// The pump is two ordinary until-end-of-turn effects on one target, the
// shape Oliphaunt uses. Plotting is a sorcery-speed special action, so
// in practice the pump lands in a main phase before combat.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "97489ef7-98c3-4700-bbe6-185215d41b25",
		Name:            "Aloe Alchemist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SpecialActions:  []game.SpecialAction{Plot("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisBecomesPlotted("Aloe Alchemist — target creature gets +3/+2 and trample", aloeAlchemistPump),
				TargetCreature("target creature"),
			),
		},
	})
}

// aloeAlchemistPump gives the still-legal target +3/+2 and trample until
// end of turn (CR 608.2b — nothing, if it left in response).
func aloeAlchemistPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := (BoostUntilEOT{Target: target, Power: 3, Toughness: 2, Label: "Aloe Alchemist — +3/+2"}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Target:   target,
		Keywords: []string{"trample"},
		Label:    "Aloe Alchemist — trample",
	}.Apply(ctx)
}

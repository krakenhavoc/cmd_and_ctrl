package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deepway Navigator — Creature — Merfolk Wizard {W}{U}, 2/2 (Lorwyn
// Eclipsed, #4458):
//
//	"Flash
//	 When this creature enters, untap each other Merfolk you control.
//	 As long as you attacked with three or more Merfolk this turn,
//	 Merfolk you control get +1/+0."
//
// The untap covers every other Merfolk PERMANENT you control (a
// changeling included), as they are when the trigger resolves. The
// static reads the turn's attack tally (AttackedWithSubtypeThisTurn):
// creatures that were Merfolk as they were declared as attackers, in any
// combat this turn, whether or not they are still attacking or still on
// the battlefield. It covers the Navigator too, and every Merfolk
// creature you control, attacking or not. DependsOnAttackingStatus wakes
// the layer pass on each declaration, and the turn's end clears it.
//
// No simplification.
func init() {
	anthem := TribalAnthem(TribeFilter{Tribes: []string{"Merfolk"}, YoursOnly: true}, 1, 0)
	merfolk := anthem.AppliesTo
	anthem.AppliesTo = func(target *game.Card, g *game.Game, source *game.Card) bool {
		return g.AttackedWithSubtypeThisTurn(source.Controller, source.TextSubtype("Merfolk")) >= 3 && merfolk(target, g, source)
	}
	anthem.DependsOnAttackingStatus = true
	anthem.Label = "As long as you attacked with three or more Merfolk this turn, Merfolk you control get +1/+0."
	Register(Spec{
		OracleID:        "812df594-75af-494f-be09-5f9633aca46a",
		Name:            "Deepway Navigator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{anthem},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Deepway Navigator — untap each other Merfolk you control", deepwayNavigatorUntap),
		},
	})
}

// deepwayNavigatorUntap untaps every other Merfolk permanent the
// trigger's controller controls. "Each other Merfolk you control" is a
// set named by group, so its members are acted on as members
// (asGroupMember), and the word is read through the Navigator's text
// (CR 612.2).
func deepwayNavigatorUntap(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	merfolk := ctx.TextSubtype("Merfolk")
	var ids []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != item.SourceCardID && c.Controller == item.Controller && c.HasSubtype(merfolk) && c.Tapped {
			ids = append(ids, c)
		}
	}
	for _, c := range ids {
		if err := (UntapTarget{Target: c.InstanceID}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

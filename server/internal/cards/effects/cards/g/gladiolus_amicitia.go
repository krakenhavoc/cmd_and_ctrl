package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gladiolus Amicitia — Legendary Creature — Human Warrior {4}{R}{G}, 6/6
// (Final Fantasy, #4462):
//
//	"When Gladiolus Amicitia enters, search your library for a land
//	 card, put it onto the battlefield tapped, then shuffle.
//	 Landfall — Whenever a land you control enters, another target
//	 creature you control gets +2/+2 and gains trample until end of
//	 turn."
//
// The enters trigger is Wight of the Reliquary's land search
// (b11FetchLandTapped): any land card, onto the battlefield tapped, then
// a shuffle, and a search that finds nothing still shuffles. The land it
// fetches is itself a landfall. The landfall pump is Eel-Hounds' attack
// trigger on a landfall: the target is chosen as the trigger is put on
// the stack (CR 603.3d), "another" excludes Gladiolus, and a target gone
// by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb5a05d1-baa5-459d-a2e5-438f4a60b82f",
		Name:         "Gladiolus Amicitia",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Gladiolus Amicitia — search for a land card and put it onto the battlefield tapped",
				effects.B11FetchLandTapped("Gladiolus Amicitia — a land card, onto the battlefield tapped")),
			effects.TriggerWithPurpose(effects.Targeting(
				effects.Landfall("Gladiolus Amicitia — another target creature you control gets +2/+2 and gains trample (landfall)",
					gladiolusLandfallPump),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "trample"))),
		},
	})
}

// gladiolusLandfallPump is "another target creature you control gets
// +2/+2 and gains trample until end of turn".
func gladiolusLandfallPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Gladiolus Amicitia — +2/+2"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Gladiolus Amicitia — trample"}.Apply(ctx)
}

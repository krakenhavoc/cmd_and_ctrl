package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Karrthus, Tyrant of Jund — Legendary Creature — Dragon {4}{B}{R}{G}, 7/7:
//
//	"Flying, haste
//	 When Karrthus enters, gain control of all Dragons, then untap all Dragons.
//	 Other Dragon creatures you control have haste."
//
// The enter trigger takes every Dragon permanent on the battlefield, whoever
// owns it and whatever else it is, for good (CR 611.2a: no duration is
// stated), then untaps every one of them, the ones you already controlled
// included. The set is fixed when the trigger resolves. The haste grant is a
// layer 6 static over the other Dragon creatures you control, read through
// the layers, so a changeling counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "242dfae4-504c-45d4-9a49-f7ab515fd23a",
		Name:            "Karrthus, Tyrant of Jund",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Karrthus, Tyrant of Jund — gain control of all Dragons, then untap all Dragons", karrthusTakeDragons),
		},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Dragon"}, Others: true, YoursOnly: true}, "haste"),
		},
	})
}

func karrthusTakeDragons(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	dragons := effects.MatchingBattlefield(ctx, effects.Subtype("Dragon"))
	for _, c := range dragons {
		if err := (effects.GainControl{
			Target:   c.InstanceID,
			Duration: game.IndefiniteDuration(),
			Label:    "Karrthus, Tyrant of Jund — gain control of all Dragons",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, c := range dragons {
		if err := (effects.UntapTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

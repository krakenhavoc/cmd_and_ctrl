package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Mimeomancer — Creature — Bird Wizard (3/1) for {1}{W}{U}:
//
//	"Flying
//	 At the beginning of your upkeep, you may put a feather counter on target creature. If you do, that creature has base power and toughness 3/1 and has flying for as long as it has a feather counter on it."
//
// ADR 0109 §2 (#1604): "base power and toughness 3/1" (layer 7b) and
// flying (layer 6) for as long as the creature has a feather counter
// on it (game.WhilePinnedHasCounter). The effect is about that
// creature, so it outlasts the Mimeomancer. "If you do" ties it to the
// counter: a declined trigger puts no counter and changes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "176c0ca2-0ca1-45e2-8412-86de5dbead48",
		Name:            "Aven Mimeomancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.AtYourUpkeep("Aven Mimeomancer — put a feather counter on target creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.CounterThenWhileItHasIt(ctx, effects.FirstLegalBattlefieldTarget(ctx), "feather",
						"Aven Mimeomancer — base 3/1 with flying while it has a feather counter",
						game.SetBasePowerMod(3), game.SetBaseToughnessMod(1), game.AddKeywordsMod("flying"))
				}), effects.TargetCreature("target creature")), "Aven Mimeomancer — put a feather counter on target creature?"),
		},
	})
}

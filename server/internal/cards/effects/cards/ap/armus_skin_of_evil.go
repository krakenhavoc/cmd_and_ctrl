package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armus, Skin of Evil — Legendary Creature — Ooze Horror {5}{B}{B}, 6/6
// (Star Trek):
//
//	Menace
//	At the beginning of your end step, target opponent discards two cards.
//	{2}{B}, Discard this card: Target creature gets -4/-4 until end of turn.
//
// The end-step trigger targets the opponent as it goes on the stack, and the
// discard is the opponent's own choice. The last ability works from the hand
// (Stegron the Dinosaur Man's shape): the discard is part of the cost, and a
// target that is gone by resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "42c77c4d-90bb-45d6-a9e0-36704b5a2b11",
		Name:            "Armus, Skin of Evil",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourEndStep("Armus, Skin of Evil — target opponent discards two cards", effects.TargetOpponentDiscards(2)),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Discard this card: Target creature gets -4/-4 until end of turn.",
			Cost:    game.AbilityCost{Mana: "{2}{B}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -4, -4)),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: -4, Toughness: -4, Label: "Armus, Skin of Evil — -4/-4 until end of turn"}.Apply(ctx)
			},
		}},
	})
}

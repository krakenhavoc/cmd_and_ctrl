package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infernal Sovereign — Creature — Demon {4}{B}{B}, 6/6 (March of the
// Machine Commander, #5477):
//
//	"Flying, trample
//	 Skip your draw step.
//	 Whenever you play a land or cast a spell, you draw a card and you
//	 lose 1 life."
//
// The skip is Necropotence's SkipYourDrawStep (CR 614.10, CR 500.11): the
// whole step is skipped, and only its controller's. The trigger is The
// Endstone's one ability watching the land play and the cast on one
// declaration (b31YouPlayedALandOrCastASpell); it draws, then loses 1
// life (Agent Venom's body; life loss is not damage, CR 119.3). A copy of
// a spell is not cast, and a land put onto the battlefield without being
// played does not trigger it. The Sovereign is on the stack when it is
// cast, so it never triggers off its own casting.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b08cb220-34fe-4c54-b654-64a3c24dba44",
		Name:            "Infernal Sovereign",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Replacements:    []game.ReplacementEffect{effects.SkipYourDrawStep()},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.OnAny([]game.EventKind{game.EventZoneMove, game.EventCast},
					func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return effects.B31YouPlayedALandOrCastASpell(ev, source, g)
					},
					"Infernal Sovereign — you draw a card and you lose 1 life",
					effects.Do(effects.DrawCards{N: 1}, effects.GainLife{Amount: -1})),
				game.Purpose{Draws: 1, DeathPayoff: true}),
		},
	})
}

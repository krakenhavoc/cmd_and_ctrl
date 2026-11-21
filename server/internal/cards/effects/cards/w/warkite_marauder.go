package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warkite Marauder — Creature — Human Pirate {1}{U}, 2/1:
//
//	"Flying
//	 Whenever this creature attacks, target creature defending player
//	 controls loses all abilities and has base power and toughness 0/1
//	 until end of turn."
//
// Two scoped layer effects until end of turn: losing abilities (layer 6)
// and base 0/1 (layer 7b), both timestamped when the trigger resolves.
// +1/+1 counters and anthems still apply on top of the new base.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8b566757-e5e4-4d47-b5ef-80ed7d853679",
		Name:            "Warkite Marauder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WithTargetsFrom(
				effects.WheneverThisAttacks("Warkite Marauder — target creature defending player controls loses all abilities and has base power and toughness 0/1 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.ScopedEffectFor{
							Target:   id,
							Mods:     append([]game.Mod{game.LoseAllAbilitiesMod()}, game.SetBasePTMods(0, 1)...),
							Duration: g.UntilEndOfTurnDuration(),
							Label:    "Warkite Marauder — loses all abilities and has base power and toughness 0/1 until end of turn",
						}.Apply(ctx)
					}),
				effects.TargetCreatureDefendingPlayerControls),
		},
	})
}

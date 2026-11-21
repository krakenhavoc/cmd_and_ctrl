package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stalactite Stalker — Creature — Goblin Rogue {B}, 1/1 (The Lost Caverns
// of Ixalan, #4470):
//
//	"Menace
//	 At the beginning of your end step, if you descended this turn, put a
//	 +1/+1 counter on this creature. (You descended if a permanent card was
//	 put into your graveyard from anywhere.)
//	 {2}{B}, Sacrifice this creature: Target creature gets -X/-X until end
//	 of turn, where X is this creature's power."
//
// The end-step trigger is the shared AtYourEndStepIfDescended, an
// intervening if (CR 603.4); the counter goes on the Stalker only while it
// is the same object on the battlefield. The sacrifice is part of the
// cost (CR 602.2b), so X is the Stalker's power as it last existed on the
// battlefield (Aerie Ouphes' read, CR 608.2h), counters and pumps
// included. The -X/-X is locked in at resolution and ends in the cleanup
// step (CR 514.2); a target no longer legal gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c8c28807-8abf-4f91-81f3-1d7d9e856c25",
		Name:            "Stalactite Stalker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfDescended("Stalactite Stalker — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice this creature: Target creature gets -X/-X until end of turn, where X is this creature's power.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				id, legal := effects.B16FirstLegalTargetCard(ctx)
				if !ok || !legal || info.Power <= 0 {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: -info.Power, Toughness: -info.Power,
					Label: "Stalactite Stalker — -X/-X"}.Apply(ctx)
			},
		}},
	})
}

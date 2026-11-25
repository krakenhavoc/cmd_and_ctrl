package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chainweb Aracnir — Creature — Spider {G}, 1/2:
//
//	"Reach
//	 When Chainweb Aracnir enters, it deals damage equal to its power to target creature with flying an opponent controls.
//	 Escape—{3}{G}{G}, Exile four other cards from your graveyard.
//	 Chainweb Aracnir escapes with three +1/+1 counters on it."
//
// The power is read when the trigger resolves, so an escaped Aracnir
// (three counters, power 4) deals 4; one that has left deals its
// last-known power (CR 608.2h).
func init() {
	effects.Register(effects.Spec{
		OracleID:         "c45423ad-1920-46f7-bcc1-ecade5880821",
		Name:             "Chainweb Aracnir",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"reach"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.EscapeWithCounters("{3}{G}{G}", 4, 3)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Chainweb Aracnir — damage equal to its power to target creature with flying an opponent controls", func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					src, ok := ctx.TriggeringPermanent()
					if !ok || len(item.Targets) == 0 {
						return nil
					}
					ref := ctx.Trigger().Object.Ref()
					return effects.DealDamage{SourceObject: &ref, Target: item.Targets[0].ID, Amount: src.Power}.Apply(ctx)
				}),
				effects.TargetCreature("target creature with flying an opponent controls", effects.OpponentControls(), effects.HasKeyword("flying")),
			),
		},
	})
}

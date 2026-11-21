package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brightfield Glider — Creature — Possum Mount {W}:
//
//	"Vigilance
//	 Whenever this creature attacks while saddled, it gets +1/+2 and gains
//	 flying until end of turn.
//	 Saddle 3"
//
// Saddle is Saddle(3) (CR 702.171, ADR 0071 amendment 2026-10-08); the attack
// trigger reads the saddled designation when the attack is declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "acff6e5c-7f25-4ae1-8672-8672450ad844",
		Name:            "Brightfield Glider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []effects.ActivatedAbility{effects.Saddle(3)},
		Triggered: []game.TriggeredAbility{
			effects.AttacksWhileSaddled("Brightfield Glider — +1/+2 and flying until end of turn", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 2, Label: "Brightfield Glider — +1/+2 and flying until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"}, Label: "Brightfield Glider — +1/+2 and flying until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return nil
			}),
		},
	})
}

package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyrewood Gearhulk — Artifact Creature — Construct {2}{R}{R}{G}{G},
// 7/7:
//
//	"Vigilance, menace
//	 When this creature enters, other creatures you control get +2/+2
//	 and gain vigilance and menace until end of turn. Damage can't be
//	 prevented this turn."
//
// The pump and the keyword grant lock their set as the trigger resolves
// (CR 611.2c); the turn grant (ADR 0107 §5) is a rule and covers every
// damage event until cleanup.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c4c6083a-b0b5-4c6b-9f11-9f4fe38417c2",
		Name:            "Pyrewood Gearhulk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pyrewood Gearhulk — others get +2/+2, vigilance and menace; damage can't be prevented", pyrewoodGearhulkEnters),
		},
	})
}

func pyrewoodGearhulkEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	others := effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID))
	if err := (effects.BoostUntilEOT{Match: others, Power: 2, Toughness: 2, Label: "Pyrewood Gearhulk — +2/+2"}).Apply(ctx); err != nil {
		return err
	}
	if err := (effects.GrantKeywordUntilEOT{Match: others, Keywords: []string{"vigilance", "menace"}, Label: "Pyrewood Gearhulk — vigilance and menace"}).Apply(ctx); err != nil {
		return err
	}
	return effects.DamageCantBePreventedThisTurn{}.Apply(ctx)
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "c4c6083a-b0b5-4c6b-9f11-9f4fe38417c2",
		Name:            "Pyrewood Gearhulk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pyrewood Gearhulk — others get +2/+2, vigilance and menace; damage can't be prevented", pyrewoodGearhulkEnters),
		},
	})
}

func pyrewoodGearhulkEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	others := And(Creature(), YouControl(), OtherThan(item.SourceCardID))
	if err := (BoostUntilEOT{Match: others, Power: 2, Toughness: 2, Label: "Pyrewood Gearhulk — +2/+2"}).Apply(ctx); err != nil {
		return err
	}
	if err := (GrantKeywordUntilEOT{Match: others, Keywords: []string{"vigilance", "menace"}, Label: "Pyrewood Gearhulk — vigilance and menace"}).Apply(ctx); err != nil {
		return err
	}
	return DamageCantBePreventedThisTurn{}.Apply(ctx)
}

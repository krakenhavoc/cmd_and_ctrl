package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul of Theros — Creature — Avatar {4}{W}{W}, 6/6:
//
//	"Vigilance
//	 {4}{W}{W}: Creatures you control get +2/+2 and gain first strike and
//	 lifelink until end of turn.
//	 {4}{W}{W}, Exile this card from your graveyard: Creatures you control
//	 get +2/+2 and gain first strike and lifelink until end of turn."
//
// Two rows sharing one body, the graveyard one declared with Zones and an
// ExileThis cost (Soul of New Phyrexia's shape). The creatures are chosen
// once, at resolution (CR 611.2c): one that enters later gets nothing.
func init() {
	Register(Spec{
		OracleID:        "d820b920-191c-43e3-ba6a-4cdceab1c06d",
		Name:            "Soul of Theros",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{
			{
				Label:   "{4}{W}{W}: Creatures you control get +2/+2 and gain first strike and lifelink until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    ManaCost("{4}{W}{W}"),
				Effect:  soulOfTherosPump,
			},
			{
				Label:   "{4}{W}{W}, Exile this card from your graveyard: Creatures you control get +2/+2 and gain first strike and lifelink until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    Plus(ManaCost("{4}{W}{W}"), ExileThis()),
				Zones:   []game.ZoneKind{game.ZoneGraveyard},
				Effect:  soulOfTherosPump,
			},
		},
	})
}

func soulOfTherosPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	const label = "Soul of Theros — creatures you control get +2/+2 and gain first strike and lifelink"
	mine := And(Creature(), YouControl())
	if err := (BoostUntilEOT{Match: mine, Power: 2, Toughness: 2, Label: label}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Match: mine, Keywords: []string{"first strike", "lifelink"}, Label: label}.Apply(ctx)
}

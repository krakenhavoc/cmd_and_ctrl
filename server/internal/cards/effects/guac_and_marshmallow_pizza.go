package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guac & Marshmallow Pizza — {G} Artifact — Food (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Flash
//	 When this artifact enters, target creature gets +2/+2 until end of
//	 turn. Untap it.
//	 {2}, {T}, Sacrifice this artifact: You gain 3 life."
//
// The enters trigger targets as it goes on the stack; on resolution the
// still-legal target (CR 608.2b) gets +2/+2 and is untapped, in printed
// order. The life ability is the printed Food ability on a nontoken
// artifact (samiFoodLifeAbility), and the Food subtype comes off the type
// line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d6d83a51-9d57-44a7-aae3-a96d43ec5ac8",
		Name:            "Guac & Marshmallow Pizza",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Guac & Marshmallow Pizza — target creature gets +2/+2, untap it", guacMarshmallowPizzaPumpAndUntap),
				TargetCreature("target creature")),
		},
		Activated: []ActivatedAbility{samiFoodLifeAbility()},
	})
}

// guacMarshmallowPizzaPumpAndUntap is "target creature gets +2/+2 until
// end of turn. Untap it."
func guacMarshmallowPizzaPumpAndUntap(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Guac & Marshmallow Pizza — +2/+2"}).Apply(ctx); err != nil {
		return err
	}
	return UntapTarget{Target: id}.Apply(ctx)
}

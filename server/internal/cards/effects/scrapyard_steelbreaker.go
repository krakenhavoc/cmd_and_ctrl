package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scrapyard Steelbreaker — Artifact Creature — Human Warrior {3}{R}, 3/4
// (Kamigawa: Neon Dynasty, S82):
//
//	"{1}, Sacrifice another artifact: This creature gets +2/+1 until end of
//	 turn."
//
// Cutthroat Centurion's sacrifice-another cost over artifacts only, with a
// mana payment and no once-a-turn limit. It may be activated any number of
// times; each resolution is a separate +2/+1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "172e7250-94a4-466f-9438-159174925012",
		Name:         "Scrapyard Steelbreaker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice another artifact: This creature gets +2/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), SacrificeAnotherN(1, "another artifact", Artifact())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 1, Label: "Scrapyard Steelbreaker — +2/+1"}.Apply(ctx)
			},
		}},
	})
}

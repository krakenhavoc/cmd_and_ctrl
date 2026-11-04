package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Baxter Stockman — {3}{U}{R} Legendary Creature — Human Scientist 3/3
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"When Baxter Stockman enters, create a 1/1 colorless Robot artifact
//	 creature token.
//	 At the beginning of combat on your turn, target artifact creature you
//	 control gets +3/+0 and gains first strike and vigilance until end of
//	 turn."
//
// The Robot is the set's 1/1 token row (Ravenous Robots). The combat trigger
// is Doc Ock, Evil Inventor's targeted beginning-of-combat trigger; the pump
// is layer 7c (CR 613.4c) and the keywords layer 6 (CR 613.1f), both ending
// at cleanup (CR 514.2). A target that is gone or no longer an artifact
// creature you control as the trigger resolves gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Baxter Stockman — target artifact creature you control gets +3/+0 and gains first strike and vigilance until end of turn"
	Register(Spec{
		OracleID:     "e610bfd4-d064-410b-929c-1d9dd82828b1",
		Name:         "Baxter Stockman",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Baxter Stockman — create a 1/1 colorless Robot artifact creature token",
				Do(CreateToken{Template: TokenCard("1/1 colorless Robot artifact"), N: 1})),
			Targeting(
				AtBeginningOfYourCombat(label, Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
					if err := (BoostUntilEOT{Target: id, Power: 3, Label: "Baxter Stockman — +3/+0"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike", "vigilance"},
						Label: "Baxter Stockman — first strike and vigilance"}.Apply(ctx)
				}))),
				TargetCreature("target artifact creature you control", Artifact(), YouControl())),
		},
	})
}

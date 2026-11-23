package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Syr Vondam, the Lucent — Legendary Creature — Human Knight
// {2}{W}{B}{B}, 4/4 (Edge of Eternities, #4461):
//
//	"Deathtouch, lifelink
//	 Whenever Syr Vondam enters or attacks, other creatures you control
//	 get +1/+0 and gain deathtouch until end of turn."
//
// One ability with two conditions (WhenThisEntersOrAttacks). The pump and
// the grant are Umaro, Raging Yeti's team mode: the affected set is the
// other creatures you control as it resolves (CR 611.2c), so a creature
// that arrives later gets neither, and both end in the cleanup step
// (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "28225bae-2d96-4352-b256-0c2ff4ef65e9",
		Name:            "Syr Vondam, the Lucent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Syr Vondam, the Lucent — other creatures you control get +1/+0 and gain deathtouch until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					others := effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID))
					if err := (effects.BoostUntilEOT{Match: others, Power: 1, Label: "Syr Vondam, the Lucent — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: others, Keywords: []string{"deathtouch"}, Label: "Syr Vondam, the Lucent — deathtouch"}.Apply(ctx)
				}),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brambleguard Veteran — Creature — Raccoon Warrior {1}{G}{G}, 3/4
// (Bloomburrow, #4467):
//
//	"Whenever you expend 4, Raccoons you control get +1/+1 and gain
//	 vigilance until end of turn. (You expend 4 as you spend your fourth
//	 total mana to cast spells during a turn.)"
//
// Expend is WheneverYouExpend (expend.go). The body is Moonstone
// Harbinger's: the Raccoons are the ones you control as the trigger
// resolves, and they keep the bonus for the turn (CR 611.2c), so a
// Raccoon that enters afterwards does not get it. The Veteran is a
// Raccoon and gets it too. The creature type is read through the
// Subtype predicate, which follows a text change (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "bc9c7e1a-5155-431a-be21-fcbd18718980",
		Name:           "Brambleguard Veteran",
		ReadsManaSpent: game.ReadsYourCasts,
		Completeness:   CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouExpend(4, "Brambleguard Veteran — Raccoons you control get +1/+1 and gain vigilance until end of turn",
				brambleguardVeteranPumpRaccoons),
		},
	})
}

func brambleguardVeteranPumpRaccoons(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	raccoons := And(Creature(), YouControl(), Subtype("Raccoon"))
	if err := (BoostUntilEOT{Match: raccoons, Power: 1, Toughness: 1, Label: "Brambleguard Veteran — +1/+1"}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Match: raccoons, Keywords: []string{"vigilance"}, Label: "Brambleguard Veteran — vigilance"}.Apply(ctx)
}

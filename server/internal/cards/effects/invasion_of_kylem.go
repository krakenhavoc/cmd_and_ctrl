package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Kylem — Battle — Siege, defense 5, for {2}{R}{W} (March of
// the Machine, #5477):
//
//	"(As a Siege enters, choose an opponent to protect it. You and
//	  others can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, up to two target creatures each get +2/+0
//	 and gain vigilance and haste until end of turn."
//
// The reminder text is engine behaviour keyed on the card type
// (game/battle.go). This Spec carries the pump and the defeated trigger
// that hands over Valor's Reach Tag Team. The targets are chosen as the
// trigger goes on the stack and each is checked again as it resolves
// (CR 608.2b); the Siege may target any creature, an opponent's too.
//
// No simplification.
func init() {
	const name = "Invasion of Kylem"
	enters := WhenThisEnters(name+" — up to two target creatures each get +2/+0 and gain vigilance and haste until end of turn",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 2, Label: name + " — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"vigilance", "haste"}, Label: name + " — vigilance and haste"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	enters.Targets = TargetCreature("up to two target creatures").WithCount(0, 2)
	Register(Spec{
		OracleID:     invasionOfKylemOracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 5,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			enters,
			DefeatedTrigger(name+" — defeated: exile it, then cast Valor's Reach Tag Team", SiegeDefeated()),
		},
	})
}

// invasionOfKylemOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const invasionOfKylemOracleID = "fee0aa18-5103-4814-94e3-87ad3eaf5a6f"

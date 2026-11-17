package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Muraganda — Battle — Siege, defense 6, for {4}{G} (March of
// the Machine, #5477):
//
//	"(As a Siege enters, choose an opponent to protect it. You and
//	  others can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, put a +1/+1 counter on target creature you
//	 control. Then that creature fights up to one target creature you
//	 don't control."
//
// The reminder text is engine behaviour keyed on the card type
// (game/battle.go). This Spec carries the counter and the fight and the
// defeated trigger that hands over Primordial Plasm. The counter is put
// first, so the fight uses the bigger creature; a creature that left or
// became illegal as the trigger resolves is skipped (CR 608.2b), and the
// fight needs both creatures still there.
//
// No simplification.
func init() {
	const name = "Invasion of Muraganda"
	enters := WhenThisEnters(name+" — put a +1/+1 counter on target creature you control, then it fights up to one target creature you don't control",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			if mine, ok := ctx.ClauseTarget(0); ok && mine.Kind == game.TargetCard {
				if err := (AddCounter{Target: mine.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return fightTargetClauses(ctx)
		})
	enters.Targets = Clauses(
		TargetCreature("target creature you control", YouControl()),
		TargetCreature("up to one target creature you don't control", OpponentControls()).WithCount(0, 1),
	)
	Register(Spec{
		OracleID:     invasionOfMuragandaOracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 6,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			enters,
			DefeatedTrigger(name+" — defeated: exile it, then cast Primordial Plasm", SiegeDefeated()),
		},
	})
}

// invasionOfMuragandaOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const invasionOfMuragandaOracleID = "2e0d12dc-55fa-4363-b28f-612e7cfa12c6"

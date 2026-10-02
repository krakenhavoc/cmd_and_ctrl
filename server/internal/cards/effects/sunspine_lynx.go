package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunspine Lynx — Creature — Elemental Cat {2}{R}{R}, 5/4:
//
//	"Players can't gain life.
//	 Damage can't be prevented.
//	 When this creature enters, it deals damage to each player equal to
//	 the number of nonbasic lands that player controls."
//
// Leyline of Punishment's two statics (ADR 0107 §5) on a creature, and
// an enters trigger that counts each player's nonbasic lands as it
// resolves. A player with none is dealt no damage (CR 120.8).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:              "479adf64-f42c-4cb6-bf25-648c044bca8b",
		Name:                  "Sunspine Lynx",
		Completeness:          CompletenessFull,
		CantGainLife:          PlayersCantGainLife(),
		DamageCantBePrevented: DamageCantBePreventedStatic(),
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sunspine Lynx — damage to each player equal to their nonbasic lands", sunspineLynxEnters),
		},
	})
}

func sunspineLynxEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	nonbasic := NonbasicLand()
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		n := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == p.ID && nonbasic(g, item.Controller, c) {
				n++
			}
		}
		if err := (DealDamage{Source: item.SourceCardID, Target: p.ID, Amount: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Full Moon's Rise — Enchantment {1}{G}:
//
//	"Werewolf creatures you control get +1/+0 and have trample.
//	 Sacrifice this enchantment: Regenerate all Werewolf creatures you
//	 control."
//
// The lord half is TribalAnthem and TribalKeywordGrant over one
// TribeFilter (Werewolves you control), so a changeling counts. The
// sacrifice is the cost; the shields go on the Werewolves you control as
// the ability resolves (CR 701.19, CR 611.2c), one per creature.
//
// No simplification.
func init() {
	f := TribeFilter{Tribes: []string{"Werewolf"}, YoursOnly: true}
	Register(Spec{
		OracleID:     "3aef097b-711a-4463-a59c-831e4dd5e69a",
		Name:         "Full Moon's Rise",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(f, 1, 0),
			TribalKeywordGrant(f, "trample"),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Regenerate all Werewolf creatures you control.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				werewolf := ctx.TextSubtype("Werewolf")
				var ids []game.Card
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.IsCreature() && c.HasSubtype(werewolf) {
						ids = append(ids, c)
					}
				}
				for _, c := range ids {
					if err := (Regenerate{Target: c.InstanceID}).Apply(ctx.asGroupMember()); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

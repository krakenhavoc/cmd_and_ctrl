package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Karn, Legacy Reforged — Legendary Artifact Creature — Golem {5}, */*:
//
//	"Karn's power and toughness are each equal to the greatest mana
//	 value among artifacts you control.
//	 At the beginning of your upkeep, add {C} for each artifact you
//	 control. This mana can't be spent to cast nonartifact spells. Until
//	 end of turn, you don't lose this mana as steps and phases end."
//
// The mana is restricted (ManaRestrictNotNonartifactSpell) and carries
// the keep mark (#2166), so it survives into the main phases. The
// restricted mana is invisible to the auto-tapper, so it is spent by hand
// — weaker than printed, never stronger. P/T is a layer 7a CDA.
func init() {
	Register(Spec{
		OracleID:     "bc8802db-8c94-4eb4-817f-5f4f5408b7a4",
		Name:         "Karn, Legacy Reforged",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, p := range g.BattlefieldCardsForEffect() {
					if p.Controller == source.Controller && p.IsArtifact() && p.ManaValue() > n {
						n = p.ManaValue()
					}
				}
				c.Power, c.Toughness = n, n
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Karn, Legacy Reforged — add {C} for each artifact you control", func(g *game.Game, item *game.StackItem) error {
				n := 0
				for _, p := range g.BattlefieldCardsForEffect() {
					if p.Controller == item.Controller && p.IsArtifact() {
						n++
					}
				}
				if n == 0 {
					return nil
				}
				return AddMana{
					Produced:     strings.Repeat("{C}", n),
					Restrictions: []string{ManaRestrictNotNonartifactSpell},
					Riders:       []game.ManaSpendRider{KeepManaUntilEndOfTurn()},
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

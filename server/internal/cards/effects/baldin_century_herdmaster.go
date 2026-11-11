package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baldin, Century Herdmaster — Legendary Creature — Human Warrior {4}{W}{W},
// 0/7 (Foundations Commander):
//
//	"During your turn, each creature assigns combat damage equal to its
//	 toughness rather than its power.
//	 Whenever Baldin attacks, up to one hundred target creatures each get +0/+X
//	 until end of turn, where X is the number of cards in your hand."
//
// The first line is Felothar's Characteristic.DamageByToughnessAlways
// (CR 510.1c: no power changes, only the amount assigned) on EVERY creature,
// yours and your opponents', while the Baldin's controller is the active player;
// it is a layer 6 static re-asked on every layer pass (Anara, Wolvid Familiar's
// "during your turn"). The attack trigger targets up to a hundred creatures (a
// clause with no minimum, CR 115.1d) and gives each +0/+X, X being your hand
// size as the ability resolves (CR 608.2h), the same for every target. A target
// that has become illegal is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "14493624-fdc7-4630-9b1e-f35668651108",
		Name:         "Baldin, Century Herdmaster",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "During your turn, each creature assigns combat damage equal to its toughness rather than its power.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && isActivePlayer(g, source.Controller)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.DamageByToughnessAlways = true
			},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Baldin, Century Herdmaster — up to one hundred target creatures each get +0/+X until end of turn, X the number of cards in your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					x := 0
					if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Hand != nil {
						x = p.Hand.Size()
					}
					if x <= 0 {
						return nil
					}
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (BoostUntilEOT{Target: t.ID, Toughness: x, Label: "Baldin, Century Herdmaster — +0/+X"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
				TargetCreature("up to one hundred target creatures").WithCount(0, 100)),
		},
	})
}

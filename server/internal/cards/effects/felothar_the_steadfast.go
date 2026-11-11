package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Felothar the Steadfast — Legendary Creature — Human Warrior {1}{W}{B}{G},
// 0/5 (Tarkir: Dragonstorm Commander):
//
//	"Each creature you control assigns combat damage equal to its toughness
//	 rather than its power.
//	 Creatures you control can attack as though they didn't have defender.
//	 {3}, {T}, Sacrifice another creature: Draw cards equal to the sacrificed
//	 creature's toughness, then discard cards equal to its power."
//
// The damage clause is Characteristic.DamageByToughnessAlways, Ghalta the
// Immovable's flag without the "when greater" test (CR 510.1c): no power
// changes, only the amount assigned. "Can attack as though they didn't have
// defender" is Ghalta's layer-6 removal of the keyword. The ability reads
// the sacrificed creature's last known toughness and power (CR 608.2h) and
// discards only after every draw has resolved.
//
// Seam: Characteristic.DamageByToughnessAlways.
func init() {
	yourCreatures := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsCreature()
	}
	Register(Spec{
		OracleID:     "2b3f791a-d672-46fb-a03a-96e19f2c3121",
		Name:         "Felothar the Steadfast",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer6Ability,
			Label:     "Each creature you control assigns combat damage equal to its toughness rather than its power.",
			AppliesTo: yourCreatures,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.DamageByToughnessAlways = true
			},
		}, {
			Layer:     game.Layer6Ability,
			Label:     "Creatures you control can attack as though they didn't have defender.",
			AppliesTo: yourCreatures,
			Apply:     loseDefenderApply,
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice another creature: Draw cards equal to the sacrificed creature's toughness, then discard cards equal to its power.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{3}"), TapCost(), SacrificeAnotherN(1, "another creature", Creature())),
			Effect: Do(Instruction(func(ctx *Context) error {
				fed, ok := b17PermanentSacrificedToPay(ctx.Game, ctx.Item)
				if !ok {
					return nil
				}
				draw, discard := departedCreatureToughness(ctx.Game, fed), departedCreaturePower(ctx.Game, fed)
				if draw <= 0 && discard <= 0 {
					return nil
				}
				return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), draw, discard, "")
			})),
		}},
	})
}

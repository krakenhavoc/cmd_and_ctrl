package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lake Silencio — Plane — Earth (Doctor Who, #4870):
//
//	"Still Point in Time — All spells have split second. (As long as a
//	 spell with split second is on the stack, players can't cast spells
//	 or activate abilities that aren't mana abilities.)
//	 Whenever chaos ensues, Lake Silencio deals 6 damage to target
//	 creature an opponent controls. If a creature dealt damage this way
//	 would die this turn, exile it instead."
//
// A plane (ADR 0216 §6). The static is a split-second grant that reaches
// every spell, whoever controls it (game.SplitSecondGrant.AllSpells), read
// live from the command zone (CR 901.7, CR 611.3a): planeswalking away
// from Lake Silencio takes split second off every spell on the stack, as
// the ruling says. The chaos ability (CR 311.7) is Lava Coil's body with
// the plane as the damage source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4dffc1cb-e9bd-4410-9b08-184eabd2c4da",
		Name:         "Lake Silencio",
		Completeness: CompletenessFull,
		SplitSecondGrants: []game.SplitSecondGrant{{
			AllSpells: true,
			Label:     "Still Point in Time — All spells have split second.",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("Lake Silencio — 6 damage to target creature an opponent controls; exile it if it would die this turn",
				func(g *game.Game, item *game.StackItem) error {
					return damageFirstTargetExileIfItDies(fixedAmount(6))(item, NewContext(g, item))
				}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

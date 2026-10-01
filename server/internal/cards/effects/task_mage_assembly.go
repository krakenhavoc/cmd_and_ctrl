package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Task Mage Assembly — Enchantment {2}{R}:
//
//	"When there are no creatures on the battlefield, sacrifice this
//	 enchantment.
//	 {2}: This enchantment deals 1 damage to target creature. Any player
//	 may activate this ability but only as a sorcery."
//
// ADR 0107 §1 (#1858), the second of the two cards the state-trigger
// row was filed with:
//
//   - The sacrifice is a CR 603.8 state trigger over every creature on
//     the battlefield, anyone's. Casting the Assembly onto an empty
//     board triggers it at once.
//   - The ping is an any-player row (ADR 0106 PR 6, CR 602.2): the {2} is
//     the activator's to pay, "only as a sorcery" is the activator's own
//     main phase with an empty stack (CR 307.1), and the Assembly deals
//     the damage.
//
// No purpose for the bot: no ActivationPurpose field describes "1
// damage to a creature of your choice", so a bot never pays for another
// player's Assembly.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dfbbfbe1-d6c1-4b5b-8d09-949d70b9ff42",
		Name:         "Task Mage Assembly",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThereAreNo(Creature(), "Task Mage Assembly — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
		Activated: []ActivatedAbility{{
			Label:        "{2}: This enchantment deals 1 damage to target creature. Any player may activate this ability but only as a sorcery.",
			Cost:         game.AbilityCost{Mana: "{2}"},
			Targets:      TargetCreature("target creature"),
			SorcerySpeed: true,
			AnyPlayer:    true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}

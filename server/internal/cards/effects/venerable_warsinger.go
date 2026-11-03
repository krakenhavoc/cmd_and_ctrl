package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venerable Warsinger — Creature — Spirit Cleric {1}{R}{W}, 3/3
// (Secrets of Strixhaven Commander, #4834):
//
//	"Vigilance, trample
//	 Whenever this creature deals combat damage to a player, you may
//	 return target creature card with mana value X or less from your
//	 graveyard to the battlefield, where X is the amount of damage this
//	 creature dealt to that player."
//
// X is the damage event's own Amount, so a Warsinger that dealt 3 can reach
// a creature card of mana value 3 or less. The bound rides the target clause
// as data (ManaValueAtMostTheEventAmount): it is fixed when the trigger
// fires, so the target is chosen under it (CR 603.3d) and re-checked against
// it on resolution (CR 608.2b). First-strike and regular damage are two
// events and two triggers, each with its own X. "Your graveyard" is the
// controller's own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e286d1aa-bb74-4568-98b5-a693ac7a369d",
		Name:            "Venerable Warsinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				WheneverThisDealsCombatDamageToAPlayer(
					"Venerable Warsinger — return a creature card with mana value X or less from your graveyard to the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						return ReturnFromGraveyard{Target: item.Targets[0].ID, Dest: game.ZoneBattlefield}.Apply(NewContext(g, item))
					}),
				ManaValueAtMostTheEventAmount(TargetCardInGraveyard(
					"target creature card with mana value X or less in your graveyard",
					YouOwn(), Creature())),
			), "Venerable Warsinger — return a creature card from your graveyard to the battlefield?"),
		},
	})
}

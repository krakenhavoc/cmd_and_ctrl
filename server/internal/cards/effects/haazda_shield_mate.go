package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Haazda Shield Mate — Creature — Human Soldier {2}{W}:
//
//	"At the beginning of your upkeep, sacrifice this creature unless you pay {W}{W}.
//	 {W}: The next time a source of your choice would deal damage to you this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
// The upkeep payment is Stasis's shape: pay {W}{W} or sacrifice it.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "49b689a3-9197-4a04-a62f-218b245d6e23",
		Name:         "Haazda Shield Mate",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Haazda Shield Mate — sacrifice unless you pay {W}{W}", func(g *game.Game, item *game.StackItem) error {
				sourceID := item.SourceCardID
				return UpkeepPayUnless{
					Chooser:  item.Controller,
					Cost:     "{W}{W}",
					Question: "Haazda Shield Mate — pay {W}{W} or sacrifice Haazda Shield Mate?",
					OnDecline: func(ctx *Context) error {
						return SacrificePermanent{Target: sourceID}.Apply(ctx)
					},
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"{W}: The next time a source of your choice would deal damage to you this turn, prevent that damage.",
			ManaCost("{W}"), nil, PreventNextDamageFromChosenSource(ShieldYou))},
	})
}

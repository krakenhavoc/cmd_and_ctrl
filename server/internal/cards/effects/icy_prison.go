package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Icy Prison — Enchantment {U}{U}:
//
//	"When this enchantment enters, exile target creature.
//	 At the beginning of your upkeep, sacrifice this enchantment unless
//	 any player pays {3}.
//	 When this enchantment leaves the battlefield, return the exiled
//	 card to the battlefield under its owner's control."
//
// The exile and the return are Oblivion Ring's (#1729): the return is
// CR 610.3's one-shot, performed as the enchantment leaves, and a
// Prison that leaves before its trigger resolves exiles nothing (CR
// 610.3b).
//
// ADR 0150 (#1903): the upkeep's {3} is offered to every player in
// APNAP order from the active player, so the Prison's controller,
// whose upkeep it is, is asked first (AnyPlayerPays). If nobody pays
// the Prison is sacrificed, which returns the card. The offer holds
// the upkeep until it is answered. Unlike the rest of the family the
// payment keeps the controller's permanent, so the offer says so
// (HelpsController) and a bot controller pays it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4e2cc1a-4212-4284-9ebe-d3c40931de50",
		Name:         "Icy Prison",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Targets:   TargetCreature("target creature"),
				Key:       "Icy Prison — exile target creature",
				Effect:    exileChosenTargetUntilThisLeaves("Icy Prison — the exiled card returns when Icy Prison leaves the battlefield"),
			},
			AtYourUpkeep("Icy Prison — sacrifice it unless any player pays {3}",
				func(g *game.Game, item *game.StackItem) error {
					return AnyPlayerPays{
						Cost:            "{3}",
						Question:        "Icy Prison — pay {3} to keep it on the battlefield?",
						HelpsController: true,
						OnNobody:        sacrificeThisIfStillHere,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

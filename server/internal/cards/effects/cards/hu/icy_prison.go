package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d4e2cc1a-4212-4284-9ebe-d3c40931de50",
		Name:         "Icy Prison",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.Self,
				Targets:   effects.TargetCreature("target creature"),
				Key:       "Icy Prison — exile target creature",
				Effect:    effects.ExileChosenTargetUntilThisLeaves("Icy Prison — the exiled card returns when Icy Prison leaves the battlefield"),
			},
			effects.AtYourUpkeep("Icy Prison — sacrifice it unless any player pays {3}",
				func(g *game.Game, item *game.StackItem) error {
					return effects.AnyPlayerPays{
						Cost:            "{3}",
						Question:        "Icy Prison — pay {3} to keep it on the battlefield?",
						HelpsController: true,
						Then:            icyPrisonSacrificed,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// icyPrisonSacrificed is "if no one does", keyed (#4264).
var icyPrisonSacrificed = effects.AnyPlayerConsequences("any-player/icy-prison-sacrifice",
	nil,
	func(ctx *effects.Context, _ game.AnyPlayerAnswer) error { return effects.SacrificeThisIfStillHere(ctx) },
	nil)

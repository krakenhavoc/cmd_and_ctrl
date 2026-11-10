package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const deconstructionHammerGrant = "deconstruction-hammer/destroy"

// Deconstruction Hammer — Artifact — Equipment {W} (The Lost Caverns of
// Ixalan, #4470):
//
//	"Equipped creature gets +1/+1 and has "{3}, {T}, Sacrifice
//	 Deconstruction Hammer: Destroy target artifact or enchantment."
//	 Equip {1}"
//
// The quoted ability is an ADR 0093 grant: the equipped creature's own
// ability, so the creature taps to pay (and CR 302.6 applies). "Sacrifice
// Deconstruction Hammer" is the grantor-sacrifice cost
// (game.AbilityCost.SacrificeGrantor, ExileGrantor's sibling): the
// Equipment is sacrificed through the ordinary sacrifice payment, so a
// creature whose controller doesn't control the Hammer can't pay it (the
// card's ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ccc2eb6d-21cf-45cf-b4e4-4b2a2517136a",
		Name:         "Deconstruction Hammer",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: deconstructionHammerGrant,
			Activated: []ActivatedAbility{{
				Label:   "{3}, {T}, Sacrifice Deconstruction Hammer: Destroy target artifact or enchantment.",
				Cost:    Plus(ManaCost("{3}"), TapCost(), SacrificeTheGrantor()),
				Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return destroyTheTargetPermanent(item, NewContext(g, item))
				},
			}},
			Text: "{3}, {T}, Sacrifice Deconstruction Hammer: Destroy target artifact or enchantment.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantAbilitiesToAttached(deconstructionHammerGrant),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Treasure Vault — Artifact Land:
//
//	"{T}: Add {C}.
//	 {X}{X}, {T}, Sacrifice this land: Create X Treasure tokens."
//
// The simplest honest test of {X} on an activated ability, and the
// reason it is worth having: "{X}{X}" is TWO X slots, so X=3 costs
// six mana and makes three Treasures. A cost model that stored a
// bool ("this ability has an X") rather than reading the parsed
// slot count would charge three and be exactly half wrong, in the
// player's favour, on the only card in the catalog that notices.
//
// The X the effect reads is the same X the cost charged — announced
// at CR 602.2b, locked onto the stack item, read back through
// ctx.X(). It cannot drift between payment and resolution, which is
// what makes "create X Treasure tokens" a fact about the
// announcement rather than about how much mana is lying around when
// the ability resolves.
//
// Land, so the sacrifice cost is paid at announce and the Treasures
// arrive when the ability resolves — a window in which the Vault is
// already gone. That is as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c43efd6-b1a8-452c-ae20-9a936c3340ab",
		Name:         "Treasure Vault",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{X}{X}, {T}, Sacrifice this land: Create X Treasure tokens.",
			Cost:  Plus(ManaCost("{X}{X}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return CreateToken{
					Controller: item.Controller,
					Template:   TreasureToken(),
					N:          ctx.X(),
				}.Apply(ctx)
			},
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ajani's Aid — Enchantment {2}{G}{W}:
//
//	"When this enchantment enters, you may search your library and/or graveyard for a card named Ajani, Valiant Protector, reveal it, and put it into your hand. If you search your library this way, shuffle.
//	 Sacrifice this enchantment: Prevent all combat damage a creature of your choice would deal this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the second ability is a
// preventFromSource shield against a creature chosen as the ability
// resolves (CR 609.7a; it does not target, its ruling), its combat damage
// prevented for the rest of the turn. The choose_source prompt offers
// creatures only, and the property is rechecked as the damage would be
// dealt (CR 615.9). It may be activated with no creature around at all.
//
// The enters trigger asks "you may" first, so declining neither searches
// nor shuffles.
//
// The search reads the library AND the graveyard (#3217,
// SearchLibrary.AlsoGraveyard). A player who says yes may search either
// zone or both, so finding nothing stays legal even with an Ajani in the
// graveyard; taking it skips the shuffle, and taking one from the
// library (or finding nothing) shuffles (declared simplification:
// finding nothing counts as a library search; a no-op in play).
func init() {
	Register(Spec{
		OracleID:     "b3779fbe-7701-433b-afad-bebd6096a3c8",
		Name:         "Ajani's Aid",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, b06SelfETB, "Ajani's Aid — search for Ajani, Valiant Protector",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Ajani's Aid — search your library and/or graveyard for a card named Ajani, Valiant Protector?",
						Then:     maySearchNamedLibraryOrGraveyard,
						Word:     "Ajani, Valiant Protector",
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"Sacrifice this enchantment: Prevent all combat damage a creature of your choice would deal this turn.",
			SacrificeThis(), nil,
			PreventDamageFromSource{Choose: true, CombatOnly: true, Protect: ShieldAnything, Queries: []game.PermanentQuery{QueryTypes("creature")}})},
	})
}

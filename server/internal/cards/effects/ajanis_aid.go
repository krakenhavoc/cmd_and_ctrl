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
// DECLARED SIMPLIFICATION (weaker than printed): the search looks at your
// library only, the gap Tower Winder and Finale of Devastation declare:
// the engine has no combined library-and-graveyard search.
func init() {
	Register(Spec{
		OracleID:     "b3779fbe-7701-433b-afad-bebd6096a3c8",
		Name:         "Ajani's Aid",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The search looks in your library only. An Ajani, Valiant Protector already in your graveyard can't be found.",
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, b06SelfETB, "Ajani's Aid — search for Ajani, Valiant Protector",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Ajani's Aid — search your library for a card named Ajani, Valiant Protector?",
						OnYes: func(ctx *Context) error {
							return SearchLibrary{
								Player:    ctx.Controller(),
								Predicate: func(c game.Card) bool { return c.Name == "Ajani, Valiant Protector" },
								Dest:      game.ZoneHand,
								Limit:     1,
								Reveal:    true,
								Shuffle:   true,
								Reason:    "Choose a card named Ajani, Valiant Protector to put into your hand",
							}.Apply(ctx)
						},
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"Sacrifice this enchantment: Prevent all combat damage a creature of your choice would deal this turn.",
			SacrificeThis(), nil,
			PreventDamageFromSource{Choose: true, CombatOnly: true, Protect: ShieldAnything, Queries: []game.PermanentQuery{QueryTypes("creature")}})},
	})
}

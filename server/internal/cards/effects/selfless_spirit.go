package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Selfless Spirit — Creature — Spirit Cleric, {1}{W}, 2/1:
//
//	"Flying"
//	"Sacrifice this creature: Creatures you control gain
//	 indestructible until end of turn."
//
// A two-mana flier that is also a Wrath insurance policy, and the
// card that makes the cleanest argument for why indestructible had to
// be enforced at the destruction path rather than anywhere near the
// zone-move plumbing: the ability's COST is a sacrifice, so Selfless
// Spirit dies to pay for the effect that saves everything else. If
// the keyword check had gone into
// `routeBattlefieldCardToOwnerGraveyardLocked` — the shared exit ramp
// — a Spirit that had already granted itself indestructible could not
// have paid its own cost (CR 701.17b: sacrifice is not destruction).
//
// The Spirit sacrifices itself as a cost, so the grant lands on
// everything ELSE you control: it is gone from the battlefield
// before the ability resolves, and `eotSnapshot` reads the
// battlefield at resolution (CR 611.2c). That is the printed
// behaviour.
//
// # Activated, not triggered
//
// `SacrificeSelf` is an `AbilityCost` component, so the engine pays
// it at announce (CR 601.2h — costs are paid before the ability goes
// on the stack, and nothing can be done in response to the payment).
// That means an opponent holding removal cannot kill the Spirit in
// response to save their Wrath: the Spirit is already dead and the
// ability is already on the stack.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "71d785a9-ddc8-472e-a778-a551b444a4bd",
		Name:            "Selfless Spirit",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Indestructible saves a permanent from single-target removal and from lethal damage, but a board wipe (\"destroy all\") still destroys it."},
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label: "Sacrifice this creature: Creatures you control gain indestructible until end of turn.",
			Cost:  game.AbilityCost{SacrificeSelf: true},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"indestructible"},
					Label:    "Selfless Spirit — indestructible",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Secluded Starforge — Land:
//
//	"{T}: Add {C}.
//	 {2}, {T}, Tap X untapped artifacts you control: Target creature
//	 gets +X/+0 until end of turn. Activate only as a sorcery.
//	 {5}, {T}: Create a 2/2 colorless Robot artifact creature token."
//
// A colorless source with a late-game token maker and, in an artifact
// deck, a finisher. The first and third abilities ship as written;
// the middle one does not, and the reason is narrow.
//
// WHAT IS MISSING. #758 gave game.AbilityCost a TapOthers component,
// so "tap an untapped creature you control" (Earthcraft, station) is
// a cost the engine validates and pays. Its count is FIXED, and that
// file states the exclusion in as many words: a VARIABLE count is not
// a matter of reading the field differently, because the number has
// to be announced the way MinX announces X, and that announce path
// does not exist for a tap component. Secluded Starforge is the
// variable case — X is how many artifacts you chose to tap, and the
// pump is that same X.
//
// Declaring the ability with a fixed count would be a different card,
// and declaring it with no tap component at all would be a two-mana
// pump with no price, which is the #259 direction. So the ability is
// left out entirely until the variable-count tap cost lands; the land
// taps for colorless and builds Robots in the meantime.
//
// "Activate only as a sorcery" is ActivatedAbility.SorcerySpeed and
// needs nothing new — it is the cost, not the timing, that is
// blocking.
func init() {
	Register(Spec{
		OracleID:     "69f55a7c-6ddf-412e-b63b-b395731a1ff2",
		Name:         "Secluded Starforge",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The ability that taps artifacts you control to give a creature +X/+0 can't be activated; the mana ability and the Robot ability both work.",
		},
		ManaAbilities: []ManaAbility{painlessColorless()},
		Activated: []ActivatedAbility{{
			Label: "{5}, {T}: Create a 2/2 colorless Robot artifact creature token.",
			Cost:  Plus(ManaCost("{5}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("2/2 colorless Robot artifact"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}

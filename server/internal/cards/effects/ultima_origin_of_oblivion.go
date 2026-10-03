package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ultima, Origin of Oblivion — Legendary Creature — God {5}, 4/4:
//
//	"Flying
//	 Whenever Ultima attacks, put a blight counter on target land. For
//	 as long as that land has a blight counter on it, it loses all land
//	 types and abilities and has '{T}: Add {C}.'
//	 Whenever you tap a land for {C}, add an additional {C}."
//
// ADR 0109 §2 decision 3 (#1604). The attack trigger targets as it goes
// on the stack (CR 603.3d). On resolution the blight counter goes on
// first and the duration is built second, in the order the card prints
// them (CounterThenWhileItHasIt), so a counter that could not be put on
// means the duration never starts and nothing else happens (CR 611.2b).
// The effect is ONE record, pinned to the land and timed by
// game.WhilePinnedHasCounter: game.LoseLandTypesMod in layer 4 (every
// land type goes, every other subtype stays, CR 205.1a), then
// game.LoseAllAbilitiesMod and the {C} grant in layer 6, the removal
// before the grant (ADR 0046 §2). It ends the moment the land's last
// blight counter goes, and for good (CR 611.2b); the land stays a land,
// and keeps its supertypes.
//
// The last line is a triggered mana ability (CR 605.1b): it doesn't use
// the stack (CR 605.4a), and it fires for a land you control whose mana
// ability produced {C} (CR 106.12a), a blighted land included, and adds
// exactly one more {C} however much {C} the land made. The auto-tap
// planner does not count the extra mana (ADR 0074 §7): it may tap one
// land more than it needed, and the surplus floats.
//
// No simplification.
const (
	ultimaOracleID  = "baa337ce-edc6-4ee5-a898-68e9dbb4ab93"
	ultimaColorless = "ultima-origin-of-oblivion/colorless"
	ultimaBlightTag = "Ultima, Origin of Oblivion — put a blight counter on target land; while it has one, it loses all land types and abilities and has \"{T}: Add {C}.\""
)

func init() {
	Register(Spec{
		OracleID:        ultimaOracleID,
		Name:            "Ultima, Origin of Oblivion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []AbilityGrant{
			TapForManaGrant(ultimaColorless, "{C}", "Add {C}", "{T}: Add {C}."),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks(ultimaBlightTag, ultimaBlight),
				TargetPermanent("target land", Land())),
		},
		ManaTriggers: []game.ManaTrigger{{
			Label:     "Ultima, Origin of Oblivion — add an additional {C}",
			AppliesTo: ultimaTappedALandForColorless,
			Produced:  AddsFixedMana("{C}"),
		}},
	})
}

// ultimaBlight is the attack trigger's resolution.
func ultimaBlight(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return CounterThenWhileItHasIt(ctx, FirstLegalBattlefieldTarget(ctx), "blight",
		"Ultima, Origin of Oblivion — no land types or abilities, and \"{T}: Add {C}\", while it has a blight counter",
		game.LoseLandTypesMod(), game.LoseAllAbilitiesMod(), game.GrantAbilitiesMod(ultimaColorless))
}

// ultimaTappedALandForColorless is "Whenever you tap a land for {C}": a
// land Ultima's controller controls, whose mana ability made {C}.
func ultimaTappedALandForColorless(prod game.ManaProduced, source *game.Card, _ *game.Game) bool {
	if !prod.Source.IsLand() || prod.Controller != source.Controller {
		return false
	}
	for _, c := range prod.Colors {
		if c == "C" {
			return true
		}
	}
	return false
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gutter Skulker // Gutter Shortcut (#1855, ADR 0107 §4) — a disturb
// card whose back face is an Aura.
//
// Front face, Creature — Spirit {3}{U}, 3/3:
//
//	"This creature can't be blocked as long as it's attacking alone.
//	 Disturb {3}{U} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Enchantment — Aura:
//
//	"Enchant creature
//	 Enchanted creature can't be blocked as long as it's attacking
//	 alone.
//	 If Gutter Shortcut would be put into a graveyard from anywhere,
//	 exile it instead."
//
// "Attacking alone" is CR 506.5's: attacking while no other creature
// is (AttackingAlone), read when blockers are declared (CR 509.1b). The
// restriction is a block rule on the creature itself (OnSelf) on the
// front face and on the enchanted creature (OnAttached) on the back,
// so the Aura's rule moves with it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         gutterSkulkerOracleID,
		Name:             "Gutter Skulker",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{3}{U}")},
		BlockRules:       []game.BlockRule{CantBeBlockedWhile(OnSelf(), AttackingAlone())},
	})
	Register(Spec{
		OracleID:     gutterSkulkerOracleID + "#1",
		Name:         "Gutter Shortcut",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		BlockRules:   []game.BlockRule{CantBeBlockedWhile(OnAttached(), AttackingAlone())},
		Replacements: []game.ReplacementEffect{DisturbedExile("Gutter Shortcut")},
	})
}

const gutterSkulkerOracleID = "d2d753d4-3bb7-4503-8cc0-8f948b7a461e"

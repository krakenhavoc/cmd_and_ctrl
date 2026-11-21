package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:         GutterSkulkerOracleID,
		Name:             "Gutter Skulker",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Disturb("{3}{U}")},
		BlockRules:       []game.BlockRule{effects.CantBeBlockedWhile(effects.OnSelf(), effects.AttackingAlone())},
	})
	effects.Register(effects.Spec{
		OracleID:     GutterSkulkerOracleID + "#1",
		Name:         "Gutter Shortcut",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		BlockRules:   []game.BlockRule{effects.CantBeBlockedWhile(effects.OnAttached(), effects.AttackingAlone())},
		Replacements: []game.ReplacementEffect{effects.DisturbedExile("Gutter Shortcut")},
	})
}

const GutterSkulkerOracleID = "d2d753d4-3bb7-4503-8cc0-8f948b7a461e"

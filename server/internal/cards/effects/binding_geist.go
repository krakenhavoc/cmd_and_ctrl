package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Binding Geist // Spectral Binding (#1855, ADR 0107 §4) — a disturb
// card whose back face is an Aura.
//
// Front face, Creature — Spirit {2}{U}, 3/1:
//
//	"Whenever this creature attacks, target creature an opponent
//	 controls gets -2/-0 until end of turn.
//	 Disturb {1}{U} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Enchantment — Aura:
//
//	"Enchant creature
//	 Enchanted creature gets -2/-0.
//	 If Spectral Binding would be put into a graveyard from anywhere,
//	 exile it instead."
//
// The attack trigger targets as it goes on the stack (CR 603.3d) and
// the shrink is a layer-7c until-end-of-turn boost on the target, if
// it is still legal. The Aura's -2/-0 is the shared attachment static;
// disturbed, the spell is the Aura (CR 712.8c) and targets through its
// EnchantCreature clause. The exile clause is DisturbedExile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         bindingGeistOracleID,
		Name:             "Binding Geist",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks(bindingGeistLabel, bindingGeistShrink),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
	Register(Spec{
		OracleID:     bindingGeistOracleID + "#1",
		Name:         "Spectral Binding",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{PumpAttached(-2, 0)},
		Replacements: []game.ReplacementEffect{DisturbedExile("Spectral Binding")},
	})
}

const (
	bindingGeistOracleID = "5e1bd17d-3825-45c0-9e7c-6887b7e2cb5c"
	bindingGeistLabel    = "Binding Geist — target creature an opponent controls gets -2/-0 until end of turn"
)

// bindingGeistShrink is the attack trigger's resolution.
func bindingGeistShrink(g *game.Game, item *game.StackItem) error {
	return vsPumpTheTarget(NewContext(g, item), -2, 0, bindingGeistLabel)
}

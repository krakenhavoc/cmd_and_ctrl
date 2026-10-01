package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twinblade Geist // Twinblade Invocation (#1855, ADR 0107 §4) — a
// disturb card whose back face is an Aura.
//
// Front face, Creature — Spirit Warrior {1}{W}, 1/1:
//
//	"Double strike
//	 Disturb {2}{W} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Enchantment — Aura:
//
//	"Enchant creature
//	 Enchanted creature has double strike.
//	 If Twinblade Invocation would be put into a graveyard from
//	 anywhere, exile it instead."
//
// Disturbed, the spell is the Aura (CR 712.8c) and targets through the
// back face's EnchantCreature clause (CR 303.4a). The double strike
// grant is the shared attachment static. The exile clause is
// DisturbedExile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         twinbladeGeistOracleID,
		Name:             "Twinblade Geist",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"double strike"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{2}{W}")},
	})
	Register(Spec{
		OracleID:     twinbladeGeistOracleID + "#1",
		Name:         "Twinblade Invocation",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{GrantToAttached("double strike")},
		Replacements: []game.ReplacementEffect{DisturbedExile("Twinblade Invocation")},
	})
}

const twinbladeGeistOracleID = "4db96d32-b4c2-44e9-a73f-aca7dad279b6"

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lantern Bearer // Lanterns' Lift (#1855, ADR 0107 §4) — a disturb
// card whose back face is an Aura.
//
// Front face, Creature — Spirit {U}, 1/1:
//
//	"Flying
//	 Disturb {2}{U} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Enchantment — Aura:
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has flying.
//	 If Lanterns' Lift would be put into a graveyard from anywhere,
//	 exile it instead."
//
// Disturbed, the spell is the Aura (CR 712.8c) and targets through the
// back face's EnchantCreature clause (CR 303.4a). The pump and the
// flying are the shared attachment statics. The exile clause is
// DisturbedExile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         lanternBearerOracleID,
		Name:             "Lantern Bearer",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{2}{U}")},
	})
	Register(Spec{
		OracleID:     lanternBearerOracleID + "#1",
		Name:         "Lanterns' Lift",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("flying"),
		},
		Replacements: []game.ReplacementEffect{DisturbedExile("Lanterns' Lift")},
	})
}

const lanternBearerOracleID = "67a025eb-6e65-435c-becb-b51085175292"

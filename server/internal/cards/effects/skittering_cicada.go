package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skittering Cicada — Creature — Insect {3}, 2/2 (Commander Masters):
//
//	"Flash
//	 You may cast colorless spells as though they had flash.
//	 Whenever you cast a colorless spell, until end of turn, this creature
//	 gains trample and gets +X/+X, where X is that spell's mana value."
//
// Liberator's cast-timing rule over the colorless filter
// (PermissionFilter.ColorlessOnly, read off the spell's effective
// colours). The trigger is Chrome Host Seedshark's "whenever you cast": X is
// the spell's mana value as it was cast (an {X} counting its chosen value,
// CR 202.3e), and both halves last until end of turn on the Cicada as the
// trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ac141679-6ada-49a8-b225-d1a20c3838db",
		Name:            "Skittering Cicada",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CastTimings: []game.CastTimingRule{
			CastKindAsThoughFlash(game.PermissionFilter{NonLandOnly: true, ColorlessOnly: true},
				"You may cast colorless spells as though they had flash."),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Colorless(),
				"Skittering Cicada — until end of turn, this creature gains trample and gets +X/+X, where X is that spell's mana value",
				skitteringCicadaGrows),
		},
	})
}

func skitteringCicadaGrows(g *game.Game, item *game.StackItem) error {
	if !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	ctx := NewContext(g, item)
	x := triggeringSpellManaValue(g, item)
	if err := (GrantKeywordUntilEOT{
		Target:   item.SourceCardID,
		Keywords: []string{"trample"},
		Label:    "Skittering Cicada — gains trample until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	return BoostUntilEOT{
		Target: item.SourceCardID, Power: x, Toughness: x,
		Label: "Skittering Cicada — +X/+X until end of turn",
	}.Apply(ctx)
}

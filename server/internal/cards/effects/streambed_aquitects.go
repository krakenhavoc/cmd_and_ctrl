package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Streambed Aquitects — Creature — Merfolk Scout {1}{U}{U}, 2/3:
//
//	"{T}: Target Merfolk creature gets +1/+1 and gains islandwalk until
//	 end of turn. (It can't be blocked as long as defending player
//	 controls an Island.)
//	 {T}: Target land becomes an Island until end of turn."
//
// The first ability is one record, a layer-7c +1/+1 and a layer-6
// islandwalk until end of turn (CR 613.4c, 613.1f, 702.14). The second is
// ADR 0109 §1's (#1881) CR 305.7 type set: until end of turn the land is an
// Island and nothing else among land types (its other subtypes stay, CR
// 205.1a), loses the abilities its rules text gives it and taps for {U}
// (CR 305.6). The two together let it make an opponent's land an Island
// for its own Merfolk's islandwalk.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61251125-1379-43e9-8fb1-ce03ca399083",
		Name:         "Streambed Aquitects",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Target Merfolk creature gets +1/+1 and gains islandwalk until end of turn. (It can't be blocked as long as defending player controls an Island.)",
				Cost:    TapCost(),
				Targets: TargetCreature("target Merfolk creature", HasSubtype("Merfolk")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					merfolk := FirstLegalBattlefieldTarget(ctx)
					if merfolk == uuid.Nil {
						return nil
					}
					return ScopedEffectFor{
						Target:   merfolk,
						Mods:     []game.Mod{game.ModifyPTMod(1, 1), game.AddKeywordsMod("islandwalk")},
						Duration: DurationUntilEndOfTurn(ctx),
						Label:    "Streambed Aquitects — +1/+1 and islandwalk",
					}.Apply(ctx)
				},
			},
			{
				Label:   "{T}: Target land becomes an Island until end of turn.",
				Cost:    TapCost(),
				Targets: TargetPermanent("target land", Land()),
				Effect:  TargetLandBecomesUntilEOT("Streambed Aquitects", "Island"),
			},
		},
	})
}

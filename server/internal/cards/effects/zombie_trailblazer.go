package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zombie Trailblazer — Creature — Zombie Scout {B}{B}{B}, 2/2:
//
//	"Tap an untapped Zombie you control: Target land becomes a Swamp
//	 until end of turn.
//	 Tap an untapped Zombie you control: Target creature gains swampwalk
//	 until end of turn. (It can't be blocked as long as defending player
//	 controls a Swamp.)"
//
// Both costs are "tap an untapped Zombie you control", not {T} (CR 302.6):
// the Trailblazer may pay with itself, summoning sick or not, and with any
// other Zombie (tapZombieCost). The land half is ADR 0109 §1's (#1881) CR
// 305.7 type set: until end of turn the land is a Swamp and nothing else
// among land types (its other subtypes stay, CR 205.1a), loses the
// abilities its rules text gives it and taps for {B} (CR 305.6). The
// swampwalk is a layer-6 grant until end of turn (CR 702.14).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5cd96cf2-c1c6-4982-8cf8-70c16704e387",
		Name:         "Zombie Trailblazer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "Tap an untapped Zombie you control: Target land becomes a Swamp until end of turn.",
				Cost:    TapAnUntapped("an untapped Zombie you control", HasSubtype("Zombie")),
				Targets: TargetPermanent("target land", Land()),
				Effect:  TargetLandBecomesUntilEOT("Zombie Trailblazer", "Swamp"),
			},
			{
				Label:   "Tap an untapped Zombie you control: Target creature gains swampwalk until end of turn. (It can't be blocked as long as defending player controls a Swamp.)",
				Cost:    TapAnUntapped("an untapped Zombie you control", HasSubtype("Zombie")),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return GrantKeywordUntilEOT{Target: FirstLegalBattlefieldTarget(ctx), Keywords: []string{"swampwalk"},
						Label: "Zombie Trailblazer — swampwalk"}.Apply(ctx)
				},
			},
		},
	})
}

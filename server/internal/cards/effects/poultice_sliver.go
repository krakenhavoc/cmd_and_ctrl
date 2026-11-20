package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Poultice Sliver — Creature — Sliver {2}{W}, 2/2 (Planar Chaos, #5595):
//
//	"All Slivers have "{2}, {T}: Regenerate target Sliver." (The next time
//	 that Sliver would be destroyed this turn, instead tap it, remove it from
//	 combat, and heal all damage on it.)"
//
// A granted activated ability (ADR 0093) on lrwB1RegenerateTargetRow: the
// shield goes on the still-legal target as the ability resolves (CR
// 701.19a). The {T} keeps a Sliver that came under its controller's control
// this turn from activating it (CR 302.6).
//
// No simplification.
const poulticeSliverGrant = "poultice-sliver/regenerate"

func init() {
	Register(Spec{
		OracleID:     "6be3cf35-f774-4afb-8c1f-5c9b616a2b3f",
		Name:         "Poultice Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: poulticeSliverGrant,
			Activated: []ActivatedAbility{lrwB1RegenerateTargetRow("{2}, {T}: Regenerate target Sliver.",
				Plus(ManaCost("{2}"), TapCost()), TargetPermanent("target Sliver", HasSubtype("Sliver")))},
			Text: "{2}, {T}: Regenerate target Sliver.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}}, poulticeSliverGrant),
		},
	})
}

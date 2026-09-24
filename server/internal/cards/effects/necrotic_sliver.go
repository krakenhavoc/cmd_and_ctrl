package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Necrotic Sliver — Creature — Sliver {1}{W}{B}, 2/2:
//
//	"All Slivers have "{3}, Sacrifice this permanent: Destroy target
//	 permanent.""
//
// A granted ACTIVATED ability (ADR 0093). "This permanent" is the
// Sliver that has the ability — the host pays {3} and is sacrificed,
// and only its controller may activate it, so an opponent's Sliver is
// an opponent's removal spell (CR 602.2).
//
// No simplification.
const necroticSliverGrant = "necrotic-sliver/destroy"

func init() {
	Register(Spec{
		OracleID:     "9655569d-bfa5-4665-9371-9f275b8d223e",
		Name:         "Necrotic Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: necroticSliverGrant,
			Activated: []ActivatedAbility{{
				Label:   "{3}, Sacrifice this permanent: Destroy target permanent",
				Cost:    Plus(ManaCost("{3}"), SacrificeThis()),
				Targets: TargetPermanent("target permanent"),
				Effect:  destroyFirstLegalCardTarget,
			}},
			Text: "{3}, Sacrifice this permanent: Destroy target permanent.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}}, necroticSliverGrant),
		},
	})
}

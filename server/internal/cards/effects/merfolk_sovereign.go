package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merfolk Sovereign — Creature — Merfolk Noble {1}{U}{U}, 2/2
// (EDHREC rank 4127):
//
//	"Other Merfolk creatures you control get +1/+1.
//	 {T}: Target Merfolk creature can't be blocked this turn."
//
// A Merfolk lord that also closes the game. The anthem half is
// ordinary; the tap ability is what a tribal deck actually casts it
// for, because it turns any one Merfolk — usually the one wearing
// every Aura — into unblockable damage for free, every turn.
//
// THE TWO HALVES HAVE DIFFERENT SCOPES, exactly as printed. The
// anthem says "OTHER Merfolk creatures YOU CONTROL", so the Sovereign
// does not pump itself and an opponent's Merfolk get nothing. The tap
// ability says only "target Merfolk creature": no "other", no "you
// control". It can unblock the Sovereign itself, and it can unblock
// an OPPONENT'S Merfolk — which is a real play in a pod, when the
// player you want dead is about to be chump-blocked by the player you
// do not.
//
// The unblockable grant is the S32 until-end-of-turn restriction
// registry, swept at cleanup (CR 514.2) with nothing on the
// battlefield recording it. CR 611.2c pins the creature at
// resolution: one flickered in response comes back a new object
// (CR 400.7) and is blockable again. Restrictions are checked at
// DECLARATION (CR 509.1b), so activating this after blockers are
// declared does nothing — that is the rule, not a gap.
//
// Merfolk is read as an EFFECTIVE subtype on both halves, so a
// changeling is a Merfolk for the anthem and a legal target for the
// ability.
//
// No simplification.
func init() {
	yourOtherMerfolk := TribeFilter{Tribes: []string{"Merfolk"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:     "16aba9af-7a12-4c3d-87eb-06278921cb7c",
		Name:         "Merfolk Sovereign",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(yourOtherMerfolk, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target Merfolk creature can't be blocked this turn.",
			Cost:    TapCost(),
			Targets: TargetCreature("target Merfolk creature", HasSubtype("Merfolk")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				// CR 608.2b: a target that left or stopped being a
				// Merfolk in response is skipped, not an error.
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Merfolk Sovereign — can't be blocked",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

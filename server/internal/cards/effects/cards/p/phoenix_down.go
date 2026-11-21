package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phoenix Down — Artifact {W} (deck request #3655):
//
//	"{1}{W}, {T}, Exile this artifact: Choose one —
//	 • Return target creature card with mana value 4 or less from your
//	   graveyard to the battlefield tapped.
//	 • Exile target Skeleton, Spirit, or Zombie."
//
// Exiling the artifact is a cost, so it is gone by the time either mode
// resolves. Each mode's target is judged again on resolution
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8cc7e418-cc5e-4aef-9757-874223fccd74",
		Name:         "Phoenix Down",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{W}, {T}, Exile this artifact: Choose one — return target creature card with mana value 4 or less from your graveyard to the battlefield tapped; or exile target Skeleton, Spirit, or Zombie.",
			Cost:  effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost(), effects.ExileThis()),
			Modes: effects.ChooseOne(
				effects.ModeDoing("Return target creature card with mana value 4 or less from your graveyard to the battlefield tapped.",
					effects.TargetCardInGraveyard("target creature card with mana value 4 or less in your graveyard",
						effects.YouOwn(), effects.Creature(), effects.ManaValueLE(4)),
					effects.ReturnTheModesTargetToBattlefieldTapped),
				effects.ModeDoing("Exile target Skeleton, Spirit, or Zombie.",
					effects.TargetPermanent("target Skeleton, Spirit, or Zombie", effects.AnySubtype("Skeleton", "Spirit", "Zombie")),
					effects.MurdocksExileTheModesTarget),
			),
		}},
	})
}

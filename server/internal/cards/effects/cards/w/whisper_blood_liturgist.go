package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whisper, Blood Liturgist — Legendary Creature — Human Cleric {3}{B},
// 2/2 (EDHREC rank 3257):
//
//	"{T}, Sacrifice two creatures: Return target creature card from
//	 your graveyard to the battlefield."
//
// Two bodies for the best one in the graveyard. The cost is a tap
// plus a sacrifice clause with a count of two (#747, SacrificeN). The
// clause is "creatures", not "other creatures", so Whisper may be one
// of the two, as on paper. The target is chosen before the cost is
// paid (CR 601.2c before 601.2h), so neither sacrificed creature can
// be the creature returned — Cauldron of Essence's ordering. The card
// returns under its owner's control, which "from your graveyard" makes
// the activator. Whisper taps, so summoning sickness applies (CR
// 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "496ab82e-24a9-4a74-ba4b-992c7309b44a",
		Name:         "Whisper, Blood Liturgist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice two creatures: Return target creature card from your graveyard to the battlefield.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(2, "two creatures", effects.Creature())),
			Targets: effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			Effect:  effects.ReturnFirstLegalGraveyardTargetToBattlefield,
		}},
	})
}

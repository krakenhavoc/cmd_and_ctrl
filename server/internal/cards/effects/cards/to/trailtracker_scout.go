package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trailtracker Scout — Creature — Raccoon Scout {1}{G}, 1/3 (EDHREC
// rank 3396):
//
//	"{T}: Add one mana of any color.
//	 Whenever you expend 8, return up to one target permanent card
//	 from your graveyard to your hand. (You expend 8 as you spend
//	 your eighth total mana to cast spells during a turn.)"
//
// A two-mana rainbow dork with a late-game regrowth attached. The
// mana ability is the Birds of Paradise shape — any colour, the
// printed width, so the pipe is not narrowed to the commander's
// identity — and summoning sickness applies, as the engine enforces
// for a creature's tap ability.
//
// The expend trigger is WheneverYouExpend (expend.go), a per-turn tally
// of mana spent on spells, with the "up to one target permanent card in
// your graveyard" clause declared on the ability: no prompt when the
// graveyard holds none (CR 603.3d), the target re-checked as the trigger
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "d38646dc-d1ea-473c-aaf1-df8ba327459f",
		Name:           "Trailtracker Scout",
		ReadsManaSpent: game.ReadsYourCasts,
		Completeness:   effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouExpend(8, "Trailtracker Scout — return up to one target permanent card from your graveyard to your hand",
					effects.ReturnFirstLegalGraveyardTargetToHand),
				effects.TargetCardInGraveyard("up to one target permanent card from your graveyard", effects.YouOwn(), effects.Permanent()).WithCount(0, 1)),
		},
	})
}

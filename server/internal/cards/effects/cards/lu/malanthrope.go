package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Malanthrope — Creature — Tyranid {1}{G}{U}, 2/2 (Warhammer 40,000
// Commander, S82 GAP batch 193, #5948):
//
//	"Flying
//	 Scavenge the Dead — When this creature enters, exile target player's
//	 graveyard. Put a +1/+1 counter on this creature for each creature card
//	 exiled this way."
//
// The graveyard is exiled as one batch and the counters count the creature
// cards that actually reached exile (a card a replacement kept elsewhere is
// not counted). The counters go on the Malanthrope only while it is still that
// permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0f99e921-61e7-4fc8-8115-4162712b4737",
		Name:            "Malanthrope",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Malanthrope — exile target player's graveyard; a +1/+1 counter on this creature for each creature card exiled this way",
				effects.GapB193ExileTargetGraveyardCountCreatures),
				effects.TargetPlayer("target player")),
		},
	})
}

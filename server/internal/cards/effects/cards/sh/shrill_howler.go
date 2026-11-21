package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shrill Howler // Howling Chorus — {2}{G} Creature — Werewolf Horror 3/1
// // Creature — Eldrazi Werewolf 3/5 (Innistrad Remastered, S82 GAP, #5948):
//
//	Front: "Creatures with power less than this creature's power can't
//	        block it.
//	        {5}{G}: Transform this creature."
//	Back:  "Creatures with power less than this creature's power can't
//	        block it.
//	        Whenever this creature deals combat damage to a player, create a
//	        3/2 colorless Eldrazi Horror creature token."
//
// The block rule is Howlgeist's: the comparison is made as blockers are
// declared, against the creature's power then. The transform is the
// permanent turning over in place (Herbology Instructor's shape, with no
// timing restriction), and the back face's trigger is a plain combat-damage
// trigger around the shared Eldrazi Horror token.
//
// No simplification.
func init() {
	const oracle = "30c3f88f-11d3-4b5f-8936-bab1c65bd693"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Shrill Howler",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.PowerLessThanSource(), effects.OnSelf(),
				"creatures with power less than Shrill Howler's can't block it"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{G}: Transform this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{5}{G}"),
			Effect:  effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Howling Chorus",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.PowerLessThanSource(), effects.OnSelf(),
				"creatures with power less than Howling Chorus's can't block it"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Howling Chorus — create a 3/2 colorless Eldrazi Horror creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/2 colorless Eldrazi Horror"), N: 1})),
		},
	})
}

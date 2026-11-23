package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kykar, Wind's Fury — Legendary Creature — Bird Wizard {1}{U}{R}{W},
// 3/3 (EDHREC rank 3001):
//
//	"Flying
//	 Whenever you cast a noncreature spell, create a 1/1 white Spirit
//	 creature token with flying.
//	 Sacrifice a Spirit: Add {R}."
//
// The Jeskai spellslinger commander. Flying rides PrintedKeywords;
// the cast trigger is b10NoncreatureSpellCastByYou (the spell's type
// read off the stack, where its type line is intact) and makes a
// b28WhiteSpiritFlyingToken; the mana ability sacrifices any Spirit
// permanent the controller controls — the tokens, or a Spirit card —
// for {R}, no tap in the cost, so a Spirit that arrived this turn can
// be cracked at once, and the dies-triggers of the sacrifice land on
// the stack after the mana is in the pool (CR 605.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "71a80491-eea2-4661-9446-26a87efbf4a8",
		Name:            "Kykar, Wind's Fury",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Kykar, Wind's Fury — create a 1/1 white Spirit with flying", effects.Do(effects.CreateToken{Template: effects.B28WhiteSpiritFlyingToken(), N: 1})),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{SacrificeOther: effects.SacrificeSpec("a Spirit", effects.HasSubtype("Spirit"))},
			Produced: "{R}",
			Label:    "Sacrifice a Spirit: Add {R}",
			Answers:  game.AnswerSacOutlet,
		}},
	})
}

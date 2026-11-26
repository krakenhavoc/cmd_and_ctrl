package cardsco

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Council of the Absolute — Creature — Human Advisor {2}{W}{U}, 2/4 (Dragon's
// Maze, S82 OLD b406, #5564):
//
//	"As this creature enters, choose a noncreature, nonland card name.
//	 Your opponents can't cast spells with the chosen name.
//	 Spells with the chosen name you cast cost {2} less to cast."
//
// The as-enters name (Phyrexian Revoker's prompt, CR 614.12) is stored on the
// permanent, and the two statics read it. The prompt has no vocabulary to
// refuse a creature's or a land's name with, so both statics also require the
// spell to be a noncreature spell: a name that "noncreature, nonland" rules out
// does nothing, rather than giving a stronger lock than printed. The lock is
// Meddling Mage's cast restriction narrowed to the source's opponents, and the
// reduction is Anointed Peacekeeper's named-spell discount for the source's
// controller only.
//
// No simplification.
func init() {
	named := func(source, card game.Card) bool {
		n := strings.TrimSpace(source.ChosenName)
		return n != "" && !card.IsCreature() && game.CardNameMatches(card, n)
	}
	effects.Register(effects.Spec{
		OracleID:     "464d5db8-5da1-4ced-aebf-33b3b6cb5bb2",
		Name:         "Council of the Absolute",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCardNameAsEnters("Council of the Absolute — choose a noncreature, nonland card name"),
		CastRestrictions: []game.CastRestriction{{
			Label: "Your opponents can't cast spells with the chosen name.",
			Forbids: func(q game.CastQuery) bool {
				return q.Controller != q.Source.Controller && named(q.Source, q.Card)
			},
		}},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(2, "Spells with the chosen name you cast cost {2} less to cast.",
				func(q game.CostQuery) bool {
					return q.Controller == q.Source.Controller && q.Ability == nil && q.SpecialAction == nil && named(q.Source, q.Card)
				}),
		},
	})
}

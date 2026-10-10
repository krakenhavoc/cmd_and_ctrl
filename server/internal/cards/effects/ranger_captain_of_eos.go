package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ranger-Captain of Eos — Creature — Human Soldier Ranger {1}{W}{W},
// 3/3:
//
//	"When this creature enters, you may search your library for a
//	 creature card with mana value 1 or less, reveal it, put it into
//	 your hand, then shuffle.
//	 Sacrifice this creature: Your opponents can't cast noncreature
//	 spells this turn."
//
// The ETB is a tutor behind the printed "you may" (SearchLibrary's
// Optional, which lets the searcher decline the card and the
// shuffle). The sacrifice ability is Mandate of Peace's cast ban on
// each opponent for the turn, narrowed to noncreature spells through
// the ban's own filter, so a creature spell is still legal. The
// sacrifice is a cost: the ban lands even if the ability is
// responded to by removing the Captain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cada3481-cc2b-4412-b9b5-0436af53aad2",
		Name:         "Ranger-Captain of Eos",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ranger-Captain of Eos — search for a creature card with mana value 1 or less", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return c.IsCreature() && c.ManaValue() <= 1
					},
					Dest:     game.ZoneHand,
					Limit:    1,
					Reveal:   true,
					Shuffle:  true,
					Optional: true,
					Reason:   "Ranger-Captain of Eos — a creature card with mana value 1 or less",
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Your opponents can't cast noncreature spells this turn.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    SacrificeThis(),
			Effect:  rangerCaptainOfEosBanNoncreatureSpells,
		}},
	})
}

func rangerCaptainOfEosBanNoncreatureSpells(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	d := DurationUntilEndOfTurn(ctx)
	for _, opp := range ctx.Opponents() {
		if err := (RestrictCasting{
			Player:   opp,
			Rule:     game.CastBanRule{Kind: game.CastBanOutright, Filter: game.PermissionFilter{NoncreatureOnly: true}},
			Label:    "Ranger-Captain of Eos — can't cast noncreature spells this turn",
			Duration: d,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

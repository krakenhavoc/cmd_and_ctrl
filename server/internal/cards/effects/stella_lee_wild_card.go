package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stella Lee, Wild Card — Legendary Creature — Human Rogue {1}{U}{R}, 2/4
// (Outlaws of Thunder Junction Commander):
//
//	"Whenever you cast your second spell each turn, exile the top card of
//	 your library. Until the end of your next turn, you may play that card.
//	 {T}: Copy target instant or sorcery spell you control. You may choose
//	 new targets for the copy. Activate only if you've cast three or more
//	 spells this turn."
//
// The trigger is the second-spell clause (the per-turn cast tally, bumped
// before the cast event) over the impulse exile every "until the end of your
// next turn" card uses (ExileTopNUntilYourNextTurn). The ability copies a
// spell its controller controls (a copy of a spell is not a cast, so the
// three-spell gate counts only real casts) through copyTargetedSpell with
// new targets offered, and the gate is read at activation (CR 602.5b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d0da2f9-0e05-4a94-8d42-fd57d06b09ba",
		Name:         "Stella Lee, Wild Card",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Impulse: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Stella Lee, Wild Card — exile the top card of your library; you may play it until the end of your next turn",
				func(g *game.Game, item *game.StackItem) error {
					return ExileTopNUntilYourNextTurn(NewContext(g, item), 1)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:     "{T}: Copy target instant or sorcery spell you control. You may choose new targets for the copy. Activate only if you've cast three or more spells this turn.",
			Cost:      TapCost(),
			Targets:   instantOrSorcerySpell("target instant or sorcery spell you control", YouControl()),
			Condition: cmdr2S41CastThreeOrMoreSpellsThisTurn,
			Effect:    copyTargetedSpell,
		}},
	})
}

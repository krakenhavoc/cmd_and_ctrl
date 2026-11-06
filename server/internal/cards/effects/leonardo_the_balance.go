package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leonardo, the Balance — Legendary Creature — Mutant Ninja Turtle {3}{W},
// 3/3 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Whenever a token you control enters, you may put a +1/+1 counter on
//	 each creature you control. Do this only once each turn.
//	 {W}{U}{B}{R}{G}: Creatures you control gain menace, trample, and
//	 lifelink until end of turn.
//	 Partner—Character select"
//
// The trigger is Belladonna Took's "a token you control enters" with
// Pantlaza's once-each-turn limit on the DOING: the "may" is asked as the
// ability resolves (CR 603.5), a "no" leaves the turn's use for a later
// token, and the use is noted (b28NoteDoneThisTurn) only on a "yes". It is
// read as the ability triggers, so no trigger is made once the use is
// spent, and again as it resolves, so the second of two triggers stacked
// before either resolved does nothing. Any token counts, not only a
// creature token. The counters are one placement per creature
// (b11PutCounterOnEachCreatureYouControl), on the creatures you control as
// it resolves.
//
// The activated ability is Spectacular Spider-Man's team grant; the set of
// creatures is fixed as it resolves (CR 611.2c).
//
// Partner—Character select is a deck-construction rule the deck importer
// reads (CR 702.124i).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46ea8d02-b087-44f7-8403-93ffc1d3a8ad",
		Name:         "Leonardo, the Balance",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, AllOf(belladonnaTokenYouControlEntered, leonardoNotYetThisTurn), leonardoLabel,
				func(g *game.Game, item *game.StackItem) error {
					if b28DoneThisTurn(g, item.SourceCardID, leonardoDone) {
						return nil
					}
					b28NoteDoneThisTurn(g, item.SourceCardID, leonardoDone)
					return b11PutCounterOnEachCreatureYouControl(g, item)
				}), "Leonardo, the Balance — put a +1/+1 counter on each creature you control?"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}{U}{B}{R}{G}: Creatures you control gain menace, trample, and lifelink until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Keywords: []string{"menace", "trample", "lifelink"}}, Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{W}{U}{B}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"menace", "trample", "lifelink"},
					Label:    "Leonardo, the Balance — menace, trample and lifelink until end of turn",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

const (
	leonardoLabel = "Leonardo, the Balance — you may put a +1/+1 counter on each creature you control"
	// leonardoDone is the once-each-turn tally's label: the turn's counters
	// have been put.
	leonardoDone = "Leonardo, the Balance — counters put this turn"
)

// leonardoNotYetThisTurn keeps the trigger from being made once the turn's
// use has been spent.
func leonardoNotYetThisTurn(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return source != nil && !b28DoneThisTurn(g, source.InstanceID, leonardoDone)
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Resplendent Angel — Creature — Angel {1}{W}{W}, 3/3 (The Lost Caverns of
// Ixalan, #4470):
//
//	"Flying
//	 At the beginning of each end step, if you gained 5 or more life this
//	 turn, create a 4/4 white Angel creature token with flying and
//	 vigilance.
//	 {3}{W}{W}{W}: Until end of turn, this creature gets +2/+2 and gains
//	 lifelink."
//
// Angelic Accord's trigger at five life with the Angel that has vigilance:
// every end step, the intervening if (CR 603.4) read when it would trigger
// and again as it resolves, over the life this turn's player gained
// whether or not the Angel was on the battlefield for it. One token however
// many times the threshold was passed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ac93092-1a57-4cf7-8f03-1dca86d5c476",
		Name:         "Resplendent Angel",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondLifeGained{AtLeast: 5},
				On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b32EndStepAndYouGainedLifeThisTurnAtLeast(ev, source, g, 5)
				}, "Resplendent Angel — create a 4/4 Angel with flying and vigilance",
					Do(CreateToken{Template: TokenCard("4/4 white Angel with flying and vigilance"), N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}{W}{W}: Until end of turn, this creature gets +2/+2 and gains lifelink.",
			Cost:    ManaCost("{3}{W}{W}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				id := item.SourceCardID
				if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Resplendent Angel — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink"}, Label: "Resplendent Angel — lifelink"}.Apply(ctx)
			},
		}},
	})
}

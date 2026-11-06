package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Needs of the Many — Instant {1}{W} (Star Trek, #4973):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Creatures you control gain hexproof and indestructible until end of
//	 turn. You gain hexproof until end of turn. Prevent all damage that
//	 would be dealt to you this turn. Exile The Needs of the Many."
//
// The sacrifice is the additional cost (SacrificeCost, Bone Splinters'),
// paid as the spell is cast (CR 601.2h), so a spell with no creature to
// sacrifice can't be cast. The resolution is four instructions in printed
// order: Spectacular Spider-Man's mass grant over the creatures you
// control as it resolves (CR 611.2c), Lazotep Plating's player hexproof,
// Riot Control's not-one-use shield on you from every source for the
// rest of the turn, and Temporal Mastery's self-exile in place of the
// graveyard. None of them prompts, so they run in one pass.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "984e0426-10ba-4c17-b26e-ce21100ec398",
		Name:           "The Needs of the Many",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Purpose:        game.Purpose{Answers: game.AnswerProtect | game.AnswerSurvive},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl()),
				Keywords: []string{"hexproof", "indestructible"},
				Label:    "The Needs of the Many — creatures you control gain hexproof and indestructible",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (GainPlayerKeyword{
				Player:   ctx.Controller(),
				Keyword:  KeywordHexproof,
				Label:    "The Needs of the Many — you gain hexproof",
				Duration: DurationUntilEndOfTurn(ctx),
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (PreventDamageFromSource{Protect: ShieldYou}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ocelot Pride — Creature — Cat {W}, 1/1 (EDHREC rank 1097):
//
//	"First strike, lifelink
//	 Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 At the beginning of your end step, if you gained life this turn,
//	 create a 1/1 white Cat creature token. Then if you have the city's
//	 blessing, for each token you control that entered this turn, create
//	 a token that's a copy of it."
//
// The "if you gained life this turn" is INTERVENING (CR 603.4): checked
// as the end step begins and again as the trigger resolves (the turn's
// tally, so a gain in response counts). The blessing is NOT intervening:
// "then if you have the city's blessing" is a clause of the effect and
// is read after the Cat is made.
//
// The set the copies are made from is taken AFTER the Cat exists, so the
// new Cat is one of the tokens that entered this turn and is copied
// too, and BEFORE any copy is made, so a copy does not itself get
// copied (they are created by the loop, not part of the set it walks).
// Tokens count whoever made them, as long as they are yours now and
// entered this turn; a token that has changed hands to you counts only
// if it entered the battlefield this turn, which the entry tally
// answers per object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b4ce2c8f-a19c-461e-8a2e-0ec2bd3ebca3",
		Name:            "Ocelot Pride",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "lifelink", game.KeywordAscend},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b24YourEndStepAndYouGainedLifeThisTurn(ev, source, g)
			}, ocelotPrideLabel, ocelotPrideEndStep),
		},
	})
}

const ocelotPrideLabel = "Ocelot Pride — create a 1/1 Cat; with the city's blessing, copy each token that entered this turn"

func ocelotPrideEndStep(g *game.Game, item *game.StackItem) error {
	if b15LifeGainedThisTurn(g, item.Controller) == 0 {
		return nil // CR 603.4: the intervening "if" is checked again
	}
	ctx := NewContext(g, item)
	if err := (CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Cat"), N: 1}).Apply(ctx); err != nil {
		return err
	}
	if !YouHaveTheCitysBlessing(g, item.Controller) {
		return nil
	}
	var tokens []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && IsToken(c) && g.EnteredThisTurn(c.InstanceID) {
			tokens = append(tokens, c.InstanceID)
		}
	}
	for _, id := range tokens {
		if err := (CreateTokenCopy{Controller: item.Controller, Copy: id, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

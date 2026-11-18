package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Protection Racket — Enchantment {2}{B} (Streets of New Capenna, #5491):
//
//	"At the beginning of your upkeep, repeat the following process for each
//	 opponent in turn order. Reveal the top card of your library. That player may
//	 pay life equal to that card's mana value. If they do, exile that card.
//	 Otherwise, put it into your hand."
//
// The opponents are walked one at a time in turn order (CR 101.4: the active
// player first, so you, then clockwise), each round revealing a fresh top card
// before the next opponent is asked. The opponent's "may pay" is an option-pick
// prompt (decline first, which a player can always take): paying is the whole
// price (CR 119.4 and CR 119.8 gate it: life at least the mana value, and no
// "can't pay life" lock), so a player who can't pay is not asked and the card
// goes to your hand. A mana value of 0 is payable for nothing and still
// exiles. Mana value is read off the card as it is in the library (CR 202.3,
// X as 0). An empty library reveals nothing, so no one is asked.
//
// The walk rides keyed continuations (the card revealed first, then the
// opponents still to come), so a table waiting on an answer is a restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b51edb1-37cc-4993-a8a5-2eb5f79fa191",
		Name:         "Protection Racket",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Protection Racket — for each opponent in turn order, reveal the top card of your library; that player may pay life equal to its mana value to exile it, otherwise it goes to your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var opps []uuid.UUID
					for _, id := range apnapPlayers(g) {
						if id != ctx.Controller() {
							opps = append(opps, id)
						}
					}
					return protectionRacketRound(ctx, opps)
				}),
		},
	})
	protectionRacketAnswered = OptionPickThen("option-pick/protection-racket-answer", func(ctx *Context, r game.OptionPicked) error {
		if len(r.Carry) == 0 || len(r.Nums) == 0 {
			return nil
		}
		card, rest := r.Carry[0], r.Carry[1:]
		if me := ctx.Game.PlayerByIDForEffect(ctx.Controller()); me != nil && me.Library.Contains(card) {
			if r.Index == 1 {
				if err := ctx.Game.PayLifeForEffect(r.Source, r.Chooser, r.Nums[0]); err == nil {
					if err := (ExileTarget{Target: card}).Apply(ctx); err != nil {
						return err
					}
					return protectionRacketRound(ctx, rest)
				}
			}
			if err := (BounceToHand{Target: card}).Apply(ctx); err != nil {
				return err
			}
		}
		return protectionRacketRound(ctx, rest)
	})
}

// protectionRacketAnswered is one opponent's answer, keyed (#4253): the seat
// that answered is the chooser, the carry is the revealed card then the
// opponents still to come, and the number is the life the card costs. An
// on-disk key: never renamed, never reused. Registered in init because the
// walk asks again from inside it.
var protectionRacketAnswered game.OptionPickThen

// protectionRacketRound reveals the next card for the next opponent still in
// the game and asks that opponent, or ends the process.
func protectionRacketRound(ctx *Context, remaining []uuid.UUID) error {
	for len(remaining) > 0 {
		opp, rest := remaining[0], remaining[1:]
		if p := ctx.Game.PlayerByIDForEffect(opp); p == nil || p.Eliminated {
			remaining = rest
			continue
		}
		var ids []uuid.UUID
		if err := (RevealTopOfLibrary{
			N:        1,
			Reason:   "Protection Racket — reveal the top card of your library",
			Revealed: &ids,
		}).Apply(ctx); err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil // nothing to reveal, for this opponent or any after
		}
		card, ok := ctx.Game.LookupCardForEffect(ids[0])
		if !ok {
			return nil
		}
		mv, _ := ctx.Game.ManaValueForEffect(card)
		mv = max(mv, 0)
		payer := ctx.Game.PlayerByIDForEffect(opp)
		if !ctx.Game.CanPayLifeLocked(payer, mv) {
			if err := (BounceToHand{Target: ids[0]}).Apply(ctx); err != nil {
				return err
			}
			remaining = rest
			continue
		}
		return PickOption{
			Player:   opp,
			Question: "Protection Racket — pay " + strconv.Itoa(mv) + " life to exile " + card.Name + "?",
			Options: []game.ChoiceOption{
				{Label: "Don't pay — it goes to the caster's hand", Cards: []uuid.UUID{ids[0]}},
				{Label: "Pay " + strconv.Itoa(mv) + " life — exile it", Cards: []uuid.UUID{ids[0]}, LifeCost: mv},
			},
			ThenKey: protectionRacketAnswered,
			Carry:   append([]uuid.UUID{ids[0]}, rest...),
			Nums:    []int{mv},
		}.Apply(ctx)
	}
	return nil
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lim-Dûl's Hex — Enchantment {1}{B}:
//
//	"At the beginning of your upkeep, for each player, this enchantment
//	 deals 1 damage to that player unless they pay {B} or {3}."
//
// "Unless they pay {B} or {3}" is one choice with three answers: pay
// {B}, pay {3}, or pay nothing and take the damage (CR 118.12a, #2854).
// Each player is asked with an option_pick whose two payments carry
// their mana cost; a payment the player cannot make is not offered
// (CR 118.3), and the chosen one is paid through the auto-tapper.
//
// Every player is asked, the Hex's controller included, the active
// player first and the rest in turn order (CR 101.4). The choices come
// first and the damage after: once the last player has answered, the
// Hex deals 1 damage to each player who did not pay, as one instance of
// damage (DealDamageEachThenForEffect), so the damage is dealt at once
// rather than between questions.
//
// The continuation is a registered key (OptionPickThen), so a table
// waiting on any of the questions is still a restore point. The prompt
// carries the players still to ask, then uuid.Nil, then the players who
// have declined so far.
//
// No simplifications.
func init() {
	limDulsHexAnswer = OptionPickThen("option-pick/lim-duls-hex-pay", limDulsHexAnswered)
	Register(Spec{
		OracleID:     "7f9116e1-9aab-470b-90eb-f51a3fbb3c0e",
		Name:         "Lim-Dûl's Hex",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Lim-Dûl's Hex — 1 damage to each player who doesn't pay {B} or {3}", func(g *game.Game, item *game.StackItem) error {
				return limDulsHexAsk(NewContext(g, item), apnapPlayers(g), nil)
			}),
		},
	})
}

// limDulsHexAnswer is the registered continuation of each question. Its
// key is an on-disk identity: never renamed, never reused.
//
// Assigned in init, not in the declaration: the continuation asks
// the next question, which names this key, and a package-level
// initializer may not refer to itself (an initialization cycle).
var limDulsHexAnswer game.OptionPickThen

// limDulsHexAsk asks the first player in `toAsk` who is still in the
// game, or, with nobody left to ask, deals the damage to `declined`.
func limDulsHexAsk(ctx *Context, toAsk, declined []uuid.UUID) error {
	for i, p := range toAsk {
		pl := ctx.Game.PlayerByIDForEffect(p)
		if pl == nil || pl.Eliminated {
			continue
		}
		carry := append(append([]uuid.UUID(nil), toAsk[i:]...), uuid.Nil)
		carry = append(carry, declined...)
		return PickOption{
			Player:   p,
			Question: "Lim-Dûl's Hex — pay {B} or {3}, or take 1 damage?",
			Options: []game.ChoiceOption{
				{Label: "Pay nothing: take 1 damage"},
				{Label: "Pay {B}", ManaCost: "{B}"},
				{Label: "Pay {3}", ManaCost: "{3}"},
			},
			ThenKey: limDulsHexAnswer,
			Carry:   carry,
		}.Apply(ctx)
	}
	if len(declined) == 0 {
		return nil
	}
	return ctx.Game.DealDamageEachThenForEffect(ctx.Source(), declined, 1, nil)
}

// limDulsHexAnswered records Carry[0]'s answer and asks the next
// player. A player who chose nothing (they left the game) is not dealt
// damage: they are no longer a player (CR 800.4a).
func limDulsHexAnswered(ctx *Context, r game.OptionPicked) error {
	sep := -1
	for i, id := range r.Carry {
		if id == uuid.Nil {
			sep = i
			break
		}
	}
	if sep < 1 {
		return nil
	}
	asked, rest := r.Carry[0], r.Carry[1:sep]
	declined := append([]uuid.UUID(nil), r.Carry[sep+1:]...)
	if r.Option != nil && r.Option.ManaCost == "" {
		declined = append(declined, asked)
	}
	return limDulsHexAsk(ctx, rest, declined)
}

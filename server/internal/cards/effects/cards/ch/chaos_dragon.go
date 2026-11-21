package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chaos Dragon — Creature — Dragon {1}{R}{R}, 4/4 (Forgotten Realms
// Commander):
//
//	"Flying, haste
//	 This creature attacks each combat if able.
//	 At the beginning of combat on your turn, each player rolls a d20. If
//	 one or more opponents had the highest result, this creature can't
//	 attack those players or planeswalkers they control this combat."
//
// The rolls (#5897). Each player still in the game rolls their own d20
// through the engine's one dice door (game.RollDiceThenForEffect), in turn
// order from the active player, so a replacement of one player's roll
// (Pixie Guide, Squid Fire Knight's sticker) applies to that roll alone and
// may pause it for its question; the next player rolls once it is settled.
// The running highest result and the players who reached it ride the keyed
// tail as plain data (its N and Carry), so a table waiting on any of those
// questions is a restore point. A player whose roll kept no die had no
// result.
//
// The restriction. Every opponent whose result equals the highest is
// protected, so a tie protects every opponent in it, and a tie with you
// still protects the opponent. If you alone had the highest, nothing
// happens. The protection is a CR 508.1c restriction on this creature
// (CantAttackPlayerByEffectMod, one per protected opponent) lasting until
// the end of this combat (CR 511.2), read by both declaration verbs, the
// CR 508.1d requirement search, the enumerator the bot plays from and the
// card's chip. So "attacks each combat if able" sends the Dragon at an
// unprotected opponent if it can and lets it stay home when every
// opponent is protected. It binds the protected player and their
// planeswalkers; a battle they control is not named. It is an effect of
// the resolved trigger, not an ability the Dragon has, so a later "loses
// all abilities" leaves it. If the Dragon has left the battlefield (or is
// a new object, CR 400.7) the players still roll and nothing is
// restricted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a04f46b4-b0b0-4aef-907f-1df1a2e261de",
		Name:            "Chaos Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Static:          []game.StaticAbility{effects.AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Chaos Dragon — each player rolls a d20; it can't attack the opponents with the highest result or their planeswalkers this combat",
				func(g *game.Game, item *game.StackItem) error {
					return chaosDragonRoll(effects.NewContext(g, item), effects.SeatsFromActive(g), 0, nil)
				}),
		},
	})
}

// chaosDragonRoll has the first of `players` roll a d20, carrying the
// highest result so far and the players who had it, or restricts the
// Dragon once everyone has rolled.
func chaosDragonRoll(ctx *effects.Context, players []uuid.UUID, highest int, holders []uuid.UUID) error {
	if len(players) == 0 {
		return chaosDragonRestrict(ctx, holders)
	}
	then := effects.WithN(ctx.ItemTail(chaosDragonRolled), highest)
	then.Carry = effects.PackIDLists(players, holders)
	return ctx.Game.RollDiceThenForEffect(game.DiceRoll{
		Player: players[0], Source: ctx.Source(), Sides: 20, N: 1, Then: then,
	})
}

// chaosDragonRolled is one player's settled d20, then the next player's
// roll. The carry is the players still to roll (the roller first), then
// the players with the highest result so far; N is that result.
// Registered in init because the body rolls the next die, which names
// the key again.
var chaosDragonRolled game.TailRef

func init() {
	chaosDragonRolled = effects.CardTail("dice/chaos-dragon-each-player-rolls", chaosDragonSettled)
}

func chaosDragonSettled(ctx *effects.Context, o game.TailOutcome) error {
	players, holders := effects.UnpackIDPair(o.Tail.Carry)
	if len(players) == 0 {
		return chaosDragonRestrict(ctx, holders)
	}
	roller, rest := players[0], players[1:]
	highest := o.Tail.N
	if len(o.Rolls) > 0 {
		switch result := effects.DiceResult(o); {
		case result > highest:
			highest, holders = result, []uuid.UUID{roller}
		case result == highest:
			holders = append(holders, roller)
		}
	}
	return chaosDragonRoll(ctx, rest, highest, holders)
}

// chaosDragonRestrict makes the Dragon unable to attack each opponent
// among `holders`, and their planeswalkers, this combat.
func chaosDragonRestrict(ctx *effects.Context, holders []uuid.UUID) error {
	g := ctx.Game
	self := ctx.Source()
	if !effects.OnBattlefield(g, self) || ctx.IsNewSourceObject(self) {
		return nil
	}
	you := ctx.Controller()
	var mods []game.Mod
	for _, p := range holders {
		if p != you {
			mods = append(mods, game.CantAttackPlayerByEffectMod(p))
		}
	}
	if len(mods) == 0 {
		return nil
	}
	d, ok := effects.DurationUntilEndOfCombat(ctx)
	if !ok {
		return nil
	}
	return effects.ScopedEffectFor{
		Target:   self,
		Mods:     mods,
		Duration: d,
		Label:    "Chaos Dragon — can't attack the opponents with the highest result or their planeswalkers this combat",
	}.Apply(ctx)
}

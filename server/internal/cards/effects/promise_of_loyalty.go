package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Promise of Loyalty — Sorcery {4}{W}:
//
//	"Each player puts a vow counter on a creature they control and
//	 sacrifices the rest. Each of those creatures can't attack you or
//	 planeswalkers you control for as long as it has a vow counter on
//	 it."
//
// Each player picks their own creature, in turn order from the active
// player, as a choice and not a target. Only once the last has answered
// do the vow counters go on and the rest of every board's creatures get
// sacrificed together, so no player sees another's choice go through
// before making their own. A player with no creature chooses nothing.
//
// "Those creatures" are the chosen ones, fixed as the spell resolves
// (CR 611.2c), and each gets a cantAttackPlayer record naming the
// caster, timed by the counter-held duration (ADR 0109 §2, CR 611.2b):
// it ends for good when the last vow counter leaves. The caster's own
// creature carries it too, but CR 506.2 already keeps it off its
// controller. A creature that is a different object later (CR 400.7), or
// one that enters afterwards, is not affected.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "e1d294e8-67b1-4a4a-86b8-f98f28a327d9",
		Name:     "Promise of Loyalty",
		// ADR 0126 §6: each player keeps one creature and sacrifices the rest.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return promiseOfLoyaltyAsk(ctx, seatsFromActive(ctx.Game), nil)
		},
	})
}

// promiseOfLoyaltyAsk has the first of `players` choose a creature they
// control, then the rest, and applies the whole spell once the last has
// answered.
func promiseOfLoyaltyAsk(ctx *Context, players, chosen []uuid.UUID) error {
	if len(players) == 0 {
		return promiseOfLoyaltyApply(ctx, chosen)
	}
	p, rest := players[0], append([]uuid.UUID(nil), players[1:]...)
	so := append([]uuid.UUID(nil), chosen...)
	return ChoosePermanents{
		Player:   p,
		Of:       []uuid.UUID{p},
		Question: "Promise of Loyalty — choose a creature you control to keep; you sacrifice the rest",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			ids := promiseOfLoyaltyCreatures(g, of)
			if len(ids) == 0 {
				return nil, 0, 0
			}
			return ids, 1, 1
		},
		Then: func(ctx *Context, picked game.PromptedPicks) error {
			return promiseOfLoyaltyAsk(ctx, rest, append(so, picked.Cards()...))
		},
	}.Apply(ctx)
}

// promiseOfLoyaltyCreatures is the creatures `player` controls.
//
// Caller holds g.mu.
func promiseOfLoyaltyCreatures(g *game.Game, player uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// promiseOfLoyaltyApply puts a vow counter on each chosen creature that
// is still there, makes it unable to attack the caster while the counter
// stays, and sacrifices every other creature.
func promiseOfLoyaltyApply(ctx *Context, chosen []uuid.UUID) error {
	g := ctx.Game
	caster := ctx.Controller()
	keep := make(map[uuid.UUID]bool, len(chosen))
	for _, id := range chosen {
		keep[id] = true
		if !onBattlefield(g, id) {
			continue
		}
		if err := CounterThenWhileItHasIt(ctx, id, "vow",
			"Promise of Loyalty — can't attack its caster or their planeswalkers while it has a vow counter",
			game.CantAttackPlayerMod(caster)); err != nil {
			return err
		}
	}
	var doomed []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && !keep[c.InstanceID] {
			doomed = append(doomed, c.InstanceID)
		}
	}
	if len(doomed) == 0 {
		return nil
	}
	return g.SacrificeAllThenForEffect(ctx.Source(), doomed, nil)
}

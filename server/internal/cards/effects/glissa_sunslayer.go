package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glissa Sunslayer — Legendary Creature — Phyrexian Zombie Elf
// {1}{B}{G}, 3/3 (EDHREC rank 1377):
//
//	"First strike, deathtouch
//	 Whenever Glissa Sunslayer deals combat damage to a player,
//	 choose one —
//	 • You draw a card and lose 1 life.
//	 • Destroy target enchantment.
//	 • Remove up to three counters from target permanent."
//
// The TARGETED half of #764's modal-trigger proof set (Gala Greeters
// is the untargeted half). Two of the three bullets target, and they
// target different things, which is exactly what a trigger could not
// express before: TriggeredAbility had a single Targets clause and no
// mode slot at all.
//
// The order the rules ask for, and the order the engine now runs:
// the mode is chosen as the ability is put on the stack (CR 603.3c),
// through a mode_pick prompt to Glissa's controller; the target is
// chosen straight after, as part of the same putting-on-the-stack
// (CR 603.3d); and a bullet whose clause has no legal target is not
// offered at all — with no enchantment and no counter on the board,
// the only bullet on the prompt is the draw. If NO bullet could be
// taken the trigger would be removed from the stack, which is the
// same CR 603.3d rule the single-clause path already applied.
//
// First strike and deathtouch are printed keywords and come off the
// imported card.
//
// The third bullet asks how many counters to remove, up to three, and of
// which kind, as it resolves (chooseCounters, CR 608.2c): one question per
// kind of counter the permanent carries, each bounded by what is left of
// the three. The bot's suggestion takes every counter off an opponent's
// permanent but a -1/-1 counter, and only a -1/-1 counter off its own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8859a692-d107-4de2-9110-4681fc2761c7",
		Name:         "Glissa Sunslayer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			b64GlissaTrigger(),
		},
	})
}

func b64GlissaTrigger() game.TriggeredAbility {
	t := WheneverThisDealsCombatDamageToAPlayer("Glissa Sunslayer — choose one",
		func(g *game.Game, item *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		// Two instructions, so the life payment waits for a prompt the
		// draw asks (CR 608.2c, #4155).
		ModeSteps("You draw a card and lose 1 life.", nil,
			DrawCards{N: 1},
			Instruction(func(ctx *Context) error {
				return ctx.Game.PayLifeForEffect(ctx.Source(), ctx.Controller(), 1)
			})),
		ModeDoing("Destroy target enchantment.",
			TargetPermanent("target enchantment", Enchantment()),
			DestroyTheModesTarget),
		ModeDoing("Remove up to three counters from target permanent.",
			TargetPermanent("target permanent", b64HasAnyCounter()),
			func(item *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return b64ChooseAndRemoveCounters(ctx, t.ID, 3)
			}),
	)
	return t
}

// b64HasAnyCounter narrows the counter-removal clause to permanents
// that actually carry a counter. Without it the bullet would be
// offered — and the trigger would sit on the stack — pointed at
// something it cannot touch, which is the "Yes that silently does
// nothing" shape ADR 0019 §6 already removed from optional triggers.
func b64HasAnyCounter() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		for _, n := range c.Counters {
			if n > 0 {
				return true
			}
		}
		return false
	}
}

// b64ChooseAndRemoveCounters is "remove up to n counters from target
// permanent": the controller chooses how many of each kind, then they are
// removed through the ordinary counter primitive.
func b64ChooseAndRemoveCounters(ctx *Context, target uuid.UUID, n int) error {
	you := ctx.Controller()
	return chooseCounters(ctx, target, n, "Glissa Sunslayer — remove up to three counters",
		func(ctx *Context, kind string, most int) int {
			mine := false
			if c, ok := ctx.Game.LookupCardForEffect(target); ok {
				mine = c.Controller == you
			}
			if (kind == "-1/-1") == mine {
				return most
			}
			return 0
		},
		func(ctx *Context, picked map[string]int) error {
			kinds := make([]string, 0, len(picked))
			for kind := range picked {
				kinds = append(kinds, kind)
			}
			sort.Strings(kinds)
			for _, kind := range kinds {
				if err := (AddCounter{Target: target, Kind: kind, N: -picked[kind]}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
}

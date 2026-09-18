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
// Declared simplification, weaker than printed: the third bullet
// removes three counters (or all of them, if fewer) rather than
// asking how many and of which kind. "Up to three" is a count prompt
// for an ability at resolution, the same seam Resourceful Defense's
// "any number" waits on; removing the maximum is a legal answer to
// the printed choice and never more than the player could have
// chosen.
func init() {
	Register(Spec{
		OracleID:     "8859a692-d107-4de2-9110-4681fc2761c7",
		Name:         "Glissa Sunslayer",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The counter-removal mode always removes as many counters as it can — you can't choose how many or which kind."},
		Triggered: []game.TriggeredAbility{
			b64GlissaTrigger(),
		},
	})
}

func b64GlissaTrigger() game.TriggeredAbility {
	t := WheneverThisDealsCombatDamageToAPlayer("Glissa Sunslayer — choose one",
		func(g *game.Game, item *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("You draw a card and lose 1 life.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return ctx.Game.PayLifeForEffect(item.SourceCardID, item.Controller, 1)
			}),
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
				return b64RemoveUpToCounters(ctx, t.ID, 3)
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

// b64RemoveUpToCounters removes up to n counters from a permanent,
// kinds in a stable order so the same board always gives the same
// answer. "Up to" is read as "as many as possible" — see the card
// comment's declared simplification.
func b64RemoveUpToCounters(ctx *Context, target uuid.UUID, n int) error {
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok || n <= 0 {
		return nil
	}
	kinds := make([]string, 0, len(c.Counters))
	for kind := range c.Counters {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	left := n
	for _, kind := range kinds {
		if left <= 0 {
			break
		}
		have := c.Counters[kind]
		if have <= 0 {
			continue
		}
		take := have
		if take > left {
			take = left
		}
		if err := (AddCounter{Target: target, Kind: kind, N: -take}).Apply(ctx); err != nil {
			return err
		}
		left -= take
	}
	return nil
}

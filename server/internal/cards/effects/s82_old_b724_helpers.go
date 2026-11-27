package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b724_helpers.go — Urza's Saga verse-counter enchantments, S82 OLD
// batch 724 (#5624): the ten "At the beginning of your upkeep, you may put a
// verse counter on this enchantment" cards and Vebulid.

// s82OldB724Verse is the counter every one of them counts.
const s82OldB724Verse = "verse"

// S82OldB724VerseUpkeep is "At the beginning of your upkeep, you may put a
// verse counter on this enchantment." (Barrin's Codex's "you may" shape).
func S82OldB724VerseUpkeep(name string) game.TriggeredAbility {
	return Optional(AtYourUpkeep(name+" — put a verse counter on this enchantment", PutACounterOnThis(s82OldB724Verse)),
		name+" — put a verse counter on it?")
}

// S82OldB724VerseCountOnAnnounce is "X, where X is the number of verse counters
// on this enchantment", read once as the ability is announced, while the
// enchantment is still on the battlefield (CR 601.2b, 602.2b), so the
// sacrifice that is part of the cost doesn't change it. The rule answers a
// target count (TargetSpec.MaxFrom) or a division (DivideBy).
var S82OldB724VerseCountOnAnnounce = game.RegisterDivideAmount("s82-old-b724/verse-counters-on-announce",
	func(g *game.Game, a game.DivideAmountArgs) game.DivideSpec {
		c, ok := g.LookupCardForEffect(a.Source)
		if !ok {
			return game.DivideSpec{}
		}
		return game.DivideSpec{Total: c.Counters[s82OldB724Verse]}
	})

// S82OldB724UpToXTargets is a target clause "up to X target <what>, where X is
// the number of verse counters on this enchantment".
func S82OldB724UpToXTargets(spec *game.TargetSpec) *game.TargetSpec {
	spec = spec.WithCount(0, 0)
	spec.MaxFrom = S82OldB724VerseCountOnAnnounce
	return spec
}

// S82OldB724X is X as the ability resolves: the verse counters the sacrificed
// enchantment had as it last existed on the battlefield (CR 608.2h).
func S82OldB724X(ctx *Context) int {
	return CountersOnThisAsItResolves(s82OldB724Verse)(ctx)
}

// S82OldB724ForEachLegalTarget runs `do` on each target that is still legal as
// the ability resolves (CR 608.2b), in the order they were chosen.
func S82OldB724ForEachLegalTarget(do func(ctx *Context, t game.TargetRef) error) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if err := do(ctx, t); err != nil {
				return err
			}
		}
		return nil
	}
}

// S82OldB724SacrificeCost is the printed cost of a verse ability: the mana
// (empty for none) and "Sacrifice this enchantment".
func S82OldB724SacrificeCost(mana string) game.AbilityCost {
	if mana == "" {
		return SacrificeThis()
	}
	return Plus(ManaCost(mana), SacrificeThis())
}

// S82OldB724CounterUnlessXPaid is Lilting Refrain's effect: counter the target
// spell unless its controller pays {X}. With X = 0 there is nothing to pay and
// nothing to counter.
func S82OldB724CounterUnlessXPaid(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := S82OldB724X(ctx)
	targets := ctx.LegalTargets()
	if x <= 0 || len(targets) == 0 {
		return nil
	}
	return CounterUnlessPaid{
		StackID:  targets[0].ID,
		Cost:     "{" + strconv.Itoa(x) + "}",
		Question: "Lilting Refrain — pay {" + strconv.Itoa(x) + "}, or your spell is countered?",
	}.Apply(ctx)
}

// S82OldB724DiscardUpToXFromTheirHand is Discordant Dirge's effect: look at the
// target opponent's hand and choose up to X cards from it, which that player
// discards. With X = 0 you still look, and choose none (Keep leaves the pick in
// the hand: the pick is a look only).
func S82OldB724DiscardUpToXFromTheirHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := S82OldB724X(ctx)
	e := ChooseFromRevealedHand{
		Player: TargetedPlayer(ctx),
		Label:  "card",
		Count:  x,
		UpTo:   true,
		Look:   true,
	}
	if x <= 0 {
		e.Count, e.Keep = 1, true
	}
	return e.Apply(ctx)
}

// S82OldB724PumpPurpose is War Dance's declaration: a pump of its target
// creature until end of turn (a benefit to it), which answers by pumping.
func S82OldB724PumpPurpose() game.Purpose {
	p := ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit})
	p.Answers = game.AnswerPump
	return p
}

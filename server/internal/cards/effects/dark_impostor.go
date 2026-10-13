package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dark Impostor — Creature — Vampire Assassin {2}{B}, 2/2:
//
//	"{4}{B}{B}: Exile target creature and put a +1/+1 counter on this
//	 creature.
//	 This creature has all activated abilities of all creature cards
//	 exiled with it."
//
// The exile is linked (ExileTarget.Linked, ADR 0155 §8): the card lands
// with Card.ExiledWith naming this Impostor object, and the borrow reads
// its lenders through that link (ExiledWithLenders), so a creature card
// another effect exiled is not one of them. A token exiled this way
// ceases to exist and lends nothing. The counter is put on whether or not
// the creature reached exile (a commander its owner moves to the command
// zone), as printed; with the target gone in response the whole ability
// does nothing (CR 608.2b).
//
// No simplification.
func init() {
	const clause = "This creature has all activated abilities of all creature cards exiled with it."
	borrow := HasActivatedAbilitiesOf(game.AbilityBorrow{
		Label:   clause,
		Lenders: ExiledWithLenders(ByController(Creature())),
	})
	borrow.DependsOnExile = true
	Register(Spec{
		OracleID:     "96e25887-e556-423a-8a69-0cd148216f6a",
		Name:         "Dark Impostor",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{borrow},
		Activated: []ActivatedAbility{{
			Label:   "{4}{B}{B}: Exile target creature and put a +1/+1 counter on this creature.",
			Cost:    ManaCost("{4}{B}{B}"),
			Targets: TargetCreature("target creature"),
			Effect:  exileLinkedTargetThenCounterOnThis,
		}},
	})
}

// exileLinkedTargetThenCounterOnThis is "exile target <card> and put a
// +1/+1 counter on this creature" (Dark Impostor, Patchwork Crawler): the
// target is exiled linked to the source, then the counter goes on the
// source, whether or not the card reached exile.
func exileLinkedTargetThenCounterOnThis(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	counter := func(ctx *Context, _ bool) error {
		return AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ExileTarget{Target: t.ID, Linked: true, Then: counter}.Apply(ctx)
	}
	return nil
}

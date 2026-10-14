package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grell Philosopher — Creature — Horror Wizard {2}{U}, 1/4:
//
//	"Aberrant Tinkering — When this creature enters and at the beginning
//	 of your upkeep, each Horror you control gains all activated abilities
//	 of target artifact an opponent controls until end of turn. You may
//	 spend blue mana as though it were mana of any color to activate those
//	 abilities."
//
// One ability with two trigger conditions. The gain is a resolved borrow
// (ADR 0155 Decision 9): the Horrors you control as it resolves are the
// affected set (CR 611.2c), and the artifact's rows as it resolves are the
// lent set (CR 608.2h), so a Horror that arrives later gains nothing and
// the Horrors keep the rows after the artifact leaves. Grell is a Horror
// itself.
//
// The spend rider is part of the same effect, so it rides the rows it
// lent (BorrowedAbility.SpendFrom) and lasts as long as they do, whatever
// happens to Grell: blue mana pays any coloured symbol of those
// abilities' costs, and a {C} still needs colourless mana.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5dc00cb-1510-4ce9-8dd9-0722f65d34b4",
		Name:         "Grell Philosopher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				OnAny([]game.EventKind{game.EventETB, game.EventBeginUpkeep}, thisEntersOrYourUpkeep,
					"Grell Philosopher — Aberrant Tinkering: each Horror you control gains the activated abilities of target artifact an opponent controls",
					grellPhilosopherBorrow),
				TargetPermanent("target artifact an opponent controls", Artifact(), OpponentControls()),
			),
		},
	})
}

// thisEntersOrYourUpkeep is "When this enters and at the beginning of your
// upkeep": the source's own entry, or its controller's upkeep.
func thisEntersOrYourUpkeep(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventETB:
		return Self(ev, source, lki, g)
	case game.EventBeginUpkeep:
		return ByYou(ev, source, lki, g)
	}
	return false
}

// grellPhilosopherBorrow lends the target artifact's activated abilities,
// as they are now, to each Horror its controller controls now, with the
// blue-mana rider.
func grellPhilosopherBorrow(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		lender, ok := g.LookupCardPtrForEffect(t.ID)
		if !ok {
			return nil
		}
		return BorrowAbilitiesFor{
			Match:     And(Subtype("Horror"), YouControl()),
			Lender:    lender,
			SpendFrom: "U",
			Label:     "Grell Philosopher — each Horror you control gains all activated abilities of " + lender.Name + " until end of turn",
		}.Apply(ctx)
	}
	return nil
}

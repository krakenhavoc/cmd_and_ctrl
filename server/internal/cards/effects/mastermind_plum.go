package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mastermind Plum — Legendary Creature — Human Wizard {2}{B}, 2/2
// (Smaug Unchained deck, #4724):
//
//	"Whenever Mastermind Plum attacks, exile up to one target card from
//	 a graveyard. If an artifact card was exiled this way, create a
//	 Treasure token.
//	 Whenever you cast a spell, if mana from a Treasure was spent to
//	 cast it, you draw a card and you lose 1 life."
//
// The attack trigger is Scavenging Ooze's exile (ExileThenIfItWas) on an
// "up to one" target clause: any graveyard, any card, or no target at
// all. Whether the card is an artifact is read before the exile, and the
// Treasure waits for the exile and is gated on it (ADR 0013 §5t): a card
// that a replacement kept out of exile was not "exiled this way", so it
// makes no Treasure.
//
// The cast trigger is Alchemist's Talent's level-3 condition
// (spellPaidWithTreasureMana) over Hired Hexblade's "you draw a card and
// you lose 1 life". It is an intervening if (CR 603.4), read off the
// cast's payment record, where each mana keeps a snapshot of what its
// source was (#1212), so a Treasure that sacrificed itself to pay still
// counts. The record cannot change between the two checks the rule asks
// for, so the check as the ability triggers is the whole of it, and a
// spell countered in response still draws. A spell paid with no
// Treasure mana puts nothing on the stack. Treasure mana is always mana
// in the pool, so a waived payment cannot hide any (ReadsPoolManaOnly).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "4605d8b3-b714-4717-8ff3-155fabfc1999",
		Name:           "Mastermind Plum",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Mastermind Plum — exile up to one target card from a graveyard",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							return ExileThenIfItWas{
								Target: t.ID,
								Was:    func(c game.Card) bool { return c.IsArtifact() },
								Keyed:  mastermindPlumIfArtifact,
							}.Apply(ctx)
						}
						return nil
					}),
				TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1),
			),
			On(game.EventCast, spellPaidWithTreasureMana,
				"Mastermind Plum — draw a card and lose 1 life",
				// The loss waits for a prompt the draw asks (a dredge
				// offer; CR 608.2c), as Hired Hexblade's does.
				Do(
					DrawCards{N: 1},
					Instruction(func(ctx *Context) error {
						return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -1)
					}),
				)),
		},
	})
}

// mastermindPlumIfArtifact is the exile's keyed "if an artifact card was
// exiled this way, create a Treasure token". An on-disk key: never
// renamed, never reused.
var mastermindPlumIfArtifact = IfItWasTail("if-it-was/mastermind-plum-treasure", func(ctx *Context) error {
	return CreateToken{
		Controller: ctx.Controller(),
		Template:   TreasureToken(),
		N:          1,
	}.Apply(ctx)
}, nil)

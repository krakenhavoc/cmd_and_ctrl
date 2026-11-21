package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bronzebeak Foragers — Creature — Dinosaur {3}{W}, 3/4 (The Lost Caverns of
// Ixalan Commander):
//
//	"When this creature enters, for each opponent, exile up to one target
//	 nonland permanent that player controls until this creature leaves the
//	 battlefield.
//	 {X}{W}: Put target card with mana value X exiled with this creature
//	 into its owner's graveyard. You gain X life."
//
// The enters trigger is Dismantling Wave's per-opponent "up to one" clauses,
// one per opponent in turn order, each exiled through ExileUntil{ThisLeaves}
// (CR 610.3) with the CR 607.2a link stamped, so the card comes back when the
// Foragers leave and is "exiled with" this object until then. A Foragers gone
// before the trigger resolves exiles nothing (CR 610.3b). The ability targets
// a card exiled with this very object whose mana value is exactly the X paid
// (a token has mana value 0 but never stays in exile); it is re-checked as it
// resolves (CR 608.2b), so a card that left in response, and the life with
// it, is lost. The card then goes to its owner's graveyard, which ends the
// hold: it does not come back. The life is X.
//
// No simplification.
func init() {
	enters := effects.Targeting(
		effects.WhenThisEnters("Bronzebeak Foragers — for each opponent, exile up to one target nonland permanent that player controls until this creature leaves the battlefield",
			bronzebeakForagersExile),
		effects.PerOpponentClauses(game.MaxPlayers-1, "nonland permanent", effects.Nonland()))
	effects.Register(effects.Spec{
		OracleID:     "0f5f6f3e-1b3c-43e2-b11d-bd454366845c",
		Name:         "Bronzebeak Foragers",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{W}: Put target card with mana value X exiled with this creature into its owner's graveyard. You gain X life.",
			Cost:    effects.ManaCost("{X}{W}"),
			Targets: effects.ExiledWithThis(effects.TargetCardInExile("target card with mana value X exiled with this creature").WithManaValueEqualsX()),
			Effect:  bronzebeakForagersGrave,
		}},
	})
}

// bronzebeakForagersExile exiles each still-legal pick until this creature
// leaves the battlefield, in announce order.
func bronzebeakForagersExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.ExileUntil{
			Target:     t.ID,
			ThisLeaves: true,
			Linked:     true,
			Label:      "Bronzebeak Foragers — returns when Bronzebeak Foragers leaves the battlefield",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// bronzebeakForagersGrave puts the target into its owner's graveyard and
// gains X life.
func bronzebeakForagersGrave(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if _, err := effects.Cmdr2S32PutExiledCardIntoGraveyard(g, t.ID); err != nil {
			return err
		}
		return effects.GainLife{Amount: ctx.X()}.Apply(ctx)
	}
	return nil
}

package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lost Legacy — Sorcery {1}{B}{B} (Kaladesh, #5539):
//
//	"Choose a nonartifact, nonland card name. Search target player's
//	 graveyard, hand, and library for any number of cards with that name and
//	 exile them. That player shuffles, then draws a card for each card exiled
//	 from their hand this way."
//
// Necromentia's resolution with the draw in place of the Zombies. The name
// is chosen as the spell resolves (the keyed card-name prompt) and is free
// text (CR 201.2), so "nonartifact, nonland" is enforced where the engine
// can: only cards that are neither an artifact nor a land are exiled, so
// naming one of those exiles nothing (CR 608.2b is not involved: the name
// itself was never a legal choice). A card has the name when any face has
// it (CR 201.2b). "Any number of cards" is all of them: exiling fewer is
// never better for the caster, and the engine has no per-card pick for a
// search that names the cards exactly. The target player shuffles after the
// search whether or not anything was found, then draws one card for each
// card that left their hand. With the target player illegal as the spell
// resolves it does nothing (CR 608.2b) and no name is chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "532f3f03-37af-46b6-bafb-031e9a3df056",
		Name:         "Lost Legacy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetPlayer {
				return nil
			}
			ctx.Game.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
				"Lost Legacy — choose a nonartifact, nonland card name",
				effects.WithCarry(ctx.ItemTail(lostLegacyNamed), legal[0].ID))
			return nil
		},
	})
}

// lostLegacyNamed is the rest of the spell once the name (the tail's Word)
// is known: search the target player's graveyard, hand and library, exile
// every nonartifact, nonland card with the name, then the shuffle and the
// draw. An on-disk key.
var lostLegacyNamed = effects.CardTail("then/lost-legacy-named", func(ctx *effects.Context, o game.TailOutcome) error {
	target, name := effects.TailCarry0(o), o.Tail.Word
	p := ctx.PlayerByID(target)
	if p == nil || name == "" {
		return nil
	}
	var fromHand, all []uuid.UUID
	exiles := func(c game.Card) bool {
		return game.CardNameMatches(c, name) && !c.IsArtifact() && !c.IsLand()
	}
	for _, c := range p.Hand.Cards {
		if exiles(c) {
			fromHand = append(fromHand, c.InstanceID)
			all = append(all, c.InstanceID)
		}
	}
	for _, c := range p.Graveyard.Cards {
		if exiles(c) {
			all = append(all, c.InstanceID)
		}
	}
	for _, c := range p.Library.Cards {
		if exiles(c) {
			all = append(all, c.InstanceID)
		}
	}
	return ctx.Game.ExileCardsKeyedForEffect(all, game.KeyedTail{
		Ref: lostLegacyExiled, Player: target, Source: ctx.Source(), Carry: fromHand,
	})
})

// lostLegacyExiled is "that player shuffles, then draws a card for each card
// exiled from their hand this way": the carried hand cards that reached
// exile are counted. An on-disk key.
var lostLegacyExiled = effects.CardTail("tail/lost-legacy-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if err := ctx.Game.ShuffleLibraryForEffect(o.Tail.Player); err != nil {
		return err
	}
	fromHand := map[uuid.UUID]bool{}
	for _, id := range o.Tail.Carry {
		fromHand[id] = true
	}
	n := 0
	for _, id := range o.IDs {
		if fromHand[id] {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return effects.DrawCards{Player: o.Tail.Player, N: n}.Apply(ctx)
})

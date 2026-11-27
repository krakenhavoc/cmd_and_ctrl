package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thought Hemorrhage — Sorcery {2}{B}{R} (Alara Reborn):
//
//	"Choose a nonland card name. Target player reveals their hand.
//	 Thought Hemorrhage deals 3 damage to that player for each card with the
//	 chosen name revealed this way. Search that player's graveyard, hand, and
//	 library for all cards with that name and exile them. Then that player
//	 shuffles."
//
// Lost Legacy's resolution: the name is chosen as the spell resolves (the
// keyed card-name prompt) and is free text (CR 201.2), so "nonland" is
// enforced where the engine can: a land named matches nothing, being
// neither revealed nor exiled. A card has the name when any face has it
// (CR 201.2b). The hand is revealed to everyone and its named cards counted
// there, then 3 damage is dealt for each; "all cards" is every one in the
// graveyard, hand and library, exiled at once. The target player shuffles
// after the search whether or not anything was found. With the target player
// illegal as the spell resolves it does nothing (CR 608.2b) and no name is
// chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5cc4ee18-675e-4707-b658-ec746a1c03d8",
		Name:         "Thought Hemorrhage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetPlayer {
				return nil
			}
			ctx.Game.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
				"Thought Hemorrhage — choose a nonland card name",
				effects.WithCarry(ctx.ItemTail(thoughtHemorrhageNamed), legal[0].ID))
			return nil
		},
	})
}

// thoughtHemorrhageNamed is the rest of the spell once the name (the tail's
// Word) is known: reveal the hand, 3 damage per named card there, then
// exile every named nonland card from the three zones. An on-disk key.
var thoughtHemorrhageNamed = effects.CardTail("then/thought-hemorrhage-named", func(ctx *effects.Context, o game.TailOutcome) error {
	target, name := effects.TailCarry0(o), o.Tail.Word
	p := ctx.PlayerByID(target)
	if p == nil || name == "" {
		return nil
	}
	named := func(c game.Card) bool { return !c.IsLand() && game.CardNameMatches(c, name) }
	ctx.Game.RevealHandForEffect(target)
	inHand := 0
	var all []uuid.UUID
	for _, c := range p.Hand.Cards {
		if named(c) {
			inHand++
			all = append(all, c.InstanceID)
		}
	}
	for _, c := range p.Graveyard.Cards {
		if named(c) {
			all = append(all, c.InstanceID)
		}
	}
	for _, c := range p.Library.Cards {
		if named(c) {
			all = append(all, c.InstanceID)
		}
	}
	if inHand > 0 {
		if err := (effects.DealDamage{Source: ctx.Source(), Target: target, Amount: 3 * inHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return ctx.Game.ExileCardsKeyedForEffect(all, game.KeyedTail{
		Ref: thoughtHemorrhageExiled, Player: target, Source: ctx.Source(),
	})
})

// thoughtHemorrhageExiled is "then that player shuffles". An on-disk key.
var thoughtHemorrhageExiled = effects.CardTail("tail/thought-hemorrhage-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	return ctx.Game.ShuffleLibraryForEffect(o.Tail.Player)
})

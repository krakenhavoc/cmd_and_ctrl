package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necromentia — Sorcery {1}{B}{B} (Core Set 2021, #5511):
//
//	"Choose a card name other than a basic land card name. Search target
//	 opponent's graveyard, hand, and library for any number of cards with that
//	 name and exile them. That player shuffles, then creates a 2/2 black
//	 Zombie creature token for each card exiled from their hand this way."
//
// The name is chosen as the spell resolves (Demonic Consultation's keyed
// card-name prompt) and refuses a basic land card's name (CR 201.4a). A card
// has the name when any face has it (CR 201.2b). "Any number of cards" is all
// of them: exiling fewer is never better for the caster, and the engine has
// no per-card pick for a search that names the cards exactly. The opponent
// shuffles after the search whether or not anything was found, and creates
// the Zombies under their own control, one for each card that left their hand.
// With the target opponent illegal as the spell resolves it does nothing
// (CR 608.2b) and no name is chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92acbbeb-f82b-413b-b243-f7a012f5400c",
		Name:         "Necromentia",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetPlayer {
				return nil
			}
			ctx.Game.QueueCardNameChoiceKeyedNotBasicForEffect(ctx.Controller(), ctx.Source(),
				"Necromentia — choose a card name other than a basic land card name",
				effects.WithCarry(ctx.ItemTail(necromentiaNamed), legal[0].ID))
			return nil
		},
	})
}

// necromentiaNamed is the rest of the spell once the name (the tail's Word)
// is known: search the target opponent's three hidden-or-not zones, exile
// every card with the name, then the shuffle and the Zombies. An on-disk key.
var necromentiaNamed = effects.CardTail("then/necromentia-named", func(ctx *effects.Context, o game.TailOutcome) error {
	opp, name := effects.TailCarry0(o), o.Tail.Word
	p := ctx.PlayerByID(opp)
	if p == nil || name == "" {
		return nil
	}
	var fromHand, all []uuid.UUID
	for _, c := range p.Hand.Cards {
		if game.CardNameMatches(c, name) {
			fromHand = append(fromHand, c.InstanceID)
			all = append(all, c.InstanceID)
		}
	}
	for _, c := range p.Graveyard.Cards {
		if game.CardNameMatches(c, name) {
			all = append(all, c.InstanceID)
		}
	}
	for _, c := range p.Library.Cards {
		if game.CardNameMatches(c, name) {
			all = append(all, c.InstanceID)
		}
	}
	return ctx.Game.ExileCardsKeyedForEffect(all, game.KeyedTail{
		Ref: necromentiaExiled, Player: opp, Source: ctx.Source(), Carry: fromHand,
	})
})

// necromentiaExiled is "that player shuffles, then creates a 2/2 black Zombie
// creature token for each card exiled from their hand this way": the carried
// hand cards that reached exile are counted. An on-disk key.
var necromentiaExiled = effects.CardTail("tail/necromentia-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	g := ctx.Game
	if err := g.ShuffleLibraryForEffect(o.Tail.Player); err != nil {
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
	return g.CreateTokenForEffect(o.Tail.Player, ctx.TextToken(effects.TokenCard("2/2 black Zombie")), n)
})

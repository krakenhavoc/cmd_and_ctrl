package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dual Nature — Enchantment {4}{G}{G} (Prophecy, S82 OLD batch 675, #5617):
//
//	"Whenever a nontoken creature enters, its controller creates a token that's
//	 a copy of that creature.
//	 Whenever a nontoken creature leaves the battlefield, exile all tokens with
//	 the same name as that creature.
//	 When this enchantment leaves the battlefield, exile all tokens created with
//	 this enchantment."
//
// Three triggers, one per printed ability. The first is Faerie Artisans'
// token copy of the entering creature (any player's), made for the creature's
// controller and stamped with the Dual Nature OBJECT that made it
// (game.Card.CreatedWith, CR 607.1d); a token copy that enters does not trigger
// it, being a token. The second reads the departing creature's name as it last
// existed (CR 603.10a) and exiles every token on the battlefield with that
// name, whoever made it. The third exiles the tokens linked to the object that
// left, read from its own item, so a Dual Nature that left and came back
// (CR 400.7) leaves the old tokens to the new object's own trigger of nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a8fcf6f-3bda-485c-b960-62a37112f0eb",
		Name:         "Dual Nature",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.B26NontokenCreatureEntered(ev, g)
				return ok
			}, "Dual Nature — its controller creates a token that's a copy of that creature", dualNatureCopy),
				game.Purpose{Tokens: 1, CreatureTokens: 1}),
			// ADR 0126 §6: the tokens named like the creature that left, a partial sweep.
			effects.TriggerWithPurpose(effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventLTB {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !effects.IsToken(c) && effects.LeftAsType(ev, c, "creature")
			}, "Dual Nature — exile all tokens with the same name as that creature", dualNatureExileSameName),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}}),
			effects.WhenThisLeaves("Dual Nature — exile all tokens created with this enchantment", dualNatureExileCreated),
		},
	})
}

// dualNatureCopy creates the copy for the entering creature's controller,
// linked to this Dual Nature.
func dualNatureCopy(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	entered, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
	if !ok {
		return nil
	}
	tmpl, ok := effects.TokenCopyTemplate(g, entered.InstanceID)
	if !ok {
		return nil
	}
	creation := game.TokenCreation{
		Controller: entered.Controller,
		Source:     ctx.Source(),
		Groups:     []game.TokenGroup{{Template: effects.CopiedToken(tmpl), Count: 1}},
	}
	if ref, ok := g.SourceObjectForEffect(item); ok {
		creation.Groups[0].CreatedWith = game.PermissionCardRef(ref)
	}
	return g.CreateTokensKeyedForEffect(creation, game.KeyedTail{})
}

// dualNatureExileSameName exiles every token on the battlefield named as the
// departed creature was.
func dualNatureExileSameName(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Object == nil {
		return nil
	}
	name := item.Trigger.Object.Name
	if name == "" {
		return nil
	}
	ctx := effects.NewContext(g, item).AsGroupMember()
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() && c.Name == name {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// dualNatureExileCreated exiles the tokens created with the Dual Nature that
// left.
func dualNatureExileCreated(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	for _, id := range g.TokensCreatedWithForEffect(game.PermissionCardRef(ref)) {
		if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

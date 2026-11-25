package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b228_helpers.go — shared pieces of the Ixalan batch 228 cards
// (#5532): Jace, Cunning Castaway's Illusion and combat-damage trigger, and
// Huatli, Warrior Poet's divided damage.

// --- Jace, Cunning Castaway ---------------------------------------------------

// OldB228IllusionToken is Jace, Cunning Castaway's 2/2 blue Illusion creature
// token with "When this creature becomes the target of a spell, sacrifice it."
func OldB228IllusionToken() game.Card { return TokenFromCatalog(printedOldB228IllusionToken) }

// printedOldB228IllusionToken is that Illusion as PRINTED. Its slug is an
// on-disk identity.
func printedOldB228IllusionToken() TokenTemplate {
	return TokenTemplate{
		Slug: "illusion-sacrifice-when-targeted",
		Card: game.Card{
			Name:      "Illusion",
			TypeLine:  "Token Creature — Illusion",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"U"},
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, SelfTargetedByASpell, "Illusion — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
		Text: "When this creature becomes the target of a spell, sacrifice it.",
	}
}

// oldB228EachPlayerYourCreaturesDamaged is the each-condition behind "whenever
// one or more creatures you control deal combat damage to a player this turn":
// the player a combat damage event from one of the delayed ability's
// controller's creatures reached. The engine emits one damage event per
// source, so a delayed trigger that also sets OncePerBatch fires once for each
// player in each damage step (CR 603.2c), however many creatures connected.
// The key is an on-disk identity.
var oldB228EachPlayerYourCreaturesDamaged = game.DelayedEachCondition("damage/your-creatures-combat-damage-each-player",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) []uuid.UUID {
		if !CombatDamageToPlayerBy(ev, dt.Controller, g) {
			return nil
		}
		return []uuid.UUID{ev.Target}
	})

// oldB228LootBody is "draw a card, then discard a card" for the delayed
// ability's controller. An on-disk key.
var oldB228LootBody = game.SimpleDelayedBody("loot/jace-cunning-castaway", B35LootOne)

// OldB228JacePlusOne is Jace, Cunning Castaway's +1: until end of turn,
// whenever one or more creatures you control deal combat damage to a player,
// draw a card, then discard a card (CR 603.7b).
func OldB228JacePlusOne(g *game.Game, item *game.StackItem) error {
	return DelayedOnEvent{
		Label:        "Jace, Cunning Castaway — draw a card, then discard a card",
		On:           []game.EventKind{game.EventDealDamage},
		Condition:    oldB228EachPlayerYourCreaturesDamaged,
		Body:         oldB228LootBody,
		Repeats:      true,
		OncePerBatch: true,
	}.Apply(NewContext(g, item))
}

// OldB228JaceCopies is Jace, Cunning Castaway's −5: two tokens that are
// copies of Jace, except they're not legendary (CR 707.9b). A Jace that has
// left the battlefield is copied as it last existed there (CR 608.2h).
func OldB228JaceCopies(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return CreateTokenCopy{
		Controller: item.Controller,
		Copy:       ctx.Source(),
		N:          2,
		Except: func(t *game.Card) {
			t.TypeLine = RemoveLegendaryFromTypeLine(t.TypeLine)
		},
	}.Apply(ctx)
}

// --- Huatli, Warrior Poet ------------------------------------------------------

// oldB228CantBlock is "creatures dealt damage this way can't block this
// turn", told each creature and the damage it took: one that took none
// (prevented) is free to block. An on-disk key.
var oldB228CantBlock = game.RegisterTail("then/huatli-warrior-poet-cant-block", func(g *game.Game, o game.TailOutcome) error {
	if o.Amount <= 0 || len(o.IDs) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(o.IDs[0])
	if !ok || !c.IsCreature() {
		return nil
	}
	return RestrictUntilEOT{
		Target:       o.IDs[0],
		Restrictions: game.CantBlock,
		Label:        "Huatli, Warrior Poet — can't block this turn",
	}.Apply(NewContext(g, &game.StackItem{Controller: o.Tail.Player, SourceCardID: o.Tail.Source}))
})

// OldB228HuatliDividedDamage is Huatli, Warrior Poet's −X: X damage divided as
// the controller chose among the targets, as one damage instance, and each
// creature that was really dealt damage can't block this turn. A target that
// is no longer legal takes nothing and keeps its share unspent (CR 608.2b).
func OldB228HuatliDividedDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	legal := ctx.LegalTargets()
	if len(legal) == 0 {
		return nil
	}
	shares := item.Distribution
	return g.DamageInstanceForEffect(func() error {
		for _, t := range legal {
			if t.Kind != game.TargetCard || shares[t.ID] <= 0 {
				continue
			}
			if err := g.DealDamageToCreatureKeyedForEffect(item.SourceCardID, t.ID, shares[t.ID],
				game.KeyedTail{Ref: oldB228CantBlock, Player: item.Controller, Source: item.SourceCardID}); err != nil {
				return err
			}
		}
		return nil
	})
}

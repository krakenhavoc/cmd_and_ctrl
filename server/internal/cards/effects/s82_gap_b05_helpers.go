package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b05_helpers.go — shared pieces of S82 GAP batch 5 (Mystery Booster
// Commander, #5948): two tokens with abilities, Lord of the Pit's upkeep body
// and the Aura-cast reader Grakk uses.

// gapB05BlackLotusToken is Arzakon's Black Lotus: a colourless artifact token
// with "{T}, Sacrifice this token: Add three mana of any one color." The three
// mana are ONE colour pick (the Gilded Lotus shape), as Kwia Vigorbloom's Lotus.
func gapB05BlackLotusToken() game.Card { return tokenFromCatalog(printedGapB05BlackLotusToken) }

// printedGapB05BlackLotusToken is that token as PRINTED. Its slug is an
// on-disk identity.
func printedGapB05BlackLotusToken() tokenTemplate {
	return tokenTemplate{
		Slug: "black-lotus",
		Card: game.Card{
			Name:     "Black Lotus",
			TypeLine: "Token Artifact",
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:       true,
			SacrificeCost: true,
			Produced:      OneColorOfAmount(3),
			Label:         "{T}, Sacrifice: Add three mana of any one color",
		}},
		Text: "{T}, Sacrifice this token: Add three mana of any one color.",
	}
}

// gapB05LordOfThePitToken is Thomil's Lord of the Pit: a 7/7 black Demon with
// flying, trample and "At the beginning of your upkeep, sacrifice another
// creature. If you can't, this token deals 7 damage to you."
func gapB05LordOfThePitToken() game.Card { return tokenFromCatalog(printedGapB05LordOfThePitToken) }

// printedGapB05LordOfThePitToken is that token as PRINTED. Its slug is an
// on-disk identity.
func printedGapB05LordOfThePitToken() tokenTemplate {
	return tokenTemplate{
		Slug: "lord-of-the-pit",
		Card: game.Card{
			Name:      "Lord of the Pit",
			TypeLine:  "Token Creature — Demon",
			Power:     7,
			Toughness: 7,
			Colors:    []string{"B"},
			Keywords:  []string{"flying", "trample"},
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Lord of the Pit — sacrifice another creature; if you can't, it deals 7 damage to you", gapB05LordUpkeep),
		},
		Text: "Flying, trample. At the beginning of your upkeep, sacrifice another creature. If you can't, this token deals 7 damage to you.",
	}
}

// gapB05LordUpkeep is the token's upkeep body. "Another creature you control"
// that can be sacrificed (CR 101.2) is a mandatory pick; with none, the token
// deals 7 damage to its controller.
func gapB05LordUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.InstanceID != item.SourceCardID &&
			!g.CantBeSacrificedForEffect(c.InstanceID) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 7}.Apply(ctx)
	}
	return SacrificeChoice{
		Player:     item.Controller,
		Candidates: candidates,
		Question:   "Lord of the Pit — sacrifice another creature",
	}.Apply(ctx)
}

// gapB05YouCastAnAuraOnAnOpponentsCreature is "whenever you cast an Aura spell
// that targets a creature an opponent controls": the cast is yours, the spell
// is an Aura, and a target it announced (CR 601.2c) is a creature on the
// battlefield an opponent of the source's controller controls.
func gapB05YouCastAnAuraOnAnOpponentsCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !spell.IsAura() {
		return false
	}
	item := g.StackItemForEffect(ev.CardID)
	if item == nil || item.Kind != game.StackItemSpell {
		return false
	}
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || !onBattlefield(g, t.ID) {
			continue
		}
		if c, ok := g.LookupCardForEffect(t.ID); ok && c.IsCreature() && c.Controller != source.Controller {
			return true
		}
	}
	return false
}

// gapB05Conjure is "conjure a card named <name> onto the battlefield": a card
// that was never in a deck is made, owned by the conjuring player, and put onto
// the battlefield through the ordinary entry pipeline (CR 614 windows, enters
// triggers), as a card put onto the battlefield from a hand. The card is made in
// its owner's hand without an event, so nothing sees it there, and
// PutFromHandOntoBattlefieldForEffect is the one door in. A replacement that
// stops the entry leaves it where it is, in the hand, a card nobody drew; a
// conjured card has no previous zone to return to.
func gapB05Conjure(ctx *Context, player uuid.UUID, card game.Card) error {
	p := ctx.Game.PlayerByIDForEffect(player)
	if p == nil {
		return nil
	}
	card.InstanceID = uuid.New()
	card.Owner, card.Controller = player, player
	p.Hand.PushTop(card)
	_, err := ctx.Game.PutFromHandOntoBattlefieldForEffect(card.InstanceID, game.HandEntryOptions{Controller: player})
	return err
}

// gapB05FoodFightCard is the Food Fight card Overcooked conjures: Wilds of
// Eldraine's {1}{R} enchantment, identified by its oracle ID like any catalog
// card, so its grant of the sacrifice ability is the catalog's.
func gapB05FoodFightCard() game.Card {
	return game.Card{
		Name:       "Food Fight",
		TypeLine:   "Enchantment",
		ManaCost:   "{1}{R}",
		Colors:     []string{"R"},
		OracleID:   "00571e7f-4e40-4a07-a529-34580e942338",
		ScryfallID: "031884be-66e9-4474-bba3-ddd6ee0f8086",
	}
}

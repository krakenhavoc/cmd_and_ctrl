package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// YouMaySpendManaAsAnyColorToCastSpellsWithTheChosenName is "You may spend
// mana as though it were mana of any color to cast spells with the chosen
// name" (Unexpected Potential, #5557, CR 609.4b): a spend grant narrowed to
// casts of a spell with the conspiracy's chosen name (game.SpellOrLandHasName),
// read through the conspiracy that grants it. It reaches no activation and no
// other cost.
func YouMaySpendManaAsAnyColorToCastSpellsWithTheChosenName() []game.AnyColorSpendStatic {
	return []game.AnyColorSpendStatic{{
		Label: "You may spend mana as though it were mana of any color to cast spells with the chosen name.",
		Whose: game.AnyColorSpendYou,
		Covers: func(ctx game.ManaSpendContext) bool {
			return ctx.Purpose == game.SpendPurposeCast
		},
		CoversFrom: func(ctx game.ManaSpendContext, source *game.Card) bool {
			for _, n := range game.ChosenNamesOn(source) {
				if ctx.SpellHasName(n) {
					return true
				}
			}
			return false
		},
	}}
}

// CreatureYouControlWithTheChosenNameEnters is the AppliesTo of a replacement
// on "each creature you control with the chosen name" as it enters the
// battlefield (Muzzio's Preparations, #5557): the entering object is a creature
// its caster controls whose name is one the conspiracy chose, read as it enters.
func CreatureYouControlWithTheChosenNameEnters(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil {
		return false
	}
	entering, ok := g.LookupCardForEffect(ev.CardID)
	return ok && entering.IsCreature() && entering.Controller == src.Controller && HasTheChosenName(entering, src)
}

// S82OldB360ChosenNameCreatureYouControlEnters is the When of "whenever a
// creature you control with the chosen name enters" (Secret Summoning).
func S82OldB360ChosenNameCreatureYouControlEnters(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB || source == nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Controller == source.Controller && HasTheChosenName(c, source)
}

// S82OldB360NamedCardsToHand is "search your library for any number of cards
// with that name, reveal them, put them into your hand, then shuffle" (Secret
// Summoning): the name is the one the trigger carried (Params.Name), compared
// as every chosen name is (game.CardNameMatches).
func S82OldB360NamedCardsToHand(g *game.Game, item *game.StackItem) error {
	name := item.Params.Name
	if name == "" {
		return nil
	}
	return SearchLibrary{
		Player:    item.Controller,
		Predicate: func(c game.Card) bool { return game.CardNameMatches(c, name) },
		Dest:      game.ZoneHand,
		Unbounded: true,
		Reveal:    true,
		Shuffle:   true,
		Source:    item.SourceCardID,
		Reason:    "Secret Summoning — any number of cards named " + name,
	}.Apply(NewContext(g, item))
}

// S82OldB360BuildNamingTheEnteringCreature is the Build of a trigger that
// carries the name of the creature that entered (Secret Summoning).
func S82OldB360BuildNamingTheEnteringCreature(label string) func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		if c, ok := g.LookupCardForEffect(ev.CardID); ok {
			item.Params.Name = c.Name
		}
		return item
	}
}

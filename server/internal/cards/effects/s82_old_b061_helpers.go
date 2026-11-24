package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB061NontokenElfOrBerserkerYouControlDied is "another nontoken Elf or
// Berserker you control dies" (Skemfar Avenger), read from the creature's
// last-known information (CR 603.10a).
func S82OldB061NontokenElfOrBerserkerYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := DiedCreature(ev, g)
	if !ok || dead.InstanceID == source.InstanceID || IsToken(dead) || LeftUnderControlOf(ev, dead) != source.Controller {
		return false
	}
	return LeftAsSubtype(ev, dead, source.TextSubtype("Elf")) || LeftAsSubtype(ev, dead, source.TextSubtype("Berserker"))
}

// --- Runeforge Champion ---------------------------------------------------

// GrantedAltCostOneRuneSpells is "You may pay {1} rather than pay the mana
// cost for Rune spells you cast." (Runeforge Champion). An on-disk key: never
// renamed, never reused.
const GrantedAltCostOneRuneSpells = "granted-one-rune-spells"

// PayOneForRuneSpellsYouCast is Runeforge Champion's static: an alternative
// cost of {1} for Rune spells from any zone their printed mana cost could be
// paid (CR 118.9a). Additional costs are still paid on top (CR 118.9d).
func PayOneForRuneSpellsYouCast() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:      GrantedAltCostOneRuneSpells,
			Label:    "Pay {1} rather than pay this spell's mana cost",
			ManaCost: "{1}",
		},
		Spells: game.PermissionFilter{SpellSubtype: "Rune"},
	}
}

// --- Pyre of Heroes ---------------------------------------------------------

// S82OldB061PyreOfHeroesSearch is Pyre of Heroes' body: "Search your library
// for a creature card that shares a creature type with the sacrificed
// creature and has mana value equal to 1 plus that creature's mana value. Put
// that card onto the battlefield, then shuffle."
//
// The creature was sacrificed as a cost, so its creature types are read as it
// last existed on the battlefield (CR 608.2h: Characteristic, post-layers,
// with the every-creature-type flag), and its mana value from the card as it
// sits now (a token's is zero, so it fetches a one-drop). With no sacrifice on
// record nothing happens.
func S82OldB061PyreOfHeroesSearch(g *game.Game, item *game.StackItem) error {
	sacrificed, ok := B17PermanentSacrificedToPay(g, item)
	if !ok {
		return nil
	}
	dead, ok := g.LookupCardForEffect(sacrificed)
	if !ok {
		return nil
	}
	want := dead.ManaValue() + 1
	var types []string
	all := false
	if info, ok := g.LastKnownPermanentForEffect(sacrificed); ok {
		all = info.Characteristic.AllCreatureTypes
		for _, s := range info.Characteristic.Subtypes {
			if game.IsCreatureType(s) {
				types = append(types, s)
			}
		}
	} else {
		all = game.HasAllCreatureTypes(&dead)
		types = game.CreatureTypesOf(&dead)
	}
	shares := func(c game.Card) bool {
		theirs := game.CreatureTypesOf(&c)
		if all {
			return len(theirs) > 0
		}
		if game.HasAllCreatureTypes(&c) {
			return len(types) > 0
		}
		for _, a := range types {
			for _, b := range theirs {
				if a == b {
					return true
				}
			}
		}
		return false
	}
	return SearchLibrary{
		Player: item.Controller,
		Predicate: func(c game.Card) bool {
			return c.IsCreature() && c.ManaValue() == want && shares(c)
		},
		Dest:    game.ZoneBattlefield,
		Limit:   1,
		Shuffle: true,
		Reason:  "Pyre of Heroes — a creature card that shares a creature type with the sacrificed creature and has mana value " + strconv.Itoa(want),
	}.Apply(NewContext(g, item))
}

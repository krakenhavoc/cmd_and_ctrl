package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b43_helpers.go — shared predicates for Standard set batches 43-46.

// APermanentYouControlWasTurnedFaceUp is "whenever a face-down permanent
// you control is turned face up" (Sumala Sentry). Unlike
// ACreatureYouControlWasTurnedFaceUp it does not ask for a creature: a
// disguised land or artifact turns up too (CR 708.8: the permanent is
// face up by the time the event is emitted).
func APermanentYouControlWasTurnedFaceUp(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller
}

// Token rows for batches 43-46, added from this file's init so the shared
// table's column alignment is not rewritten (#4452). The keys read like the
// printed text, as the table's do.
func init() {
	for key, card := range map[string]game.Card{
		// Soul Search, Teysa, Opulent Oligarch.
		"1/1 white and black Spirit with flying": {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1, Colors: []string{"W", "B"}, Keywords: []string{"flying"}},
		// Intrude on the Mind: the card's own Thopter is a 0/0 that takes counters.
		"0/0 colorless Thopter artifact with flying": {Name: "Thopter", TypeLine: "Token Artifact Creature — Thopter", Power: 0, Toughness: 0, Keywords: []string{"flying"}, PrintedPTKnown: true},
		// Tolsimir, Midnight's Light.
		"5/5 green and white Legendary Wolf named Voja Fenstalker": {Name: "Voja Fenstalker", TypeLine: "Token Legendary Creature — Wolf", Power: 5, Toughness: 5, Colors: []string{"G", "W"}, Keywords: []string{"trample"}},
	} {
		if _, dup := tokenTable[key]; dup {
			panic("std_b43_helpers: token key already in the table: " + key)
		}
		tokenTable[key] = card
	}
}

// differentColorPairsAmong is the number of different color pairs among the
// permanents `controller` controls that are exactly two colors (Niv-Mizzet,
// Guildpact): the distinct two-color sets, read through the layered view, out
// of the ten there are (the ruling). A colorless, monocolored or three-color
// permanent has no pair.
func differentColorPairsAmong(g *game.Game, controller uuid.UUID) int {
	pairs := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller {
			continue
		}
		colors := append([]string(nil), c.EffectiveColors()...)
		if len(colors) != 2 {
			continue
		}
		sort.Strings(colors)
		pairs[colors[0]+colors[1]] = true
	}
	return len(pairs)
}

// FaceUpCostLess is "This cost is reduced by {1} for each …" printed on a
// morph or disguise cost (Fugitive Codebreaker, #4452): count is read as the
// turn-face-up cost is priced, for the player turning the permanent up, and
// reduces its generic mana (never below zero). `ac` is copied, never
// modified.
func FaceUpCostLess(ac game.AlternativeCost, count func(g *game.Game, controller uuid.UUID) int) game.AlternativeCost {
	if ac.FaceDown == nil {
		panic("effects.FaceUpCostLess: " + ac.Key + " is not a face-down cast")
	}
	fd := *ac.FaceDown
	fd.FaceUpCostLess = &game.FaceUpCostReduction{Count: count}
	ac.FaceDown = &fd
	return ac
}

// instantAndSorceryCardsInGraveyard is how many instant and sorcery cards
// `player`'s graveyard holds.
func instantAndSorceryCardsInGraveyard(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsInstant() || c.IsSorcery() {
			n++
		}
	}
	return n
}

// DealtDamageAsASourceThisTurn is "a creature that dealt damage this turn"
// (Treacherous Greed): the permanent was the source of a damage event this
// turn that dealt more than 0 (damage actually dealt, after prevention). CR
// 400.7: an entry after the damage makes a new object that dealt none, so an
// entry wipes the record. Combat and noncombat damage both count, and the
// recipient does not have to still exist (the ruling).
func DealtDamageAsASourceThisTurn() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		dealt := false
		for _, ev := range g.EventsThisTurn() {
			switch {
			case ev.Kind == game.EventZoneMove && ev.CardID == c.InstanceID && ev.NewZone == game.ZoneBattlefield:
				dealt = false
			case ev.Kind == game.EventDealDamage && ev.Source == c.InstanceID && ev.Amount > 0:
				dealt = true
			}
		}
		return dealt
	}
}

package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b07_helpers.go — Streets of New Capenna, S82 batch 7 (#5491).

// printedS82SncB07FishToken is the 1/1 blue Fish with "This token can't be
// blocked." that Reservoir Kraken, Sleep with the Fishes, Expendable Lackey
// and Exotic Pets make. A catalog template, not a tokens_table row, because
// the clause is a static the token always prints (Reef Worm's Fish is a
// different token). Its slug is an on-disk identity: never renamed.
func printedS82SncB07FishToken() tokenTemplate {
	return tokenTemplate{
		Slug: "fish-cant-be-blocked",
		Card: game.Card{
			Name:      "Fish",
			TypeLine:  "Token Creature — Fish",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"U"},
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Text:   "This token can't be blocked.",
	}
}

// S82SncB07FishToken is that Fish's card value for CreateToken.
func S82SncB07FishToken() game.Card { return tokenFromCatalog(printedS82SncB07FishToken) }

// sncB07Kinds lists, sorted, the kinds of counter that at least one of the
// cards has (CR 122.1: a kind is a counter's name; a keyword counter is its
// own kind).
func sncB07Kinds(cards []game.Card) []string {
	seen := map[string]bool{}
	for _, c := range cards {
		for kind, n := range c.Counters {
			if n > 0 {
				seen[kind] = true
			}
		}
	}
	kinds := make([]string, 0, len(seen))
	for kind := range seen {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return kinds
}

// sncB07PermanentsYouControl is the battlefield cards a player controls,
// the creatures only when creaturesOnly.
func sncB07PermanentsYouControl(g *game.Game, player uuid.UUID, creaturesOnly bool) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player || (creaturesOnly && !c.IsCreature()) {
			continue
		}
		out = append(out, c)
	}
	return out
}

// sncB07KindsAmongPermanentsYouControl is "each kind of counter among
// permanents you control", read now.
func sncB07KindsAmongPermanentsYouControl(g *game.Game, player uuid.UUID) []string {
	return sncB07Kinds(sncB07PermanentsYouControl(g, player, false))
}

// sncB07KindsAmongCreaturesYouControl is "each kind of counter among
// creatures you control", read now.
func sncB07KindsAmongCreaturesYouControl(g *game.Game, player uuid.UUID) []string {
	return sncB07Kinds(sncB07PermanentsYouControl(g, player, true))
}

// sncB07CreaturePairs lists every (creature, kind of counter on it) pair as an
// option: the label names the kind and the creature, the carry holds the
// creature, and the words hold the kind, in the same order.
func sncB07CreaturePairs(g *game.Game, player uuid.UUID) (options []game.ChoiceOption, creatures []uuid.UUID, kinds []string) {
	for _, c := range sncB07PermanentsYouControl(g, player, true) {
		var own []string
		for kind, n := range c.Counters {
			if n > 0 {
				own = append(own, kind)
			}
		}
		sort.Strings(own)
		for _, kind := range own {
			options = append(options, game.ChoiceOption{Label: "A " + kind + " counter on " + c.Name})
			creatures = append(creatures, c.InstanceID)
			kinds = append(kinds, kind)
		}
	}
	return options, creatures, kinds
}

// sncB07HasCounters reports whether any kind has a counter.
func sncB07HasCounters(counters map[string]int) bool {
	for _, n := range counters {
		if n > 0 {
			return true
		}
	}
	return false
}

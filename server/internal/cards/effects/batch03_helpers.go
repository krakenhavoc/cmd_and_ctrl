package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch03_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 03 (#296, `edhrec_rank` 361–471).
//
// Its own file rather than helpers.go, per the convention #231 set:
// concurrent card batches collide on shared helper files. Every name
// here is b03-prefixed for the same reason — batches 02 and 04 are
// being written at the same time on sibling branches.

// b03TriMana is the "{T}: Add {A}, {B}, or {C}" every tri-land
// prints — a three-option pipe, the same shape the Temple cycle's
// two-option pipe and Birds of Paradise's five-option one take.
func b03TriMana(a, b, c string) ManaAbility {
	return ManaAbility{
		Cost:     ManaAbilityCost{Tap: true},
		Produced: "{" + a + "|" + b + "|" + c + "}",
		Label:    "Add {" + a + "}, {" + b + "}, or {" + c + "}",
	}
}

// b03ManaValueIs passes when the card's mana value is exactly n —
// Mental Misstep's "target spell with mana value 1". Same
// arithmetic as ManaValueLE / manaValueOf; {X} counts 0, so a spell
// with X in its printed cost is mana value 1 only when the rest of
// the cost says so (the sandbox reads the printed cost here; Mana
// Drain's refund is the one place the announced X is read).
func b03ManaValueIs(n int) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return manaValueOf(c) == n
	}
}

// b03Nonbasic passes for a land without the basic supertype —
// Demolition Field's "target nonbasic land".
func b03Nonbasic() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return !IsBasicLand(c)
	}
}

// b03NotNamed excludes a card by name. It is how "sacrifice ANOTHER
// creature" is written on Warren Soultrader: TargetSpec.CardOK never
// receives the source (#350 lists "another/other" as an open gap),
// so the source cannot be excluded by instance — but it can by
// name, which in a singleton format is the same creature. A token
// copy of the Soultrader would also be excluded, which is WEAKER
// than printed, never stronger.
func b03NotNamed(name string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.Name != name
	}
}

// b03IsLandCard / b03IsArtifactCard are SearchLibrary predicates
// (which take a bare game.Card rather than a CardPredicate).
func b03IsLandCard(c game.Card) bool     { return c.IsLand() }
func b03IsArtifactCard(c game.Card) bool { return c.IsArtifact() }

// b03ArtifactsControlled counts the artifacts `player` controls —
// Dispatch's metalcraft check, evaluated at resolution. IsArtifact
// reads post-layer types since #354, so an animated artifact and a
// Mycosynth Lattice board both count.
func b03ArtifactsControlled(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsArtifact() {
			n++
		}
	}
	return n
}

// b03LandsControlled counts the lands `player` controls (post-layer
// types, like every Is* read since #354).
func b03LandsControlled(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsLand() {
			n++
		}
	}
	return n
}

// b03OpponentControlsMoreLands is Land Tax's intervening-if: "if an
// opponent controls more lands than you". Eliminated seats are not
// opponents.
func b03OpponentControlsMoreLands(g *game.Game, controller uuid.UUID) bool {
	mine := b03LandsControlled(g, controller)
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == controller {
			continue
		}
		if b03LandsControlled(g, p.ID) > mine {
			return true
		}
	}
	return false
}

// b03CastSpellHasType reports whether the spell an EventCast names
// has any of the given lowercase type-line words — Sram's "an Aura,
// Equipment, or Vehicle spell". The card is read off the stack,
// where its printed type line is intact. Needles must be lowercase
// (containsFoldASCII folds the haystack only).
func b03CastSpellHasType(ev game.Event, g *game.Game, words ...string) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	for _, w := range words {
		if containsFoldASCII(c.TypeLine, w) {
			return true
		}
	}
	return false
}

// b03InstantOrSorceryCastByYou is the magecraft-shaped condition
// Guttersnipe shares with Storm-Kiln Artist: the controller cast an
// instant or sorcery.
func b03InstantOrSorceryCastByYou(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && (spell.IsInstant() || spell.IsSorcery())
}

// b03GreenBeast44Token is Rampaging Baloths' 4/4 green Beast — a
// different token from Beast Within's 3/3 GreenBeastToken, so it
// gets its own template rather than a parameter on that one.
func b03GreenBeast44Token() game.Card {
	return game.Card{
		Name:      "Beast",
		TypeLine:  "Token Creature — Beast",
		Power:     4,
		Toughness: 4,
	}
}

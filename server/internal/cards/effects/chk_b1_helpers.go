package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// chk_b1_helpers.go — the shapes Champions of Kamigawa, batch 1 (#5603)
// prints more than once: the Honden cycle's Shrine count, the "Spirit or
// Arcane spell" cast trigger, and the soulshift and bushido bodies.

// chkB1SpiritOrArcane is the spell half of "Whenever you cast a Spirit or
// Arcane spell" and of "target Spirit or Arcane spell": a creature type
// or a spell type on the spell's type line (the ki flip cards' predicate).
func chkB1SpiritOrArcane() CardPredicate {
	return Or(Subtype("Spirit"), Subtype("Arcane"))
}

// chkB1ShrinesYouControl is "the number of Shrines you control", counted
// as the Honden's upkeep trigger resolves (CR 608.2h), the Honden itself
// among them while it is still on the battlefield. "Shrine" is a word of
// the source's text and reads through a text change (CR 612.2), as
// Battlefield Medic's "Cleric" does.
func chkB1ShrinesYouControl(g *game.Game, item *game.StackItem) int {
	shrine := g.TextReaderForEffect(item).TextSubtype("Shrine")
	return countControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtypeWord(shrine) })
}

// chkB1Honden is a Honden: "At the beginning of your upkeep, <effect> for
// each Shrine you control." `effect` is handed the count.
func chkB1Honden(label string, effect func(g *game.Game, item *game.StackItem, shrines int) error) game.TriggeredAbility {
	return AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
		n := chkB1ShrinesYouControl(g, item)
		if n <= 0 {
			return nil
		}
		return effect(g, item, n)
	})
}

// chkB1Soulshifter registers a creature whose only rules text is its
// printed keywords and Soulshift n (CR 702.46a): the five plain kami.
func chkB1Soulshifter(oracleID, name string, n int, keywords ...string) {
	Register(Spec{
		OracleID:        oracleID,
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: keywords,
		Triggered:       []game.TriggeredAbility{Soulshift(n, name)},
	})
}

// chkB1Bushido registers a creature whose only rules text is its printed
// keywords and Bushido n (CR 702.45a).
func chkB1Bushido(oracleID, name string, n int, keywords ...string) {
	Register(Spec{
		OracleID:        oracleID,
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: keywords,
		Triggered:       []game.TriggeredAbility{Bushido(n, name)},
	})
}

// chkB1SnakeDealtCombatDamageToAPlayer is Seshiro's "Whenever a Snake you
// control deals combat damage to a player": one trigger per Snake that
// deals it. "Snake" is a word of the source's text (CR 612.2).
func chkB1SnakeDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && src.HasSubtypeWord(source.TextSubtype("Snake"))
}

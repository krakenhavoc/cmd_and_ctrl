package game

// prowess.go — CR 702.108, the first KEYWORD TRIGGER the engine derives
// from an object's ability list rather than from its catalog entry
// (#706; ADR 0014 amendment 2026-09-24).
//
//	702.108a Prowess is a triggered ability. "Prowess" means
//	         "Whenever you cast a noncreature spell, this creature
//	         gets +1/+1 until end of turn."
//	702.108b If a creature has multiple instances of prowess, each
//	         triggers separately.
//
// WHY A TOKEN AND NOT A CONSTRUCTOR. Ward and cascade are card-side
// constructors (effects.Ward, effects.Cascade) and never joined
// canonicalKeywords. Ward stayed out because it carries a cost a bare
// token has nowhere to put (ADR 0038 §7). Prowess has no parameter,
// and it is GRANTED far more often than ward is: Bria's "other
// creatures you control have prowess", Sokka's Allies, Narset,
// Wizard's Staff, and every Monk and Otter token that is printed "with
// prowess". A constructor on Spec.Triggered reaches none of those. A
// token in Characteristic.Abilities reaches all of them through
// machinery that already exists: the deck importer stamps a printed
// prowess off Scryfall (no catalog entry needed), a token template's
// Keywords carry it, and a layer-6 grant appends it.
//
// So the TOKEN is the declaration and this file is its consumer:
// TriggersForCard asks keywordTriggersFor for one TriggeredAbility per
// prowess token in the object's effective ability list, and the
// harvester treats them exactly like catalog triggers — the same
// AppliesTo / Build pair, the same trigger doublers, the same APNAP
// drain and the same response window.
//
// MULTIPLE INSTANCES. CR 702.108b is why prowess is a CUMULATIVE
// keyword (KeywordIsCumulative): AppendKeywordAbility keeps every
// granted instance instead of deduping it, so a Ty Lee under Sokka has
// two, and triggers twice. A PRINTED "Prowess, prowess" (Thor Odinson,
// Ruric Thar, Biomagus) also reaches the list twice: the deck
// importer's oracle-line scan counts a cumulative keyword's repeats
// (deck.printedKeywords, #1510) rather than taking Scryfall's
// `keywords` array — a set — at its word, and printedCharacteristic's
// merge (characteristic.go) keeps the higher of the catalog's and the
// import's own counts for a cumulative token instead of deduping it
// to one.

import "github.com/google/uuid"

// KeywordProwess is CR 702.108's canonical token.
const KeywordProwess = "prowess"

// prowessLabel is the stack label every prowess trigger carries.
// Identical on every instance of one creature on purpose: CR 603.3b's
// ordering prompt (seatNeedsTriggerOrder) is skipped when every item a
// player has is the same source and label, and asking someone to order
// two identical +1/+1s would be a question with no meaning. Instances
// on DIFFERENT creatures skip it too, through StackItem.Commutes
// (#1511) rather than the label.
const prowessLabel = "Prowess — +1/+1 until end of turn"

// prowessTrigger is the one TriggeredAbility a single instance of
// prowess is. A package-level value, never mutated: keywordTriggersFor
// hands out copies of it.
var prowessTrigger = TriggeredAbility{
	Keyword: KeywordProwess,
	Watches: []EventKind{EventCast},
	// "Whenever you cast a noncreature spell." EventCast is emitted by
	// CastSpell and by nothing else, so a COPY of a spell — storm,
	// Twincast, a Kitsa activation — never triggers it (CR 707.10: a
	// copy is put on the stack, not cast). A spell cast by any route
	// — from hand, by cascade, from exile — does.
	AppliesTo: func(ev Event, source *Card, _ Characteristic, g *Game) bool {
		if ev.Actor != source.Controller || ev.CardID == uuid.Nil {
			return false
		}
		spell, ok := g.LookupCardForEffect(ev.CardID)
		return ok && !spell.IsCreature()
	},
	Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
		// ADR 0041 P9 (#1497, tier 4): prowess has no catalog row, so
		// its item is keyed directly rather than through a catalog
		// AbilityRef — a table with one on the stack is still a
		// restore point.
		item := NewKeyedTriggeredItem(source, prowessLabel, prowessPumpBody, EffectParams{})
		// #1511: prowess instances commute with each other, so a batch
		// of nothing but prowess needs no CR 603.3b ordering prompt.
		// Each one reads only its own source (is this object still on
		// the battlefield, and the same object) and writes only a
		// layer-7c +1/+1 pinned to that source; no prowess changes
		// what another reads, layer-7c modifications add whatever
		// their timestamps, and registering one emits no event. The
		// full argument, and what it leaves out, is ADR 0018's #1511
		// amendment.
		item.Commutes = true
		return item
	},
}

// prowessPumpBody is registered under "prowess/pump" (ADR 0041 P9,
// #1497, tier 4), so a prowess trigger waiting on the stack restores
// through the running binary rather than blocking the snapshot.
var prowessPumpBody = SimpleDelayedBody("prowess/pump", resolveProwess)

// resolveProwess is "this creature gets +1/+1 until end of turn". A
// package-level func, so it captures nothing and survives Clone.
//
// "This creature" is the OBJECT that triggered (CR 400.7, #1432): a
// creature that left the battlefield in response, or left and came
// back as a new object, gets nothing. The pump is pinned to the
// object's instance ID and battlefield-entry stamp — the same key
// effects.BoostUntilEOT uses — so a flicker after resolution ends it
// too.
func resolveProwess(g *Game, item *StackItem) error {
	if g.AbilitySourceIsNewObjectForEffect(item) {
		return nil
	}
	src := findBattlefieldCard(g, item.SourceCardID)
	if src == nil {
		return nil
	}
	// A data record (ADR 0041 phase 3, tier 3a): a table holding a
	// prowess pump is still a restore point.
	g.RegisterScopedEffectForEffect(src.InstanceID,
		[]AffectedObject{PinObject(src.InstanceID, src.EnteredBattlefieldAt)},
		[]Mod{ModifyPTMod(1, 1)}, g.UntilEndOfTurnDuration(), prowessLabel)
	return nil
}

// keywordTriggersFor is the keyword triggers an object has right now:
// one prowess trigger per prowess token in its ability list (CR
// 702.108b).
//
// It reads the list through forEachAbilityToken, the same walk
// HasKeyword uses, so the answer follows the layer engine on the
// battlefield — a printed prowess, a token's, and a layer-6 grant all
// count, and a CR 613.1f "loses all abilities" empties the list in its
// own timestamp slot (ability_removal.go). A face-down permanent has
// no tokens and so no prowess.
//
// Off the battlefield the walk reads the card's printed keywords, and
// the ability it returns declares no Zones — so it is a battlefield
// ability and the zone harvesters skip it, exactly as they skip a
// catalog trigger that did not declare their zone.
func keywordTriggersFor(c *Card) []TriggeredAbility {
	n := 0
	forEachAbilityToken(c, func(a string) bool {
		if a == KeywordProwess {
			n++
		}
		return true
	})
	if n == 0 {
		return nil
	}
	out := make([]TriggeredAbility, n)
	for i := range out {
		out[i] = prowessTrigger
	}
	return out
}

// ProwessCount reports how many instances of prowess the card has — a
// read for tests, the bot and the view, through the same walk the
// harvester uses.
func ProwessCount(c *Card) int {
	return len(keywordTriggersFor(c))
}

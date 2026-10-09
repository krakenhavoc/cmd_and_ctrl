package game

// modular.go — CR 702.43, modular as a keyword the engine derives from
// an object's ability list (#2012).
//
//	702.43a Modular represents both a static ability and a triggered
//	        ability. "Modular N" means "This permanent enters with N
//	        +1/+1 counters on it" and "When this permanent is put into
//	        a graveyard from the battlefield, you may put a +1/+1
//	        counter on target artifact creature for each +1/+1 counter
//	        on this permanent."
//	702.43b If a permanent has multiple instances of modular, each one
//	        works separately.
//
// NUMBERED like annihilator: the tokens are "modular N", minted by
// CanonicalModularToken, and a bare "modular" is refused (Scryfall's
// keyword array carries only the word; the number is on the line).
// CUMULATIVE (CR 702.43b).
//
// THE STATIC is an entry replacement (CR 614.1c), so it is gathered the
// way riot's and unleash's are (riot.go): the entry look-ahead counts
// the instances the permanent would have on the battlefield (CR 614.12)
// and the gather adds one mandatory self-replacement per instance, each
// putting its N +1/+1 counters on the entry event. The counters land
// with the permanent and are drained through the ordinary counter
// window, so Hardened Scales sees them. Every entry path gets them — a
// cast, a reanimation, a flicker — because the gather runs on every
// battlefield entry, which is the difference from sunburst's cast-only
// seed. A modular instance commutes with every other entry replacement
// (it only adds counters), so one is applied alone, after the CR 616.1b
// control tier, rather than asking the player to order it.
//
// THE TRIGGER is a dies keyword like undying (undying_persist.go):
// harvestLTB derives one per instance from the LAST-KNOWN ability list
// (CR 603.10a), so a granted modular it died wearing counts and one
// that died having lost its abilities has none. "For each +1/+1 counter
// on this permanent" is its last-known counters (the departed object's
// PermanentInfo record), read on resolution. It is a "you may" (CR
// 603.5) with a target artifact creature; the target is re-checked on
// resolution (CR 608.2b), and the counters are put on through the
// ordinary counter door, so Hardened Scales applies.

import (
	"strconv"

	"github.com/google/uuid"
)

// KeywordModular is CR 702.43's family key. The tokens are
// "modular N"; see CanonicalModularToken.
const KeywordModular = "modular"

// maxModularValue bounds what ModularValue accepts. Printed modular
// tops out at 6 (Arcbound Overseer).
const maxModularValue = 999

// modularReplacementIDBase is the sixth stride of the self-replacement
// range: riot's is the third and unleash's the fourth (riot.go), and
// read ahead's single ID opens the fifth (read_ahead.go).
const modularReplacementIDBase = selfReplacementIDBase + 5*MaxCatalogReplacementSlots

// ModularValue parses one ability token as "modular N".
func ModularValue(token string) (int, bool) {
	return numberedKeywordValue(token, KeywordModular, maxModularValue)
}

// CanonicalModularToken normalises one printed modular clause to the
// engine's form — "Modular 2" becomes "modular 2" — reporting whether
// it is one at all.
func CanonicalModularToken(s string) (string, bool) {
	n, ok := ModularValue(s)
	if !ok {
		return "", false
	}
	return KeywordModular + " " + strconv.Itoa(n), true
}

// ModularAmounts is the N of every modular instance the card has, in
// ability-list order. A read for tests, the bot and the view.
func ModularAmounts(c *Card) []int {
	var out []int
	forEachAbilityToken(c, func(a string) bool {
		if n, ok := ModularValue(a); ok {
			out = append(out, n)
		}
		return true
	})
	return out
}

// modularReplacement is one modular N instance's entry replacement:
// "this permanent enters with N +1/+1 counters on it".
func modularReplacement(n int) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.AddCounterAtETB(CounterPlusOne, n)
			return nil
		},
		Controller:      entryKeywordController,
		SelfReplacement: true,
		Label:           "Modular " + strconv.Itoa(n),
		entryKeyword:    KeywordModular,
	}
}

// firstModular is the index of the first modular instance among the
// applicable replacements, or -1.
func firstModular(applicable []activeReplacement) int {
	for i, a := range applicable {
		if a.effect.entryKeyword == KeywordModular {
			return i
		}
	}
	return -1
}

// modularMoveCountersBody is the dies trigger's body, registered under
// "modular/move-counters" (ADR 0041 P9): an on-disk identity, never
// renamed.
var modularMoveCountersBody = SimpleDelayedBody("modular/move-counters", resolveModular)

// modularLabel is the stack label of the dies trigger.
const modularLabel = "Modular — put its +1/+1 counters on target artifact creature"

// modularTargetSpec is "target artifact creature".
func modularTargetSpec() *TargetSpec {
	return &TargetSpec{
		Mode:  "creature",
		Label: "target artifact creature",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature() && c.IsArtifact()
		},
		Min: 1, Max: 1,
	}
}

// modularTrigger is the one TriggeredAbility a single instance of
// modular is. A package-level value, never mutated:
// ltbKeywordTriggersFor hands out copies.
var modularTrigger = TriggeredAbility{
	Keyword: KeywordModular,
	Watches: []EventKind{EventLTB},
	// "When this permanent is put into a graveyard from the
	// battlefield". A permanent exiled instead was never put into a
	// graveyard; a token was, before it ceased to exist (CR 111.7).
	AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
		return ev.CardID != uuid.Nil && ev.CardID == source.InstanceID && ev.NewZone == ZoneGraveyard
	},
	OptionalPrompt: &TriggerOptionalPrompt{Question: "Modular — put its +1/+1 counters on target artifact creature?"},
	Targets:        modularTargetSpec(),
	Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
		return NewKeyedTriggeredItem(source, modularLabel, modularMoveCountersBody, EffectParams{})
	},
}

// resolveModular puts one +1/+1 counter on the target for each +1/+1
// counter the departed permanent last had (CR 603.10a), if the target
// is still an artifact creature on the battlefield (CR 608.2b). A
// package-level func, so it captures nothing and survives Clone.
func resolveModular(g *Game, item *StackItem) error {
	if item.Trigger == nil || item.Trigger.Object == nil || len(item.Targets) == 0 {
		return nil
	}
	departed, ok := g.PermanentForEffect(item.Trigger.Object.Ref())
	if !ok {
		return nil
	}
	n := departed.Counters[CounterPlusOne]
	if n <= 0 {
		return nil
	}
	target := findBattlefieldCard(g, item.Targets[0].ID)
	if target == nil || !target.IsCreature() || !target.IsArtifact() {
		return nil
	}
	return g.AddCounterForEffect(target.InstanceID, CounterPlusOne, n)
}

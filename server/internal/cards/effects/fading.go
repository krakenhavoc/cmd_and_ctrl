package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fading.go — CR 702.32 (#5802).
//
//	702.32a Fading is a keyword that represents two abilities. "Fading N"
//	        means "This permanent enters with N fade counters on it" and
//	        "At the beginning of your upkeep, remove a fade counter from
//	        this permanent. If you can't, sacrifice the permanent."
//
// Vanishing's sibling (cmdr2B17Vanishing*, CR 702.63a), built the same
// way: the entry counters are a self-replacement on the card's own entry
// (b10EntersWithCounters), so they go through the counter window and a
// Doubling Season doubles them, and the upkeep half is one trigger. The
// two keywords differ in when the permanent goes:
//
//   - Vanishing's upkeep trigger has an intervening "if" (it needs a time
//     counter to trigger at all) and a SECOND trigger sacrifices it when
//     the last time counter is removed. A Vanishing 3 permanent is gone
//     at the third upkeep.
//   - Fading's upkeep trigger has no "if": it always triggers, and as it
//     resolves it removes a fade counter, or sacrifices the permanent if
//     there is none to remove. Taking the last counter off does nothing
//     by itself, so a Fading 3 permanent survives three upkeeps and is
//     sacrificed at the fourth. A "Remove a fade counter" cost spends the
//     same counters, and a permanent that has spent them all is
//     sacrificed at the next upkeep.
//
// Nobody chooses anything: whether there is a counter to remove is read
// as the trigger resolves. A permanent that left the battlefield, or
// left and came back (a new object, CR 400.7), is neither touched nor
// sacrificed.
//
// Not a canonical keyword token, for echo's reason: the number lives on
// the entry replacement. The trigger carries KeywordFading in
// TriggeredAbility.Keyword, which is how a card's fading is found
// (TestEveryFadingCardHasBothHalves).

// KeywordFading is the Keyword of a fading upkeep trigger.
const KeywordFading = "fading"

// fadingUpkeepLabel is the fading trigger's row: its catalog name, so
// never renamed.
const fadingUpkeepLabel = "Fading — remove a fade counter from it; if you can't, sacrifice it"

// Fading is the first of fading's two abilities (CR 702.32a): "This
// permanent enters with N fade counters on it." Put it in Replacements
// and FadingUpkeep() in Triggered:
//
//	Replacements: []game.ReplacementEffect{Fading(3)},
//	Triggered:    []game.TriggeredAbility{FadingUpkeep()},
func Fading(n int) game.ReplacementEffect {
	r := b10EntersWithCounters(game.CounterFade, n, "Fading: enters with fade counters on it")
	r.RowLabel = fmt.Sprintf("Fading %d", n)
	return r
}

// FadingEntersTapped is Fading(n) for a permanent that also "enters
// tapped" (Skyshroud Behemoth). One replacement doing both, as
// SelfEntersTappedWithCounters does, because two self-replacements on
// the same entry would ask the controller to order them for a choice
// that changes nothing (CR 616.1).
func FadingEntersTapped(n int) game.ReplacementEffect {
	r := Fading(n)
	r.SelfEntersTapped = true
	r.RowLabel = fmt.Sprintf("Fading %d. Enters tapped.", n)
	r.Replace = func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
		ev.EntersTapped = true
		ev.AddCounterAtETB(game.CounterFade, n)
		return nil
	}
	return r
}

// FadingUpkeep is the second (CR 702.32a): "At the beginning of your
// upkeep, remove a fade counter from this permanent. If you can't,
// sacrifice the permanent."
func FadingUpkeep() game.TriggeredAbility {
	t := AtYourUpkeep(fadingUpkeepLabel, fadingRemoveOrSacrifice)
	t.Keyword = KeywordFading
	return t
}

// fadingRemoveOrSacrifice is the upkeep trigger's resolution. "If you
// can't" is read now, not as the upkeep began: a counter spent in
// response (Jolting Merfolk's cost) is one the trigger can no longer
// remove.
func fadingRemoveOrSacrifice(g *game.Game, item *game.StackItem) error {
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	if c.Counters[game.CounterFade] > 0 {
		return AddCounter{Target: item.SourceCardID, Kind: game.CounterFade, N: -1}.Apply(NewContext(g, item))
	}
	return SacrificeThisIfStillOnBattlefield(g, item)
}

// RemoveAFadeCounterFromThis is the "Remove a fade counter from this
// <permanent>" cost several fading cards print beside the keyword. It is
// the ordinary counter-removal cost, paid as the ability is activated
// (CR 602.2b), so it can't be paid with no fade counter left.
func RemoveAFadeCounterFromThis() game.AbilityCost {
	return RemoveCountersFromThis(game.CounterFade, 1)
}

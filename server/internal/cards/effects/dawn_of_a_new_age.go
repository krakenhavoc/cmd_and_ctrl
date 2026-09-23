package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dawn of a New Age — Enchantment {1}{W} (EDHREC rank 4147):
//
//	"This enchantment enters with a hope counter on it for each
//	 creature you control.
//	 At the beginning of your end step, remove a hope counter from
//	 this enchantment. If you do, draw a card. Then if this
//	 enchantment has no hope counters on it, sacrifice it and you
//	 gain 4 life."
//
// Two mana that draws you one card a turn for as many turns as you
// had creatures when it landed, and then pays 4 life on the way out.
// A token deck casting this after a Secure the Wastes is drawing
// five or six cards for two mana; cast on an empty board it is a
// two-mana "gain 4 life" at the end of the turn, which is the
// printed floor and is not a bug.
//
// THE COUNT IS TAKEN AS IT ENTERS, ONCE, AND NEVER AGAIN. Creatures
// that arrive later add nothing — it is an entry replacement
// (CR 614.1c), not a static. The count runs inside the replacement,
// so a creature that died in response is already gone and is not
// counted, exactly as printed. Dawn of a New Age is an enchantment
// and never counts itself.
//
// THE END-STEP TRIGGER HAS NO INTERVENING "IF", so it fires on every
// one of its controller's end steps regardless of how many counters
// are left — which is what makes the last chapter work. Read the
// three clauses in order:
//
//  1. "Remove a hope counter" — with none left, nothing is removed.
//  2. "IF YOU DO, draw a card" — conditional on the removal actually
//     happening. No counter, no draw.
//  3. "THEN if it has no hope counters, sacrifice it and gain 4" —
//     checked AFTER the removal, so the turn that spends the last
//     counter both draws the card and cashes the enchantment in.
//     That ordering is the whole card and it is why this is one
//     trigger body rather than three primitives in a Do().
//
// "YOUR END STEP" is the controller's, matched on the event's actor,
// so an opponent's end step does nothing. A Dawn that has already
// left the battlefield when the trigger resolves does nothing at all.
//
// The sacrifice is the enchantment's own, so it is a real sacrifice:
// aristocrats payoffs see it, and indestructible would not save it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "37df5ade-6a21-4b0c-89f8-e2e917589a8d",
		Name:         "Dawn of a New Age",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(b39HopeCounter, func(g *game.Game, src *game.Card) int {
				return b39CreaturesControlledBy(g, src.Controller)
			}, "Dawn of a New Age: enters with a hope counter for each creature you control"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			}, "Dawn of a New Age — remove a hope counter and draw", b39DawnOfANewAgeEndStep),
		},
	})
}

// b39HopeCounter is Dawn of a New Age's printed counter kind. Counter
// kinds are free-form strings (Gemstone Mine's "mining" is the
// precedent); naming it here keeps the entry replacement and the end
// step reading the same one.
const b39HopeCounter = "hope"

// b39DawnOfANewAgeEndStep is the end-step body, in printed order:
// remove one counter if there is one, draw only if a counter actually
// came off, then check for zero and cash the enchantment in.
func b39DawnOfANewAgeEndStep(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	source := item.SourceCardID

	if b39CountersOn(g, source, b39HopeCounter) > 0 {
		// #1282: the draw and the zero check are the removal's
		// continuation. The removal runs the CR 614 counter window,
		// and a window that pauses on a CR 616 ordering prompt removes
		// nothing until the prompt is answered — read on the next
		// line, the check would still see the last hope counter and
		// never cash the enchantment in.
		return g.AddCounterThenForEffect(source, b39HopeCounter, -1, func(g *game.Game, _ int) error {
			if err := (DrawCards{Player: item.Controller, N: 1}).Apply(NewContext(g, item)); err != nil {
				return err
			}
			return b39DawnOfANewAgeCashIn(g, item)
		})
	}
	return b39DawnOfANewAgeCashIn(g, item)
}

// b39DawnOfANewAgeCashIn is the end step's last sentence: with no hope
// counters left, sacrifice the enchantment and gain 4 life.
func b39DawnOfANewAgeCashIn(g *game.Game, item *game.StackItem) error {
	if b39CountersOn(g, item.SourceCardID, b39HopeCounter) > 0 {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Player: item.Controller, Amount: 4}.Apply(ctx)
}

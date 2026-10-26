package effects

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// route_batch_tails.go — #4253 slice 9: the "then" of a route batch (an
// exile, a destroy, a sacrifice, a bounce, a tuck, a mill, a discard) and
// of manifest dread written as a registered tail (game.RegisterTail) plus
// the plain data it reads, so a table waiting on a CR 616 prompt in the
// middle of the batch is a restore point.
//
// A batch's tail is told the cards that went THIS way (CR 400.7) as
// TailOutcome.IDs, in batch order: the cards that reached exile, the
// permanents destroyed, the cards milled. A single-card door's tail is
// told one ID or none. Register a body with CardTail; the helpers below
// are the shapes several cards share. Keys are on-disk identities
// (ledgered "tail <key>"): never renamed, never reused.

// MassTail registers the "for each permanent <verb> this way" clause of a
// mass verb (DestroyAllMatching, ExileAllMatching, BounceAllMatching): the
// body is told the pre-move copies of the permanents that went this way
// (CR 603.10, the batch's last-known information) and how many there were.
func MassTail(key string, body func(ctx *Context, swept []game.Card, moved int) error) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		return body(ctx, ctx.Game.SimultaneousExitCardsForEffect(o.IDs), len(o.IDs))
	})
}

// --- manifest dread ----------------------------------------------------------

// ManifestDreadTail registers a manifest dread's "then", told what the
// action did (game.ManifestDreadResult): the permanent that entered face
// down and the card put into the graveyard, either uuid.Nil.
func ManifestDreadTail(key string, body func(ctx *Context, res game.ManifestDreadResult, o game.TailOutcome) error) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		return body(ctx, game.ManifestDreadOutcome(o), o)
	})
}

// manifestDreadAttachSource is "then attach this Equipment to that
// creature" (Conductive Machete, Cursed Windbreaker, Dissection Tools,
// Killer's Mask): the source is the tail's Source.
var manifestDreadAttachSource = ManifestDreadTail("manifest-dread/attach-source", func(ctx *Context, res game.ManifestDreadResult, o game.TailOutcome) error {
	if res.Manifested == uuid.Nil {
		return nil
	}
	return ctx.Game.AttachForEffect(o.Tail.Source, game.TargetRef{Kind: game.TargetCard, ID: res.Manifested})
})

// manifestDreadPutCounters is "put <counters> on that creature": the
// counters are the tail's Word, "kind=n" pairs joined by ",".
var manifestDreadPutCounters = ManifestDreadTail("manifest-dread/put-counters", func(ctx *Context, res game.ManifestDreadResult, o game.TailOutcome) error {
	if res.Manifested == uuid.Nil || !b15OnBattlefield(ctx.Game, res.Manifested) {
		return nil
	}
	for _, c := range decodeCounterAmounts(o.Tail.Word) {
		if err := ctx.Game.AddCounterForEffect(res.Manifested, c.Kind, c.N); err != nil {
			return err
		}
	}
	return nil
})

// manifestDreadAgain is "manifest dread N times": N is how many are left
// after this one. Set in init, because its body names itself.
var manifestDreadAgain game.TailRef

func init() {
	manifestDreadAgain = ManifestDreadTail("manifest-dread/again", func(ctx *Context, _ game.ManifestDreadResult, o game.TailOutcome) error {
		if o.Tail.N <= 0 {
			return nil
		}
		return ctx.Game.ManifestDreadKeyedForEffect(o.Tail.Player, o.Tail.Source,
			game.KeyedTail{Ref: manifestDreadAgain, Player: o.Tail.Player, Source: o.Tail.Source, N: o.Tail.N - 1})
	})
}

// encodeCounterAmounts writes counters as a tail's Word.
func encodeCounterAmounts(counters []CounterAmount) string {
	parts := make([]string, 0, len(counters))
	for _, c := range counters {
		parts = append(parts, c.Kind+"="+strconv.Itoa(c.N))
	}
	return strings.Join(parts, ",")
}

// decodeCounterAmounts reads encodeCounterAmounts back.
func decodeCounterAmounts(w string) []CounterAmount {
	var out []CounterAmount
	for _, part := range strings.Split(w, ",") {
		kind, n, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		v, err := strconv.Atoi(n)
		if err != nil {
			continue
		}
		out = append(out, CounterAmount{Kind: kind, N: v})
	}
	return out
}

// --- mill runs' stop clauses -------------------------------------------------

// The stop clauses of "mills a card, then repeats this process until …"
// (and "exile cards from the top until …"), registered so a run is data
// (game.RegisterMillUntil, ledgered "milluntil <key>"). Each is asked with
// the cards that have landed so far, as they were in the library.
var (
	untilNonlandRef = game.RegisterMillUntil("until/nonland-card", func(landed []game.Card, _ int) bool {
		return len(landed) > 0 && !landed[len(landed)-1].IsLand()
	})
	untilLandRef = game.RegisterMillUntil("until/land-card", func(landed []game.Card, _ int) bool {
		return len(landed) > 0 && landed[len(landed)-1].IsLand()
	})
	untilLegendaryRef = game.RegisterMillUntil("until/legendary-card", func(landed []game.Card, _ int) bool {
		return len(landed) > 0 && landed[len(landed)-1].IsLegendary()
	})
	untilTotalManaValueRef = game.RegisterMillUntil("until/total-mana-value", func(landed []game.Card, n int) bool {
		total := 0
		for _, c := range landed {
			total += c.ManaValue()
		}
		return total >= n
	})
	// Helm of Obedience: "until a creature card or X cards have been put
	// into their graveyard this way, whichever comes first" (#1161).
	untilCreatureOrCountRef = game.RegisterMillUntil("until/creature-card-or-count", func(landed []game.Card, n int) bool {
		if len(landed) == 0 {
			return false
		}
		return landed[len(landed)-1].IsCreature() || (n > 0 && len(landed) >= n)
	})
)

// UntilNonland is "until they exile a nonland card".
var UntilNonland = game.MillUntil{Ref: untilNonlandRef}

// UntilLand is "until a land card".
var UntilLand = game.MillUntil{Ref: untilLandRef}

// UntilLegendary is "until a legendary card".
var UntilLegendary = game.MillUntil{Ref: untilLegendaryRef}

// UntilTotalManaValueOf is "until they have exiled cards with total mana
// value N or greater".
func UntilTotalManaValueOf(n int) game.MillUntil {
	return game.MillUntil{Ref: untilTotalManaValueRef, N: n}
}

// UntilCreatureOrCount is "until a creature card or N cards".
func UntilCreatureOrCount(n int) game.MillUntil {
	return game.MillUntil{Ref: untilCreatureOrCountRef, N: n}
}

// withCarry sets the IDs a tail's body reads.
func withCarry(t game.KeyedTail, ids ...uuid.UUID) game.KeyedTail {
	t.Carry = append([]uuid.UUID(nil), ids...)
	return t
}

// tailCarry0 is the first carried ID, or uuid.Nil.
func tailCarry0(o game.TailOutcome) uuid.UUID {
	if len(o.Tail.Carry) == 0 {
		return uuid.Nil
	}
	return o.Tail.Carry[0]
}

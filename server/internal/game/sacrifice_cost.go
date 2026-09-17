package game

import (
	"sort"

	"github.com/google/uuid"
)

// sacrifice_cost.go — #747 (ADR 0020 addendum §12–§15, ADR 0021
// addendum): a sacrifice cost of N permanents. "Sacrifice two
// artifacts" (Sai), "Sacrifice three Foods" (Samwise Gamgee),
// "Sacrifice five Treasures" (Magda, Brazen Outlaw).
//
// There is no new cost component. The count lives on the sacrifice
// clause itself, as the TargetSpec's Min == Max, so the three cost
// sites that already carry the clause — AbilityCost.SacrificeOther,
// ManaAbilityShape.SacrificeOther and AdditionalCost.Sacrifice — get
// the count with no new field, and a cost merge (effects.Plus) cannot
// separate the count from the thing it counts.
//
// Three pieces live here because all three sites and both readers
// (the legal enumerator and the protocol view) share them:
//
//   - SacrificeCostCount reads the count off a clause.
//   - payCostSacrificesLocked pays one payment as one simultaneous
//     exit.
//   - SacrificePaymentOrderForEffect is the policy-neutral order the
//     enumerator takes its one payment from, and the order the view
//     ships the options in, so the client's "Choose for me" fills the
//     picker with the same permanents a bot would pay.

// SacrificeCostCount is how many permanents a sacrifice clause pays:
// the spec's Max. effects.Register refuses a sacrifice clause whose
// Min differs from its Max, or whose count is below 1, so for every
// catalog clause Min == Max == N. A hand-built fixture clause that
// never set a count is read as 1, the count every sacrifice cost had
// before #747. Nil-safe: a nil clause pays nothing.
func SacrificeCostCount(spec *TargetSpec) int {
	if spec == nil {
		return 0
	}
	if spec.Max < 1 {
		return 1
	}
	return spec.Max
}

// payCostSacrificesLocked sacrifices every permanent of one cost
// payment — the source when the cost sacrifices it, and the N
// permanents of the clause — as ONE simultaneous battlefield exit.
//
// Each permanent still goes through sacrificePermanentLocked, so each
// emits its own EventSacrifice and takes its own route to the
// graveyard: "whenever you sacrifice a permanent" and dies triggers
// fire once per permanent. What the batch adds is what the watchers
// see. Leaves-the-battlefield and sacrifice triggers look back in time
// (CR 603.10a), so a Blood Artist sacrificed alongside a Goblin must
// see both deaths. Paid one at a time with no batch, the Artist would
// see the Goblin only when the Goblin happened to go first, and the
// trigger count would depend on the order of IDs on the wire.
//
// beginSimultaneousExitLocked has no indestructible filter (that lives
// in DestroyPermanentsForEffect, which this does not call), so an
// indestructible permanent is still sacrificed, as CR 701.21a says.
//
// Caller must hold g.mu in write mode, and must have validated the
// payment with validateSacrificeCostLocked.
func (g *Game) payCostSacrificesLocked(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	defer g.beginSimultaneousExitLocked(ids)()
	for _, id := range ids {
		if err := g.sacrificePermanentLocked(id); err != nil {
			return err
		}
	}
	return nil
}

// SacrificePaymentOrderForEffect returns `ids` — candidates for a
// sacrifice cost — in the order a payment takes them from (ADR 0020
// addendum §15):
//
//  1. tokens before nontokens;
//  2. then lower mana value;
//  3. then the ability's own source last, when it is a candidate
//     (pass uuid.Nil when there is no source, as for a spell's
//     additional cost);
//  4. then the order given, which callers pass in board order, so the
//     result is stable.
//
// The order is policy-neutral on purpose. The legal enumerator offers
// one sacrifice-N payment per move — the first N of this order —
// rather than every combination, and `legal` must not import a bot
// policy (#687). The protocol view ships its options in the same
// order, so the owner's "Choose for me" button (#747) fills the picker
// with exactly the permanents a bot would have paid.
//
// A new slice; `ids` is not modified. IDs not on the battlefield sort
// as nontokens of mana value 0, which no caller passes.
//
// Caller must hold g.mu (read or write).
func (g *Game) SacrificePaymentOrderForEffect(ids []uuid.UUID, sourceID uuid.UUID) []uuid.UUID {
	type key struct {
		id     uuid.UUID
		token  bool
		mv     int
		source bool
	}
	keys := make([]key, 0, len(ids))
	for _, id := range ids {
		k := key{id: id, source: sourceID != uuid.Nil && id == sourceID}
		if c := findBattlefieldCard(g, id); c != nil {
			k.token = typeLineHas(c.TypeLine, "token")
			k.mv = c.ManaValue()
		}
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.token != b.token {
			return a.token
		}
		if a.mv != b.mv {
			return a.mv < b.mv
		}
		if a.source != b.source {
			return b.source
		}
		return false
	})
	out := make([]uuid.UUID, len(keys))
	for i, k := range keys {
		out[i] = k.id
	}
	return out
}

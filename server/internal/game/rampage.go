package game

// rampage.go — CR 702.23, rampage as a NUMBERED keyword trigger the
// engine derives from an object's ability list rather than from its
// catalog entry (#5769). Annihilator (annihilator.go) is the model for
// the number and melee (melee.go) for the count read on resolution.
//
//	702.23a Rampage is a triggered ability. "Rampage N" means "Whenever
//	        this creature becomes blocked, it gets +N/+N until end of
//	        turn for each creature blocking it beyond the first." (See
//	        rule 509, "Declare Blockers Step.")
//	702.23b The rampage bonus is calculated only once per combat, when
//	        the triggered ability resolves. Adding or removing blockers
//	        later in combat won't change the bonus.
//	702.23c If a creature has multiple instances of rampage, each
//	        triggers separately.
//
// THE TOKEN. Rampage carries a number, so it is stored the way
// annihilator is: KeywordRampage is the FAMILY key canonicalKeywords
// holds, the wire tokens are "rampage N" (minted only by
// CanonicalRampageToken), and CanonicalKeywords refuses a bare
// "rampage". Scryfall's keywords array says only "Rampage", so the deck
// importer reads the number off the oracle line. It is CUMULATIVE (CR
// 702.23c): AppendKeywordAbility keeps a granted "rampage 2" beside a
// printed one, and each is its own trigger. A token is what lets a
// creature whose only text is keywords and rampage (Craw Giant,
// Wolverine Pack, …) need no card file, and what reaches a granted
// rampage through the same walk.
//
// THE TRIGGER. "Whenever this creature becomes blocked" watches
// EventBecomesBlocked for the source itself, which the block lock-in
// emits once per attacker however many creatures block it (CR 509.3c),
// and which a creature put onto the battlefield blocking emits only for
// an attacker that was unblocked (put_blocking.go).
//
// THE COUNT (CR 702.23b). The creatures blocking it are counted as the
// ability RESOLVES: a blocker that has left the battlefield, or been
// removed from combat, by then is not blocking it, and one added later
// changes nothing, because the bonus is registered once as a fixed
// +N/+N. A creature that is itself no longer attacking is blocked by
// nothing (CR 506.4). The bonus is N times the blockers beyond the
// first, so one blocker or none is a +0/+0 that is not registered.
//
// THE PUMP. The creature is fixed as the ability triggers, by instance
// ID and object epoch (Params.Object): one that left the battlefield
// and came back is a new object and gets nothing (CR 400.7). The
// +N/+N is a layer-7c scoped effect until end of turn.
//
// ORDERING (CR 603.3b). Rampage triggers commute: each reads the
// current blockers, which no rampage changes, and writes only a +N/+N
// to its own pinned creature, so a batch of nothing but rampage
// triggers needs no ordering prompt.

import (
	"strconv"

	"github.com/google/uuid"
)

// KeywordRampage is the bare word CR 702.23 spells with a number after
// it. Like KeywordAnnihilator it is the family key, not a token: the
// tokens are "rampage N".
const KeywordRampage = "rampage"

// RampageValue parses one ability token as CR 702.23's numbered
// keyword, reporting the N it carries. "rampage" alone, "rampage 0" and
// "rampage two" answer false.
func RampageValue(token string) (int, bool) {
	return numberedKeywordValue(token, KeywordRampage)
}

// CanonicalRampageToken normalises one printed rampage clause to the
// engine's wire form — "Rampage 02" becomes "rampage 2" — reporting
// whether it is one at all. The only thing that mints the token;
// CanonicalKeywords calls it.
func CanonicalRampageToken(s string) (string, bool) {
	n, ok := RampageValue(s)
	if !ok {
		return "", false
	}
	return KeywordRampage + " " + strconv.Itoa(n), true
}

// RampageAmounts is the N of every rampage instance the card has, in
// ability-list order, through the walk keywordTriggersFor reads. A read
// for tests, the bot and the view.
func RampageAmounts(c *Card) []int {
	var out []int
	forEachAbilityToken(c, func(a string) bool {
		if n, ok := RampageValue(a); ok {
			out = append(out, n)
		}
		return true
	})
	return out
}

// rampageLabel is the stack label of one rampage N trigger, and the
// label of its +N/+N.
func rampageLabel(n int) string {
	s := strconv.Itoa(n)
	return "Rampage " + s + " — +" + s + "/+" + s + " until end of turn for each creature blocking it beyond the first"
}

// rampagePumpKey is the body every rampage trigger names (ADR 0041 P9):
// an on-disk identity, never renamed or reused. Params.Object is the
// blocked creature the ability is on and Params.Amount is N.
const rampagePumpKey = "rampage/pump"

// rampagePumpBody is the registered body's reference, built from the
// key and registered in init, as melee's is.
var rampagePumpBody = BodyRef{key: rampagePumpKey}

func init() {
	DelayedBody(rampagePumpKey, func(g *Game, item *StackItem, p EffectParams) error {
		return g.resolveRampageLocked(item, p.Object, p.Amount)
	})
}

// rampageTriggerFor is the one TriggeredAbility a single instance of
// "rampage N" is.
func rampageTriggerFor(n int) TriggeredAbility {
	label := rampageLabel(n)
	return TriggeredAbility{
		Keyword: KeywordRampage,
		Watches: []EventKind{EventBecomesBlocked},
		AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
			return ev.Kind == EventBecomesBlocked && ev.CardID != uuid.Nil && ev.CardID == source.InstanceID
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			// Keyed directly, like annihilator: rampage has no catalog
			// row, and a table with one waiting on the stack is still a
			// restore point.
			self := ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}
			item := NewKeyedTriggeredItem(source, label, rampagePumpBody, EffectParams{Object: self, Amount: n})
			// Rampage instances commute with each other: see the file
			// comment.
			item.Commutes = true
			return item
		},
	}
}

// CreaturesBlockingForEffect is how many creatures are blocking the
// attacker `id` right now: CR 702.23a's count, read as a rampage
// trigger resolves. A blocker that has left the battlefield or been
// removed from combat is not blocking it, and an attacker that is no
// longer attacking is blocked by nothing (CR 506.4).
//
// Caller must hold g.mu (read or write).
func (g *Game) CreaturesBlockingForEffect(id uuid.UUID) int {
	atk := findBattlefieldCard(g, id)
	if atk == nil || atk.AttackingTarget == uuid.Nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.IsBlockingAttacker(id) && c.IsCreature() {
			n++
		}
	}
	return n
}

// resolveRampageLocked is "it gets +N/+N until end of turn for each
// creature blocking it beyond the first" for the creature the trigger
// pinned, counting as the ability resolves (CR 702.23b). A creature no
// longer on the battlefield as the same object gets nothing (CR 400.7).
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) resolveRampageLocked(item *StackItem, self ObjectRef, n int) error {
	if item == nil || self.ID == uuid.Nil || n <= 0 {
		return nil
	}
	c := findBattlefieldCard(g, self.ID)
	if c == nil || c.ObjectEpoch != self.Epoch {
		return nil
	}
	beyond := g.CreaturesBlockingForEffect(c.InstanceID) - 1
	if beyond <= 0 {
		return nil
	}
	bonus := n * beyond
	g.RegisterScopedEffectForEffect(item.SourceCardID,
		[]AffectedObject{PinObject(c.InstanceID, c.EnteredBattlefieldAt)},
		[]Mod{ModifyPTMod(bonus, bonus)}, g.UntilEndOfTurnDuration(), rampageLabel(n))
	return nil
}

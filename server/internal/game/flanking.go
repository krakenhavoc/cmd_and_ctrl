package game

// flanking.go — CR 702.25, flanking as a KEYWORD TRIGGER the engine
// derives from an object's ability list rather than from its catalog
// entry (#5784). Exalted (exalted.go) and melee (melee.go) are the
// models.
//
//	702.25a Flanking is a triggered ability that triggers during the
//	        declare blockers step. (See rule 509, "Declare Blockers
//	        Step.") "Flanking" means "Whenever this creature becomes
//	        blocked by a creature without flanking, the blocking
//	        creature gets -1/-1 until end of turn."
//	702.25b If a creature has multiple instances of flanking, each
//	        triggers separately.
//
// WHY A TOKEN. #5757 made flanking a canonical keyword, but its trigger
// stayed a catalog row (effects.Flanking) on each printed card. A row
// reaches only the card that declares it: not a creature an Aura gives
// flanking to (Agility), not one that loses flanking (Barbed Foliage's
// "it loses flanking until end of turn" left the row firing), and not an
// uncatalogued card the deck importer stamped with Scryfall's
// "Flanking". A token in Characteristic.Abilities reaches all of them
// through one walk, keywordTriggersFor, and a creature that loses the
// keyword, or all its abilities (CR 613.1f), loses the trigger with it.
//
// THE OLD ROWS. The flanking cards catalogued before #5784 still carry
// their old row (effects.RetiredFlankingRow), but it no longer
// triggers: it is kept so a restore point written before #5784, with a
// row's trigger waiting on the stack, still resolves through the catalog
// (ADR 0041 P9) and the #522 parity count does not see a row go
// missing. Every new flanking trigger is this file's, so no instance
// triggers twice.
//
// MULTIPLE INSTANCES. CR 702.25b: each instance triggers separately.
// Flanking is CUMULATIVE (KeywordIsCumulative), so AppendKeywordAbility
// keeps a granted flanking beside a printed one, and each is its own
// trigger and its own -1/-1 per blocker.
//
// THE TRIGGER. "Becomes blocked by a creature" triggers once for each
// creature that blocks it (CR 509.3d). EventBlock names the blocker in
// CardID and the attacker in Target, one per pair of the final
// declaration, so a creature blocked by two creatures without flanking
// shrinks both. Whether the blocker has flanking is read as it blocks:
// a trigger condition is checked when the event happens (CR 603.2), and
// the ability has no intervening "if" to read again on resolution.
//
// THE SHRINK. The blocker is fixed as the ability triggers, by instance
// ID and object epoch (Params.Object). A blocker that leaves the
// battlefield in response, or leaves and comes back as a new object,
// gets nothing (CR 400.7). The -1/-1 is a layer-7c scoped effect until
// end of turn.
//
// ORDERING (CR 603.3b). Flanking triggers commute: each reads only
// whether its pinned blocker is still the same object and writes only a
// -1/-1 to it, so a batch of nothing but flanking triggers needs no
// ordering prompt.

import "github.com/google/uuid"

// flankingLabel is the stack label every flanking trigger carries, and
// the label of its -1/-1.
const flankingLabel = "Flanking — the blocking creature gets -1/-1 until end of turn"

// flankingShrinkKey is the body every flanking trigger names (ADR 0041
// P9): an on-disk identity, never renamed or reused. An older binary
// refuses a restore point naming it with ErrUnknownEffectKey, which is
// the designed rollback case. Params.Object is the blocker.
const flankingShrinkKey = "flanking/shrink"

// flankingShrinkBody is the registered body's reference, built from the
// key and registered in init, as exalted's is.
var flankingShrinkBody = BodyRef{key: flankingShrinkKey}

func init() {
	DelayedBody(flankingShrinkKey, func(g *Game, item *StackItem, p EffectParams) error {
		return g.resolveFlankingLocked(item, p.Object)
	})
}

// flankingTrigger is the one TriggeredAbility a single instance of
// flanking is. A package-level value, never mutated: keywordTriggersFor
// hands out copies of it.
var flankingTrigger = TriggeredAbility{
	Keyword: KeywordFlanking,
	Watches: []EventKind{EventBlock},
	AppliesTo: func(ev Event, source *Card, _ Characteristic, g *Game) bool {
		if ev.Kind != EventBlock || ev.CardID == uuid.Nil || ev.Target != source.InstanceID || ev.CardID == source.InstanceID {
			return false
		}
		blocker, ok := g.LookupCardForEffect(ev.CardID)
		return ok && !HasKeyword(&blocker, KeywordFlanking)
	},
	Build: func(ev Event, source *Card, _ Characteristic, g *Game) *StackItem {
		blocker := ObjectRef{ID: ev.CardID}
		if c := findBattlefieldCard(g, ev.CardID); c != nil {
			blocker.Epoch = c.ObjectEpoch
		}
		item := NewKeyedTriggeredItem(source, flankingLabel, flankingShrinkBody, EffectParams{Object: blocker})
		// Flanking instances commute with each other: see the file
		// comment.
		item.Commutes = true
		return item
	},
}

// resolveFlankingLocked is "the blocking creature gets -1/-1 until end
// of turn" for the blocker the trigger pinned. A blocker no longer on
// the battlefield as the same object gets nothing (CR 400.7). The
// flanking creature leaving does not matter: the ability has no "if".
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) resolveFlankingLocked(item *StackItem, blocker ObjectRef) error {
	if item == nil || blocker.ID == uuid.Nil {
		return nil
	}
	c := findBattlefieldCard(g, blocker.ID)
	if c == nil || c.ObjectEpoch != blocker.Epoch {
		return nil
	}
	g.RegisterScopedEffectForEffect(item.SourceCardID,
		[]AffectedObject{PinObject(c.InstanceID, c.EnteredBattlefieldAt)},
		[]Mod{ModifyPTMod(-1, -1)}, g.UntilEndOfTurnDuration(), flankingLabel)
	return nil
}

// FlankingCount reports how many instances of flanking the card has — a
// read for tests, the bot and the view, through the same walk the
// harvester uses.
func FlankingCount(c *Card) int {
	return countAbilityTokens(c, KeywordFlanking)
}

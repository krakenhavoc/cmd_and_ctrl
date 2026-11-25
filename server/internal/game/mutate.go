package game

import (
	"strings"

	"github.com/google/uuid"
)

// mutate.go — CR 702.140, mutate, and CR 730, merging with permanents
// (ADR 0223, #5910).
//
//	702.140a "Mutate [cost]" means "You may pay [cost] rather than pay
//	this spell's mana cost. If you do, it becomes a mutating creature
//	spell and targets a non-Human creature with the same owner as this
//	spell."
//	702.140b As a mutating creature spell begins resolving, if its
//	target is illegal, it ceases to be a mutating creature spell and
//	continues resolving as a creature spell …
//	702.140c As a mutating creature spell resolves, if its target is
//	legal, it doesn't enter the battlefield. Rather, it merges with the
//	target creature … The spell's controller chooses whether the spell
//	is put on top of the creature or on the bottom.
//	702.140d An ability that triggers whenever a creature mutates
//	triggers when a spell merges with a creature as a result of a
//	resolving mutating creature spell.
//	702.140e A mutated permanent has all abilities of each card and
//	token that represents it. Its other characteristics are derived
//	from the topmost card or token.
//
// # The shape
//
// The cast is an ordinary alternative cost keyed MutateKey whose
// Targets is the "non-Human creature with the same owner as this
// spell" clause (effects.Mutate). A stack item is a mutating creature
// spell while it was cast for that cost and still has its target.
//
// The permanent it merges with is the target's own Card: same
// InstanceID, counters, damage, attachments and effects (CR 730.2c). It
// carries every component in Card.MergedFrom, topmost first; the entry
// with the permanent's own InstanceID stands for the card it was
// before it first merged. The characteristics are CR 730.2a's copiable
// effect, written the way copy.go writes one (applyMergedValues): the
// topmost component's copiable values, its token status (CR 730.2d),
// and for each other component a composite-key part naming that
// component's catalog entry (MergedPartKey), which lends every one of
// its abilities, with its printed keywords unioned in. A Clone copies
// the flat fields, so it copies all of that and is not itself merged.
//
// Leaving the battlefield is meld's exit, generalised (meld.go):
// MoveCard splits the permanent into its components, the permanent's
// own card carrying the object's ID and the others landed as
// passengers (CR 730.3).

// MutateKey is the wire key of the mutate offer: CastSpellParams.
// AlternativeCost and StackItem.AltCost.
const MutateKey = "mutate"

// MutateGraveyardKey is the wire key of the mutate offer bound to the
// graveyard: "You may cast this card from your graveyard using its mutate
// ability" (Brokkos, Apex of Forever). The same price and the same target as
// the mutate offer from the hand, a second offer so each is bound to its one
// zone as blitz's graveyard offer is. A cast claiming either is a mutating
// creature spell (CR 702.140a).
const MutateGraveyardKey = "mutate-graveyard"

// MergedPartPrefix marks a granted-ability part as "every ability of
// the component whose catalog key follows" — a card or token under a
// mutated permanent (CR 702.140e). It carries GrantKeyPrefix, so
// GrantKey leaves it alone.
const MergedPartPrefix = GrantKeyPrefix + "merged:"

// MergedPartKey is the part for a component whose catalog key is
// `key` (its own copy grants are carried as parts of their own). Empty
// for an empty key.
func MergedPartKey(key string) string {
	base := BaseCatalogKey(key)
	if base == "" {
		return ""
	}
	return MergedPartPrefix + base
}

// MergedPartLender is the catalog key a MergedPartKey part names.
func MergedPartLender(part string) (string, bool) {
	if !strings.HasPrefix(part, MergedPartPrefix) {
		return "", false
	}
	return part[len(MergedPartPrefix):], true
}

// MergedPartDef is the definition a MergedPartKey part resolves to:
// the lender's whole definition, marked MergedComponent so its
// replacement effects merge too, with `text` as the line a player
// reads. The triggered rows keep the lender's catalog stamp, so a
// trigger harvested off the mutated permanent is keyed under the
// component's own key and restores with no new effect key. Nil for a
// nil lender.
func MergedPartDef(lender *CardDef, text string) *CardDef {
	if lender == nil {
		return nil
	}
	d := *lender
	d.GrantText = text
	d.MergedComponent = true
	return &d
}

// IsMergedPtr reports whether this object is a merged (mutated)
// permanent: one object represented by more than one card or token
// (CR 730.2).
func IsMergedPtr(c *Card) bool { return len(c.MergedFrom) > 0 }

// IsMerged is IsMergedPtr on a value.
func (c Card) IsMerged() bool { return IsMergedPtr(&c) }

// isMultiCardPtr reports whether this object is one permanent made of
// several cards: melded (ADR 0145) or merged (ADR 0223).
func isMultiCardPtr(c *Card) bool { return IsMeldedPtr(c) || IsMergedPtr(c) }

// componentCount is how many cards or tokens represent a melded or
// merged permanent; 0 for every other object.
func componentCount(c Card) int {
	if IsMergedPtr(&c) {
		return len(c.MergedFrom)
	}
	return len(c.MeldedFrom)
}

// MergedComponentsForEffect returns the cards and tokens that make up
// a mutated permanent, topmost first ("the cards that make up" it),
// or nil for any other object. The entry with the permanent's own ID
// stands for its own card. A copy, never the live slice.
func MergedComponentsForEffect(c Card) []Card {
	if !IsMergedPtr(&c) {
		return nil
	}
	return append([]Card(nil), c.MergedFrom...)
}

// splitComponentsForMove is MoveCard's split of a melded or merged
// object landing in `dst`: the cards it becomes, in the order they are
// pushed (the last ends on top).
func splitComponentsForMove(m Card, dst ZoneKind) []Card {
	if IsMergedPtr(&m) {
		return splitMergedForMove(m, dst)
	}
	return splitMeldedForMove(m, dst)
}

// splitMergedForMove is CR 730.3 for MoveCard: one permanent leaves and
// each component arrives. The permanent's own card (the carrier) is the
// moved object itself, with its own commander status back, so it keeps
// the object's ID and everything MoveCard already did to it; every
// other component keeps its own ID and starts a new object (CR 400.7).
// Pushed bottom first, so the topmost component ends on top.
func splitMergedForMove(m Card, dst ZoneKind) []Card {
	out := make([]Card, 0, len(m.MergedFrom))
	for i := len(m.MergedFrom) - 1; i >= 0; i-- {
		part := m.MergedFrom[i]
		if part.InstanceID == m.InstanceID {
			carrier := m
			carrier.MergedFrom = nil
			carrier.IsCommander = part.IsCommander
			carrier.NeedsEffect = part.NeedsEffect
			carrier.CommanderReturnDue = commanderReturnDueOn(carrier, dst)
			out = append(out, carrier)
			continue
		}
		part.MergedFrom = nil
		part.ObjectEpoch++
		part.Tapped = false
		part.Counters = nil
		part.effective = nil
		part.CommanderReturnDue = commanderReturnDueOn(part, dst)
		out = append(out, part)
	}
	return out
}

// mergedPassengers is meldPassengers for a merged permanent: the IDs of
// every component but the carrier, topmost first.
func mergedPassengers(moved Card) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(moved.MergedFrom))
	for _, part := range moved.MergedFrom {
		if part.InstanceID != moved.InstanceID {
			out = append(out, part.InstanceID)
		}
	}
	return out
}

// mutatingSpell reports whether a stack item is a mutating creature
// spell: cast for its mutate cost and still targeting (CR 702.140a,
// 702.140b).
func mutatingSpell(item *StackItem) bool {
	return item != nil && item.Kind == StackItemSpell && (item.AltCost == MutateKey || item.AltCost == MutateGraveyardKey) && len(item.Targets) > 0
}

// mutateTargetIllegalLocked is CR 702.140b's question as the spell
// begins to resolve: is it a mutating creature spell whose target is
// illegal? The CR 608.2b re-check's own answer.
//
// Caller must hold g.mu.
func (g *Game) mutateTargetIllegalLocked(item *StackItem) bool {
	return mutatingSpell(item) && spellAllTargetsIllegalLocked(g, item)
}

// endMutateOnStackLocked is CR 702.140b: the spell ceases to be a
// mutating creature spell and continues resolving as a creature spell,
// so its target goes (CR 608.3a puts a permanent spell with no target
// onto the battlefield).
//
// Caller must hold g.mu.
func (g *Game) endMutateOnStackLocked(item *StackItem) {
	item.Targets = nil
}

// mutateHostLocked is the creature a mutating creature spell merges
// with: its target, on the battlefield as the object it was targeted
// as. Nil when there is none.
//
// Caller must hold g.mu.
func (g *Game) mutateHostLocked(item *StackItem) *Card {
	if !mutatingSpell(item) {
		return nil
	}
	t := item.Targets[0]
	if t.Kind != TargetCard {
		return nil
	}
	host := findBattlefieldCard(g, t.ID)
	if host == nil || (t.Epoch != 0 && host.ObjectEpoch != t.Epoch) {
		return nil
	}
	return host
}

// mutateOverUnderThen is CR 702.140c's question answered: option 0 puts
// the spell on top, option 1 under. Carry is the spell's stack item ID;
// the answer is written onto the held resolution's item, which the
// merge reads. An on-disk key: never renamed, never reused.
var mutateOverUnderThen = RegisterOptionPickThen("mutate/over-under", func(g *Game, r OptionPicked) error {
	if len(r.Carry) != 1 {
		return nil
	}
	t := g.pausedResolutionTail
	if t == nil || t.item == nil || t.item.ID != r.Carry[0] {
		return nil
	}
	t.item.MutateUnder = r.Index == 1
	return nil
})

// askMutateOverUnderLocked asks a mutating creature spell's controller
// whether it goes over or under (CR 702.140c). The resolution's tail is
// held behind the question (resolution_tail.go), so the merge runs
// after the answer. Nothing is asked for a spell that is not merging.
//
// Caller must hold g.mu, inside the resolution.
func (g *Game) askMutateOverUnderLocked(top Card, item *StackItem) {
	if item == nil || item.IsCopy {
		return
	}
	host := g.mutateHostLocked(item)
	if host == nil {
		return
	}
	hostName := host.Name
	if host.FaceDown || hostName == "" {
		hostName = "the creature"
	}
	g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  item.Controller,
		Source:   top.InstanceID,
		Question: "Mutate: put " + top.Name + " over or under " + hostName + "?",
		Options: []ChoiceOption{
			{Label: "Over: the creature becomes " + top.Name + " plus all abilities from under it"},
			{Label: "Under: the creature stays " + hostName + " and gains " + top.Name + "'s abilities"},
		},
		ThenKey: mutateOverUnderThen,
		Carry:   []uuid.UUID{item.ID},
	})
}

// mergeMutatingSpellLocked is CR 702.140c and 608.3d: a mutating
// creature spell whose target is still legal leaves the stack and
// merges with it, on top or underneath as its controller chose, and
// "whenever this creature mutates" triggers (CR 702.140d). Reports
// whether it merged; false leaves the spell to the ordinary exits (a
// copy, CR 707.10, enters as a token creature: ADR 0223's out of
// scope).
//
// Nothing enters the battlefield (CR 730.2b): no zone move, no ETB.
//
// Caller must hold g.mu, inside a resolution function.
func (g *Game) mergeMutatingSpellLocked(top Card, item *StackItem) bool {
	if item == nil || item.IsCopy || g.Stack == nil {
		return false
	}
	host := g.mutateHostLocked(item)
	if host == nil || !g.Stack.Contains(top.InstanceID) {
		return false
	}
	hostID := host.InstanceID
	// The card leaves the stack as MoveCard would take it out: a new
	// object (CR 400.7), with everything that belonged to the spell
	// stripped. It lands nowhere: it is part of the permanent now.
	scratch := &Zone{Kind: ZoneExile}
	comp, err := MoveCard(g.Stack, scratch, top.InstanceID)
	if err != nil {
		return false
	}
	comp = scratch.Cards[0]
	comp.CommanderReturnDue = false
	comp.Controller = comp.Owner
	comp.ManaAbilities, comp.ActivatedAbilities = nil, nil
	host = findBattlefieldCard(g, hostID)
	if host == nil {
		return false
	}
	under := item.MutateUnder
	g.mergeIntoLocked(host, comp, under)
	ev := Event{
		Kind:   EventMutate,
		CardID: hostID,
		Source: comp.InstanceID,
		Actor:  item.Controller,
		Label:  comp.Name,
	}
	if under {
		ev.Amount = 1
	}
	g.EmitEvent(ev)
	return true
}

// mergeIntoLocked merges `comp` with the permanent `host`, on top of it
// or under it (CR 730.2), and rewrites the permanent's characteristics
// (applyMergedValues). The first merge records the permanent's own card
// as a component: its copiable values as they are now, under its own
// ID.
//
// Caller must hold g.mu.
func (g *Game) mergeIntoLocked(host *Card, comp Card, under bool) {
	if !IsMergedPtr(host) {
		self := Card{
			InstanceID:  host.InstanceID,
			ObjectEpoch: host.ObjectEpoch,
			Owner:       host.Owner,
			Controller:  host.Owner,
			IsCommander: host.IsCommander,
			NeedsEffect: host.NeedsEffect,
		}
		self.setPrintedValues(CopiableValuesOf(*host))
		host.MergedFrom = []Card{self}
	}
	if under {
		host.MergedFrom = append(host.MergedFrom, comp)
	} else {
		host.MergedFrom = append([]Card{comp}, host.MergedFrom...)
	}
	host.applyMergedValues()
	g.layerVersion.Add(1)
}

// applyMergedValues writes a merged permanent's characteristics onto
// its flat printed fields (CR 702.140e, 730.2a): the topmost
// component's copiable values and token status (CR 730.2d), plus every
// other component's abilities as a MergedPartKey part and its printed
// keywords. The permanent's own values are stashed in PrintedSelf the
// way applyCopy stashes them, so the battlefield-leave path puts them
// back (CR 400.7). A commander component makes the permanent a
// commander (CR 903.3c).
func (c *Card) applyMergedValues() {
	if !IsMergedPtr(c) {
		return
	}
	top := c.MergedFrom[0]
	v := CopiableValuesOf(top).withTokenStatus(IsTokenPtr(&top))
	needs := false
	commander := false
	for i := range c.MergedFrom {
		part := &c.MergedFrom[i]
		needs = needs || part.NeedsEffect
		commander = commander || part.IsCommander
		if i == 0 {
			continue
		}
		if p := MergedPartKey(catalogKeyOf(part)); p != "" {
			v.GrantedAbilities = append(v.GrantedAbilities, p)
		}
		// A component's own copy grants (CR 707.9a) are its abilities
		// too.
		_, grants := splitCatalogKeyGrants(underlyingTextKeyOf(part))
		v.GrantedAbilities = append(v.GrantedAbilities, grants...)
		for _, kw := range part.Keywords {
			if !containsFold(v.Keywords, kw) {
				v.Keywords = append(v.Keywords, kw)
			}
		}
	}
	if c.PrintedSelf == nil {
		own := printedValuesOf(*c)
		c.PrintedSelf = &own
	}
	c.setPrintedValues(v)
	c.NeedsEffect = needs
	c.IsCommander = commander
}

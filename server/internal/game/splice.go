package game

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
)

// splice.go — CR 702.47, splice (#1989, ADR 0202).
//
// "Splice onto [quality] [cost]" means "You may reveal this card from
// your hand as you cast a [quality] spell. If you do, that spell gains
// the text of this card's rules text and you pay [cost] as an
// additional cost to cast that spell" (CR 702.47a). The card stays in
// the hand.
//
// # The shape
//
//   - A card with splice declares CardDef.Splice: the quality it is
//     spliced onto and its splice cost. Its rules text is the rest of its
//     catalog row — the target clause and the resolution body the card
//     already has for being cast on its own.
//   - The announcement names the revealed cards in
//     CastSpellParams.Splices, in the order their effects will happen
//     (CR 702.47b). It is checked with the rest of CR 601.2b, after the
//     modes and before the targets.
//   - The targets for a spliced text ride the one flat target list,
//     each ref carrying TargetRef.Splice (CR 702.47d, 601.2c). Each text
//     is validated against its own clause list, judged from the spell
//     being cast (CR 702.47c: the spell keeps the main spell's
//     characteristics, so Glacial Ray spliced onto a blue spell can
//     target a creature with protection from red).
//   - The splice costs' mana joins the total in the one pricer
//     (printedCostLocked) at the additional-cost point of CR 601.2f, so
//     the cast, the preview and the view read one price.
//   - The stack item carries the texts in StackItem.Spliced. A CR 707.10
//     copy carries them; they vanish with the item (CR 702.47e).
//   - At resolution the main spell's text runs first, then each spliced
//     text in order (CR 702.47b), each waiting for a prompt the one
//     before it opened (the bullet walk of mode_walk.go). A spliced text
//     runs against a VIEW of the item whose Targets are its own, so its
//     card file's body reads its targets exactly as it does when the
//     card is cast alone; the view's source is still the spell
//     (CR 702.47c: "Glacial Ray deals 2 damage" is the spell dealing
//     it).

// SpliceOnto is the quality a splice card is spliced onto (CR 702.47a's
// "[quality]").
type SpliceOnto string

const (
	// SpliceOntoArcane is "Splice onto Arcane" (Kamigawa block).
	SpliceOntoArcane SpliceOnto = "arcane"
	// SpliceOntoInstantOrSorcery is "Splice onto instant or sorcery"
	// (Modern Horizons 2, Everdream, Fell Beast's Shriek).
	SpliceOntoInstantOrSorcery SpliceOnto = "instant_or_sorcery"
)

// Accepts reports whether a spell, as it is being cast (the face
// materialised, ADR 0034), has the quality.
func (o SpliceOnto) Accepts(spell Card) bool {
	switch o {
	case SpliceOntoArcane:
		return spell.HasSubtype("Arcane")
	case SpliceOntoInstantOrSorcery:
		return spell.IsInstant() || spell.IsSorcery()
	}
	return false
}

// Words is the quality as the card prints it.
func (o SpliceOnto) Words() string {
	switch o {
	case SpliceOntoArcane:
		return "Arcane"
	case SpliceOntoInstantOrSorcery:
		return "instant or sorcery"
	}
	return string(o)
}

// SpliceAbility is a card's "Splice onto [quality] [cost]".
//
// Only a mana cost is expressed. The five Kamigawa splice cards whose
// cost is not mana (Veil of Secrecy, Horobi's Whisper, Roar of Jukai,
// Hundred-Talon Strike, Torrent of Stone) stay out of the catalog until
// a non-mana splice cost is built (ADR 0202, "Not covered").
type SpliceAbility struct {
	Onto     SpliceOnto
	ManaCost string
}

// Label is the ability as printed: "Splice onto Arcane {1}{R}".
func (s *SpliceAbility) Label() string {
	if s == nil {
		return ""
	}
	return "Splice onto " + s.Onto.Words() + " " + s.ManaCost
}

// SpliceAbilityFor is the splice ability a card's catalog row declares,
// or nil.
func SpliceAbilityFor(key string) *SpliceAbility {
	if d := catalogDef(key); d != nil {
		return d.Splice
	}
	return nil
}

// SplicedText is one card's rules text spliced onto a spell.
type SplicedText struct {
	// CardID is the card revealed from the hand. It stays there
	// (CR 702.47a) and may since have moved; nothing reads it but the
	// view.
	CardID uuid.UUID `json:"cardId"`
	// Key is the catalog key of the text, read as the card was revealed.
	// It is the text's identity on disk: the resolution reads the body
	// and the target clause from the catalog by it, so a card discarded
	// to pay the spell's own discard cost (CR 702.47a's example) still
	// adds its text.
	Key string `json:"key"`
	// Name is the spliced card's name, for the stack and the log.
	Name string `json:"name"`
	// Targets are the targets chosen for this text (CR 702.47d), each
	// with Splice set to this text's 1-based position.
	Targets []TargetRef `json:"targets,omitempty"`
}

// cloneSplicedTexts deep-copies a Spliced list.
func cloneSplicedTexts(in []SplicedText) []SplicedText {
	if len(in) == 0 {
		return nil
	}
	out := make([]SplicedText, len(in))
	for i, sp := range in {
		out[i] = sp
		out[i].Targets = append([]TargetRef(nil), sp.Targets...)
	}
	return out
}

// AllTargets is the spell's whole target list: its own text's targets,
// then each spliced text's (CR 702.47d). The readers that ask about
// "the spell's targets" as an object — what became a target, how many
// targets it has — read this; a resolution body reads its own text's.
func (it *StackItem) AllTargets() []TargetRef {
	if it == nil {
		return nil
	}
	if len(it.Spliced) == 0 {
		return it.Targets
	}
	out := append([]TargetRef(nil), it.Targets...)
	for _, sp := range it.Spliced {
		out = append(out, sp.Targets...)
	}
	return out
}

// SplicedNames is the names of the spliced cards, in the order their
// effects happen.
func (it *StackItem) SplicedNames() []string {
	if it == nil || len(it.Spliced) == 0 {
		return nil
	}
	out := make([]string, len(it.Spliced))
	for i, sp := range it.Spliced {
		out[i] = sp.Name
	}
	return out
}

// spliceRefsOf splits a flat target list into the main text's refs and
// each spliced text's, keeping their order. ok is false when a ref names
// a spliced text that was not announced.
func spliceRefsOf(targets []TargetRef, n int) (main []TargetRef, spliced [][]TargetRef, ok bool) {
	spliced = make([][]TargetRef, n)
	for _, t := range targets {
		switch {
		case t.Splice == 0:
			main = append(main, t)
		case t.Splice < 0 || t.Splice > n:
			return nil, nil, false
		default:
			spliced[t.Splice-1] = append(spliced[t.Splice-1], t)
		}
	}
	return main, spliced, true
}

// spliceTextSteps is one spliced text's target steps: its card's own
// clause list.
func spliceTextSteps(sp SplicedText) []AnnouncedClause {
	return AnnouncedClauses(TargetSpecFor(sp.Key), nil, nil)
}

// announcedSplicesLocked checks the cards a cast reveals to splice
// (CR 601.2b, 702.47a-b) and returns them as texts with no targets yet.
//
// Each must be a card in the caster's hand other than the spell itself,
// named once (CR 702.47b: "You can't splice any one card onto the same
// spell more than once"), with a splice ability whose quality the spell
// being cast has. `spell` is the face-materialised announce copy, so a
// face-down cast (no subtypes, a creature) is spliced onto by nothing.
//
// Caller must hold g.mu.
func (g *Game) announcedSplicesLocked(caster uuid.UUID, spell Card, spellID uuid.UUID, ids []uuid.UUID) ([]SplicedText, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	p := g.playerByIDLocked(caster)
	if p == nil || p.Hand == nil {
		return nil, ErrInvalidParam
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	out := make([]SplicedText, 0, len(ids))
	for _, id := range ids {
		if id == spellID || seen[id] {
			return nil, fmt.Errorf("%w: a card is spliced onto a spell at most once, and never onto itself", ErrInvalidParam)
		}
		seen[id] = true
		c, ok := cardInHand(p, id)
		if !ok {
			return nil, fmt.Errorf("%w: a spliced card must be in your hand", ErrInvalidParam)
		}
		key := CatalogKey(c)
		sa := SpliceAbilityFor(key)
		if sa == nil {
			return nil, fmt.Errorf("%w: %s has no splice ability", ErrInvalidParam, c.Name)
		}
		if !sa.Onto.Accepts(spell) {
			return nil, fmt.Errorf("%w: %s splices onto %s spells only", ErrInvalidParam, c.Name, sa.Onto.Words())
		}
		out = append(out, SplicedText{CardID: id, Key: key, Name: c.Name})
	}
	return out, nil
}

// validateSplicedTargetsLocked checks each spliced text's refs against
// its own clause list (CR 702.47d, 601.2c) and returns them stamped with
// their clause coordinates. A text that has a required target it cannot
// be given cannot be spliced (CR 702.47b), which is the count check of
// the same gate.
//
// Caller must hold g.mu.
func (g *Game) validateSplicedTargetsLocked(src TargetSource, spliced []SplicedText, refs [][]TargetRef) error {
	for i := range spliced {
		steps := spliceTextSteps(spliced[i])
		mine := assignAnnouncedSlots(steps, refs[i])
		if len(steps) == 0 && len(mine) > 0 {
			return fmt.Errorf("%w: %s's text takes no target", ErrInvalidParam, spliced[i].Name)
		}
		if err := g.validateAnnouncedTargetsLocked(src, steps, mine); err != nil {
			return err
		}
		for j := range mine {
			mine[j].Splice = i + 1
		}
		spliced[i].Targets = mine
	}
	return nil
}

// spliceManaLocked adds the splice costs of the announced cards to a
// cast's cost, at the additional-cost point of CR 601.2f (ADR 0073 §3):
// after the alternative cost and the commander tax, before the cost
// modifiers. A card the announcement names that is not a splice card in
// the caster's hand adds nothing — CastSpell refuses that announcement
// before anything is priced, so this is only ever a quote.
//
// Caller must hold g.mu.
func (g *Game) spliceManaLocked(cost ParsedCost, caster uuid.UUID, ids []uuid.UUID) (ParsedCost, error) {
	if len(ids) == 0 {
		return cost, nil
	}
	p := g.playerByIDLocked(caster)
	if p == nil || p.Hand == nil {
		return cost, nil
	}
	for _, id := range ids {
		c, ok := cardInHand(p, id)
		if !ok {
			continue
		}
		sa := SpliceAbilityFor(CatalogKey(c))
		if sa == nil || sa.ManaCost == "" {
			continue
		}
		add, err := ParseCost(sa.ManaCost)
		if err != nil {
			return cost, fmt.Errorf("%w for %s: %w", ErrUnparseableCost, sa.Label(), err)
		}
		cost.Generic += add.Generic
		cost.Required = append(cost.Required, add.Required...)
		cost.HasPhyrexian = cost.HasPhyrexian || add.HasPhyrexian
		cost.HasSnow = cost.HasSnow || add.HasSnow
	}
	return cost, nil
}

// SpliceOffer is one card in a caster's hand that may be spliced onto a
// spell they are casting: the view's offer and its target clause.
type SpliceOffer struct {
	CardID  uuid.UUID
	Name    string
	Ability SpliceAbility
	// Targets is the spliced text's clause list, nil when it has none.
	Targets *TargetSpec
}

// SpliceOffersLocked lists the cards in the caster's hand that could be
// spliced onto `spell` as it is cast: a splice card other than the spell
// whose quality the spell has, and whose required targets can all be
// chosen from the spell (CR 702.47b, judged under CR 702.47c from `src`,
// the spell). The order is the hand's. Mana is not asked: CR 601.2g lets
// the caster make it after the costs are settled.
//
// Caller must hold g.mu.
func (g *Game) SpliceOffersLocked(caster uuid.UUID, spell Card, src TargetSource) []SpliceOffer {
	p := g.playerByIDLocked(caster)
	if p == nil || p.Hand == nil {
		return nil
	}
	var out []SpliceOffer
	for _, c := range p.Hand.Cards {
		if c.InstanceID == spell.InstanceID {
			continue
		}
		key := CatalogKey(c)
		sa := SpliceAbilityFor(key)
		if sa == nil || !sa.Onto.Accepts(spell) {
			continue
		}
		spec := TargetSpecFor(key)
		if spec != nil && g.anyClauseUnfillableLocked(src, AnnouncedClauses(spec, nil, nil)) {
			continue
		}
		out = append(out, SpliceOffer{CardID: c.InstanceID, Name: c.Name, Ability: *sa, Targets: spec})
	}
	return out
}

// splicedView is the item a spliced text resolves against: the
// spell, with that text's targets and clause list in place of its own.
// A shallow copy — the body reads the item, and nothing a spell body
// writes onto its item is read after the spell's own text has run.
func splicedView(item *StackItem, k int) *StackItem {
	sp := item.Spliced[k]
	v := *item
	v.Targets = sp.Targets
	v.targetSpec = TargetSpecFor(sp.Key)
	v.modeSpec = nil
	v.Modes = nil
	v.Distribution = nil
	v.Spliced = nil
	return &v
}

// spellTargetStatusLocked is the CR 608.2b reading of one text's targets:
// whether it had any, whether any is still legal, and whether its clause
// says it resolves anyway (#2182).
func spellTargetStatusLocked(g *Game, item *StackItem) (had, anyLegal, resolvesAnyway bool) {
	if item == nil || len(item.Targets) == 0 {
		return false, false, false
	}
	for _, step := range g.itemAnnouncedClauses(item) {
		if step.Clause.ResolvesIfIllegal {
			return true, true, true
		}
	}
	for _, t := range item.Targets {
		switch t.Kind {
		case TargetPlayer, TargetCard:
			had = true
			if g.TargetStillLegalForEffect(item, t) {
				anyLegal = true
			}
		}
	}
	return had, anyLegal, false
}

// textTargetsAllIllegalLocked reports whether one text of a spell had
// targets and every one of them is now illegal — the CR 608.2b test the
// whole spell has always been put to, put to one text. Such a text does
// nothing as the spell resolves: no part of it may act on an illegal
// target, and every instruction of every printed splice card is about
// its target.
func textTargetsAllIllegalLocked(g *Game, item *StackItem) bool {
	return spellAllTargetsIllegalLocked(g, item)
}

// spellWithSplicedTextAllTargetsIllegalLocked is CR 608.2b for a spell
// that may carry spliced text: it is countered by the rules when it had
// targets and every one of them, in its own text and in every spliced
// text, is illegal (CR 702.47d: "a spell with one or more targets won't
// resolve if all of its targets are illegal on resolution"). With
// nothing spliced it is spellAllTargetsIllegalLocked.
//
// Caller must hold g.mu.
func spellWithSplicedTextAllTargetsIllegalLocked(g *Game, item *StackItem) bool {
	if item == nil || len(item.Spliced) == 0 {
		return spellAllTargetsIllegalLocked(g, item)
	}
	had, legal := false, false
	texts := []*StackItem{item}
	for k := range item.Spliced {
		texts = append(texts, splicedView(item, k))
	}
	for _, t := range texts {
		h, l, anyway := spellTargetStatusLocked(g, t)
		if anyway {
			return false
		}
		had, legal = had || h, legal || l
	}
	return had && !legal
}

// runSplicedTextLocked carries out spliced text k of a resolving spell
// (CR 702.47b): the spliced card's body, run against the view of the
// item that holds its own targets.
//
// Caller must hold g.mu in write mode, inside the resolution function.
func (g *Game) runSplicedTextLocked(item *StackItem, k int) {
	if k < 0 || k >= len(item.Spliced) {
		return
	}
	v := splicedView(item, k)
	if textTargetsAllIllegalLocked(g, v) {
		return
	}
	if EffectResolver == nil {
		return
	}
	if err := EffectResolver(g, v, item.Spliced[k].Key); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    item.Controller,
			Source:   item.SourceCardID,
			ErrorMsg: err.Error(),
		})
	}
}

// revealSplicesLocked shows the spliced cards to the table (CR 702.47a,
// "reveal them all at once", 702.47b), with the spell on the stack.
//
// Caller must hold g.mu in write mode.
func (g *Game) revealSplicesLocked(caster, spellID uuid.UUID, spellName string, spliced []SplicedText) {
	if len(spliced) == 0 {
		return
	}
	ids := make([]uuid.UUID, len(spliced))
	names := make([]string, len(spliced))
	for i, sp := range spliced {
		ids[i] = sp.CardID
		names[i] = sp.Name
	}
	g.RevealForEffect(RevealSpec{
		Player: caster,
		Source: spellID,
		Reason: spellName + " — splice " + strings.Join(names, ", then ") + " onto it",
		Cards:  ids,
	})
	slog.Debug("cast_spell: spliced", "spell", spellName, "spliced", names)
}

// cardInHand is the card with this ID in the player's hand.
func cardInHand(p *Player, id uuid.UUID) (Card, bool) {
	if p == nil || p.Hand == nil {
		return Card{}, false
	}
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return Card{}, false
}

// splitSplicedRefs separates a flat target list into the refs of the
// spell's own text and those of its spliced texts, each in order.
func splitSplicedRefs(targets []TargetRef) (main, spliced []TargetRef) {
	for _, t := range targets {
		if t.Splice != 0 {
			spliced = append(spliced, t)
			continue
		}
		main = append(main, t)
	}
	if len(spliced) == 0 {
		return targets, nil
	}
	return main, spliced
}

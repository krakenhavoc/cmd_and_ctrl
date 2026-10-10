package game

import (
	"strings"

	"github.com/google/uuid"
)

// transform.go — CR 701.27, the keyword action that turns a permanent
// over to its other face. ADR 0079.
//
// The representation was already here: ADR 0034 gave Card its printed
// Faces, an ActiveFace index, and SetFace as the one writer that
// re-materialises the flat printed fields from the face that is up.
// CatalogKey already appends "#<face>" for a non-zero face, so a back
// face registered as "<oracle_id>#1" brings its own triggers, statics,
// replacements, mana abilities and activated abilities with it the
// instant the index moves. MoveCard already puts a card back on its
// front face when it lands anywhere but the battlefield or the stack
// (CR 712.8a). The wire, the snapshot and cloneCard already carry
// Layout, Faces and ActiveFace.
//
// What was missing was the verb, and CR 712.18 is the whole of its
// contract:
//
//	When a double-faced permanent transforms or converts, it doesn't
//	become a new object. Any effects that applied to that permanent
//	will continue to apply to it.
//
// So this file does NOT go anywhere near MoveCard. Nothing here mints
// an InstanceID, bumps ObjectEpoch, re-stamps EnteredBattlefieldAt or
// emits a zone move: counters, marked damage, attachments, tap state,
// the CR 613.7 timestamp, summoning sickness, the controller and every
// Game-side per-object registry keyed by instance ID all survive
// untouched, because the object never ended.
//
// The OTHER shape — "exile it, then return it to the battlefield
// transformed" (Fable of the Mirror-Breaker's chapter III, CR 712.14a)
// — is two zone changes and a genuinely new object, and it lives at
// the bottom of this file under a name that cannot be confused with
// the first one. Reaching for the wrong half is the mistake ADR 0079
// exists to prevent: it gives Fable a Saga that kept its lore counters,
// or gives an in-place transform a permanent that re-triggers its own
// ETBs.

// CanTransform reports whether this permanent is one an instruction to
// transform can do anything to (CR 712.9).
//
// Five refusals, each with a rule behind it:
//
//   - A face-down permanent can't transform (CR 712.15a). It also has
//     no text at all (CR 708.2a), so nothing it prints could have
//     asked.
//   - Only a permanent represented by a double-faced token or a
//     double-faced card can (CR 712.9), and this engine has no
//     double-faced tokens (ADR 0079, out of scope 1). Two printed
//     faces is the shape.
//   - The LAYOUT allowlist is the load-bearing line. An `adventure`
//     card has two Faces and a `split` card has two Faces, and neither
//     is a double-faced card — turning Foulmire Knight into Profane
//     Insight on the battlefield is not a thing that can happen.
//     Keying on len(Faces) == 2 alone would have made it one.
//   - A meld card can't transform (CR 712.4c), and neither can a
//     melded permanent (ADR 0145); the layout check refuses both, since
//     neither has a `transform` or `modal_dfc` layout.
//   - An instruction whose target face is an instant or sorcery face
//     does nothing (CR 701.27d, CR 712.10).
func CanTransform(c Card) bool {
	if c.FaceDownIsPermanent() {
		return false
	}
	if !isDoubleFacedPermanent(c) {
		return false
	}
	return !faceIsInstantOrSorcery(c.Faces[1-c.ActiveFace])
}

// isDoubleFacedPermanent is "is this permanent represented by a
// double-faced card" (CR 712.2), and it is the layout allowlist the
// comment above calls the load-bearing line, pulled out so there is
// ONE definition of it.
//
// Two rules ask, from opposite ends of the tree: CR 712.9 (only a
// double-faced permanent can transform) and CR 712.16 ("Melded
// permanents and other double-faced permanents can't be turned face
// down", #1209). Two copies of an allowlist that an `adventure` and a
// `split` card both fall foul of is two places to forget one.
func isDoubleFacedPermanent(c Card) bool {
	if len(c.Faces) != 2 {
		return false
	}
	switch c.Layout {
	case LayoutTransform, LayoutModalDFC:
		return true
	}
	return false
}

// IsDoubleFacedForEffect reports whether `c` is a double-faced card or
// permanent (CR 712.1): a transforming or modal double-faced card, a
// meld card, or a melded permanent — Tetzin, Gnome Champion's "another
// double-faced artifact" and The Golden-Gear Colossus's target (#2709).
// Whether it can then TRANSFORM is CanTransform's question; the
// Colossus may target a modal double-faced artifact that will not.
func IsDoubleFacedForEffect(c Card) bool {
	return isDoubleFacedPermanent(c) || c.IsMeldCard() || c.IsMelded()
}

// faceIsInstantOrSorcery is CR 712.10's guard, asked of a face rather
// than of a card because the card is still showing the other one.
//
// A substring scan over the face's own type line, which is safe here
// in a way the whole-card scan ADR 0034 was written about is not: a
// Face.TypeLine is one face's line ("Sorcery"), never the joined
// "Sorcery // Land" that made IsLand match a modal DFC's front.
func faceIsInstantOrSorcery(f Face) bool {
	line := strings.ToLower(f.TypeLine)
	return strings.Contains(line, "instant") || strings.Contains(line, "sorcery")
}

// TransformPermanentForEffect turns the named battlefield permanent
// over so that its other face is up (CR 701.27a).
//
// Returns nil and emits nothing for a permanent CanTransform refuses.
// That is CR 701.27c and CR 701.27d in one line — "nothing happens" —
// and it is deliberately not an error and deliberately not an
// EventEffectError: the catalog soak fails a game on any effect error
// (AISEAT_CATALOG_GAMES), and an instruction that legally does nothing
// is a resolution, not a card that threw.
//
// ErrCardNotFound is still returned for an ID that names no
// battlefield permanent, because that is a caller holding a stale ID
// rather than a rules outcome.
//
// Two writes beyond the face itself, and neither is optional:
//
//   - Card.effective is nilled. A face change is a PRINTED-value
//     change and Effective() returns the cached Characteristic
//     verbatim when it is warm, so a reader between here and the next
//     layer pass would otherwise get the old face's type line, P/T and
//     ability list. stampBattlefieldEntryLocked nils it for the same
//     reason on entry.
//   - EventTransform is emitted, which is what bumps g.layerVersion
//     (see layerVersionBump.OnEvent) and what makes "whenever this
//     transforms" writable. The bump lives in the listener rather than
//     inline because the listener is the one place that lists what
//     invalidates the layer engine.
//
// Caller must hold g.mu. Added in S46 (ADR 0079, #343).
func (g *Game) TransformPermanentForEffect(cardID uuid.UUID) error {
	card := findBattlefieldCard(g, cardID)
	if card == nil {
		return ErrCardNotFound
	}
	if !CanTransform(*card) {
		return nil
	}
	// CR 702.145b / 702.145e (ADR 0132): a permanent with daybound or
	// nightbound "can't transform except due to" that ability, so an
	// instruction to transform it does nothing, exactly as CR 701.27c
	// treats an object that can't. The day/night rules themselves turn
	// it over through transformInPlaceLocked, which has no such guard.
	if hasDayNightKeyword(card) {
		return nil
	}
	g.transformInPlaceLocked(card)
	return nil
}

// transformInPlaceLocked is the body of the verb with every guard
// already answered: the face goes over, the printed-value cache is
// dropped, EventTransform announces it and the new face's "as this
// transforms into" clause runs. TransformPermanentForEffect is the
// guarded entry for an instruction; the day/night rules (CR 702.145c,
// 702.145f) call this directly, because they are the one thing a
// daybound or nightbound permanent can transform for.
//
// Caller must hold g.mu, and must have checked CanTransform.
func (g *Game) transformInPlaceLocked(card *Card) {
	cardID := card.InstanceID
	from := card.Name
	to := 1 - card.ActiveFace
	controller := card.Controller
	card.SetFace(to)
	card.effective = nil
	g.EmitEvent(Event{
		Kind:   EventTransform,
		CardID: cardID,
		Actor:  controller,
		Amount: to,
		Label:  from,
	})
	g.runAsTransformsIntoLocked(cardID)
}

// runAsTransformsIntoLocked runs the "As this permanent transforms into
// <face>, …" clause of the face that is now up (#1574, ADR 0079
// amendment 2026-09-24) — Sephiroth, One-Winged Angel's Super Nova.
//
// Here, in the verb, and not in the card that asked for the transform,
// because the clause belongs to the FACE: Sephiroth's own fourth drain,
// a Moonmist, or anything else that turns it over must all make the
// emblem, and a clause the front face's text ran would make it only for
// the first of those. It is a static ability that applies as the
// permanent turns over (not a trigger — nothing goes on the stack and
// nobody gets a window), so it runs synchronously, the way the AsEnters
// hook runs for an entry, and after EventTransform for the same reason
// that hook runs after EventETB: the event is the moment, and the
// layer version it bumps is what lets the lookup below see the new face.
//
// Only the in-place verb calls it. "Exile it, then return it
// transformed" (ExileAndReturnTransformedForEffect) does not: that
// permanent ENTERS on its back face, and a permanent that enters
// transformed never transformed (ADR 0079 decision 5), so its "as this
// transforms" clause does nothing.
//
// The key is CatalogAbilityKey, read after a layer refresh: CR 712.18
// keeps every effect that applied to the permanent applying after it
// turns over, so one that removed all its abilities has removed the
// new face's clause too. An error is published as EventEffectError,
// exactly as fireETBHookLocked publishes an AsEnters one — the
// transform has already happened and cannot be unwound.
//
// Caller must hold g.mu.
func (g *Game) runAsTransformsIntoLocked(cardID uuid.UUID) {
	g.RecomputeLayersIfStaleLocked()
	card := findBattlefieldCard(g, cardID)
	if card == nil {
		return
	}
	d := catalogDef(catalogAbilityKeyOf(card))
	if d == nil || d.AsTransformsInto == nil {
		return
	}
	if err := d.AsTransformsInto(g, cardID); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   cardID,
			ErrorMsg: err.Error(),
		})
	}
}

// ExileAndReturnTransformedForEffect is the OTHER verb: "exile this
// permanent, then return it to the battlefield transformed under your
// control" (Fable of the Mirror-Breaker III, CR 712.14a).
//
// Two zone changes, so CR 400.7 applies in full and what comes back is
// a NEW OBJECT — new InstanceID, no counters, no marked damage, no
// attachments, a fresh CR 613.7 timestamp, summoning sick again, and
// every "enters the battlefield" ability on its new face fires. That
// is the whole difference from TransformPermanentForEffect above, and
// it is why the two are separate functions rather than a flag.
//
// Built out of the two moves that already exist:
//
//  1. ExileCardThenForEffect, the shared exit primitive, so a
//     commander that would be exiled is offered the command zone
//     (CR 903.9) and the second half waits for the answer instead of
//     firing into a paused move. When the exile is replaced away
//     — the commander took the command zone, a replacement redirected
//     it — `exiled` is false, nothing returns, and the permanent stays
//     where the replacement put it. That is right: the sentence's
//     second half is conditional on its first.
//  2. The face is set on the card WHILE IT IS STILL IN EXILE, then
//     ReturnFromExileToBattlefieldForEffect mints the new object and
//     runs the CR 614 entry pipeline.
//
// Step 2's ordering is the same move CastSpell makes through
// setFaceInZoneLocked, and for the same reason: the replacement
// pipeline resolves the entering card by ID out of its SOURCE zone, so
// a self-replacement on the back face (an "enters tapped" clause) has
// to be findable under the back face's catalog key before the pipeline
// runs. It leaves a card in exile on its back face for the width of
// one locked mutation, which CR 712.8a says should not happen; nothing
// observes it, and the alternative — threading a face through
// ReplacementEvent and the resumable entry tail — is a much larger
// change for a state no viewer can see. Recorded in ADR 0079 decision
// 5 so the next reader knows it was a choice.
//
// A permanent whose other face is not a permanent face is refused up
// front (CR 712.14b's shape): it would be exiled and then stranded,
// which is strictly worse than not exiling it. CanTransform already
// asks exactly that question.
//
// A permanent that is no longer on the battlefield is not an error
// either: the ability does nothing (#2497). "This Saga" names one
// object, and once it has left the battlefield it is gone (CR 400.7)
// — a copy of chapter III resolving after the original already
// exiled and returned it (Gogo, Master of Mimicry: a copy refers to
// the same object as the original, CR 707.10b), or a Saga bounced or
// destroyed in response. Nothing is exiled, so nothing returns, and
// the new object the first resolution made is left alone. Before
// #2497 this returned ErrCardNotFound and every such copy logged an
// effect error.
//
// controller is who it returns under; uuid.Nil means its owner.
//
// Caller must hold g.mu. Added in S46 (ADR 0079, #343).
func (g *Game) ExileAndReturnTransformedForEffect(cardID, controller uuid.UUID) error {
	card := findBattlefieldCard(g, cardID)
	if card == nil {
		return nil
	}
	if !CanTransform(*card) {
		return nil
	}
	face := 1 - card.ActiveFace
	return g.ExileCardThenForEffect(cardID, func(g *Game, exiled bool) error {
		if !exiled {
			return nil
		}
		setFaceInZoneLocked(g.Exile, cardID, face)
		_, err := g.ReturnFromExileToBattlefieldForEffect(cardID, controller, false)
		return err
	})
}

// ReturnCraftedFromExileForEffect is craft's resolution (CR 702.167a,
// ADR 0137): "Return this card to the battlefield transformed under its
// owner's control", said of the card the ability's own cost has just
// exiled, plus CR 702.167c's link to the materials that cost exiled.
//
// It is the exile form of ReturnFromGraveyardTransformedForEffect, and
// the second half of ExileAndReturnTransformedForEffect with the first
// half paid as a cost: the back face is set on the card WHILE IT IS IN
// EXILE (CR 712.8a puts it there front face up), for the reason those
// two give, and the ordinary exile return mints the new object (CR
// 400.7) through the CR 614 entry pipeline. What enters is a new
// permanent: summoning sick, no counters, a fresh timestamp, and every
// enters ability on its back face fires.
//
// `materials` are the instance IDs the cost exiled (PaidCost.Exiled).
// Each is linked as the object it is in exile NOW, and one that is not
// in exile any more — a commander its owner moved to the command zone
// (CR 903.9a), a token that ceased to exist there (CR 111.8) — is not
// linked at all: CR 702.167c names "cards in exile that were exiled to
// pay the cost". The link rides the entry (entryTail.craftedWith) and
// is stamped before EventETB, so it is there for the new permanent's
// own enters trigger on the inline path and the resumed one alike.
//
// Nothing happens, and that is not an error, when the card is no longer
// in exile (it is a new object the ability does not know, CR 400.7),
// when it is a token (CR 111.8), or when it cannot transform — a copy
// of a craft card that is not itself double-faced stays in exile, as
// ExileAndReturnTransformedForEffect leaves a permanent it cannot
// transform (ADR 0079 decision 5). If the
// entry is cancelled or redirected the front face is put back, so no
// card sits in exile showing its back; a PAUSED entry keeps the back
// face, because the resume needs it.
//
// Returns the new permanent's ID, or uuid.Nil when nothing entered (or
// the entry paused). Under its owner's control (CR 702.167a), whoever
// activated the ability.
//
// Caller must hold g.mu.
func (g *Game) ReturnCraftedFromExileForEffect(cardID uuid.UUID, materials []uuid.UUID) (uuid.UUID, error) {
	if g.Exile == nil {
		return uuid.Nil, nil
	}
	var card *Card
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == cardID {
			card = &g.Exile.Cards[i]
			break
		}
	}
	if card == nil || card.IsToken() {
		return uuid.Nil, nil
	}
	front := *card
	front.ActiveFace = 0
	if !CanTransform(front) {
		return uuid.Nil, nil
	}
	var link []ObjectRef
	for _, id := range materials {
		if id == cardID {
			continue
		}
		for i := range g.Exile.Cards {
			if c := &g.Exile.Cards[i]; c.InstanceID == id && !c.IsToken() {
				link = append(link, ObjectRef{ID: id, Epoch: c.ObjectEpoch})
				break
			}
		}
	}
	setFaceInZoneLocked(g.Exile, cardID, 1)
	newController := card.Owner
	card.Controller = newController
	entered, err := g.enterBattlefieldThroughPipelineLocked(&ReplacementEvent{
		Kind:           RepEventMove,
		Actor:          newController,
		CardID:         cardID,
		OldZone:        ZoneExile,
		NewZone:        ZoneBattlefield,
		entryResumable: true,
		entryTail:      &entryTail{newObject: true, craftedWith: link},
	})
	if entered == uuid.Nil && len(g.PendingChoices) == 0 && g.Exile.Contains(cardID) {
		setFaceInZoneLocked(g.Exile, cardID, 0)
	}
	return entered, err
}

// CraftMaterialsForEffect resolves CR 702.167c's "the exiled cards used
// to craft it" (Card.CraftedWith, PermanentInfo.CraftedWith, ADR 0137):
// the cards still in exile as the objects the craft cost put there, in
// the order named, as value copies. The delve link's reader under the
// craft link's name — one rule (CR 400.7: a card that left exile and
// came back is a new object nothing links to), applied in one place.
//
// Caller must hold g.mu.
func (g *Game) CraftMaterialsForEffect(refs []ObjectRef) []Card {
	return g.DelvedCardsForEffect(refs)
}

// ReturnFromGraveyardTransformedForEffect is the graveyard form of the
// other verb: "return it to the battlefield [tapped] transformed under
// its owner's control" (the Ojer gods, CR 712.14a), said of a double-faced
// card that is already in a graveyard (#1900, ADR 0079 amendment).
//
// CR 712.8a puts a card outside the battlefield and the stack on its
// FRONT face, so the card sits in the graveyard showing its front and
// this sets the back face on it there, for the width of one locked
// mutation, and then runs the ordinary graveyard entry. Setting it
// first is the same ordering ExileAndReturnTransformedForEffect and
// CastSpell use, for the same reason: the CR 614 pipeline resolves the
// entering card by ID out of its SOURCE zone, so a self-replacement on
// the back face has to be findable under the back face's catalog key
// before the pipeline runs. Unlike the exile return, the graveyard
// return keeps the card's InstanceID (it is the same zone-change path
// every reanimation takes); the battlefield entry stamps the fresh
// CR 613.7 timestamp.
//
// A card whose other face is not a permanent face (or that is not
// double-faced at all) is returned as itself would NOT be right, so it
// is left in the graveyard: the clause "transformed" has nothing to turn
// it into, and a front-face return would be the stronger-than-printed
// outcome (#259). It does nothing and returns uuid.Nil with no error.
//
// If the entry is cancelled or redirected the face is put back, so no
// card is left in a graveyard showing its back (CR 712.8a). The one
// case left alone is a paused entry (a CR 616 ordering prompt or an
// entry choice is open): the resume needs the back face, and a card
// that is still in a graveyard with a prompt open is mid-move.
//
// `counters` ride the entry event as "enters with" counters
// (Ojer Pakpatiq's three time counters, CR 614.1c); nil names none.
// controller uuid.Nil means "under its owner's control". Returns
// ErrCardNotFound when the card is not in a graveyard any more, the
// CR 400.7 answer a "return it" trigger swallows.
//
// Caller must hold g.mu.
func (g *Game) ReturnFromGraveyardTransformedForEffect(cardID, controller uuid.UUID, tapped bool, counters map[string]int) (uuid.UUID, error) {
	src := g.findCardZoneLocked(cardID)
	if src == nil || src.Kind != ZoneGraveyard {
		return uuid.Nil, ErrCardNotFound
	}
	var card *Card
	for i := range src.Cards {
		if src.Cards[i].InstanceID == cardID {
			card = &src.Cards[i]
			break
		}
	}
	if card == nil {
		return uuid.Nil, ErrCardNotFound
	}
	// The card is on its front face (CR 712.8a); ask the transform
	// question of that shape, whatever a stray ActiveFace says.
	front := *card
	front.ActiveFace = 0
	if !CanTransform(front) {
		return uuid.Nil, nil
	}
	setFaceInZoneLocked(src, cardID, 1)
	id, err := g.returnFromGraveyardFaceLocked(cardID, ZoneBattlefield, controller, tapped, FaceDownNone, nil, counters)
	if id == uuid.Nil && len(g.PendingChoices) == 0 {
		if z := g.findCardZoneLocked(cardID); z != nil && z.Kind == ZoneGraveyard {
			setFaceInZoneLocked(z, cardID, 0)
		}
	}
	return id, err
}

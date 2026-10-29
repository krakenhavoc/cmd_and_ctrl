package game

import (
	"strings"

	"github.com/google/uuid"
)

// manifest_dread_keyed.go — #4253 slice 9: manifest dread (CR 701.62a)
// with the rest of the action as data.
//
// The closure form (ManifestDreadThenForEffect, deleted in #4253's last
// part) composed a choose_cards pick, a face-down entry and a graveyard
// route batch, each handing the next a closure, and ended in the card's own
// closure. The keyed form is the same three steps
// over registered continuations: the pick is a keyed choose_cards prompt
// (manifest-dread/pick), the entry's tail routes the other card to the
// graveyard through a keyed route batch (engine/manifest-dread-entered),
// and that batch's tail emits EventManifestDread and runs the card's tail
// (engine/manifest-dread-finished). A table waiting on the pick, on the
// face-down entry's prompt or on the graveyard card's prompt is then a
// restore point.

// ManifestDreadKeyedForEffect carries out CR 701.62a for `playerID` and
// tells `then`, a registered tail, what happened: TailOutcome.IDs =
// [manifested, graveyarded], either uuid.Nil when it did not happen
// (ManifestDreadOutcome reads them back). A zero tail is "no
// continuation". `source` is the permanent or spell asking and names the
// prompt's source; it may be uuid.Nil.
//
// The prompt is skipped when the library shows one card (nothing to
// choose) and the whole action is skipped, with `then` still told an empty
// result, when it shows none: a "then put counters on that creature"
// continuation must run so it can find it has no creature. (The closure
// form, ManifestDreadThenForEffect, was deleted in #4253's last part once
// only tests called it.)
//
// The tail rides the pick prompt packed into its carried data, so it may
// not set Next (ErrInvalidParam).
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ManifestDreadKeyedForEffect(playerID, source uuid.UUID, then KeyedTail) error {
	if g.playerByIDLocked(playerID) == nil {
		return ErrPlayerNotFound
	}
	if then.Next != nil || strings.Contains(then.Ref.key, manifestDreadWordSep) {
		return ErrInvalidParam
	}
	looked := g.LookAtTopOfLibraryForEffect(playerID, 2)
	if len(looked) == 0 {
		return then.Run(g, []uuid.UUID{uuid.Nil, uuid.Nil})
	}
	if len(looked) == 1 {
		return g.finishManifestDreadKeyedLocked(playerID, source, looked, looked, then)
	}
	g.QueueChooseCardsForEffect(ChooseCardsPrompt{
		Chooser:  playerID,
		Source:   source,
		Question: "Manifest dread — put one onto the battlefield face down; the other goes to your graveyard",
		Cards:    looked,
		Min:      1,
		Max:      1,
		Zone:     ZoneLibrary,
		ThenKey:  manifestDreadPick,
		// The looked-at pair, then the card's tail: its player, its
		// source and its own carry. Its key (and its word) is the word
		// and its N the second number.
		Carry:     append(copyUUIDs(looked), append([]uuid.UUID{then.Player, then.Source}, then.Carry...)...),
		CarryWord: packManifestDreadWord(then),
		CarryNums: []int{len(looked), then.N},
		Asked:     copyAskedItem(then.Asked),
	})
	return nil
}

// manifestDreadPick is the keyed pick's answer: the two moves, with the
// card's tail unpacked from the prompt's carried data.
var manifestDreadPick = RegisterChooseCardsThen("manifest-dread/pick", func(g *Game, r CardsChosen) error {
	n, tailN := 0, 0
	if len(r.Nums) >= 2 {
		n, tailN = r.Nums[0], r.Nums[1]
	}
	if n > len(r.Carry) {
		n = len(r.Carry)
	}
	if n == 0 {
		return nil
	}
	looked := copyUUIDs(r.Carry[:n])
	rest := r.Carry[n:]
	var then KeyedTail
	if r.Word != "" && len(rest) >= 2 {
		key, word, _ := strings.Cut(r.Word, manifestDreadWordSep)
		then = KeyedTail{Ref: TailRef{key: key}, Player: rest[0], Source: rest[1], Carry: copyUUIDs(rest[2:]), N: tailN, Word: word}
		if r.Asked != nil {
			then.Asked = NewAskedItem(r.Asked)
		}
	}
	picked := r.Picked
	if !r.Answered {
		picked = nil
	}
	return g.finishManifestDreadKeyedLocked(r.Chooser, r.Source, looked, picked, then)
})

// finishManifestDreadKeyedLocked is finishManifestDreadLocked over keyed
// tails: the entry first, its tail routing the other card to the
// graveyard, whose tail emits the event and runs the card's.
//
// Caller must hold g.mu in write mode.
func (g *Game) finishManifestDreadKeyedLocked(playerID, source uuid.UUID, looked, picked []uuid.UUID, then KeyedTail) error {
	chosen := looked[0]
	if len(picked) > 0 {
		chosen = picked[0]
	}
	var rest []uuid.UUID
	for _, id := range g.cardsInALibraryLocked(looked) {
		if id != chosen {
			rest = append(rest, id)
		}
	}
	entered := KeyedTail{Ref: manifestDreadEntered, Player: playerID, Source: source, Carry: rest}
	if !then.IsZero() {
		t := copyKeyedTail(then)
		entered.Next = &t
	}
	if len(g.cardsInALibraryLocked([]uuid.UUID{chosen})) == 0 {
		// The chosen card left the library between the look and the
		// answer; nothing enters.
		return entered.Run(g, nil)
	}
	return g.PutCardsFromLibraryOntoBattlefieldKeyedForEffect([]uuid.UUID{chosen},
		ZoneEntryOptions{Controller: playerID, FaceDown: FaceDownManifested}, entered)
}

// manifestDreadEntered is the face-down entry's end: the other card (Carry)
// goes to the graveyard, as a keyed batch whose tail finishes the action.
var manifestDreadEntered = RegisterTail("engine/manifest-dread-entered", func(g *Game, o TailOutcome) error {
	manifested := uuid.Nil
	if len(o.IDs) > 0 {
		manifested = o.IDs[0]
	}
	t := o.Tail
	return g.routeAllKeyedLocked(zoneRoute{Dst: ZoneGraveyard, Actor: t.Player, Source: t.Source}, t.Carry,
		KeyedTail{Ref: manifestDreadFinished, Player: t.Player, Source: t.Source, Carry: []uuid.UUID{manifested}, Next: t.Next})
})

// manifestDreadFinished emits EventManifestDread and runs the card's tail
// (Next), told [manifested, graveyarded].
var manifestDreadFinished = RegisterTail("engine/manifest-dread-finished", func(g *Game, o TailOutcome) error {
	res := ManifestDreadResult{Player: o.Tail.Player}
	if len(o.Tail.Carry) > 0 {
		res.Manifested = o.Tail.Carry[0]
	}
	if len(o.IDs) > 0 {
		res.Graveyarded = o.IDs[0]
	}
	g.EmitEvent(Event{
		Kind:   EventManifestDread,
		Actor:  res.Player,
		Source: o.Tail.Source,
		CardID: res.Manifested,
		Target: res.Graveyarded,
	})
	if o.Tail.Next == nil {
		return nil
	}
	return g.runKeyedTailLocked(*o.Tail.Next, []uuid.UUID{res.Manifested, res.Graveyarded}, 0)
})

// ManifestDreadOutcome reads a keyed manifest dread's result back off its
// tail's outcome.
func ManifestDreadOutcome(o TailOutcome) ManifestDreadResult {
	res := ManifestDreadResult{Player: o.Tail.Player}
	if len(o.IDs) > 0 {
		res.Manifested = o.IDs[0]
	}
	if len(o.IDs) > 1 {
		res.Graveyarded = o.IDs[1]
	}
	return res
}

// manifestDreadWordSep separates a packed tail's key from its word on the
// pick prompt (packManifestDreadWord). No registered key holds it.
const manifestDreadWordSep = "|"

// packManifestDreadWord is the pick prompt's word for a card's tail: its
// key, then its word when it has one.
func packManifestDreadWord(t KeyedTail) string {
	if t.Word == "" {
		return t.Ref.key
	}
	return t.Ref.key + manifestDreadWordSep + t.Word
}

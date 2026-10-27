package game

import (
	"sort"

	"github.com/google/uuid"
)

// PendingChoiceProliferate is CR 701.34a's choice: "choose any number
// of permanents and/or players with counters on them". The proliferating
// player answers it with {card_ids: []string}, the same payload every
// card-set pick uses, and the list may name SEATS as well as permanents
// (ChoosePlayers are the seats on offer; a seat is named by its player
// ID). Floor zero: "any number" includes none.
//
// It joins isCardSetPickKind, so the bounds, candidate and duplicate
// checks, the enumerator's subset walk and the wire's Options
// projection are the card-set picks' own, written once. What it adds is
// ChooseSuggested, the engine's beneficial pick (everything of yours a
// counter helps, everything of theirs one hurts), which the client
// pre-selects, the enumerator always offers and the bot answers with.
//
// Each proliferate is its own prompt: "proliferate twice" (Tekuthal)
// asks twice, and the second prompt is built from the board the first
// answer left, as it is in paper.
const PendingChoiceProliferate PendingChoiceKind = "proliferate"

// proliferateQuestion is the prompt's header.
const proliferateQuestion = "Proliferate: choose any number of permanents and players with counters"

// harmfulCardCounters are the counter kinds a permanent's controller
// does not want more of. Everything else — +1/+1, loyalty, charge,
// shield, lore, defense on a battle you are not fighting — is either
// wanted or harmless, and proliferate gives one of EACH kind a
// permanent has, so a single unwanted kind rules the permanent out
// entirely rather than being skipped.
//
// Lore is a judgement call: another lore counter advances a Saga
// towards its final chapter and its sacrifice, which is usually the
// point (you want the chapter abilities) — it is treated as wanted.
var harmfulCardCounters = map[string]bool{
	CounterMinusOne: true,
	CounterStun:     true,
}

// harmfulCardCounter is harmfulCardCounters widened to every P/T
// counter kind (#1664, CR 122.1a): a -2/-1 or a -0/-1 shrinks the
// creature exactly as a -1/-1 does, and a +1/+0 or a +0/+1 grows it.
// A P/T kind is harmful when it takes more than it gives.
func harmfulCardCounter(name string) bool {
	if harmfulCardCounters[name] {
		return true
	}
	p, t, ok := ParsePTCounter(name)
	return ok && p+t < 0
}

// HarmfulCardCounter is harmfulCardCounter for the catalog: the counter
// kinds a permanent's controller would rather not have. Sin, Unending
// Cataclysm's suggested answer (ADR 0170) reads it to know which of its
// controller's own counters are worth stripping.
func HarmfulCardCounter(name string) bool { return harmfulCardCounter(name) }

// harmfulPlayerCounters are the player-level counters you do not want
// on yourself and do want on an opponent.
var harmfulPlayerCounters = map[string]bool{
	CounterPoison: true,
	CounterRad:    true,
}

// ProliferateCandidatesForEffect is everything CR 701.34a lets a
// proliferate choose from: the battlefield permanents with at least one
// counter on them and the seated players with at least one counter,
// in battlefield then seat order.
//
// Caller must hold g.mu (read or write).
func (g *Game) ProliferateCandidatesForEffect() (cards []uuid.UUID, players []uuid.UUID) {
	for _, c := range g.BattlefieldCardsForEffect() {
		if len(sortedCounterKinds(c.Counters)) > 0 {
			cards = append(cards, c.InstanceID)
		}
	}
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || len(sortedCounterKinds(p.Counters)) == 0 {
			continue
		}
		players = append(players, p.ID)
	}
	return cards, players
}

// ProliferateSuggestionForEffect is the pick a proliferate should
// default to on `controller`'s behalf: everything they control whose
// counters all help, plus every other permanent whose counters all
// hurt, and the same test applied to each player's own counters.
//
// It is a SUGGESTION and nothing more now (#2525): the prompt
// pre-selects it, the enumerator always offers it and the bot answers
// it, but the player may choose any other set. The catalog's
// BeneficialProliferateChoice is this function under the name the
// tests know it by.
//
// Deterministic: battlefield order then seat order, so the same board
// always produces the same event stream.
//
// Caller must hold g.mu (read or write).
func (g *Game) ProliferateSuggestionForEffect(controller uuid.UUID) (cards []uuid.UUID, players []uuid.UUID) {
	for _, c := range g.BattlefieldCardsForEffect() {
		if len(c.Counters) == 0 {
			continue
		}
		if wantsMoreCounters(c.Counters, harmfulCardCounter, c.Controller == controller) {
			cards = append(cards, c.InstanceID)
		}
	}
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || len(p.Counters) == 0 {
			continue
		}
		if wantsMoreCounters(p.Counters, func(name string) bool { return harmfulPlayerCounters[name] }, p.ID == controller) {
			players = append(players, p.ID)
		}
	}
	return cards, players
}

// wantsMoreCounters is the shared test both halves of the pick use.
// For something of yours (`mine`): choose it only when NO kind on it
// is harmful. For something that isn't: choose it only when EVERY
// kind on it is harmful — an opponent's creature with a -1/-1 counter
// is a fine choice; the same creature also carrying a +1/+1 counter
// is not, because proliferate would hand them both.
func wantsMoreCounters(counters map[string]int, harmful func(string) bool, mine bool) bool {
	seen := false
	for name, n := range counters {
		if n <= 0 {
			continue
		}
		seen = true
		if wanted := !harmful(name); wanted != mine {
			// Mine and harmful, or theirs and helpful.
			return false
		}
	}
	return seen
}

// proliferateAskLocked takes `remaining` proliferates one prompt at a
// time, then runs `then` (the rest of the sentence: Steady Progress's
// "draw a card"). A proliferate with nothing on the board to choose
// asks nothing and simply ends the sequence — "any number" includes
// zero, and a prompt with no candidates would have no answer but the
// empty one.
//
// The continuation captures scalars and the card's own `then` only —
// the StackItem.Effect contract — and resolves against whichever
// *Game it is handed, so an undone-then-redone answer works.
//
// Caller must hold g.mu.
func (g *Game) proliferateAskLocked(actor, source uuid.UUID, remaining int, then func(g *Game) error) error {
	finish := func() error {
		if then == nil {
			return nil
		}
		return then(g)
	}
	if remaining <= 0 {
		return finish()
	}
	cards, players := g.ProliferateCandidatesForEffect()
	if len(cards)+len(players) == 0 {
		return finish()
	}
	sc, sp := g.ProliferateSuggestionForEffect(actor)
	n := len(cards) + len(players)
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:            PendingChoiceProliferate,
		Chooser:         actor,
		FromPlayer:      actor,
		Count:           n,
		Source:          source,
		Reason:          proliferateQuestion,
		ChooseCards:     cards,
		ChoosePlayers:   players,
		ChooseSuggested: append(sc, sp...),
		ChooseMin:       0,
		ChooseMax:       n,
		chooseCardsResume: &chooseCardsFrame{
			then: proliferateThen(actor, source, remaining-1, then),
		},
	})
	if id == uuid.Nil {
		// The proliferating player is gone, so there is nobody to ask
		// and nothing to give; the rest of the sentence still runs.
		return finish()
	}
	return nil
}

// proliferateThen is a proliferate prompt's continuation: apply the
// picks, then, once every counter has landed, ask for the next
// proliferate of the settled count.
func proliferateThen(actor, source uuid.UUID, remaining int, then func(g *Game) error) func(g *Game, picked []uuid.UUID) error {
	return func(g *Game, picked []uuid.UUID) error {
		var cardIDs, playerIDs []uuid.UUID
		for _, id := range picked {
			if g.playerByIDLocked(id) != nil {
				playerIDs = append(playerIDs, id)
			} else {
				cardIDs = append(cardIDs, id)
			}
		}
		return g.applyProliferateThenLocked(actor, cardIDs, playerIDs, func(g *Game) error {
			return g.proliferateAskLocked(actor, source, remaining, then)
		})
	}
}

// ResolveProliferate answers a PendingChoiceProliferate. `picks` may
// name permanents and seats; both are checked against the candidates
// the prompt offered.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveProliferate(choiceID, chooserID uuid.UUID, picks []uuid.UUID) error {
	return g.resolveCardSetPick(PendingChoiceProliferate, choiceID, chooserID, picks)
}

// proliferate.go is CR 701.34: "Choose any number of permanents
// and/or players with counters on them, then give each another
// counter of each kind already there."
//
// The rule splits cleanly into a CHOICE and an APPLICATION.
// applyProliferateThenLocked is the application: it takes the chosen
// permanents and players as arguments and does not decide anything.
// The choice is the PendingChoiceProliferate prompt below (#2525): the
// proliferating player picks any number of the permanents and players
// that have counters, with the engine's beneficial pick offered as the
// suggested default, and the answer feeds the same application.
//
// The keyword ACTION — the CR 614 window that "if you would
// proliferate, proliferate twice instead" replaces, and the entry
// point every catalog proliferate goes through — is
// ProliferateForEffect in keyword_action.go (#976). It calls this
// once per time the window settled on.
//
// Two properties the application has to get right:
//
//   - "another counter of EACH KIND already there" — a creature with
//     a +1/+1 and a shield counter gets one of each, and a permanent
//     with no counters at all is not a legal choice and gets nothing.
//     Kinds are snapshotted before any counter is placed, so a
//     replacement that adds a NEW kind mid-loop can't cascade.
//   - The counters go through AddCounterByForEffect, not a raw map
//     write, so the CR 614 replacement pipeline sees them: Doubling
//     Season doubles a proliferated counter exactly as it doubles any
//     other, which is the paper interaction. Since ADR 0056 that is
//     true of the PLAYER half too — poison and energy used to be a bare
//     map write with no window, no event and no layer bump; see
//     player_counters.go.
//   - The PROLIFERATING PLAYER is the placer on every counter it
//     places, on permanents and on players alike (CR 701.34: "GIVE
//     each another counter"). That is what makes Vorinclex halve an
//     opponent's proliferate onto your creature and double your own.

// applyProliferateThenLocked gives each named permanent and each named
// player one additional counter of every kind they already have —
// ONE proliferate, the application half of CR 701.34 with the choice
// already made and the CR 614 window already settled — and then runs
// `then`.
//
// Both lists may be empty — "any number" includes zero, and a
// proliferate with nothing worth choosing is a legal no-op rather
// than an error. IDs that name something that has left the
// battlefield, or a player who is not seated, are skipped: the
// choice is made when the effect starts resolving and the board can
// have moved underneath it.
//
// #3930: the counters are placed one at a time, each behind the last
// one's continuation (AddCounterByThenForEffect), in the order the
// loop this replaced placed them: permanents in the order named, kinds
// sorted, then players. On the ordinary path every continuation runs
// inline, so nothing changes. When a placement pauses on a CR 616
// ordering prompt (a Doubling Season beside a Hardened Scales), the
// rest of the proliferate, and `then`, wait for the answer instead of
// running while the counter is still owed. That is what lets `then`
// read what the proliferate really did (ProliferateChoosingThenKeyedForEffect).
//
// Caller must hold g.mu (it is an effect-time helper).
func (g *Game) applyProliferateThenLocked(placer uuid.UUID, cardIDs, playerIDs []uuid.UUID, then func(g *Game) error) error {
	return g.proliferateCardsLocked(placer, cardIDs, playerIDs, 0, then)
}

// proliferateCardsLocked proliferates onto cardIDs[i:], then onto the
// players. Kinds are snapshotted as each permanent's turn comes, so a
// replacement that adds a NEW kind mid-placement can't cascade.
func (g *Game) proliferateCardsLocked(placer uuid.UUID, cardIDs, playerIDs []uuid.UUID, i int, then func(g *Game) error) error {
	for ; i < len(cardIDs); i++ {
		c, ok := g.LookupCardForEffect(cardIDs[i])
		if !ok {
			continue
		}
		if kinds := sortedCounterKinds(c.Counters); len(kinds) > 0 {
			return g.proliferateCardKindsLocked(placer, cardIDs, playerIDs, i, kinds, 0, then)
		}
	}
	return g.proliferatePlayersLocked(placer, playerIDs, 0, then)
}

// proliferateCardKindsLocked gives cardIDs[i] one counter of kinds[k],
// and in its continuation the next kind, then the next permanent.
func (g *Game) proliferateCardKindsLocked(placer uuid.UUID, cardIDs, playerIDs []uuid.UUID, i int, kinds []string, k int, then func(g *Game) error) error {
	if k >= len(kinds) {
		return g.proliferateCardsLocked(placer, cardIDs, playerIDs, i+1, then)
	}
	return g.AddCounterByThenForEffect(placer, cardIDs[i], kinds[k], 1, func(g *Game, _ int) error {
		return g.proliferateCardKindsLocked(placer, cardIDs, playerIDs, i, kinds, k+1, then)
	})
}

// proliferatePlayersLocked is proliferateCardsLocked's player half:
// playerIDs[i:], one counter of each kind each player already has,
// then `then`.
func (g *Game) proliferatePlayersLocked(placer uuid.UUID, playerIDs []uuid.UUID, i int, then func(g *Game) error) error {
	for ; i < len(playerIDs); i++ {
		p := g.playerByIDLocked(playerIDs[i])
		if p == nil || p.Eliminated {
			continue
		}
		if kinds := sortedCounterKinds(p.Counters); len(kinds) > 0 {
			return g.proliferatePlayerKindsLocked(placer, playerIDs, i, kinds, 0, then)
		}
	}
	if then == nil {
		return nil
	}
	return then(g)
}

// proliferatePlayerKindsLocked gives playerIDs[i] one counter of
// kinds[k], and in its continuation the next kind, then the next
// player.
func (g *Game) proliferatePlayerKindsLocked(placer uuid.UUID, playerIDs []uuid.UUID, i int, kinds []string, k int, then func(g *Game) error) error {
	if k >= len(kinds) {
		return g.proliferatePlayersLocked(placer, playerIDs, i+1, then)
	}
	return g.addPlayerCounterByThenLocked(placer, playerIDs[i], kinds[k], 1, func(g *Game, _ int) error {
		return g.proliferatePlayerKindsLocked(placer, playerIDs, i, kinds, k+1, then)
	})
}

// Proliferated is what one proliferate keyword action really gave
// counters to (#3930): every permanent and every player that received
// at least one counter, across every proliferate the CR 614 window
// settled on ("proliferate twice" is one action), each named once, in
// the order they first received one.
//
// "Really" is the point. A permanent that was chosen but had its
// counter replaced away (Vorinclex halving one counter to none), or
// that left before its turn came, got nothing and is not here; that is
// what "had a counter put on them this way" (Ripples of Potential)
// reads.
type Proliferated struct {
	Permanents []uuid.UUID
	Players    []uuid.UUID
}

// proliferatedSinceLocked is the permanents and players `player` put
// counters on after event `since`: each EventCounterPlaced naming them
// as placer that raised the count, and each EventPlayerCounterPlaced
// naming them with a positive delta.
//
// Caller must hold g.mu.
func (g *Game) proliferatedSinceLocked(player uuid.UUID, since uint64) Proliferated {
	var out Proliferated
	seen := map[uuid.UUID]bool{}
	for i := range g.Events {
		ev := &g.Events[i]
		if ev.Seq <= since || ev.Actor != player || ev.Target == uuid.Nil || seen[ev.Target] {
			continue
		}
		switch ev.Kind {
		case EventCounterPlaced:
			if ev.Amount > g.counterAmountBeforeLocked(i) {
				seen[ev.Target] = true
				out.Permanents = append(out.Permanents, ev.Target)
			}
		case EventPlayerCounterPlaced:
			if ev.Amount > 0 {
				seen[ev.Target] = true
				out.Players = append(out.Players, ev.Target)
			}
		}
	}
	return out
}

// counterAmountBeforeLocked is the count of g.Events[i]'s counter kind
// on its permanent before that event changed it: the most recent
// earlier EventCounterPlaced for the same permanent and kind, or zero
// when the permanent's arrival is more recent (CR 400.7) or nothing
// earlier is logged.
//
// Caller must hold g.mu.
func (g *Game) counterAmountBeforeLocked(i int) int {
	ev := g.Events[i]
	for j := i - 1; j >= 0; j-- {
		prev := g.Events[j]
		if prev.CardID == ev.Target {
			switch prev.Kind {
			case EventETB, EventTokenCreated:
				return 0
			case EventZoneMove:
				if prev.NewZone == ZoneBattlefield {
					return 0
				}
			}
		}
		if prev.Kind == EventCounterPlaced && prev.Target == ev.Target && prev.Label == ev.Label {
			return prev.Amount
		}
	}
	return 0
}

// sortedCounterKinds snapshots the counter names present on a
// counter map, in a stable order. Sorted rather than map order so a
// proliferate over several kinds emits its EventCounterPlaced
// stream identically on every run — map iteration order in Go is
// randomised, and an event log that reorders between runs makes
// replay diffs unreadable.
func sortedCounterKinds(counters map[string]int) []string {
	if len(counters) == 0 {
		return nil
	}
	out := make([]string, 0, len(counters))
	for name, n := range counters {
		if n <= 0 {
			continue
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

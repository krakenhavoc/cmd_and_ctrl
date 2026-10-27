package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// put_from_library.go — "put a [card] from among them onto the
// battlefield", where "them" are cards revealed or looked at on top of
// a library (#745).
//
// The clause is two halves and the catalog owns one of them, exactly
// as it does for the hand version (put_from_hand.go, #654). The MOVE
// is engine: game.PutFromLibraryOntoBattlefieldForEffect and its batch
// twin run the CR 614 entry pipeline, move the card and fire the ETB
// hook, without the search event or the shuffle a tutor owes. The
// PICK is card text — which of the revealed cards qualify, "a" or
// "any number", "you may", tapped or not, and where the rest go — and
// it lives here, on the choose_cards prompt with Zone: ZoneLibrary.
//
// Hidden information is the caller's first line, not this file's:
// a card that REVEALS calls RevealTopOfLibrary / revealUntil (every
// seat becomes a knower), and a card that LOOKS calls
// LookAtTopOfLibraryForEffect (only the looker does). The prompt then
// shows the chooser what they are entitled to and nobody else anything
// at all — protocol.FilterViewFor drops a choose_cards prompt's
// options, and its bounds, for every seat but the chooser.

// PutFromLibraryOntoBattlefield offers `Player` the cards in Cards that
// pass Match and puts the chosen ones onto the battlefield without
// casting them.
//
// # The three shapes a card prints
//
//   - "Put all land cards from among them onto the battlefield"
//     (Animist's Awakening): All. No prompt — there is no choice.
//   - "You may put a Dragon creature card from among them" (Ureni):
//     Optional with Max 1.
//   - "You may put any number of permanent cards … from among them"
//     (Genesis Wave): Optional with Max 0, meaning no ceiling.
//
// A mandatory clause whose only legal answer is every candidate ("put
// that card onto the battlefield" after a reveal-until) is not asked
// either: a prompt with one possible answer is a click, not a choice.
//
// # Simultaneity
//
// Everything chosen enters in ONE batch
// (PutCardsFromLibraryOntoBattlefieldForEffect): every entry
// replacement is evaluated against the board as it stood before any of
// them entered, and every ETB event fires after all of them have
// landed, so "any number" behaves as the simultaneous entry it is
// rather than as a sequence of entries that can see each other. The
// engine function's comment names the one residue (AsEnters choices
// run in sequence after the batch lands).
//
// # It queues; it does not finish
//
// Apply returns once the prompt is queued. The next step of a Do that
// is the item's own Effect waits for the answer (#4012, do_walk.go),
// but it cannot read the answer, and a Do run from inside a hand-written
// closure does not wait at all. So "the rest on the bottom of your
// library in a random order" — which must know which cards were NOT
// chosen — goes in ThenTail. ThenTail runs in every case:
// after the answer, immediately when there was nothing to choose from,
// immediately when All or a forced answer made the choice for the
// player, and even when the put itself returned an error (with whatever
// did enter), so the rest never stay on top because one card failed.
type PutFromLibraryOntoBattlefield struct {
	// Player chooses, and is who the permanents enter under. uuid.Nil
	// means the resolving item's controller.
	Player uuid.UUID

	// LibraryOwner is whose library Cards sit in, when that is not the
	// chooser's own: Lonis, Cryptozoologist's "target opponent reveals
	// the top X cards of their library. You may put a nonland
	// permanent card … from among them onto the battlefield under
	// your control". uuid.Nil (every other card) means the chooser's.
	// It is the prompt's FromPlayer, which the submit-time zone
	// re-check reads: without it a pick out of an opponent's library
	// is refused as "not in your library". The card enters under
	// Player's control and stays its owner's (CR 108.3), so it dies to
	// the owner's graveyard.
	LibraryOwner uuid.UUID

	// Cards are the revealed or looked-at cards, top card first — the
	// "them" of "from among them". A card that has since left its
	// library is ignored.
	Cards []uuid.UUID

	// Match is the clause's filter ("a land card", "a Dragon creature
	// card"). Nil means any permanent card; nonpermanent cards (CR
	// 110.4) and tokens (CR 108.2, CR 111.8) are dropped whatever
	// Match says.
	Match CardPredicate

	// Max is "a" (1) or "any number" (0). Ignored with All.
	Max int

	// Optional is the printed "you MAY put". It sets the prompt's
	// floor to zero.
	Optional bool

	// All is "put ALL [matching] cards from among them": no prompt.
	All bool

	// Tapped is the clause's own "onto the battlefield tapped".
	Tapped bool
	// Counters are the counters the put says each card enters with, as
	// data (a kind and a count) — the Undercity's Throne of the Dead Three:
	// "Put a creature card from among them onto the battlefield with three
	// +1/+1 counters on it" (ADR 0203). Seeded onto the entry, so they are
	// there as the permanent enters and go through the CR 614 counter
	// pipeline. With IfManaValueAtMost they are decided by the card picked:
	// "If that card has mana value 3 or less, it enters with three
	// additional +1/+1 counters on it" (Turntimber Symbiosis, #4170), read
	// off the picked card as it sits in the library.
	//
	// A gated clause puts ONE card (Max 1, not All): the counters belong to
	// the card, and one entry batch takes one set of options, so Apply
	// refuses IfManaValueAtMost on a clause that can put more than one.
	Counters PutCounters

	// Rule is the clause's rule about the picked SET, as opposed to
	// Match's rule about each card: "any number of nonland permanent cards
	// with TOTAL MANA VALUE 4 OR LESS from among them" (Ao, the Dawn Sky).
	// The zero value means the bounds and Match are the whole rule, which
	// is every other card that prints this sentence.
	//
	// It is a registered set rule (RegisterPutRule, #4253): a function of
	// the cards alone, looked up by the pick's key each time an answer is
	// checked, so the prompt carries no closure. It is called with the
	// picks in submitted order as live value copies, never for an empty
	// pick (so "put none of them" stays the answer nothing can refuse), and
	// it runs BOTH on the submit path and inside legal.EnumerateFor — so a
	// set the resolver would refuse is never offered to a bot seat, and a
	// set a client submits anyway comes back as ErrChoiceSetRejected with
	// the prompt still open.
	//
	// It also switches OFF the "the only legal answer is every candidate"
	// shortcut below: with a set rule, the legal answers are the SUBSETS
	// that pass it, so a mandatory clause that would otherwise be answered
	// for the player becomes a real prompt. It is ignored by All, which is
	// not a choice at all.
	Rule PutRule

	// Label is the prompt header, "<card> — <the printed clause>".
	Label string

	// ThenTail runs after the put with what happened, as a registered
	// tail (PutFromLibraryTail; #4253 slice 6): a key plus plain data, so
	// a table waiting on a put card's entry question is a restore point.
	// It is the only correct place for "the rest" and for any rider on
	// the result. It runs with g.mu held and may queue further prompts.
	// The tail is told the revealed cards (to compute "the rest"), the
	// chooser and the source card; ThenN is a number it reads, ThenCarry
	// IDs it reads (ahead of the revealed cards on the tail: register it
	// with PutFromLibraryTailCarrying), and ThenItem carries the
	// resolving item (its targets, its X, its text) as data for a body
	// that reads more than that.
	ThenTail  game.TailRef
	ThenN     int
	ThenItem  bool
	ThenCarry []uuid.UUID
}

// PutFromLibraryResult is what ThenTail is told about the put.
type PutFromLibraryResult struct {
	// Source is the card whose clause this was.
	Source uuid.UUID
	// Player is who was asked.
	Player uuid.UUID
	// Entered are the permanents that arrived, in the order given.
	// Empty when nothing did: a declined "you may", no candidate, or a
	// replacement that kept every card off the battlefield.
	Entered []uuid.UUID
	// Rest are the cards of Cards that are still in a library — "the
	// rest", "all cards revealed this way that weren't put onto the
	// battlefield" — in the order Cards gave them.
	Rest []uuid.UUID
	// N and Carry are what the keyed tail carried
	// (PutFromLibraryOntoBattlefield.ThenN, ThenCarry).
	N     int
	Carry []uuid.UUID
}

func (p PutFromLibraryOntoBattlefield) Apply(ctx *Context) error {
	player := p.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	source := ctx.Source()
	cards := append([]uuid.UUID(nil), p.Cards...)
	var keyed game.KeyedTail
	if p.ThenTail.Key() != "" {
		keyed = game.KeyedTail{Ref: p.ThenTail, Player: player, Source: source, Carry: append(append([]uuid.UUID(nil), p.ThenCarry...), cards...), N: p.ThenN}
		if p.ThenItem {
			keyed.Asked = game.NewAskedItem(ctx.Item)
		}
	}
	if p.Counters.IfManaValueAtMost && (p.All || p.Max != 1) {
		return fmt.Errorf("PutFromLibraryOntoBattlefield: counters gated on the picked card need a clause that puts exactly one card (Max 1, not All)")
	}

	candidates := libraryCardsMatching(ctx.Game, player, cards, p.Match)
	if len(candidates) == 0 {
		return putFromLibraryEnter(ctx.Game, player, nil, p.Tapped, p.Counters, keyed)
	}
	if p.All {
		return putFromLibraryEnter(ctx.Game, player, candidates, p.Tapped, p.Counters, keyed)
	}
	hi := p.Max
	if hi <= 0 || hi > len(candidates) {
		hi = len(candidates)
	}
	lo := hi
	if p.Optional {
		lo = 0
	}
	if lo == len(candidates) && p.Rule.IsZero() {
		// The only legal answer is every candidate. Not so with a
		// set rule: it is the rule, not the count, that decides
		// which subsets are answers, and taking the shortcut would
		// perform a set the prompt would have refused.
		return putFromLibraryEnter(ctx.Game, player, candidates, p.Tapped, p.Counters, keyed)
	}
	label := p.Label
	if label == "" {
		label = "Put a card from among them onto the battlefield"
	}
	// #4253: the pick is keyed — its set rule is the key's registered rule,
	// the entry's "tapped" and counters are its numbers and word, and the
	// put's tail is its tail — so a table waiting on the answer is a
	// restore point. Every pick is re-checked against the library on
	// submit (LiveZone).
	pick := putFromLibraryPicked
	if !p.Rule.IsZero() {
		pick = p.Rule.pick
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		Source:     source,
		Question:   label,
		Cards:      candidates,
		Min:        lo,
		Max:        hi,
		LiveZone:   game.ZoneLibrary,
		FromPlayer: p.LibraryOwner,
		ThenKey:    pick,
		CarryNums:  p.Counters.nums(p.Tapped),
		CarryWord:  p.Counters.Kind,
		Tail:       keyed,
	})
	return nil
}

// PutCounters is "with N Kind counters on it" on a put from among
// revealed or looked-at cards (PutFromLibraryOntoBattlefield.Counters):
// plain data, so the pick that waits for the answer carries it. With
// IfManaValueAtMost, only a picked card whose mana value is at most
// ManaValueAtMost gets them (Turntimber Symbiosis). The zero value is no
// counters.
type PutCounters struct {
	Kind              string
	N                 int
	IfManaValueAtMost bool
	ManaValueAtMost   int
}

// none reports whether the put places no counters.
func (c PutCounters) none() bool { return c.Kind == "" || c.N <= 0 }

// nums is the counters and the entry's "tapped" as the pick's numbers:
// [tapped] alone with no counters, as the pick has always carried it, else
// [tapped, N, gated, mana value]. The kind is the pick's word.
func (c PutCounters) nums(tapped bool) []int {
	t := 0
	if tapped {
		t = 1
	}
	if c.none() {
		return []int{t}
	}
	gated := 0
	if c.IfManaValueAtMost {
		gated = 1
	}
	return []int{t, c.N, gated, c.ManaValueAtMost}
}

// putCountersFromPick is nums' inverse, with the pick's word as the kind.
func putCountersFromPick(word string, nums []int) (tapped bool, c PutCounters) {
	tapped = len(nums) > 0 && nums[0] == 1
	if len(nums) >= 4 && word != "" {
		c = PutCounters{Kind: word, N: nums[1], IfManaValueAtMost: nums[2] == 1, ManaValueAtMost: nums[3]}
	}
	return tapped, c
}

// putFromLibraryEnter puts `picked` onto the battlefield under `chooser`'s
// control through the keyed door (#1322, #4253 slice 6), with the put's
// tail as the entry's tail, or runs the tail at once when nothing was
// picked. A card of the batch may stop to ask its own question (a
// shockland's life, a Clone's copy), and until it is answered the card is
// still in the library; "the rest" is computed when the entry is complete,
// so a card still waiting to enter is never put on the bottom with the
// pile. An error still reports whatever did enter and the tail still runs,
// so the revealed or looked-at rest is never left on top with its knowers.
//
// Caller holds g.mu.
func putFromLibraryEnter(g *game.Game, chooser uuid.UUID, picked []uuid.UUID, tapped bool, c PutCounters, tail game.KeyedTail) error {
	if len(picked) == 0 {
		return tail.Run(g, nil)
	}
	opts := game.LibraryEntryOptions{Controller: chooser, Tapped: tapped}
	if !c.none() && (!c.IfManaValueAtMost || (len(picked) == 1 && manaValueInLibraryAtMost(g, picked[0], c.ManaValueAtMost))) {
		opts.EntersWithCounters = map[string]int{c.Kind: c.N}
	}
	return g.PutCardsFromLibraryOntoBattlefieldKeyedForEffect(picked, opts, tail)
}

// manaValueInLibraryAtMost reports whether the card's mana value, read off
// it as it sits in the library, is at most n.
func manaValueInLibraryAtMost(g *game.Game, id uuid.UUID, n int) bool {
	card, ok := g.LookupCardForEffect(id)
	if !ok {
		return false
	}
	mv, ok := g.ManaValueForEffect(card)
	return ok && mv <= n
}

// putFromLibraryAnswered is the keyed answer to "put a card from among
// them onto the battlefield": the picks enter under the chooser's control,
// tapped and with counters as the prompt's numbers and word say, and the
// prompt's tail is the put's ThenTail. Nothing picked runs the tail at
// once.
func putFromLibraryAnswered(g *game.Game, r game.CardsChosen) error {
	tapped, counters := putCountersFromPick(r.Word, r.Nums)
	return putFromLibraryEnter(g, r.Chooser, r.Picked, tapped, counters, r.Tail)
}

// putFromLibraryPicked is the keyed answer with no set rule (#4253 slice
// 10b). An on-disk key: never renamed, never reused.
var putFromLibraryPicked = game.RegisterChooseCardsThen("choose-cards/put-from-library-onto-battlefield", putFromLibraryAnswered)

// PutRule names a registered set rule for PutFromLibraryOntoBattlefield's
// pick (PutFromLibraryOntoBattlefield.Rule). Hold one only from
// RegisterPutRule.
type PutRule struct{ pick game.ChooseCardsThen }

// IsZero reports whether no rule is named.
func (r PutRule) IsZero() bool { return r.pick.Key() == "" }

// RegisterPutRule registers a set rule for a put's pick under key (#4253):
// the pick's continuation is the put's, registered under the key with
// `rule` (game.RegisterChooseCardsThenRule), so a binary that does not
// know the key refuses the file and one that knows it has the rule. The
// key is an on-disk identity (ledgered "choose <key>"): never renamed,
// never reused. Call it once, from a package-level var. The rule must be
// a function of the cards alone and capture nothing that lives in a game.
func RegisterPutRule(key string, rule game.SetRuleFunc) PutRule {
	if rule == nil {
		panic("effects.RegisterPutRule: a rule is required")
	}
	return PutRule{pick: game.RegisterChooseCardsThenRule(key, rule, putFromLibraryAnswered)}
}

// libraryCardsMatching filters `ids` to the permanent cards still in a
// library that pass `pred`, in the order given. Characteristics are
// read off the card in the library, where they are printed.
//
// A token is never a candidate, whatever its type line says: it is not
// a "card" (CR 108.2), and a token that has left the battlefield can't
// come back onto it (CR 111.8). One only gets into a library because
// the engine has no CR 704.5d sweep (Chaos Warp tucks a token), and the
// move itself refuses it (game.PutCardsFromLibraryOntoBattlefieldForEffect).
//
// Caller holds g.mu.
func libraryCardsMatching(g *game.Game, player uuid.UUID, ids []uuid.UUID, pred CardPredicate) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range ids {
		z := g.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneLibrary {
			continue
		}
		c, ok := g.LookupCardForEffect(id)
		if !ok || !isPermanentCard(c) {
			continue
		}
		if pred != nil && !pred(g, player, c) {
			continue
		}
		out = append(out, id)
	}
	return out
}

// isPermanentCard is "a permanent card" as the library-to-battlefield
// family prints it: a permanent type (CR 110.4) on something that is a
// card rather than a token (CR 108.2, CR 111.8).
func isPermanentCard(c game.Card) bool {
	return c.IsPermanent() && !c.IsToken()
}

// cardsStillInALibrary is "the rest": the cards of `ids` that have not
// left a library, in order.
//
// Caller holds g.mu.
func cardsStillInALibrary(g *game.Game, ids []uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneLibrary {
			out = append(out, id)
		}
	}
	return out
}

// PutFromLibraryTail registers a PutFromLibraryOntoBattlefield rider as
// a keyed tail (#4253 slice 6): the body is handed what Then would be,
// with a Context rebuilt from the tail (the asking item's copy when the
// primitive's ThenItem carried it, else a stand-in with the chooser as
// controller and the source card). The key is an on-disk identity
// (ledgered "tail <key>"): never renamed, never reused. Call it once,
// from a package-level var.
func PutFromLibraryTail(key string, body func(ctx *Context, res PutFromLibraryResult) error) game.TailRef {
	return PutFromLibraryTailCarrying(key, 0, body)
}

// PutFromLibraryTailCarrying is PutFromLibraryTail for a tail that reads
// `extra` IDs the card put in PutFromLibraryOntoBattlefield.ThenCarry
// (handed to the body as PutFromLibraryResult.Carry).
func PutFromLibraryTailCarrying(key string, extra int, body func(ctx *Context, res PutFromLibraryResult) error) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		carry := o.Tail.Carry
		extra := extra
		if extra > len(carry) {
			extra = len(carry)
		}
		return body(ctx, PutFromLibraryResult{
			Source:  o.Tail.Source,
			Player:  o.Tail.Player,
			Entered: o.IDs,
			Rest:    cardsStillInALibrary(ctx.Game, carry[extra:]),
			N:       o.Tail.N,
			Carry:   carry[:extra],
		})
	})
}

// PutRestOnBottomInRandomOrderTail and PutRestIntoGraveyardTail are the shared riders
// as keyed tails: "put the rest on the bottom of your library in a
// random order" and "put the rest into your graveyard". On-disk keys.
var (
	PutRestOnBottomInRandomOrderTail = PutFromLibraryTail("put-from-library/rest-on-bottom-in-random-order",
		func(ctx *Context, res PutFromLibraryResult) error { return PutRestOnBottomInRandomOrder(ctx.Game, res) })
	PutRestIntoGraveyardTail = PutFromLibraryTail("put-from-library/rest-into-graveyard",
		func(ctx *Context, res PutFromLibraryResult) error { return PutRestIntoGraveyard(ctx.Game, res) })
)

// PutRestOnBottomInRandomOrder is the Then for "put the rest on the
// bottom of your library in a random order" — The Regalia, Ureni, Atla
// Palani, Animist's Awakening.
func PutRestOnBottomInRandomOrder(g *game.Game, res PutFromLibraryResult) error {
	return g.PutOnBottomInRandomOrderForEffect(res.Player, game.ZoneLibrary, res.Rest)
}

// PutRestIntoGraveyard is the Then for "put all cards revealed this way
// that weren't put onto the battlefield into your graveyard" (Genesis
// Wave). Not a mill: see game.PutIntoGraveyardForEffect.
//
// A token among the rest stays where it is: a token that has left the
// battlefield can't move to another zone (CR 111.8), and moving it to a
// graveyard would hand a reanimator an object the rules have already
// taken off the table.
//
// An error on one card does not keep the rest out of the graveyard: the
// loop carries on and the first error is returned, as the random-order
// bottom does.
func PutRestIntoGraveyard(g *game.Game, res PutFromLibraryResult) error {
	return restIntoGraveyard(g, res.Rest)
}

// restIntoGraveyard is the body PutRestIntoGraveyard and
// TakeRestIntoGraveyard share: "the rest" into their owners'
// graveyards, one routed move each.
//
// game.PutIntoGraveyardForEffect routes, so Rest in Peace, Leyline of
// the Void and CR 903.9 all see the arrival; it is not a mill (CR
// 701.17a counts off the TOP of a library), so it emits an ordinary
// zone move.
//
// A token stays where it is (CR 111.8), and an error on one card does
// not keep the others out: the loop carries on and the first error is
// returned.
func restIntoGraveyard(g *game.Game, ids []uuid.UUID) error {
	var firstErr error
	for _, id := range ids {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsToken() {
			continue
		}
		if err := g.PutIntoGraveyardForEffect(id); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// revealUntil reveals cards from the top of `player`'s library until
// it reveals one that passes `pred` — "reveal cards from the top of
// your library until you reveal a land card" — and returns every card
// revealed (top first, the hit last) and the hit, which is uuid.Nil
// when the library ran out first.
//
// Nothing moves. Revealing is not drawing, so an exhausted library is
// not a loss. The whole run is announced as one reveal, which is what
// the table sees happen.
//
// A token in the library is revealed with the rest but never stops the
// run: "until you reveal a creature card" asks for a card, and a token
// is not one (CR 108.2) — nor could "put that card onto the
// battlefield" bring it back (CR 111.8).
//
// Caller holds g.mu.
func revealUntil(ctx *Context, player uuid.UUID, pred func(game.Card) bool, reason string) ([]uuid.UUID, uuid.UUID) {
	run, hits := revealUntilCount(ctx, player, pred, 1, reason)
	if len(hits) == 0 {
		return run, uuid.Nil
	}
	return run, hits[0]
}

// revealUntilCount is revealUntil for "reveal cards from the top of your
// library until you reveal N creature cards" (Selvala's Stampede, a
// creature card for each wild vote; ADR 0146 amendment 2026-10-13): it
// stops at the Nth card that passes `pred` and returns every card
// revealed (top first, the last hit last) and the hits, oldest first —
// fewer than n when the library ran out first. n below 1 reveals
// nothing and announces nothing. The same token rule applies: a token
// is revealed with the rest but is never a hit (CR 108.2).
//
// Caller holds g.mu.
func revealUntilCount(ctx *Context, player uuid.UUID, pred func(game.Card) bool, n int, reason string) ([]uuid.UUID, []uuid.UUID) {
	p := ctx.PlayerByID(player)
	if p == nil || p.Library == nil || n < 1 {
		return nil, nil
	}
	var hits []uuid.UUID
	run := make([]uuid.UUID, 0, 8)
	for i := len(p.Library.Cards) - 1; i >= 0; i-- {
		c := p.Library.Cards[i]
		run = append(run, c.InstanceID)
		if !c.IsToken() && pred(c) {
			hits = append(hits, c.InstanceID)
			if len(hits) == n {
				break
			}
		}
	}
	ctx.Game.RevealForEffect(game.RevealSpec{
		Player: player,
		Source: ctx.Source(),
		Reason: reason,
		Cards:  run,
	})
	return run, hits
}

// RevealUntilThenPutOntoBattlefield is the reveal-until family's whole
// sentence: "reveal cards from the top of your library until you
// reveal a [Match] card. Put that card onto the battlefield [tapped]
// and the rest on the bottom of your library in a random order" (The
// Regalia, Atla Palani, The Prismatic Bridge).
//
// A library with no such card reveals itself entirely and every card
// goes to the bottom in a random order, which is the printed outcome:
// "that card" does not exist, and "the rest" is everything revealed.
type RevealUntilThenPutOntoBattlefield struct {
	// Match picks the card the reveal stops on.
	Match func(game.Card) bool
	// Tapped is "put that card onto the battlefield tapped".
	Tapped bool
	// Reason is the reveal banner.
	Reason string
}

func (r RevealUntilThenPutOntoBattlefield) Apply(ctx *Context) error {
	player := ctx.Controller()
	run, hit := revealUntil(ctx, player, r.Match, r.Reason)
	// With no hit the match refuses everything, nothing enters, and
	// Then still bottoms the whole run.
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    run,
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
		All:      true,
		Tapped:   r.Tapped,
		ThenTail: PutRestOnBottomInRandomOrderTail,
	}.Apply(ctx)
}

// revealTopThenPutIfMatch is "reveal the top card of [player's]
// library. If it's a [match] card, put it onto the battlefield
// [tapped]" — Coiling Oracle, Thrasios, Lurking Predators, Chaos
// Warp. It returns the revealed card (uuid.Nil for an empty library),
// so the card can write its own "otherwise" about what the card is.
//
// The permanent enters under `player`'s control, which on every card
// that prints this sentence is the library's owner.
//
// A revealed token is "not a [match] card" (CR 108.2) and is left where
// it is (CR 111.8): the card's "otherwise" branch sees it and must not
// move it either (see IsToken).
//
// Caller holds g.mu.
func revealTopThenPutIfMatch(g *game.Game, source, player uuid.UUID, match func(game.Card) bool, tapped bool, reason string) (uuid.UUID, error) {
	ids := g.RevealTopOfLibraryForEffect(player, source, 1, reason)
	if len(ids) == 0 {
		return uuid.Nil, nil
	}
	top := ids[0]
	c, ok := g.LookupCardForEffect(top)
	if !ok || !isPermanentCard(c) || !match(c) {
		return top, nil
	}
	_, err := g.PutFromLibraryOntoBattlefieldForEffect(top, game.LibraryEntryOptions{
		Controller: player,
		Tapped:     tapped,
	})
	return top, err
}

// LookAtTopThenMayPutOntoBattlefield is "look at the top N cards of
// your library. You may put [a / any number of] [Match] card[s] from
// among them onto the battlefield. Put the rest on the bottom of your
// library in a random order" — Ureni of the Unwritten, Gilgamesh.
//
// Max is 1 for "a" and 0 for "any number".
//
// No Rule parameter: a clause that also rules on the picked SET
// ("with total mana value 4 or less", Ao, the Dawn Sky) states
// PutFromLibraryOntoBattlefield directly, the way Armored Skyhunter
// does for its own rider. Widening this signature for a field two of
// its three callers would pass nil to buys nothing.
func LookAtTopThenMayPutOntoBattlefield(n int, match CardPredicate, max int, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		player := ctx.Controller()
		looked := g.LookAtTopOfLibraryForEffect(player, n)
		return PutFromLibraryOntoBattlefield{
			Player:   player,
			Cards:    looked,
			Match:    match,
			Max:      max,
			Optional: true,
			Label:    label,
			ThenTail: PutRestOnBottomInRandomOrderTail,
		}.Apply(ctx)
	}
}

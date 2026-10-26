package game

import (
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

// sticker.go — Unfinity's stickers and ticket counters (#2138, ADR 0208).
//
//	107.17   The ticket symbol is {TK}. It represents one ticket counter.
//	107.17a  A ticket symbol with a number inside it represents a ticket
//	         cost. To pay that cost, a player removes that many ticket
//	         counters from themselves.
//	123.1    A sticker is a marker placed on an object that modifies its
//	         characteristics and/or interacts with a rule, ability, or
//	         effect. Stickers are not objects. … Changes to an object
//	         from stickers are not part of its copiable values. There are
//	         four kinds of stickers: name stickers; ability stickers;
//	         power and toughness stickers; and art stickers.
//	123.2a   In constructed play, a player who chooses to play with
//	         stickers must start the game with at least ten sticker
//	         sheets … each of their sticker sheets must be unique. …
//	         Each player playing with sticker sheets reveals all of their
//	         sticker sheets and chooses three of them at random.
//	123.3    If an effect instructs a player to put a sticker on an
//	         object, that player chooses a sticker that is not currently
//	         on any objects they own from among the stickers they have
//	         access to and puts it on that object.
//	123.3b   A player can't put a sticker on an object that they don't own.
//	123.3c   … In order to put a sticker with a ticket cost on an object,
//	         the player who owns that object must pay that much {TK}.
//	123.5    Stickers on an object are not retained as that object moves
//	         to a hidden zone. Stickers are retained as that object moves
//	         to a public zone …; this is an exception to rule 400.7.
//	123.6    A name sticker … causes the word on that sticker to be added
//	         to the text of that object's name. This is a text-changing
//	         effect (613.1c).
//	123.7    An ability sticker … causes that object to gain the ability
//	         that is printed on that sticker (613.1f).
//	123.8    A power and toughness sticker on a creature … sets that
//	         object's power and toughness to the values printed on that
//	         sticker (613.4b). … use timestamp order (613.7).
//	123.9    An art sticker on a permanent has no effect on game play
//	         other than to act as a marker …
//	613.7k   A sticker receives a new timestamp each time it's put on an
//	         object. If the object a sticker is on receives a new
//	         timestamp, the sticker receives a new timestamp immediately
//	         after that one.
//
// # The shape
//
// Tickets are a player counter, CounterTicket, beside energy and rad: "you
// get {TK}" is AddPlayerCounterByForEffect and a sticker's ticket cost is
// payTicketsLocked, the one path that removes them.
//
// A sticker sheet is not a card and has no characteristics (CR 123.2). It
// is plain data: StickerSheet, the sheet's printed stickers in printed
// order. The deck importer builds it from the Scryfall record and hands it
// to the game on a pseudo-card (Card.StickerSheet) that seatDeckCard turns
// into Player.StickerSheetPool and never puts in a zone. As the game
// begins, dealAndBeginLocked chooses three of the pool at random, on a
// stream of their own (CR 103.2d), into Player.StickerSheets: the stickers
// the player has access to.
//
// A sticker on an object is plain data on the card, Card.Stickers: which
// sheet, which sticker, what it says, when it was put on and (for a name
// sticker) where in the name it went. Nothing is a closure and nothing is
// a new container of cards, so the snapshot carries it as JSON and the
// closure census has nothing new to count.
//
// What a sticker DOES is a continuous effect gathered by the layer pass
// from that data (stickerContinuousEffectsLocked), one more source list
// beside the keyword counters: a name sticker in layer 3, an ability
// sticker in layer 6 and a power/toughness sticker in layer 7b, each at
// the sticker's CR 613.7k timestamp. An ability sticker's ability is the
// catalog's (CatalogStickerAbility: its keywords, or a grant bundle for a
// triggered or activated ability), looked up by the sheet's oracle ID and
// the sticker's place among the sheet's ability stickers. A sticker has no
// source permanent, so CR 613.6's silencing never touches it; a "loses all
// abilities" with a later timestamp takes a sticker's ability away, an
// earlier one does not. Off the battlefield the same three are folded
// onto the baseline (applyStickersOffBattlefield), as keyword counters are.
//
// MoveCard drops every sticker on a card that moves to a hidden zone or
// out of the game (CR 123.5), and keeps them on every other move. A token
// or a copy never inherits them (CR 123.1, 707.2).

// StickerKickerKey is the optional-cost key sticker kicker is announced
// under (CR 702.33h, #4307). It is a kicker ability, but not the kicker a
// card prints: an "if it was kicked" ability is linked to the kicker
// printed on its own object (CR 702.33e), so it has a key of its own, which
// KickedTimesPaid does not count, and which a granted copy
// (effects.StickerKicker, Wicker Picker) can sit beside a printed kicker
// without the offer dropping it as a duplicate.
const StickerKickerKey = "sticker-kicker"

// StickerKind is one of CR 123.1's four kinds.
type StickerKind string

const (
	StickerName    StickerKind = "name"
	StickerAbility StickerKind = "ability"
	StickerPT      StickerKind = "pt"
	StickerArt     StickerKind = "art"
)

// StickerPrint is one sticker as it is printed on its sheet.
type StickerPrint struct {
	Kind StickerKind `json:"kind"`
	// Text is what the sticker says: a name sticker's word or words, an
	// ability sticker's ability (reminder text dropped), a power and
	// toughness sticker's "2/3". Empty for an art sticker.
	Text string `json:"text,omitempty"`
	// TicketCost is the number in the sticker's ticket symbol (CR
	// 123.3c), 0 for a sticker that has none.
	TicketCost int `json:"ticketCost,omitempty"`
	// Power and Toughness are a power and toughness sticker's numbers.
	Power     int `json:"power,omitempty"`
	Toughness int `json:"toughness,omitempty"`
	// Ordinal is the sticker's place among its sheet's stickers of the
	// same kind, from 0. An ability sticker's ability is looked up by it.
	Ordinal int `json:"ordinal,omitempty"`
}

// StickerSheet is one sticker sheet (CR 123.2): its identity and its
// stickers in printed order. Not a card, not an object, no
// characteristics.
type StickerSheet struct {
	OracleID   string         `json:"oracleId"`
	ScryfallID string         `json:"scryfallId,omitempty"`
	Name       string         `json:"name"`
	Stickers   []StickerPrint `json:"stickers"`
}

// PlacedSticker is one sticker on an object.
type PlacedSticker struct {
	// Sheet is the sheet's oracle ID and Index the sticker's position on
	// it: together, which sticker this is (CR 123.3a).
	Sheet     string `json:"sheet"`
	SheetName string `json:"sheetName,omitempty"`
	Index     int    `json:"index"`
	// Print is what the sticker says.
	Print StickerPrint `json:"print"`
	// PlacedAt is the CR 613.7k timestamp it got as it was put on.
	PlacedAt int64 `json:"placedAt,omitempty"`
	// After is a name sticker's position: the number of words that were
	// before it in the name when it was put on (CR 123.6b, 123.6c).
	After int `json:"after,omitempty"`
}

// StickerAbilityDef is what an ability sticker gives the object it is on:
// keyword tokens, and the key of a grant bundle (ADR 0093) for an ability
// that is not a keyword. The catalog declares it per sheet.
type StickerAbilityDef struct {
	Keywords []string
	Grant    string
	// GraveyardCastLife, when above zero, makes the sticker give the card
	// "You may cast this card from your graveyard by paying N life in
	// addition to paying its other costs" (Eldrazi Guacamole Tightrope's
	// second sticker): a static ability that functions from the graveyard
	// (CR 123.7: an ability sticker gives its ability to a card in any
	// zone). It is a derived cast permission (stickerGraveyardPermissionLocked),
	// not a layered ability, so it needs neither keywords nor a bundle.
	GraveyardCastLife int
}

// CatalogStickerAbility is the catalog's answer for the ordinal-th
// ability sticker of the sheet `sheet` (an oracle ID). Set by the effects
// package at init. False, or a nil hook, means the catalog does not run
// that sticker: it is placed and shown, and it does nothing — weaker than
// printed, the posture of any card the catalog does not run.
var CatalogStickerAbility func(sheet string, ordinal int) (StickerAbilityDef, bool)

// stickerIDSpace namespaces the stable IDs a sticker rides a prompt as.
var stickerIDSpace = uuid.MustParse("6a3f0f61-5d4e-4b8e-9b55-2138a5c1e7d1")

// StickerID is the stable ID of the sticker at `index` on the sheet
// `sheet`: the same on every binary and every restore, so a prompt can
// carry it (OptionPickPrompt.Carry). A player's sheets are unique (CR
// 123.2a), so it names one sticker of theirs.
func StickerID(sheet string, index int) uuid.UUID {
	return uuid.NewSHA1(stickerIDSpace, []byte(sheet+"#"+strconv.Itoa(index)))
}

// HasStickers reports whether the object is "stickered" (CR 123.4).
func (c *Card) HasStickers() bool { return c != nil && len(c.Stickers) > 0 }

// StickersOf returns a copy of the stickers on the object, oldest first.
func StickersOf(c *Card) []PlacedSticker {
	if c == nil || len(c.Stickers) == 0 {
		return nil
	}
	return append([]PlacedSticker(nil), c.Stickers...)
}

// StickerCountOfKind is how many stickers of `kind` are on the object.
func StickerCountOfKind(c *Card, kind StickerKind) int {
	n := 0
	if c == nil {
		return 0
	}
	for _, s := range c.Stickers {
		if s.Print.Kind == kind {
			n++
		}
	}
	return n
}

// PlayerTickets is how many ticket counters p has. Nil has none.
func PlayerTickets(p *Player) int {
	if p == nil {
		return 0
	}
	return p.Counters[CounterTicket]
}

// payTicketsLocked removes n ticket counters from `payer` to pay a ticket
// cost (CR 107.17a). It is the only path that pays tickets, and it changes
// nothing and reports false when the payer has fewer than n (CR 118.3).
// Zero or less pays nothing.
//
// Caller must hold g.mu.
func (g *Game) payTicketsLocked(payer uuid.UUID, n int, source uuid.UUID) bool {
	if n <= 0 {
		return true
	}
	p := g.playerByIDLocked(payer)
	if p == nil || PlayerTickets(p) < n {
		return false
	}
	before := PlayerTickets(p)
	after := before - n
	setPlayerCounterLocked(p, CounterTicket, after)
	g.emitPlayerCounterDeltaLocked(payer, CounterTicket, before, after, payer, source)
	return true
}

// seatStickerSheet adds one sheet to the player's revealed pool (CR
// 123.2a), from the deck's pseudo-card.
func seatStickerSheet(p *Player, sheet StickerSheet) {
	p.StickerSheetPool = append(p.StickerSheetPool, cloneStickerSheet(sheet))
}

// StickerSheetsChosen is the number of sheets a constructed player
// chooses at random from those they revealed (CR 103.2d, 123.2a).
const StickerSheetsChosen = 3

// chooseStickerSheetsLocked chooses the sheets `p` has access to this
// game: three of their revealed pool at random (CR 103.2d), on the
// sticker_sheets stream, so a deck with sheets shuffles its library
// exactly as it would without them. A pool of three or fewer is used
// whole, in pool order; the deck checker has already said whether the
// pool was legal (CR 123.2a), and the importer seats none of an illegal
// one.
//
// Caller must hold g.mu in write mode.
func (g *Game) chooseStickerSheetsLocked(p *Player) {
	if p == nil {
		return
	}
	p.StickerSheets = nil
	if len(p.StickerSheetPool) == 0 {
		return
	}
	order := make([]int, len(p.StickerSheetPool))
	for i := range order {
		order[i] = i
	}
	if len(order) > StickerSheetsChosen {
		r := g.randForLocked(rngStream{kind: rngStreamStickerSheets, player: p.ID})
		r.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
		order = order[:StickerSheetsChosen]
		sort.Ints(order)
	}
	for _, i := range order {
		p.StickerSheets = append(p.StickerSheets, cloneStickerSheet(p.StickerSheetPool[i]))
	}
}

// StickerChoice is one sticker a player may put on an object now.
type StickerChoice struct {
	ID        uuid.UUID
	Sheet     string
	SheetName string
	Index     int
	Print     StickerPrint
}

// Label is how a prompt names the sticker.
func (s StickerChoice) Label() string {
	return StickerLabel(s.Print, s.SheetName)
}

// StickerLabel is a sticker's name on a prompt or in the log: what it
// says, its kind, its ticket cost and its sheet.
func StickerLabel(p StickerPrint, sheet string) string {
	var b strings.Builder
	switch p.Kind {
	case StickerName:
		b.WriteString("Name sticker “" + p.Text + "”")
	case StickerAbility:
		b.WriteString("Ability sticker: " + p.Text)
	case StickerPT:
		b.WriteString("Power and toughness sticker " + p.Text)
	case StickerArt:
		b.WriteString("Art sticker " + strconv.Itoa(p.Ordinal+1))
	}
	if p.TicketCost > 0 {
		b.WriteString(" (" + strconv.Itoa(p.TicketCost) + " " + stickerPlural(p.TicketCost, "ticket", "tickets") + ")")
	}
	if sheet != "" {
		b.WriteString(" — " + sheet)
	}
	return b.String()
}

func stickerPlural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// stickersInUseLocked are the stickers of `owner` currently on objects
// they own, by sticker ID (CR 123.3: such a sticker can't be put on
// anything else). Every public zone is walked, the phased-out permanents
// and the junkyard included; a hidden zone holds no stickers (CR 123.5).
//
// Caller must hold g.mu.
func (g *Game) stickersInUseLocked(owner uuid.UUID) map[uuid.UUID]bool {
	used := map[uuid.UUID]bool{}
	note := func(z *Zone) {
		if z == nil {
			return
		}
		for i := range z.Cards {
			c := &z.Cards[i]
			if c.Owner != owner || len(c.Stickers) == 0 {
				continue
			}
			for _, s := range c.Stickers {
				used[StickerID(s.Sheet, s.Index)] = true
			}
		}
	}
	note(g.Battlefield)
	note(g.Stack)
	note(g.Exile)
	note(g.PhasedOut)
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		note(p.Graveyard)
		note(p.Command)
		note(p.Junkyard)
	}
	return used
}

// StickersInUseLocked is stickersInUseLocked for the view: the IDs
// (StickerID) of `owner`'s stickers that are on an object now. Caller
// holds g.mu (read is enough).
func (g *Game) StickersInUseLocked(owner uuid.UUID) map[uuid.UUID]bool {
	return g.stickersInUseLocked(owner)
}

// StickerLimit narrows the stickers an instruction allows beyond their
// kind: Pin Collection's "an ability sticker with ticket cost X or less on
// it without paying that sticker's ticket cost".
type StickerLimit struct {
	// MaxCost caps the sticker's ticket cost when HasMax is set.
	MaxCost int
	HasMax  bool
	// Free puts the sticker on without paying its ticket cost, so it is
	// offered whatever the owner has (CR 123.3c asks the payment only of
	// the one who must make it).
	Free bool
}

// AvailableStickersLocked are the stickers `owner` may put on an object
// they own right now, in sheet then printed order: on a sheet they have
// access to, not on any object they own (CR 123.3), of one of `kinds`
// (every kind when empty), and with a ticket cost they can pay (CR
// 123.3c).
//
// Caller must hold g.mu.
func (g *Game) AvailableStickersLocked(owner uuid.UUID, kinds []StickerKind) []StickerChoice {
	return g.availableStickersLimitedLocked(owner, kinds, StickerLimit{})
}

// availableStickersLimitedLocked is AvailableStickersLocked under a limit.
func (g *Game) availableStickersLimitedLocked(owner uuid.UUID, kinds []StickerKind, lim StickerLimit) []StickerChoice {
	p := g.playerByIDLocked(owner)
	if p == nil || len(p.StickerSheets) == 0 {
		return nil
	}
	used := g.stickersInUseLocked(owner)
	tickets := PlayerTickets(p)
	var out []StickerChoice
	for _, sheet := range p.StickerSheets {
		for i, s := range sheet.Stickers {
			if len(kinds) > 0 && !stickerKindIn(s.Kind, kinds) {
				continue
			}
			id := StickerID(sheet.OracleID, i)
			if used[id] {
				continue
			}
			if lim.HasMax && s.TicketCost > lim.MaxCost {
				continue
			}
			if !lim.Free && s.TicketCost > tickets {
				continue
			}
			out = append(out, StickerChoice{ID: id, Sheet: sheet.OracleID, SheetName: sheet.Name, Index: i, Print: s})
		}
	}
	return out
}

// AvailableStickersForEffect is AvailableStickersLocked for the catalog.
// Caller must hold g.mu.
func (g *Game) AvailableStickersForEffect(owner uuid.UUID, kinds []StickerKind) []StickerChoice {
	return g.AvailableStickersLocked(owner, kinds)
}

func stickerKindIn(k StickerKind, kinds []StickerKind) bool {
	for _, x := range kinds {
		if x == k {
			return true
		}
	}
	return false
}

// stickerChoiceByIDLocked finds one of `owner`'s available stickers by ID,
// under a limit.
func (g *Game) stickerChoiceByIDLocked(owner, id uuid.UUID, lim StickerLimit) (StickerChoice, bool) {
	for _, s := range g.availableStickersLimitedLocked(owner, nil, lim) {
		if s.ID == id {
			return s, true
		}
	}
	return StickerChoice{}, false
}

// EventStickerPlaced — a player put a sticker on an object (CR 123.3).
// Actor is the player, CardID the object, Source the card whose effect
// it was, Label the sticker's kind and text ("ability:Menace", read with
// StickerEventKindText) and Amount the tickets paid. "Whenever
// you put a sticker on …" (Wee Champion, Goblin Airbrusher) watches it.
const EventStickerPlaced EventKind = "sticker_placed"

// stickerHostLocked finds the object a sticker may go on: a card in a
// public zone (CR 123.5 — a hidden zone keeps no stickers).
func (g *Game) stickerHostLocked(cardID uuid.UUID) *Card {
	z, pos := g.locateCardLocked(cardID)
	if z == nil || pos < 0 {
		return nil
	}
	switch z.Kind {
	case ZoneLibrary, ZoneHand:
		return nil
	}
	return &z.Cards[pos]
}

// placeStickerLocked puts `owner`'s sticker `id` on the object `cardID`,
// paying its ticket cost. A name sticker goes after `after` words of the
// object's current name. Nothing happens, and false comes back, when the
// object is not in a public zone, is not `owner`'s (CR 123.3b), or the
// sticker is not available or its cost can't be paid (CR 123.3c).
//
// Caller must hold g.mu in write mode.
func (g *Game) placeStickerLocked(owner, cardID, id uuid.UUID, after int, source uuid.UUID) bool {
	return g.placeStickerUnderLocked(owner, cardID, id, after, source, StickerLimit{})
}

// placeStickerUnderLocked is placeStickerLocked under a limit: a Free one
// pays no ticket cost.
//
// Caller must hold g.mu in write mode.
func (g *Game) placeStickerUnderLocked(owner, cardID, id uuid.UUID, after int, source uuid.UUID, lim StickerLimit) bool {
	host := g.stickerHostLocked(cardID)
	if host == nil || host.Owner != owner {
		return false
	}
	choice, ok := g.stickerChoiceByIDLocked(owner, id, lim)
	if !ok {
		return false
	}
	paid := choice.Print.TicketCost
	if lim.Free {
		paid = 0
	}
	if !g.payTicketsLocked(owner, paid, source) {
		return false
	}
	// The payment may have moved nothing, but look again: the counter
	// event is announced and nothing can have moved the object since.
	if host = g.stickerHostLocked(cardID); host == nil {
		return false
	}
	if after < 0 {
		after = 0
	}
	placed := PlacedSticker{
		Sheet:     choice.Sheet,
		SheetName: choice.SheetName,
		Index:     choice.Index,
		Print:     choice.Print,
		PlacedAt:  timeNowUnixNano(),
	}
	if choice.Print.Kind == StickerName {
		placed.After = after
	}
	host.Stickers = append(host.Stickers[:len(host.Stickers):len(host.Stickers)], placed)
	host.effective = nil
	g.layerVersion.Add(1)
	g.recordStickerPlacedLocked(source, cardID, placed)
	g.EmitEvent(Event{
		Kind:   EventStickerPlaced,
		Actor:  owner,
		CardID: cardID,
		Source: source,
		Label:  string(choice.Print.Kind) + ":" + choice.Print.Text,
		Amount: paid,
		Words:  nameStickerWords(host.Stickers),
	})
	return true
}

// StickerPlacedRecordKey is the step record (#4163) a placement files
// for the instruction that asked for it, so a later step of the same
// resolution can read "that sticker": "put a name sticker on it. Add {R}
// for each unique vowel on that sticker." It is filed under the asking
// card (the placement's source) and lives for the one resolution.
const StickerPlacedRecordKey = "sticker/placed"

// recordStickerPlacedLocked files the placement for the asking card:
// IDs gets the object, Strs the "kind:text" of the sticker, in the order
// they went on. A placement with no source files nothing.
//
// Caller must hold g.mu in write mode.
func (g *Game) recordStickerPlacedLocked(source, object uuid.UUID, placed PlacedSticker) {
	if source == uuid.Nil {
		return
	}
	holder := &StackItem{ID: source}
	rec, _ := g.StepRecordForEffect(holder, StickerPlacedRecordKey)
	rec.IDs = append(rec.IDs, object)
	rec.Strs = append(rec.Strs, string(placed.Print.Kind)+":"+placed.Print.Text)
	g.RecordStepForEffect(holder, StickerPlacedRecordKey, rec)
}

// StickerPlacedByForEffect is the latest sticker the card `source` put
// on an object in the resolution that is open now: its kind, what it
// says, the object it went on and whether there is one. False when the
// instruction put none on (declined, or nothing qualified).
//
// Caller must hold g.mu.
func (g *Game) StickerPlacedByForEffect(source uuid.UUID) (kind StickerKind, text string, object uuid.UUID, ok bool) {
	rec, found := g.StepRecordForEffect(&StackItem{ID: source}, StickerPlacedRecordKey)
	if !found || len(rec.IDs) == 0 || len(rec.Strs) == 0 {
		return "", "", uuid.Nil, false
	}
	i := min(len(rec.IDs), len(rec.Strs)) - 1
	k, t, _ := strings.Cut(rec.Strs[i], ":")
	return StickerKind(k), t, rec.IDs[i], true
}

// nameStickerWords are the words of the name stickers in `stickers`, in
// the order they went on.
func nameStickerWords(stickers []PlacedSticker) []string {
	var out []string
	for _, s := range stickers {
		if s.Print.Kind == StickerName {
			out = append(out, s.Print.Text)
		}
	}
	return out
}

// stickerKickerPaidLocked is how many sticker kicker costs were paid for
// the spell `card` that was just cast: its own (by index) and any a
// permanent granted (by key).
func (g *Game) stickerKickerPaidLocked(card Card, paid PaidCost) int {
	return OptionalCostTimesPaid(card, paid.OptionalCosts, StickerKickerKey) + GrantedOptionalTimesPaid(paid, StickerKickerKey)
}

// payStickerKickerLocked is the second half of CR 702.33h: "As an
// additional cost to cast this spell, if it's kicked, you get a ticket
// counter and you may put a sticker on this spell." Called as the spell
// becomes cast, with every cost paid and before EventCast, so the sticker
// is on the spell when the cast triggers look at it and before anyone has
// priority. One ticket and one question per sticker kicker cost paid. The
// sticker goes on the spell card on the stack, which keeps it as the
// spell becomes a permanent or goes to the graveyard (CR 123.5); a spell
// the caster does not own takes none (CR 123.3b).
//
// Caller must hold g.mu in write mode.
func (g *Game) payStickerKickerLocked(player uuid.UUID, card Card) {
	item := g.StackMeta[card.InstanceID]
	if item == nil {
		return
	}
	n := g.stickerKickerPaidLocked(card, item.Paid)
	for i := 0; i < n; i++ {
		_ = g.AddPlayerCounterByForEffect(player, player, CounterTicket, 1)
		g.PutStickerForEffect(StickerRequest{
			Player:   player,
			Source:   card.InstanceID,
			Objects:  []uuid.UUID{card.InstanceID},
			Question: "Sticker kicker: you may put a sticker on " + card.Name + ". Choose it, or none.",
		})
	}
}

// StickerEventKindText reads an EventStickerPlaced's Label: the kind of
// sticker that was put on and what it says.
func StickerEventKindText(ev Event) (StickerKind, string) {
	kind, text, _ := strings.Cut(ev.Label, ":")
	return StickerKind(kind), text
}

// PlaceStickerForTest puts a sticker on an object directly, for tests.
func (g *Game) PlaceStickerForTest(owner, cardID, id uuid.UUID, after int) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.placeStickerLocked(owner, cardID, id, after, uuid.Nil)
}

// ---------------------------------------------------------------------
// Names (CR 123.6)

// isBlankNameToken reports whether a token of a name is a blank line
// ("_____"), which is not a word (CR 123.6a).
func isBlankNameToken(t string) bool {
	return t != "" && strings.Trim(t, "_") == ""
}

// NameWordCount is the number of words in a name (CR 123.6a): runs of
// non-space characters, blank lines not counted.
func NameWordCount(name string) int {
	n := 0
	for _, t := range strings.Fields(name) {
		if !isBlankNameToken(t) {
			n++
		}
	}
	return n
}

// InsertNameWord adds `word` to `name` after `after` of its words (CR
// 123.6b), or at the end when the name has fewer words than that now (CR
// 123.6c). A name with no text becomes the word. Blank lines stay where
// they are and are not counted.
func InsertNameWord(name, word string, after int) string {
	word = strings.TrimSpace(word)
	if word == "" {
		return name
	}
	tokens := strings.Fields(name)
	if len(tokens) == 0 {
		return word
	}
	at := len(tokens)
	if after <= 0 {
		at = 0
	} else {
		seen := 0
		for i, t := range tokens {
			if isBlankNameToken(t) {
				continue
			}
			seen++
			if seen == after {
				at = i + 1
				break
			}
		}
	}
	out := make([]string, 0, len(tokens)+1)
	out = append(out, tokens[:at]...)
	out = append(out, word)
	out = append(out, tokens[at:]...)
	return strings.Join(out, " ")
}

// ---------------------------------------------------------------------
// Counting letters on a name sticker (CR 123.6d, 123.6e)

// StickerLetters is the number of letters on a name sticker's text: its
// words without spaces, hyphens or other punctuation.
func StickerLetters(text string) int {
	n := 0
	for _, r := range text {
		if unicode.IsLetter(r) {
			n++
		}
	}
	return n
}

// StickerLetterCount is how many times `letter` appears on a name
// sticker's text. A lowercase letter and its uppercase equivalent are the
// same letter (CR 123.6d).
func StickerLetterCount(text string, letter rune) int {
	want := unicode.ToLower(letter)
	n := 0
	for _, r := range text {
		if unicode.ToLower(r) == want {
			n++
		}
	}
	return n
}

// StickerUniqueVowels is the number of different vowels on a name
// sticker's text, however many times each appears (CR 123.6e). The
// vowels are A, E, I, O, U and Y; case does not matter.
func StickerUniqueVowels(text string) int {
	seen := map[rune]bool{}
	for _, r := range text {
		switch l := unicode.ToLower(r); l {
		case 'a', 'e', 'i', 'o', 'u', 'y':
			seen[l] = true
		}
	}
	return len(seen)
}

// StickerBeginsWith reports whether a name sticker's first letter is
// `letter` (CR 123.6d: case does not matter).
func StickerBeginsWith(text string, letter rune) bool {
	for _, r := range text {
		if unicode.IsLetter(r) {
			return unicode.ToLower(r) == unicode.ToLower(letter)
		}
	}
	return false
}

// ---------------------------------------------------------------------
// The layer pass (CR 613.1c, 613.1f, 613.4b, 613.7k)

// stickerEffect is one sticker on one permanent as a continuous effect.
type stickerEffect struct {
	target    uuid.UUID
	sticker   PlacedSticker
	ability   StickerAbilityDef
	timestamp int64
}

func (e stickerEffect) Layer() (Layer, SubLayer) {
	switch e.sticker.Print.Kind {
	case StickerName:
		return Layer3Text, 0
	case StickerAbility:
		return Layer6Ability, 0
	}
	return Layer7PT, SubLayer7B_Set
}

func (e stickerEffect) Timestamp() int64 { return e.timestamp }

func (e stickerEffect) AppliesTo(target *Card, _ *Game) bool {
	if target == nil || target.InstanceID != e.target {
		return false
	}
	if e.sticker.Print.Kind == StickerPT {
		// CR 123.8: on a CREATURE. Layer 4 has settled what it is.
		return target.effective != nil && typeListHas(target.effective.Types, "Creature")
	}
	return true
}

func (e stickerEffect) Apply(c *Characteristic, target *Card, _ *Game) {
	switch e.sticker.Print.Kind {
	case StickerName:
		c.Name = InsertNameWord(c.Name, e.sticker.Print.Text, e.sticker.After)
	case StickerAbility:
		applyStickerAbility(c, e.ability, target)
	case StickerPT:
		c.Power = e.sticker.Print.Power
		c.Toughness = e.sticker.Print.Toughness
	}
}

func (e stickerEffect) RemovesAbilities() bool      { return false }
func (e stickerEffect) ContinuesAfterRemoval() bool { return false }

// applyStickerAbility writes an ability sticker's ability onto `c`: its
// keywords, and its bundle granted with `recipient` as "this creature".
func applyStickerAbility(c *Characteristic, ab StickerAbilityDef, recipient *Card) {
	for _, kw := range ab.Keywords {
		c.Abilities = AppendKeywordAbility(c.Abilities, kw)
	}
	if ab.Grant != "" && recipient != nil {
		c.GrantAbility(ab.Grant, recipient.InstanceID, recipient.Controller)
	}
}

// GiveAbilityOfSticker is "<recipient> has the ability of that ability
// sticker" (CR 123.7a): the ability the sticker grants to the object it is
// on, written onto the recipient's characteristic with the recipient as
// "this creature" and its controller as "you". It reads the sticker, not
// the object it is on, so it is the ability even when that object has lost
// its abilities. A sticker of another kind, or one the catalog does not
// run, gives nothing. For a layer-6 static's Apply (Pin Collection,
// Clandestine Chameleon).
func GiveAbilityOfSticker(c *Characteristic, s PlacedSticker, recipient *Card) {
	if ab, ok := stickerAbilityOf(s); ok {
		applyStickerAbility(c, ab, recipient)
	}
}

// stickerTimestamp is CR 613.7k: the later of the moment the sticker was
// put on and just after the object's own latest timestamp.
func stickerTimestamp(c *Card, s PlacedSticker) int64 {
	return max(s.PlacedAt, c.layerTimestamp()+1)
}

// stickerAbilityOf is the catalog's ability for an ability sticker, or
// nothing.
func stickerAbilityOf(s PlacedSticker) (StickerAbilityDef, bool) {
	if s.Print.Kind != StickerAbility || CatalogStickerAbility == nil {
		return StickerAbilityDef{}, false
	}
	return CatalogStickerAbility(s.Sheet, s.Print.Ordinal)
}

// stickerContinuousEffectsLocked is the layer pass's source list for
// stickers: one effect per name, ability and power/toughness sticker on
// each battlefield permanent. An art sticker does nothing (CR 123.9), and
// neither does an ability sticker the catalog does not run.
//
// Caller must hold g.mu in write mode (the layer pass does).
func (g *Game) stickerContinuousEffectsLocked() []ContinuousEffect {
	if g.Battlefield == nil {
		return nil
	}
	var out []ContinuousEffect
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if len(c.Stickers) == 0 {
			continue
		}
		for _, s := range c.Stickers {
			e := stickerEffect{target: c.InstanceID, sticker: s, timestamp: stickerTimestamp(c, s)}
			switch s.Print.Kind {
			case StickerArt:
				continue
			case StickerAbility:
				ab, ok := stickerAbilityOf(s)
				if !ok {
					continue
				}
				e.ability = ab
			}
			out = append(out, e)
		}
	}
	return out
}

// applyStickersOffBattlefield is the stickers' half of effectiveOf, for a
// card in a public zone other than the battlefield, where no layer pass
// runs (CR 123.6, 123.7, 123.8 all reach "a card in a zone other than the
// battlefield"): the name words, the ability stickers' keywords and the
// latest power and toughness sticker on a creature or Vehicle card, in
// the order they were put on. It never writes through a shared baseline:
// every changed list is a new slice.
func applyStickersOffBattlefield(c *Card, ch *Characteristic) {
	if c == nil || len(c.Stickers) == 0 {
		return
	}
	stickers := append([]PlacedSticker(nil), c.Stickers...)
	sort.SliceStable(stickers, func(i, j int) bool { return stickers[i].PlacedAt < stickers[j].PlacedAt })
	creatureOrVehicle := typeListHas(ch.Types, "Creature") || typeListHas(ch.Subtypes, "Vehicle")
	for _, s := range stickers {
		switch s.Print.Kind {
		case StickerName:
			ch.Name = InsertNameWord(ch.Name, s.Print.Text, s.After)
		case StickerAbility:
			ab, ok := stickerAbilityOf(s)
			if !ok {
				continue
			}
			for _, kw := range ab.Keywords {
				ch.Abilities = AppendKeywordAbility(ch.Abilities[:len(ch.Abilities):len(ch.Abilities)], kw)
			}
		case StickerPT:
			if creatureOrVehicle {
				ch.Power = s.Print.Power
				ch.Toughness = s.Print.Toughness
				ch.PTDefined = true
			}
		}
	}
}

// ---------------------------------------------------------------------
// Copies

func cloneStickerSheet(s StickerSheet) StickerSheet {
	s.Stickers = append([]StickerPrint(nil), s.Stickers...)
	return s
}

func cloneStickerSheets(in []StickerSheet) []StickerSheet {
	if len(in) == 0 {
		return nil
	}
	out := make([]StickerSheet, len(in))
	for i, s := range in {
		out[i] = cloneStickerSheet(s)
	}
	return out
}

// StickerSheetsOf is a copy of the sheets `p` has access to this game.
func StickerSheetsOf(p *Player) []StickerSheet {
	if p == nil {
		return nil
	}
	return cloneStickerSheets(p.StickerSheets)
}

// ---------------------------------------------------------------------
// "You may cast this card from your graveyard" (CR 123.7, #4307)

// AltCostKeyStickerGraveyard is the offer the derived permission below
// is claimed under: the card's mana cost and N life.
const AltCostKeyStickerGraveyard = "sticker-graveyard"

// stickerGraveyardLifeOf is N in "by paying N life" for a card with an
// ability sticker that lets it be cast from the graveyard, or 0.
func stickerGraveyardLifeOf(c Card) int {
	for _, s := range c.Stickers {
		if ab, ok := stickerAbilityOf(s); ok && ab.GraveyardCastLife > 0 {
			return ab.GraveyardCastLife
		}
	}
	return 0
}

// stickerGraveyardPermissionLocked is the cast permission an ability
// sticker gives its card while the card is in its owner's graveyard:
// "You may cast this card from your graveyard by paying 2 life in addition
// to paying its other costs." It is derived on every query, never stored,
// so it lasts exactly as long as the sticker is on the card (a sticker
// stays on a card that goes to a graveyard, CR 123.5, and comes off one
// that goes to a hand or a library).
//
// The extra life is the offer's: the cast claims AltCostKeyStickerGraveyard,
// which charges the card's printed mana cost and the life
// (CastPermission.AlternativeCostFor). Timing, targets and every other cost
// are the spell's own; a land has no cast to permit.
//
// Caller must hold g.mu.
func (g *Game) stickerGraveyardPermissionLocked(playerID uuid.UUID, card Card) *CastPermission {
	if len(card.Stickers) == 0 || card.Owner != playerID || card.IsLand() || stickerGraveyardLifeOf(card) == 0 {
		return nil
	}
	return &CastPermission{
		Player:     playerID,
		Zone:       ZoneGraveyard,
		Scope:      ScopeCards,
		Cards:      []PermissionCardRef{{ID: card.InstanceID, Epoch: card.ObjectEpoch}},
		Duration:   IndefiniteDuration(),
		CastOnly:   true,
		AltCostKey: AltCostKeyStickerGraveyard,
		Source:     card.InstanceID,
		SourceName: card.Name,
		Label:      card.Name + " — cast from your graveyard, paying " + strconv.Itoa(stickerGraveyardLifeOf(card)) + " life",
	}
}

// anyStickerGraveyardCastLocked is AnyCastPermissionsForEffect's fast
// positive for the derived permission above: a graveyard holds a card with
// a sticker that gives it.
func (g *Game) anyStickerGraveyardCastLocked() bool {
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for i := range p.Graveyard.Cards {
			c := &p.Graveyard.Cards[i]
			if len(c.Stickers) > 0 && stickerGraveyardLifeOf(*c) > 0 {
				return true
			}
		}
	}
	return false
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_return_together.go — #1867. Every card that returns more
// than one card from a graveyard to the battlefield does it through
// ReturnFromGraveyardTogether, one simultaneous entry, rather than a
// loop over ReturnFromGraveyard. #1872 did the same for exile:
// ReturnFromExileTogether, over the same body. Append-only: the
// helpers below are the shapes those cards share (the clone gate's
// rule).

// ReturnFromGraveyardTogether puts several cards from graveyards onto
// the battlefield as ONE simultaneous entry (#1867) — Reveillark's "up
// to two target creature cards", Replenish's "all enchantment cards",
// Exhume's one card per player. Every card's CR 614 window runs against
// the board as it was before any of them entered (CR 614.12), all of
// them land, and only then is any announced, so each newcomer sees the
// others enter (CR 603.6a). A loop over ReturnFromGraveyard is not the
// same thing: it announces the first card before the second moves, so
// "whenever another creature enters" on the second never sees the
// first.
//
// Cards that are no longer in a graveyard (a target exiled in
// response, CR 608.2b), tokens and nonpermanent cards are skipped
// rather than refusing the whole entry, which is what the engine's
// batch door does with them.
//
// Controller is ReturnFromGraveyard's: zero is "under its owner's
// control", and ctx.Controller() is "under your control". Tapped is
// the effect's own "to the battlefield tapped", riding each card's
// entry event.
//
// ThenTail, when set, is the rest of the effect, and it runs exactly once
// with the IDs that entered (a graveyard card keeps its ID) — after
// the last entry question has been answered when one was asked, so it
// must not be written on the line after Apply. It is a registered tail
// (TogetherTail; #4253 slice 6): a key plus plain data, so a table waiting
// on an entering card's question is a restore point.
type ReturnFromGraveyardTogether struct {
	Targets    []uuid.UUID
	Controller uuid.UUID
	Tapped     bool
	// AttachedTo is "… to the battlefield attached to <host>" (Mantle of
	// the Ancients, #3158): every card enters attached to it. Zero for a
	// return that names no host, whose Auras choose theirs as they enter
	// (CR 303.4f).
	AttachedTo game.TargetRef

	// The resolving item rides the tail as data (its targets, its modes,
	// the words of its text), and ThenN is a number it reads.
	ThenTail game.TailRef
	ThenN    int
}

func (r ReturnFromGraveyardTogether) Apply(ctx *Context) error {
	return putCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, r.Targets,
		game.ZoneEntryOptions{Controller: r.Controller, Tapped: r.Tapped, AttachedTo: r.AttachedTo}, r.ThenTail, r.ThenN)
}

// TogetherTail registers the rest of a ReturnFromGraveyardTogether or
// ReturnFromExileTogether as a keyed tail (#4253 slice 6). The body runs
// against the resolving item's copy (the asking item rides the tail), told
// the IDs that entered (a graveyard card keeps its ID, an exiled one has a
// new one) and the tail's N. The key is an on-disk identity (ledgered
// "tail <key>"): never renamed, never reused. Call it once, from a
// package-level var.
func TogetherTail(key string, body func(ctx *Context, entered []uuid.UUID, n int) error) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		return body(ctx, o.IDs, o.Tail.N)
	})
}

// legalTargetCardIDs is the spell's or ability's still-legal card
// targets (CR 608.2b), in announce order — the list a "return up to N
// target cards" hands to ReturnFromGraveyardTogether.
func legalTargetCardIDs(ctx *Context) []uuid.UUID {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	return ids
}

// graveyardCardIDs is the cards in `player`'s graveyard that match, read
// once before anything moves — "return all land cards from your
// graveyard". Nil when the player or graveyard is gone.
func graveyardCardIDs(ctx *Context, player uuid.UUID, match func(game.Card) bool) []uuid.UUID {
	p := ctx.PlayerByID(player)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if match(c) {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// allGraveyardsCardIDs is graveyardCardIDs over every seat, in seat
// order — "all creature cards from all graveyards".
func allGraveyardsCardIDs(ctx *Context, match func(game.Card) bool) []uuid.UUID {
	var ids []uuid.UUID
	for _, p := range ctx.Game.Seats {
		if p == nil {
			continue
		}
		ids = append(ids, graveyardCardIDs(ctx, p.ID, match)...)
	}
	return ids
}

// ReturnFromExileTogether is ReturnFromGraveyardTogether for cards in
// exile (#1872): Living Death's "puts all cards they exiled this way
// onto the battlefield", a delayed blink's "return those cards", an
// Oblivion Ring's return of everything it exiled. All of them land
// before any is announced, so each newcomer sees the others enter
// (CR 603.6a). A loop over ReturnFromExile announces the first card
// before the second moves.
//
// Each card returns as a NEW OBJECT (CR 400.7), as ReturnFromExile's
// does, so the IDs ThenTail is told are the new ones, not the ones in
// Targets.
//
// A card no longer in exile is skipped: a delayed return does not
// follow a card that has moved on (CR 603.7c). So is a token, which
// cannot come back to the battlefield (CR 111.8), and a nonpermanent
// card. Controller, Tapped and ThenTail mean what they mean on
// ReturnFromGraveyardTogether: zero Controller is "under its owner's
// control" (CR 610.3c for an "until" return).
//
// CountersIfType is the return's own "each of them enters with an
// additional <counter> on it if it's a <type>" (Semester's End, #4193):
// each card is judged by what it would be on the battlefield
// (CR 614.12), and the return stays one event.
type ReturnFromExileTogether struct {
	Targets        []uuid.UUID
	Controller     uuid.UUID
	Tapped         bool
	CountersIfType []game.EntryCountersIfType
	// ThenTail and ThenN are ReturnFromGraveyardTogether's.
	ThenTail game.TailRef
	ThenN    int
}

func (r ReturnFromExileTogether) Apply(ctx *Context) error {
	return putCardsFromZoneTogetherWith(ctx, game.ZoneExile, r.Targets, game.ZoneEntryOptions{
		Controller:               r.Controller,
		Tapped:                   r.Tapped,
		EntersWithCountersIfType: r.CountersIfType,
	}, r.ThenTail, r.ThenN)
}

// putCardsFromZoneTogether is the body of ReturnFromGraveyardTogether
// and ReturnFromExileTogether: keep the targets still in a zone of kind
// `from` that can enter (not a token, a permanent card), and put them
// onto the battlefield as one entry through the engine's batch door.
func putCardsFromZoneTogether(ctx *Context, from game.ZoneKind, targets []uuid.UUID, controller uuid.UUID, tapped bool) error {
	return putCardsFromZoneTogetherWith(ctx, from, targets, game.ZoneEntryOptions{Controller: controller, Tapped: tapped}, game.TailRef{}, 0)
}

// putCardsFromZoneTogetherWith is putCardsFromZoneTogether with the whole
// set of entry options.
func putCardsFromZoneTogetherWith(ctx *Context, from game.ZoneKind, targets []uuid.UUID, opts game.ZoneEntryOptions,
	tail game.TailRef, tailN int) error {
	entries := make([]game.BatchEntry, 0, len(targets))
	for _, id := range targets {
		z := ctx.Game.FindCardZoneForEffect(id)
		if z == nil || z.Kind != from {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsToken() || !c.IsPermanent() {
			continue
		}
		entries = append(entries, game.BatchEntry{CardID: id, From: from})
	}
	var then game.KeyedTail
	if tail.Key() != "" {
		then = withN(ctx.ItemTail(tail), tailN)
	}
	return ctx.Game.PutOntoBattlefieldTogetherKeyedForEffect(entries, opts, then)
}

// exiledCardIDs is the cards in exile that match, read once before
// anything moves — Oblivion Sower's "all land cards that player owns
// from exile".
func exiledCardIDs(ctx *Context, match func(game.Card) bool) []uuid.UUID {
	if ctx.Game.Exile == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range ctx.Game.Exile.Cards {
		if match(c) {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// ReturnChosenFromGraveyard is "return N <kind> cards from your
// graveyard to the battlefield" when the cards are CHOSEN as the effect
// resolves rather than targeted at cast (#2863, ADR 0013 amendment of
// 2026-10-09): Rise of the Witch-king's "you may return another
// permanent card", Lich-Knights' Conquest's "return that many creature
// cards".
//
// The candidates are read from the graveyard as it is WHEN THIS RUNS,
// so a card that arrived earlier in the same resolution is on offer.
// That is the point of it: put it in the continuation of whatever came
// first (a sacrifice run's `then`, a SacrificeAllKeyedForEffect), and
// the creature just sacrificed is one of the cards that can come back.
// It is not a target, so nothing is announced at cast and opponents
// have no pick to respond to.
//
// One choose_cards prompt to Player (zero is the resolving effect's
// controller) over their own graveyard's permanent cards that Match
// (nil matches every one), bounded by Min and Max. Max <= 0 is "all of
// them", and the engine clamps Min to what is on offer, so a mandatory
// "that many" with too few cards takes them all (CR 608.2). No
// candidate, no prompt. Except leaves one card out of the offer:
// "another" card. The chosen cards enter together (#1867) under their
// owner's control, tapped when Tapped is set, and the Zone re-check on
// submit refuses a card that left the graveyard while the question was
// open.
//
// ToHand (#2941) sends the chosen cards to their owner's hand instead:
// "mill three, then you may return a creature or land card from your
// graveyard to your hand" (Grapple with the Past). The offer is then
// every non-token card Match accepts, permanent or not, since a hand
// return has no battlefield to be a permanent for. Each card goes
// through ReturnFromGraveyard, the exit the targeted spells use.
type ReturnChosenFromGraveyard struct {
	Player   uuid.UUID
	Question string
	Match    func(game.Card) bool
	Except   uuid.UUID
	Min, Max int
	Tapped   bool
	ToHand   bool
}

func (r ReturnChosenFromGraveyard) Apply(ctx *Context) error {
	player := r.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	var candidates []uuid.UUID
	for _, id := range graveyardCardIDs(ctx, player, func(c game.Card) bool {
		return (r.ToHand || c.IsPermanent()) && !c.IsToken() && (r.Match == nil || r.Match(c))
	}) {
		if id != r.Except {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	dest := game.ChooseOntoBattlefield
	if r.Tapped {
		dest = game.ChooseOntoBattlefieldTapped
	}
	toHand := r.ToHand
	if toHand {
		dest = ""
	}
	// The pick is keyed (#4253 slice 8): the rest of the card is a registered
	// continuation, and the two flags it needs are numbers on the prompt.
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:     player,
		FromPlayer:  player,
		Source:      ctx.Source(),
		Question:    r.Question,
		Cards:       candidates,
		Min:         r.Min,
		Max:         r.Max,
		Destination: dest,
		ThenKey:     returnChosenFromGraveyardThen,
		LiveZone:    game.ZoneGraveyard,
		CarryNums:   []int{boolToInt(r.Tapped), boolToInt(toHand)},
		Asked:       game.NewAskedItem(ctx.Item),
	})
	return nil
}

// returnChosenFromGraveyardThen is ReturnChosenFromGraveyard's answer:
// the chosen cards come back together, or go to hand one by one. Its
// numbers are Tapped and ToHand. A card that left the graveyard while the
// question was open is skipped by the entry itself, and a prompt whose
// candidates all left is withdrawn with nothing picked.
var returnChosenFromGraveyardThen = game.RegisterChooseCardsThen("choose-cards/return-chosen-from-graveyard",
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) == 0 {
			return nil
		}
		ctx, _ := askedContext(g, r.Asked, r.Chooser, r.Source, nil)
		tapped, toHand := len(r.Nums) > 0 && r.Nums[0] != 0, len(r.Nums) > 1 && r.Nums[1] != 0
		if toHand {
			for _, id := range r.Picked {
				if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		}
		return ReturnFromGraveyardTogether{Targets: r.Picked, Tapped: tapped}.Apply(ctx)
	})

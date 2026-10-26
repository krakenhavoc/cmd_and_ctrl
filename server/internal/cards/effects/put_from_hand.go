package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// put_from_hand.go — "put a [land / Equipment / permanent] card from
// your hand onto the battlefield" (#654).
//
// The clause is two halves and the catalog owns one of them. The
// MOVE is engine: game.PutFromHandOntoBattlefieldForEffect runs the
// CR 614 entry pipeline, moves the card and fires the ETB hook, the
// same way the search, reanimation and exile-return paths do. The
// PICK is card text — which cards qualify, whether it is "you may",
// whether the permanent arrives tapped — and it lives here, on top of
// the pick-from-hand prompt with a continuation that shipped in #552
// (game.ChooseCardsPrompt with Zone: ZoneHand and Then).
//
// Five catalog cards print the same sentence with different framing
// (Eureka Moment, Broken Bond, Chulane, Insidious Fungus's third
// mode, Spelunking), which is why the shape is one primitive plus two
// named constructors rather than five copies of a prompt literal.

// PutFromHandOntoBattlefield asks `Player` to pick one card from
// their hand and puts it onto the battlefield without casting or
// playing it.
//
// Used as a primitive value — Do(PutFromHandOntoBattlefield{…}) in a
// trigger, or applied directly inside an OnResolve closure. Player
// defaults to the resolving item's controller like every other
// primitive in the catalog.
//
// # It queues; it does not finish
//
// Apply returns as soon as the prompt is QUEUED. The next step of a Do
// that is the item's own Effect waits for the answer (#4012,
// do_walk.go), but it cannot read the answer, and a Do run from inside a
// hand-written closure does not wait at all. So a rider that depends on
// what arrived ("if you put a Cave onto the battlefield this way, you
// gain 4 life") goes in Then, not in the next slot of the Do.
//
// A player with no matching card in hand is never prompted — there is
// nothing to ask about, and QueueChooseCardsForEffect deliberately
// does not short-circuit an empty candidate set (its doc says why).
// Then still runs in that case, with a nil permanent, so a card whose
// rider is unconditional stays correct.
type PutFromHandOntoBattlefield struct {
	// Player picks and, unless a printed card ever says otherwise,
	// is who the permanent enters under. uuid.Nil means the
	// resolving item's controller.
	Player uuid.UUID

	// Match filters the hand to the cards the clause names — Land()
	// for the land-drop family, OfSubtype("Equipment") for
	// Stoneforge Mystic. Nil means every permanent card in hand.
	//
	// Nonpermanent cards are dropped whatever Match says: an instant
	// cannot become a permanent (CR 110.4), so offering one would be
	// offering a pick the engine has to refuse.
	Match CardPredicate

	// Tapped is the clause's own "onto the battlefield tapped"
	// (Insidious Fungus, Arboreal Grazer). It rides the entry event
	// rather than tapping afterwards, so the permanent ARRIVES
	// tapped and nothing that watches for a tap sees one.
	Tapped bool

	// Optional is the printed "you MAY put". It sets the prompt's
	// floor to zero, so declining is a real answer; a mandatory
	// clause ("put an Equipment card from your hand onto the
	// battlefield") leaves it false.
	Optional bool

	// Max is how many cards the clause puts: 0 and 1 are "a card" (the
	// default), N is "up to N" with Optional and exactly N without
	// ("you may put a permanent card from your hand onto the
	// battlefield FOR EACH free vote", Selvala's Stampede). More than
	// one card enters as one simultaneous batch, so every CR 614
	// window sees the board as it stood before any of them arrived and
	// every ETB fires after all have landed. The ceiling is the number
	// of matching cards in hand, so a player holding fewer than N is
	// asked about the cards they have.
	Max int

	// Label is the prompt's header, "<card> — <the printed clause>".
	// It is what the player and the bot's move list read, so write
	// the card's sentence rather than a description of the mechanic.
	Label string

	// Validate is a rule on the picked SET, for a clause no per-card
	// Match can say: Yuna's Decision's "a creature card and/or a land
	// card" (at most one of each). Nil accepts any set within Max.
	Validate func(picked []game.Card) bool

	// ThenTail, when set, runs after the answer, as a registered tail
	// (PutFromHandTail; #4253 slice 6): a key plus plain data, so a table
	// waiting on a put card's entry question is a restore point. It is
	// the only correct place for a rider on the result ("if you put a
	// Cave onto the battlefield this way…"), because everything
	// sequenced after this primitive has already run by the time it is
	// called. It runs with g.mu held and may queue further prompts.
	// ThenN is a number the tail reads and ThenItem carries the
	// resolving item as data.
	ThenTail game.TailRef
	ThenN    int
	ThenItem bool
}

// PutFromHandResult is what Then is told about the put: everything
// the continuation would otherwise have to capture, handed to it as
// scalars so it can be a package-level function rather than a closure
// over the resolving item.
type PutFromHandResult struct {
	// Source is the card whose clause this was — what a life gain or
	// a damage rider is attributed to.
	Source uuid.UUID

	// Player is who was asked.
	Player uuid.UUID

	// Entered is the permanent that arrived, or uuid.Nil when
	// nothing did: a declined "you may", no matching card in hand,
	// or a replacement that canceled the entry. With Max above one it
	// is the first of EnteredAll.
	Entered uuid.UUID

	// EnteredAll is every permanent that arrived, in the order
	// picked: one entry for a single put, up to Max for a batch.
	EnteredAll []uuid.UUID

	// N is the number the keyed tail carried (PutFromHandOntoBattlefield.ThenN).
	N int
}

func (p PutFromHandOntoBattlefield) Apply(ctx *Context) error {
	player := p.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	source := ctx.Source()
	var keyed game.KeyedTail
	if p.ThenTail.Key() != "" {
		keyed = game.KeyedTail{Ref: p.ThenTail, Player: player, Source: source, N: p.ThenN}
		if p.ThenItem {
			keyed.Asked = game.NewAskedItem(ctx.Item)
		}
	}
	candidates := handCardsMatching(ctx.Game, player, p.Match)
	if len(candidates) == 0 {
		return keyed.Run(ctx.Game, nil)
	}
	max := p.Max
	if max < 1 {
		max = 1
	}
	if max > len(candidates) {
		max = len(candidates)
	}
	floor := max
	if p.Optional {
		floor = 0
	}
	label := p.Label
	if label == "" {
		label = "Put a card from your hand onto the battlefield"
	}
	tapped := p.Tapped
	// #2680: say where the pick goes, so a policy does not price
	// putting a land onto the battlefield as giving it up.
	dest := game.ChooseOntoBattlefield
	if tapped {
		dest = game.ChooseOntoBattlefieldTapped
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Destination: dest,
		Chooser:     player,
		Source:      source,
		Question:    label,
		Cards:       candidates,
		Min:         floor,
		Max:         max,
		Validate:    p.Validate,
		// Re-checked on submit: the pick must still be in that
		// player's hand when the answer arrives.
		Zone: game.ZoneHand,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return keyed.Run(g, nil)
			}
			if len(picked) > 1 {
				// One simultaneous entry (entry_batch.go), told what
				// entered when the last card's window has settled.
				return g.PutCardsFromHandOntoBattlefieldKeyedForEffect(picked, game.HandEntryOptions{
					Controller: player,
					Tapped:     tapped,
				}, keyed)
			}
			// Through the keyed door (#1322, #4253 slice 6): the land may
			// stop to ask its own question (a shockland's life), and a
			// rider on the result (Spelunking's Cave) has to wait for the
			// answer rather than read an entry that has not happened.
			return g.PutFromHandOntoBattlefieldKeyedForEffect(picked[0], game.HandEntryOptions{
				Controller: player,
				Tapped:     tapped,
			}, keyed)
		},
	})
	return nil
}

// PutFromHandTail registers a PutFromHandOntoBattlefield rider as a keyed
// tail (#4253 slice 6): the body is handed what the rider is told, with a
// Context rebuilt from the tail. The key is an on-disk identity
// (ledgered "tail <key>"): never renamed, never reused. Call it once,
// from a package-level var.
func PutFromHandTail(key string, body func(ctx *Context, res PutFromHandResult) error) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		res := PutFromHandResult{Source: o.Tail.Source, Player: o.Tail.Player, EnteredAll: o.IDs, N: o.Tail.N}
		if len(o.IDs) > 0 {
			res.Entered = o.IDs[0]
		}
		return body(ctx, res)
	})
}

// MayPutALandFromHand is the printed clause five catalog cards share
// word for word — "you may put a land card from your hand onto the
// battlefield" (Eureka Moment, Broken Bond, Chulane, Spelunking, and
// Growth Spiral when it is written).
//
// `cardName` is the card's own name, which is what makes the prompt
// header say who is asking on a board with two of these in play.
func MayPutALandFromHand(cardName string) PutFromHandOntoBattlefield {
	return PutFromHandOntoBattlefield{
		Match:    Land(),
		Optional: true,
		Label:    cardName + " — you may put a land card from your hand onto the battlefield",
	}
}

// MayPutALandFromHandTapped is the same clause with the printed
// "tapped" rider (Insidious Fungus's third mode, Arboreal Grazer).
func MayPutALandFromHandTapped(cardName string) PutFromHandOntoBattlefield {
	p := MayPutALandFromHand(cardName)
	p.Tapped = true
	p.Label = cardName + " — you may put a land card from your hand onto the battlefield tapped"
	return p
}

// handCardsMatching returns the cards in a player's hand that pass
// `pred`, in hand order, skipping anything that could not be a
// permanent (CR 110.4) and any token (CR 108.2, CR 111.8) — the same
// "permanent card" test libraryCardsMatching applies.
//
// A nil predicate means "every permanent card". Characteristics are
// read off the card in hand, which is where they are printed — a
// layer effect on the battlefield has nothing to say about it.
//
// Caller holds g.mu.
func handCardsMatching(g *game.Game, playerID uuid.UUID, pred CardPredicate) []uuid.UUID {
	p := g.PlayerByIDForEffect(playerID)
	if p == nil || p.Hand == nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range p.Hand.Cards {
		if !isPermanentCard(c) {
			continue
		}
		if pred != nil && !pred(g, playerID, c) {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

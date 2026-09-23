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
// Apply returns as soon as the prompt is QUEUED. Anything sequenced
// after it in the same Do(…) runs BEFORE the player answers, which is
// the property that matters when writing a card: a rider that must
// happen after the land arrives ("if you put a Cave onto the
// battlefield this way, you gain 4 life") goes in Then, not in the
// next slot of the Do.
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

	// Label is the prompt's header, "<card> — <the printed clause>".
	// It is what the player and the bot's move list read, so write
	// the card's sentence rather than a description of the mechanic.
	Label string

	// Then, when set, runs after the answer. It is the only correct
	// place for a rider on the result ("if you put a Cave onto the
	// battlefield this way…"), because everything sequenced after
	// this primitive has already run by the time it is called.
	//
	// Same contract as any other continuation: it runs with g.mu
	// held, may queue further prompts, and must capture only scalars
	// — never a *Card or a pointer into a zone.
	Then func(g *game.Game, res PutFromHandResult) error
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
	// or a replacement that canceled the entry.
	Entered uuid.UUID
}

func (p PutFromHandOntoBattlefield) Apply(ctx *Context) error {
	player := p.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	then := p.Then
	source := ctx.Source()
	candidates := handCardsMatching(ctx.Game, player, p.Match)
	if len(candidates) == 0 {
		if then != nil {
			return then(ctx.Game, PutFromHandResult{Source: source, Player: player})
		}
		return nil
	}
	floor := 1
	if p.Optional {
		floor = 0
	}
	label := p.Label
	if label == "" {
		label = "Put a card from your hand onto the battlefield"
	}
	tapped := p.Tapped
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   source,
		Question: label,
		Cards:    candidates,
		Min:      floor,
		Max:      1,
		// Re-checked on submit: the pick must still be in that
		// player's hand when the answer arrives.
		Zone: game.ZoneHand,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			report := func(g *game.Game, entered uuid.UUID) error {
				if then != nil {
					return then(g, PutFromHandResult{Source: source, Player: player, Entered: entered})
				}
				return nil
			}
			if len(picked) == 0 {
				return report(g, uuid.Nil)
			}
			// Through the Then door (#1322): the land may stop to ask
			// its own question (a shockland's life), and a rider on
			// the result (Spelunking's Cave) has to wait for the
			// answer rather than read an entry that has not happened.
			return g.PutFromHandOntoBattlefieldThenForEffect(picked[0], game.HandEntryOptions{
				Controller: player,
				Tapped:     tapped,
			}, report)
		},
	})
	return nil
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

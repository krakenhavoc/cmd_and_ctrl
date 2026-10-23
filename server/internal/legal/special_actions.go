package legal

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// specialActionParams is the special_action wire payload this package
// emits. Strict + auto_tap for the same reason every cast is: the
// enumerator decides affordability here, so the engine never has to
// reject a move a bot was offered (#544).
type specialActionParams struct {
	CardID  string `json:"card_id"`
	Kind    string `json:"kind"`
	Strict  bool   `json:"strict,omitempty"`
	AutoTap bool   `json:"auto_tap,omitempty"`
	// Cost picks between two offers of one kind on one card (#1391).
	Cost string `json:"cost,omitempty"`
	// Door is the door an unlock unlocks (ADR 0103).
	Door string `json:"door,omitempty"`
	// SacrificeID is the permanent an ignore_effect pays (CR 116.2d,
	// #1986). The enumerator names the first of the payment order, the
	// one a "choose for me" would give up.
	SacrificeID string `json:"sacrifice_id,omitempty"`
	// Route names which turn-face-up offer this is, sent only when the
	// card offers turn_face_up more than once (ADR 0189 decision 4).
	Route string `json:"route,omitempty"`
}

// specialActionMoves enumerates the CR 116.2 special actions the seat
// may take on cards in its own hand — foretell (CR 702.143a) and
// suspend (CR 702.62a).
//
// IT DOES NOT COPY THE SPLIT-SECOND EARLY RETURN that opens
// castMoves and activatedMoves, and that is the whole reason ADR 0062
// Decision 4 wrote this method down before anything implemented it.
// CR 702.61b stops casts and activations of non-mana abilities; a
// special action is neither, so foretelling under a Trickbind is
// legal. Suspend is barred under split second — but by its own
// wording (CR 702.62c imports every restriction on beginning to cast
// the card), which the engine's per-kind timing table says and this
// loop asks rather than assumes.
//
// One question per (card, kind): the engine's own
// SpecialActionTimingOKLocked, so the enumerator and the engine
// cannot drift apart the way two copies of a timing rule always do.
func (e *enumerator) specialActionMoves() {
	g, p := e.g, e.p
	if p == nil {
		return
	}
	e.outsidePlayMove()
	e.ignoreEffectMoves()
	if p.Hand != nil {
		for i := range p.Hand.Cards {
			e.specialActionMovesForCard(p.Hand.Cards[i], game.ZoneHand)
		}
	}
	// #1391: the top of the seat's own library, the one card a granted
	// zone reaches (Fblthp, Lost on the Range's plot). Without a grant
	// SpecialActionsOfferedLocked answers nothing here, so this costs
	// one lookup per seat in every game without a Fblthp.
	if p.Library != nil && len(p.Library.Cards) > 0 {
		e.specialActionMovesForCard(p.Library.Cards[len(p.Library.Cards)-1], game.ZoneLibrary)
	}
	// ADR 0172 §7: the seat's companion, outside the game. The same
	// walk as the hand, asked about ZoneOutside, so the engine's own
	// timing table and pricing decide the move (#544).
	if c, ok := p.CompanionCard(); ok {
		e.specialActionMovesForCard(c, game.ZoneOutside)
	}
	// ADR 0082 decision 6: turn_face_up is the first kind whose card
	// is not in a hand. The seat's own FACE-DOWN permanents are the
	// only battlefield cards that offer anything —
	// SpecialActionsOfferedByCard derives the row from the face-down
	// kind and answers nothing for a face-up permanent — so the walk
	// filters on the one predicate rather than pricing every
	// permanent on the board.
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			card := g.Battlefield.Cards[i]
			// ADR 0103: and a face-up Room offers an unlock per
			// locked door (CR 709.5e).
			if !(card.FaceDownIsPermanent() || game.HasSharedTypeLine(card)) || card.Controller != e.seat {
				continue
			}
			e.specialActionMovesForCard(card, game.ZoneBattlefield)
		}
	}
}

// specialActionMovesForCard emits one move per special action `card`
// offers in `zone` that the seat may take and afford right now.
func (e *enumerator) specialActionMovesForCard(card game.Card, zone game.ZoneKind) {
	g := e.g
	// ADR 0122 §6.2: the scope a Source filter asks about.
	if !e.enter(scope{source: card.InstanceID}) {
		return
	}
	offers := g.SpecialActionsOfferedLocked(e.seat, card, zone)
	for _, sa := range offers {
		if !game.SpecialActionKindBuilt(sa.Kind) {
			continue
		}
		if !g.SpecialActionTimingOKLocked(e.seat, card, sa.Kind) {
			continue
		}
		// The cost is mana and nothing else, and it is the only
		// way a special action can be unaffordable — there is no
		// target to be missing and no choice to be unanswerable.
		// Priced through the same CR 601.2f pass the engine charges
		// with (#1319: SpecialActionManaCostForEffect), so a discount
		// like Ranar's "first foretell each turn costs {0}" is never
		// offered as an ordinary {2} the engine then charges less
		// for, or the reverse. Priced with the same zero spend
		// context the engine pays with, so restricted mana counts
		// here exactly as little as it does there.
		if sa.Cost != "" {
			cost, err := g.SpecialActionManaCostForEffect(e.seat, card, sa.Kind, sa)
			if err != nil {
				continue
			}
			if !e.canPay(cost, 0, game.SpecialActionSpendContext(sa.Kind)) {
				continue
			}
		}
		label := sa.Label
		if label == "" {
			label = string(sa.Kind)
		}
		// A face-down permanent has NO NAME (CR 708.2), so the row
		// would read "Turn face up {1}{U} " with a dangling space.
		// The label alone is the whole of what the seat is choosing
		// between.
		switch {
		case sa.Kind == game.SpecialActionCompanion:
			// ADR 0172 §7: one sentence a seat can choose by.
			label = "Pay " + sa.Cost + ": put " + card.Name + " into your hand (companion)"
		case card.Name != "" && sa.Kind != game.SpecialActionUnlock:
			label += " " + card.Name
		}
		params := specialActionParams{
			CardID:  card.InstanceID.String(),
			Kind:    string(sa.Kind),
			Strict:  true,
			AutoTap: true,
			Door:    game.DoorName(sa.Door),
		}
		// #1391: an offer a grant opened in another zone names its
		// price, because the card may offer the same kind twice
		// there (a Djinn of Fool's Fall under Fblthp: its own plot
		// cost and its mana cost). A card in its own zone offers each
		// kind once, and its params stay exactly as they were.
		if sa.Zone != "" {
			params.Cost = sa.Cost
		}
		// ADR 0189: a manifested or cloaked card whose card also has
		// morph or disguise offers turn_face_up twice, perhaps at the
		// same price, so each move names its route.
		if sa.Kind == game.SpecialActionTurnFaceUp && sa.Route != "" && turnFaceUpOffers(offers) > 1 {
			params.Route = string(sa.Route)
		}
		if sa.Zone == game.ZoneLibrary {
			label += " from the top of your library"
		}
		e.add(Move{
			Type:   TypeSpecialAction,
			Player: e.seat,
			Kind:   KindSpecialAction,
			Label:  label,
			Source: card.InstanceID,
			Params: mustJSON(params),
		})
	}
}

// ignoreEffectMoves offers CR 116.2d's special action for every static
// ability that bears on the seat and may be ignored (Damping Engine,
// #1986): one move per source, paying with the first permanent of the
// payment order. The engine's own accessor decides what is on offer
// (IgnoreEffectOffersLocked), so this cannot disagree with it; the action
// has no mana cost and no timing beyond priority, so there is nothing
// else to ask. Under a split-second spell the move is still offered: a
// special action is neither a cast nor an activation (CR 702.61b).
func (e *enumerator) ignoreEffectMoves() {
	for _, offer := range e.g.IgnoreEffectOffersLocked(e.seat) {
		if !e.enter(scope{source: offer.Source}) {
			continue
		}
		e.add(Move{
			Type:   TypeSpecialAction,
			Player: e.seat,
			Kind:   KindSpecialAction,
			Label:  offer.Label,
			Source: offer.Source,
			Params: mustJSON(specialActionParams{
				CardID:      offer.Source.String(),
				Kind:        string(game.SpecialActionIgnoreEffect),
				SacrificeID: offer.Candidates[0].String(),
			}),
		})
	}
}

// outsidePlayMove offers the Wish right's action (#3955): name a card from
// the card database to play this turn. Offered while the seat holds the
// right (game.CanNameOutsideCardLocked) and its container is empty. The
// engine accepts the action whenever the right is live; the empty
// container is policy for an automated seat, which would otherwise name
// a fresh card at every priority it holds a card it cannot yet cast (a
// sorcery on another player's turn). A human has the button regardless.
func (e *enumerator) outsidePlayMove() {
	g, p := e.g, e.p
	if !g.CanNameOutsideCardLocked(e.seat) || (p.Sideboard != nil && p.Sideboard.Size() > 0) {
		return
	}
	right := g.OutsidePlayRightLocked(e.seat)
	if right == nil {
		return
	}
	e.add(Move{
		Type:   TypeOutsidePlay,
		Player: e.seat,
		Kind:   KindSpecialAction,
		Label:  right.SourceName + ": name a card you own from outside the game to play this turn",
		Params: mustJSON(struct {
			Kind string `json:"kind"`
		}{Kind: "outside_play"}),
	})
}

// turnFaceUpOffers counts the turn_face_up offers in a card's list.
func turnFaceUpOffers(offers []game.SpecialAction) int {
	n := 0
	for _, sa := range offers {
		if sa.Kind == game.SpecialActionTurnFaceUp {
			n++
		}
	}
	return n
}

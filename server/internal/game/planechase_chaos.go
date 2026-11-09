package game

import (
	"github.com/google/uuid"
)

// planechase_chaos.go is "chaos ensues" as an effect, and the planar deck
// walks the planes that reveal from it need (ADR 0216 slice 4, #4870):
// Norn's Seedcore, The Fertile Lands of Saulvinia. CR 311.7:
//
//	"Each plane card has a triggered ability that triggers “Whenever chaos
//	 ensues.” ... This ability triggers if the chaos symbol is rolled on the
//	 planar die (see rule 901.9b), if a resolving spell or ability says that
//	 chaos ensues, or if a resolving spell or ability states that chaos
//	 ensues for a particular object. In the last case, the chaos ability can
//	 trigger even if that plane card is still in the planar deck but
//	 revealed. A chaos ability is controlled by the current planar
//	 controller."
//
// An effect that says "chaos ensues" emits EventChaosEnsues with no card, and
// every face-up plane's chaos ability triggers; "chaos ensues on that plane"
// emits it naming the card, and that plane's chaos ability triggers, from
// the planar deck when it is still there (harvestFromPlanesLocked).

// EventChaosEnsues — a resolving spell or ability said chaos ensues (CR
// 311.7). Actor is the player it ensued for; CardID is the one plane card it
// ensued on, or uuid.Nil for "chaos ensues" (every face-up plane).
const EventChaosEnsues EventKind = "chaos_ensues"

// ChaosEnsuesForEffect is "chaos ensues" (plane == uuid.Nil) or "chaos
// ensues on that plane" (plane the card).
//
// Caller must hold g.mu (write).
func (g *Game) ChaosEnsuesForEffect(player, plane uuid.UUID) {
	if g.Planechase == nil {
		return
	}
	g.EmitEvent(Event{Kind: EventChaosEnsues, Actor: player, CardID: plane})
}

// RevealPlanarUntilPlaneForEffect is "reveal cards from the top of your
// planar deck until you reveal a plane card": the cards revealed, top card
// first, and the plane card the reveal stopped on, uuid.Nil when the deck
// holds none. Nothing moves (a reveal is not a zone change, CR 701.20b); the
// cards stay on the planar deck until the effect puts them somewhere.
//
// Caller must hold g.mu (write).
func (g *Game) RevealPlanarUntilPlaneForEffect(player, source uuid.UUID, reason string) ([]uuid.UUID, uuid.UUID) {
	deck := g.planarDeckForLocked(player)
	if deck == nil {
		return nil, uuid.Nil
	}
	var run []uuid.UUID
	plane := uuid.Nil
	for i := len(deck.Cards) - 1; i >= 0; i-- {
		c := &deck.Cards[i]
		run = append(run, c.InstanceID)
		if !isPhenomenonCard(c) {
			plane = c.InstanceID
			break
		}
	}
	if len(run) == 0 {
		return nil, uuid.Nil
	}
	seq := g.eventSeq + 1
	for _, id := range run {
		g.EmitEvent(Event{
			Kind: EventRevealCards, Actor: player, Source: source, CardID: id,
			Label: reason, OldZone: ZoneCommand, RevealSeq: seq,
		})
	}
	return run, plane
}

// PlaneswalkToCardForEffect is "planeswalk to it, except don't planeswalk
// away from any plane": the planar deck card `cardID` is turned face up on
// top of the face-up planes, which stay (Norn's Seedcore, and its ruling:
// until a player planeswalks the game continues on several planes). The
// player planeswalked to it (CR 901.11): "when you planeswalk to" abilities
// trigger. ErrInvalidParam when the card is not on the planar deck.
//
// Caller must hold g.mu (write).
func (g *Game) PlaneswalkToCardForEffect(player, cardID uuid.UUID) error {
	if g.Planechase == nil || !g.playerInGameLocked(player) {
		return ErrInvalidParam
	}
	deck := g.planarDeckForLocked(player)
	c, err := deck.Remove(cardID)
	if err != nil {
		return ErrInvalidParam
	}
	viewers := make([]uuid.UUID, 0, len(g.Seats))
	for _, p := range g.Seats {
		if p != nil {
			viewers = append(viewers, p.ID)
		}
	}
	c.AddKnowersAll(viewers)
	c.EnteredBattlefieldAt = timeNowUnixNano()
	g.planarFaceUpLocked().PushTop(c)
	g.refreshSplitSecondForBoardChangeLocked()
	g.EmitEvent(Event{Kind: EventPlaneswalkTo, Actor: player, CardID: c.InstanceID, PublicNames: map[uuid.UUID]string{c.InstanceID: c.Name}})
	return nil
}

const planarBottomAnyOrderKey = "planechase/bottom-any-order"

var planarBottomAnyOrderThen OptionPickThen

func init() {
	planarBottomAnyOrderThen = RegisterOptionPickThen(planarBottomAnyOrderKey, planarBottomAnyOrderAnswered)
}

// PutPlanarCardsOnBottomInAnyOrderForEffect is "put the rest on the bottom
// of your planar deck in any order": the planar controller picks the card
// that goes to the bottom next until one is left (a keyed pick per card, so
// the pause is a restore point). The cards named are on the planar deck.
//
// Caller must hold g.mu (write).
func (g *Game) PutPlanarCardsOnBottomInAnyOrderForEffect(ids []uuid.UUID) {
	g.putPlanarBottomNextLocked(g.planarCardsStillOnDeckLocked(ids))
}

// planarCardsStillOnDeckLocked is the cards of `ids` that are on the planar
// deck, in order.
func (g *Game) planarCardsStillOnDeckLocked(ids []uuid.UUID) []uuid.UUID {
	deck := g.planarDeckForLocked(uuid.Nil)
	var out []uuid.UUID
	for _, id := range ids {
		if deck != nil && deck.Contains(id) {
			out = append(out, id)
		}
	}
	return out
}

// putPlanarBottomNextLocked asks which of `ids` goes to the bottom next, or
// puts the last one there.
func (g *Game) putPlanarBottomNextLocked(ids []uuid.UUID) {
	deck := g.planarDeckForLocked(uuid.Nil)
	if deck == nil || len(ids) == 0 {
		return
	}
	if len(ids) == 1 || !g.playerInGameLocked(g.Planechase.Controller) {
		for _, id := range ids {
			if c, err := deck.Remove(id); err == nil {
				deck.PushBottom(c)
			}
		}
		return
	}
	options := make([]ChoiceOption, 0, len(ids))
	for _, id := range ids {
		name := "a plane card"
		for _, c := range deck.Cards {
			if c.InstanceID == id {
				name = c.Name
			}
		}
		options = append(options, ChoiceOption{Label: name + " goes to the bottom next"})
	}
	g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  g.Planechase.Controller,
		Question: "Put the revealed cards on the bottom of the planar deck in any order. Which goes next?",
		Options:  options,
		ThenKey:  planarBottomAnyOrderThen,
		Carry:    append([]uuid.UUID(nil), ids...),
	})
}

// planarBottomAnyOrderAnswered puts the chosen card on the bottom and asks
// about the rest; unanswered (the planar controller left), the cards go in
// the order they were revealed.
func planarBottomAnyOrderAnswered(g *Game, r OptionPicked) error {
	if g.Planechase == nil {
		return nil
	}
	ids := g.planarCardsStillOnDeckLocked(r.Carry)
	if len(ids) == 0 {
		return nil
	}
	pick := 0
	if r.Index >= 0 && r.Index < len(ids) {
		pick = r.Index
	}
	deck := g.planarDeckForLocked(uuid.Nil)
	if c, err := deck.Remove(ids[pick]); err == nil {
		deck.PushBottom(c)
	}
	g.putPlanarBottomNextLocked(append(append([]uuid.UUID(nil), ids[:pick]...), ids[pick+1:]...))
	return nil
}

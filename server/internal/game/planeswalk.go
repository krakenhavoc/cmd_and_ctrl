package game

import (
	"github.com/google/uuid"
)

// planeswalk.go is planeswalking (CR 901.8, 901.11, ADR 0216 §5, slice
// 2, #4870). The rules, quoted from the September 25, 2026 edition
// AGENTS.md §6 pins:
//
//	901.7a A face-up plane card or phenomenon card that's turned face
//	       down becomes a new object.
//	901.8. Planechase games have an inherent triggered ability known as
//	       the "planeswalking ability." The full text of this ability is
//	       "Whenever you roll the Planeswalker symbol on the planar die,
//	       planeswalk." (See rule 701.31, "Planeswalk.") This ability has
//	       no source and is controlled by the player whose planar die
//	       roll caused it to trigger. This is an exception to rule 113.8.
//	901.11. After the game has started, if a player moves the top card
//	       of their planar deck off that planar deck and turns it face
//	       up, that player has "planeswalked." Continuous effects with
//	       durations that last until a player planeswalks end.
//	       Abilities that trigger when a player planeswalks trigger. See
//	       rule 701.31.
//	901.11b The plane card that's turned face up is the plane the player
//	       planeswalks to. The plane card or phenomenon card that's
//	       turned face down, or that leaves the game, is the plane or
//	       phenomenon the player planeswalks away from.
//	901.11c If a player planeswalks when there is more than one face-up
//	       plane card, that player planeswalks away from all such
//	       planes.
//
// What the text settles that ADR 0216 §5 guessed:
//
//   - The planeswalking ability has NO SOURCE and is controlled by the
//     roller (901.8). So one roll is one trigger, whoever rolled and
//     however many planes are face up, not one per face-up plane. Its
//     stack item names no card (SourceCardID is nil) and its body is
//     keyed, so a table with one waiting is a restore point.
//   - Every face-up card goes, and only then is the top card turned up
//     (901.11b, 901.11c). With more than one face-up card, the planar
//     controller (their owner, CR 901.15b) chooses the order they go to
//     the bottom, one keyed pick per card, so the pause is a restore
//     point. A face-up card turned face down is a new object (901.7a):
//     its knowers are cleared.
//
// Rule 701.31 (the keyword action) is not quoted here: the session that
// wrote this could not reach the pinned text, and 901 is what was
// given. Slice 4, whose cards say "planeswalk", checks 701.31 against
// this order.

// EventPlaneswalkFrom — a player planeswalked away from a planar card
// (CR 901.11b). Actor is the player who planeswalked, CardID the card
// that left the face-up set (now on the bottom of the planar deck).
// Emitted once per card, before EventPlaneswalkTo.
const EventPlaneswalkFrom EventKind = "planeswalk_from"

// EventPlaneswalkTo — a player planeswalked to a plane (CR 901.11b).
// Actor is the player who planeswalked, CardID the card turned face up.
// Not emitted for the starting plane (CR 901.5: that is not a
// planeswalk).
const EventPlaneswalkTo EventKind = "planeswalk_to"

// The planeswalking keys. On-disk identities: never renamed, never
// reused (ADR 0083, #1497).
const (
	planeswalkBodyKey  = "planechase/planeswalk"
	bottomOrderPickKey = "planechase/bottom-order"
)

var bottomOrderThen OptionPickThen

func init() {
	bottomOrderThen = RegisterOptionPickThen(bottomOrderPickKey, bottomOrderAnswered)
}

// planeswalkBody is the planeswalking ability resolving: Params.Player,
// who rolled, planeswalks.
var planeswalkBody = DelayedBody(planeswalkBodyKey, func(g *Game, _ *StackItem, p EffectParams) error {
	return g.PlaneswalkForEffect(p.Player)
})

// planechaseTriggers is the Listener for CR 901.8's planeswalking
// ability. Registered from NewGame beside initiativeTriggers; a table
// without Planechase never reaches past the nil check.
type planechaseTriggers struct{}

// OnEvent runs under g.mu in write mode from notifyListenersLocked.
func (planechaseTriggers) OnEvent(g *Game, ev Event) {
	if g.Planechase == nil {
		return
	}
	if ev.Kind == EventRollDie && ev.PlanarFace == PlanarFacePlaneswalk {
		g.queuePlaneswalkingAbilityLocked(ev.Actor)
	}
}

// queuePlaneswalkingAbilityLocked puts CR 901.8's ability on the way to
// the stack: no source, controlled by the roller.
//
// Caller must hold g.mu (write).
func (g *Game) queuePlaneswalkingAbilityLocked(roller uuid.UUID) {
	p := g.playerByIDLocked(roller)
	if p == nil || p.Eliminated {
		return
	}
	params := EffectParams{Player: roller}
	g.queueHarvestedTriggerLocked(&StackItem{
		Kind:       StackItemTriggered,
		Controller: roller,
		Owner:      roller,
		Label:      "planeswalking ability — " + p.Name + " planeswalks",
		Body:       planeswalkBody.Key(),
		Params:     params,
		Effect:     bodyEffect(planeswalkBody.Key(), params),
	})
}

// PlaneswalkForEffect is "planeswalk" (CR 901.11): `player` planeswalks
// away from every face-up planar card, then to the top card of the
// planar deck. The planeswalking ability resolves through it, and so
// will a card that says "planeswalk". ErrInvalidParam when the table is
// not playing Planechase.
//
// It may pause for the bottom order (more than one face-up card); the
// answer finishes it.
//
// Caller must hold g.mu (write).
func (g *Game) PlaneswalkForEffect(player uuid.UUID) error {
	if g.Planechase == nil || !g.playerInGameLocked(player) {
		return ErrInvalidParam
	}
	g.continuePlaneswalkLocked(player)
	return nil
}

// continuePlaneswalkLocked puts face-up cards on the bottom until one
// is left (asking the order while there is a choice), then that one,
// then turns the top card up.
//
// Caller must hold g.mu (write).
func (g *Game) continuePlaneswalkLocked(player uuid.UUID) {
	up := g.planarFaceUpLocked()
	if len(up.Cards) > 1 && g.playerInGameLocked(g.Planechase.Controller) {
		g.askBottomOrderLocked(player)
		return
	}
	for len(up.Cards) > 0 {
		g.planeswalkAwayFromLocked(player, up.Cards[0].InstanceID)
	}
	g.planeswalkToTopLocked(player)
}

// askBottomOrderLocked asks the planar controller which face-up card
// goes to the bottom of the planar deck next. The cards still face up
// are the pause's state; Carry is the planeswalking player.
//
// Caller must hold g.mu (write).
func (g *Game) askBottomOrderLocked(player uuid.UUID) {
	up := g.PlanarFaceUpCards()
	options := make([]ChoiceOption, 0, len(up))
	for _, c := range up {
		options = append(options, ChoiceOption{Label: c.Name + " goes to the bottom next"})
	}
	g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  g.Planechase.Controller,
		Question: "Planeswalking away from every face-up plane. Which goes to the bottom of the planar deck first?",
		Options:  options,
		ThenKey:  bottomOrderThen,
		Carry:    []uuid.UUID{player},
	})
}

// bottomOrderAnswered puts the chosen card on the bottom and goes on.
// Unanswered (the planar controller left), the cards go in the order
// they were turned up.
func bottomOrderAnswered(g *Game, r OptionPicked) error {
	if g.Planechase == nil || len(r.Carry) != 1 {
		return nil
	}
	player := r.Carry[0]
	up := g.PlanarFaceUpCards()
	pick := 0
	if r.Index >= 0 && r.Index < len(up) {
		pick = r.Index
	}
	if len(up) > 0 {
		g.planeswalkAwayFromLocked(player, up[pick].InstanceID)
	}
	g.continuePlaneswalkLocked(player)
	return nil
}

// planeswalkAwayFromLocked moves one face-up card to the bottom of the
// planar deck, face down and a new object (CR 901.7a), ends every
// "until you planeswalk away" effect it made (CR 901.11) and announces
// it (CR 901.11b).
//
// Caller must hold g.mu (write).
func (g *Game) planeswalkAwayFromLocked(player, cardID uuid.UUID) {
	up := g.planarFaceUpLocked()
	c, err := up.Remove(cardID)
	if err != nil {
		return
	}
	// CR 901.7a: turned face down, it is a new object.
	c.ClearKnown()
	c.Counters = nil
	c.ObjectEpoch++
	c.EnteredBattlefieldAt = 0
	g.planarDeckForLocked(c.Owner).PushBottom(c)
	g.endPlaneswalkAwayDurationsLocked(cardID)
	// Permanents phased out "until a player planeswalks" (Unyaro) phase in.
	g.sweepPhaseInLocksLocked()
	// A face-down plane grants nothing: a spell its static gave split
	// second (Lake Silencio) loses it here (ADR 0216 slice 4).
	g.refreshSplitSecondForBoardChangeLocked()
	g.EmitEvent(Event{Kind: EventPlaneswalkFrom, Actor: player, CardID: cardID, PublicNames: map[uuid.UUID]string{cardID: c.Name}})
}

// planeswalkToTopLocked turns the top card of the planar deck face up
// (CR 901.11): the plane `player` planeswalks to. Public, like every
// face-up card in the command zone.
//
// Caller must hold g.mu (write).
func (g *Game) planeswalkToTopLocked(player uuid.UUID) {
	deck := g.planarDeckForLocked(player)
	c, err := deck.PopTop()
	if err != nil {
		return
	}
	viewers := make([]uuid.UUID, 0, len(g.Seats))
	for _, p := range g.Seats {
		if p != nil {
			viewers = append(viewers, p.ID)
		}
	}
	c.AddKnowersAll(viewers)
	// Its static abilities' CR 613.7 timestamp, as an emblem's is.
	c.EnteredBattlefieldAt = timeNowUnixNano()
	g.planarFaceUpLocked().PushTop(c)
	g.refreshSplitSecondForBoardChangeLocked()
	g.EmitEvent(Event{Kind: EventPlaneswalkTo, Actor: player, CardID: c.InstanceID, PublicNames: map[uuid.UUID]string{c.InstanceID: c.Name}})
}

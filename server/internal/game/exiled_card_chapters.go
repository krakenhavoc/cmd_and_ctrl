package game

import "github.com/google/uuid"

// exiled_card_chapters.go — the two verbs The Creation of Avacyn's later
// chapters need on "the exiled card" (#4077, CR 607.2a):
//
//	TurnExiledFaceUpForEffect                  II: "Turn the exiled card face up."
//	PutExiledCreatureOntoBattlefieldOrHand...  III: "You may put the exiled card
//	                                           onto the battlefield if it's a
//	                                           creature card. If you don't put it
//	                                           onto the battlefield, put it into
//	                                           its owner's hand."
//
// Chapter I is an ordinary SearchLibrary with SearchLibrarySpec.LinkTo and
// FaceDownHidden, so the card arrives linked to the Saga and looked at only
// by its controller; these two read it back through ExiledWithForEffect.

// TurnExiledFaceUpForEffect turns the face-down exiled card `cardID` face
// up (CR 406.3b), as an effect rather than a special action, and shows it to
// the table. Not a "reveal" of a hidden zone: the card is in exile, a public
// zone, and once it is face up everyone may read it, so every seat becomes a
// knower and the log carries the same reveal line a search does.
//
// A card that is not in exile, or is already face up, is left alone and the
// answer is false: "turn the exiled card face up" has nothing to do.
//
// Caller must hold g.mu in write mode.
func (g *Game) TurnExiledFaceUpForEffect(source, cardID uuid.UUID, reason string) bool {
	if g.Exile == nil {
		return false
	}
	for i := range g.Exile.Cards {
		c := &g.Exile.Cards[i]
		if c.InstanceID != cardID {
			continue
		}
		if !c.FaceDown {
			return false
		}
		c.ClearFaceDown()
		c.effective = nil
		g.RevealForEffect(RevealSpec{
			Player: c.Owner,
			Source: source,
			Reason: reason,
			Cards:  []uuid.UUID{cardID},
		})
		return true
	}
	return false
}

// PutExiledCreatureOntoBattlefieldOrHandForEffect is the third chapter's
// instruction. A non-creature exiled card goes straight to its owner's hand
// (the "if you don't" clause, with no choice to make). A creature card
// offers `player` the battlefield: accepting puts it onto the battlefield
// under their control, declining puts it into its owner's hand, and an
// accepted entry that a replacement turned away leaves the card in exile —
// "if you don't put it onto the battlefield" — so it goes to the hand too.
// A card that has already left exile (a response took it) is left alone.
//
// Caller must hold g.mu in write mode.
func (g *Game) PutExiledCreatureOntoBattlefieldOrHandForEffect(player, source, cardID uuid.UUID) error {
	c, ok := g.cardInZoneLocked(g.Exile, cardID)
	if !ok {
		return nil
	}
	if !c.IsCreature() {
		return g.putExiledIntoHandLocked(player, source, cardID)
	}
	_, err := g.QueueConfirmKeyedForEffect(ConfirmPrompt{
		Chooser:      player,
		Source:       source,
		Question:     "Put " + c.Name + " onto the battlefield? If you don't, it goes to its owner's hand.",
		AcceptLabel:  "Put it onto the battlefield",
		DeclineLabel: "Put it into hand",
	}, exiledCreatureOrHandAnswer, []uuid.UUID{cardID})
	return err
}

// exiledCreatureOrHandAnswer is the answer to the third chapter's offer,
// a registered key (#4298): the exiled card rides the prompt as its carried
// ID, and the player and the Saga are its chooser and source. Accepting
// puts the card onto the battlefield, and an entry a replacement turned away
// leaves it in exile, so it goes to the hand too; declining puts it into its
// owner's hand.
var exiledCreatureOrHandAnswer = RegisterConfirmThen("confirm/exiled-creature-battlefield-or-hand", func(g *Game, r ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	cardID, player, source := r.Carry[0], r.Chooser, r.Source
	if !r.Accepted {
		return g.putExiledIntoHandLocked(player, source, cardID)
	}
	return g.ReturnFromExileToBattlefieldKeyedForEffect(cardID, player, false,
		KeyedTail{Ref: exiledCreatureOrHandRest, Player: player, Source: source, Carry: []uuid.UUID{cardID}})
})

// exiledCreatureOrHandRest is the entry's tail as a registered tail (#4253
// slice 6): an entry a replacement turned away leaves the card in exile,
// and "if you don't put it onto the battlefield" puts it into the hand. The
// exiled card is the tail's Carry, the player and the Saga its Player and
// Source. An on-disk key: never renamed, never reused.
var exiledCreatureOrHandRest = RegisterTail("engine/exiled-creature-or-hand", func(g *Game, o TailOutcome) error {
	if len(o.IDs) > 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	cardID := o.Tail.Carry[0]
	if g.Exile == nil || !g.Exile.Contains(cardID) {
		return nil
	}
	return g.putExiledIntoHandLocked(o.Tail.Player, o.Tail.Source, cardID)
})

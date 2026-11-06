package game

import (
	"slices"

	"github.com/google/uuid"
)

// final_board.go is #4912: what an eliminated player had when they left
// the game, kept as plain data so the table can still look at it.
//
// CR 800.4a takes every object the player owns out of the game the moment
// they leave, so the seat's zones are empty in every later frame. The
// record below is written once, by leaveGameLocked, BEFORE any of that
// happens (before even the stack cleanup), and is never changed again. It
// is not game state any rule reads: nothing in the engine looks at it, and
// it holds no object, only what each card looked like.
//
// What it covers, per the issue: the battlefield (what the player
// controlled, with counters, tapped state and what each attachment was
// on), their hand, graveyard, exile and command zone. Never their library:
// its order and contents stay hidden for good.
//
// #4547: it also names the player's commanders, wherever each was, for
// the tombstone's commander line. The zones alone do not: a commander an
// opponent had stolen is on nobody's record (the battlefield is what the
// player controlled), one being cast is on the stack, one shuffled away
// is in the library, and an opponent's commander the player had stolen
// IS on the record, with no owner to tell it apart. The card keeps its
// commander designation in every zone (CR 903.3), and it is public from
// the start of the game, which puts it face up into the command zone
// (CR 903.6), so naming one that is in the library tells the table
// nothing it did not know.
//
// Every card is recorded face up. CR 708.9 reveals a face-down permanent
// as it leaves the battlefield, every face-down permanent the leaving
// player owns, and every face-down permanent at the end of a game; and the
// engine reveals the player's own face-down exiled cards as they leave
// (revealFaceDownOwnedByLocked, ADR 0069 decision 5). The one card no
// rule or engine step reveals is a face-down exiled card when the
// departure ENDS the game (leaveGameObjectsLocked does not run then, ADR
// 0060 Decision 5), so that card is recorded face down with its identity
// left out.
//
// Who sees it: everyone, hand included. The record only exists once the
// player has left, and the owner decided (2026-11-04, on #4912) that
// opponents may see the last hand then, "so others can also know if they
// made the right call of taking them out first".

// FinalBoard is an eliminated player's zones at the moment they left the
// game (#4912). Nil on a player who is still in, and on one restored from
// a snapshot written before the record existed. Additive within snapshot
// v9: it is carried on the player as is.
type FinalBoard struct {
	Battlefield []FinalCard `json:"battlefield,omitempty"`
	Hand        []FinalCard `json:"hand,omitempty"`
	Graveyard   []FinalCard `json:"graveyard,omitempty"`
	Exile       []FinalCard `json:"exile,omitempty"`
	Command     []FinalCard `json:"command,omitempty"`
	// Commanders is the names of the commanders the player owned, each
	// once, wherever it was (#4547, commanderNamesLocked). Absent for a
	// player with none, and on a record written before #4547.
	Commanders []string `json:"commanders,omitempty"`
}

// FinalCard is one card of a FinalBoard: what it looked like, nothing it
// could do. The instance ID is the one it had, kept so an attachment can
// name its host on the same board; it is not an object any more.
type FinalCard struct {
	InstanceID uuid.UUID `json:"instanceId"`
	// Name is the active face's name, as the seat showed it. Empty, with
	// ScryfallID, for a face-down card nothing revealed.
	Name       string `json:"name,omitempty"`
	ScryfallID string `json:"scryfallId,omitempty"`
	Layout     string `json:"layout,omitempty"`
	ActiveFace int    `json:"activeFace,omitempty"`
	TypeLine   string `json:"typeLine,omitempty"`
	// Power and Toughness are a creature card's, counters included, as
	// the seat showed them. Zero for anything else.
	Power       int            `json:"power,omitempty"`
	Toughness   int            `json:"toughness,omitempty"`
	Tapped      bool           `json:"tapped,omitempty"`
	FaceDown    bool           `json:"faceDown,omitempty"`
	IsToken     bool           `json:"isToken,omitempty"`
	IsCommander bool           `json:"isCommander,omitempty"`
	Counters    map[string]int `json:"counters,omitempty"`
	// AttachedTo is the card an Aura, Equipment or Fortification was on,
	// and AttachedToName what that host was called (a card's name, or a
	// player's for a Curse), so the host can be named even when it was
	// another player's permanent and is not on this board.
	AttachedTo     string `json:"attachedTo,omitempty"`
	AttachedToName string `json:"attachedToName,omitempty"`
}

// recordFinalBoardLocked writes p.FinalBoard from the board as it stands.
// revealed reports whether the player's face-down exiled cards will be
// revealed as they leave (the departure does not end the game).
//
// Caller must hold g.mu.
func (g *Game) recordFinalBoardLocked(p *Player, revealed bool) {
	if p == nil {
		return
	}
	// The layers are read as they stand, not recomputed: a recompute can
	// change combat state, and the record must not move the game.
	fb := &FinalBoard{}
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller == p.ID {
				fb.Battlefield = append(fb.Battlefield, g.finalCardLocked(c, true, true))
			}
		}
	}
	fb.Hand = g.finalCardsLocked(p.Hand, true)
	fb.Graveyard = g.finalCardsLocked(p.Graveyard, true)
	fb.Command = g.finalCardsLocked(p.Command, true)
	if g.Exile != nil {
		for i := range g.Exile.Cards {
			c := &g.Exile.Cards[i]
			if c.Owner == p.ID {
				fb.Exile = append(fb.Exile, g.finalCardLocked(c, false, revealed || !c.FaceDown))
			}
		}
	}
	fb.Commanders = g.commanderNamesLocked(p)
	p.FinalBoard = fb
}

// commanderNamesLocked is the names of the commanders p owns, each once:
// those in the command zone first, then the battlefield, the stack, the
// graveyard, exile, the hand and the library. The name is the card's
// printed name (the face that is up), not what a copy effect made it.
//
// Caller must hold g.mu.
func (g *Game) commanderNamesLocked(p *Player) []string {
	var names []string
	add := func(z *Zone) {
		if z == nil {
			return
		}
		for i := range z.Cards {
			c := &z.Cards[i]
			if !c.IsCommander || c.Owner != p.ID || c.Name == "" || slices.Contains(names, c.Name) {
				continue
			}
			names = append(names, c.Name)
		}
	}
	for _, z := range []*Zone{p.Command, g.Battlefield, g.Stack, p.Graveyard, g.Exile, p.Hand, p.Library} {
		add(z)
	}
	return names
}

// finalCardsLocked records every card of one of the player's own zones.
func (g *Game) finalCardsLocked(z *Zone, revealed bool) []FinalCard {
	if z == nil || len(z.Cards) == 0 {
		return nil
	}
	out := make([]FinalCard, 0, len(z.Cards))
	for i := range z.Cards {
		out = append(out, g.finalCardLocked(&z.Cards[i], false, revealed))
	}
	return out
}

// finalCardLocked records one card. onBattlefield adds what only a
// permanent has: tapped, and what it is attached to. revealed false
// leaves the card's identity out.
func (g *Game) finalCardLocked(c *Card, onBattlefield, revealed bool) FinalCard {
	fc := FinalCard{
		InstanceID:  c.InstanceID,
		IsToken:     IsTokenPtr(c),
		IsCommander: c.IsCommander,
	}
	if len(c.Counters) > 0 {
		fc.Counters = make(map[string]int, len(c.Counters))
		for k, v := range c.Counters {
			fc.Counters[k] = v
		}
	}
	if !revealed {
		fc.FaceDown = true
		return fc
	}
	eff := EffectiveOf(c)
	fc.Name = eff.Name
	if fc.Name == "" {
		fc.Name = c.Name
	}
	fc.ScryfallID = c.ScryfallID
	fc.Layout = c.Layout
	fc.ActiveFace = c.ActiveFace
	fc.TypeLine = c.TypeLine
	if IsCreaturePtr(c) && c.ToughnessIsKnown() {
		fc.Power, _, fc.Toughness = PowerToughnessFrom(c, eff)
	}
	if !onBattlefield {
		return fc
	}
	fc.Tapped = c.Tapped
	if c.AttachedTo.ID != uuid.Nil {
		fc.AttachedTo = c.AttachedTo.ID.String()
		switch c.AttachedTo.Kind {
		case TargetPlayer:
			if host := g.playerByIDLocked(c.AttachedTo.ID); host != nil {
				fc.AttachedToName = host.Name
			}
		default:
			if host := g.findCardByIDLocked(c.AttachedTo.ID); host != nil {
				fc.AttachedToName = EffectiveOf(host).Name
				if fc.AttachedToName == "" {
					fc.AttachedToName = host.Name
				}
			}
		}
	}
	return fc
}

// cloneFinalBoard deep-copies a record, for clonePlayer.
func cloneFinalBoard(fb *FinalBoard) *FinalBoard {
	if fb == nil {
		return nil
	}
	cp := func(in []FinalCard) []FinalCard {
		if in == nil {
			return nil
		}
		out := make([]FinalCard, len(in))
		for i, c := range in {
			out[i] = c
			if c.Counters != nil {
				out[i].Counters = make(map[string]int, len(c.Counters))
				for k, v := range c.Counters {
					out[i].Counters[k] = v
				}
			}
		}
		return out
	}
	return &FinalBoard{
		Battlefield: cp(fb.Battlefield),
		Hand:        cp(fb.Hand),
		Graveyard:   cp(fb.Graveyard),
		Exile:       cp(fb.Exile),
		Command:     cp(fb.Command),
		Commanders:  slices.Clone(fb.Commanders),
	}
}

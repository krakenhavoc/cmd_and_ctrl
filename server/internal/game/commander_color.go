package game

import (
	"fmt"

	"github.com/google/uuid"
)

// commander_color.go — "If <this> is your commander, choose a color
// before the game begins. <This> is the chosen color." (CR 903.4b,
// #5948): The Prismatic Piper, Faceless One, Clara Oswald.
//
//	903.4b If a commander has a static ability that causes a player to
//	choose its color before the game begins, that choice applies during
//	deck construction and throughout the game, even as the commander
//	changes zones. That choice affects the commander's color identity.
//	The player reveals that choice as they put their commander into the
//	command zone before the game begins.
//
// # The choice
//
// The rule puts the choice at deck construction, and a decklist has no
// place to write it. So the deck says it the only way it can: by the
// cards it holds. A deck whose other cards need one colour outside the
// rest of its commanders' identity has chosen that colour (the deck
// checker accepts exactly such a deck, internal/deck/commander_color.go),
// and the engine records it as the game begins without asking. A deck
// that needs no colour outside that identity made no choice the cards can
// show, so its owner is asked, with every colour offered (CR 105.4).
//
// The question is the ordinary choose_color prompt (color_choice.go), so
// the client's colour picker and every bot tier answer it the way they
// answer Coldsteel Heart. It is asked when the mulligan window closes,
// before the opening-hand actions (CR 103.2c comes before CR 103.6), in
// turn order from the starting player, and it holds the table like any
// open choice. That is the narrowing ADR 0133 makes for the opening-hand
// actions, and it costs nothing here either: nothing reads a commander's
// colour before the first turn.
//
// # The colour
//
// The answer is Card.CommanderColor, an attribute of the card like
// IsCommander, kept through every zone change. Its second sentence is a
// characteristic-defining ability (CR 604.3, linked to the choice by CR
// 607.2p), so it is read where devoid is read: in the printed colour
// baseline (printedInputsOf, Card.EffectiveColors), which holds in every
// zone and which a layer-5 effect still applies on top of. It is gated on
// the card's catalog row still declaring the ability, so a commander that
// becomes a copy of something else takes that thing's colours, and a copy
// of the commander (which never has the field, CR 707.2) is colourless.
// The colour is also added to the commander's colour identity
// (printedIdentityOf), which Command Tower and its kin read.

// CatalogChoosesCommanderColor reports whether a card's catalog row
// declares the CR 903.4b ability (CardDef.ChoosesCommanderColor).
// Installed at init; a game-package test can replace it.
var CatalogChoosesCommanderColor func(key string) bool

// choosesCommanderColor reports whether the card has the ability now.
func choosesCommanderColor(c *Card) bool {
	return CatalogChoosesCommanderColor != nil && CatalogChoosesCommanderColor(catalogKeyOf(c))
}

// commanderColorCDA is the colour the CR 903.4b ability makes the card,
// or "" when it defines none: no colour chosen, or the card no longer has
// the ability (it became a copy of something else).
func (c *Card) commanderColorCDA() string {
	if c.CommanderColor == "" || c.FaceDownIsPermanent() {
		return ""
	}
	if !choosesCommanderColor(c) {
		return ""
	}
	return c.CommanderColor
}

// offerCommanderColorChoicesLocked records or asks every commander's
// CR 903.4b colour, in turn order from the starting seat. Called once,
// from the close of the mulligan window. A commander that already has a
// colour (a restarted game, CR 727) keeps it: the choice applies
// throughout the game and was made at deck construction.
//
// Caller must hold g.mu.
func (g *Game) offerCommanderColorChoicesLocked() {
	if CatalogChoosesCommanderColor == nil {
		return
	}
	n := len(g.Seats)
	for k := 0; k < n; k++ {
		p := g.Seats[(g.StartingSeat+k)%n]
		if p == nil || p.Eliminated {
			continue
		}
		type pending struct {
			id   uuid.UUID
			name string
		}
		var asks []pending
		for _, z := range commanderColorZones(p) {
			for i := range z.Cards {
				c := &z.Cards[i]
				if c.IsCommander && c.Owner == p.ID && c.CommanderColor == "" && choosesCommanderColor(c) {
					asks = append(asks, pending{id: c.InstanceID, name: c.Name})
				}
			}
		}
		for _, a := range asks {
			options := g.commanderColorOptionsLocked(p)
			if len(options) == 1 {
				// Chosen at deck construction: revealed, not asked. The
				// reveal is the same history line an answered prompt
				// writes (ResolveColorChoice).
				if g.setCommanderColorLocked(p.ID, a.id, options[0]) {
					g.EmitEvent(Event{Kind: EventColorChosen, Actor: p.ID, CardID: a.id, Label: options[0]})
				}
				continue
			}
			g.QueueColorChoiceThenForEffect(ColorPrompt{
				Chooser:  p.ID,
				Source:   a.id,
				Question: fmt.Sprintf("%s is your commander: choose its color", a.name),
				Options:  options,
				// No purpose (#780): none of the five fits "the card becomes
				// this colour", and the undeclared prompt keeps the neutral
				// copy on the client and the pre-#780 rule for a bot, the
				// colour its kept hand asks for most, which is the sensible
				// answer before the first turn's empty board.
				ThenTail: KeyedTail{Ref: commanderColorChosenRef, Player: p.ID, Source: a.id},
			})
		}
	}
}

// commanderColorZones are the zones a seat's own cards are in as the game
// begins (and, on a restart, wherever CR 727 put them).
func commanderColorZones(p *Player) []*Zone {
	out := make([]*Zone, 0, 4)
	for _, z := range []*Zone{p.Command, p.Hand, p.Library, p.Graveyard} {
		if z != nil {
			out = append(out, z)
		}
	}
	return out
}

// commanderColorOptionsLocked is the colours the seat's deck allows its
// CR 903.4b commander to be. The deck's other cards need the union of
// their colour identities; the commanders without the ability provide
// theirs. A deck that needs exactly one colour more chose it at deck
// construction, and that is the only option. A deck that needs none made
// no choice its cards can show, and every colour is open (CR 105.4). A
// deck that needs more than one is not a legal Commander deck; it is
// offered the colours it needs, the nearest thing to the rule.
//
// Caller must hold g.mu.
func (g *Game) commanderColorOptionsLocked(p *Player) []string {
	need := map[string]bool{}
	have := map[string]bool{}
	add := func(into map[string]bool, c *Card) {
		colors, _ := printedIdentityOf(c)
		for _, col := range colors {
			into[col] = true
		}
	}
	for _, z := range commanderColorZones(p) {
		for i := range z.Cards {
			c := &z.Cards[i]
			if c.Owner != p.ID {
				continue
			}
			switch {
			case c.IsCommander && choosesCommanderColor(c):
				// Its own identity is what is being chosen.
			case c.IsCommander:
				add(have, c)
			default:
				add(need, c)
			}
		}
	}
	if p.Companion != nil {
		// The companion must fit the identity too (CR 702.139c's
		// deck, internal/deck/companion.go).
		for i := range p.Companion.Cards {
			add(need, &p.Companion.Cards[i])
		}
	}
	var extra []string
	for _, col := range AllColors {
		if need[col] && !have[col] {
			extra = append(extra, col)
		}
	}
	if len(extra) == 0 {
		return append([]string(nil), AllColors...)
	}
	return extra
}

// commanderColorChosenRef is the answer to the CR 903.4b prompt: the
// commander (the tail's Source) of the seat (Player) becomes the colour
// told as Word. An empty Word (the chooser left) records nothing. An
// on-disk key: never renamed, never reused.
var commanderColorChosenRef = RegisterTail("engine/commander-color-chosen", func(g *Game, o TailOutcome) error {
	if o.Tail.Word == "" {
		return nil
	}
	g.setCommanderColorLocked(o.Tail.Player, o.Tail.Source, o.Tail.Word)
	return nil
})

// setCommanderColorLocked writes the chosen colour onto the owner's
// commander, wherever it is (CR 903.4b), and reports whether it did. A
// colour that is not one of the five, or a card that is not that seat's
// commander, is ignored.
//
// Caller must hold g.mu.
func (g *Game) setCommanderColorLocked(owner, cardID uuid.UUID, color string) bool {
	if !isColorLetter(color) {
		return false
	}
	c := g.ownedCommanderLocked(owner, cardID)
	if c == nil {
		return false
	}
	c.CommanderColor = color
	g.layerVersion.Add(1)
	return true
}

// ownedCommanderLocked finds the seat's commander by instance ID in any
// zone it can be in.
func (g *Game) ownedCommanderLocked(owner, cardID uuid.UUID) *Card {
	p := g.playerByIDLocked(owner)
	if p == nil {
		return nil
	}
	zones := append(commanderColorZones(p), g.Stack, g.Battlefield, g.Exile)
	for _, z := range zones {
		if z == nil {
			continue
		}
		for i := range z.Cards {
			c := &z.Cards[i]
			if c.InstanceID == cardID && c.IsCommander && c.Owner == owner {
				return c
			}
		}
	}
	return nil
}

// isColorLetter reports whether s is one of the five colour letters.
func isColorLetter(s string) bool {
	for _, col := range AllColors {
		if col == s {
			return true
		}
	}
	return false
}

// SetCommanderColorForTest records a commander's CR 903.4b colour
// directly, for a test that does not play out the mulligan window.
func (g *Game) SetCommanderColorForTest(owner, cardID uuid.UUID, color string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.setCommanderColorLocked(owner, cardID, color)
}

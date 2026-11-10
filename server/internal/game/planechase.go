package game

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// planechase.go is the Planechase table option's state and its communal
// planar deck (CR 901, ADR 0216 slice 1, #4870). The rules this file
// implements, quoted from the September 25, 2026 edition AGENTS.md §6
// pins:
//
//	901.4. All plane and phenomenon cards remain in the command zone
//	       throughout the game, both while they're part of a planar
//	       deck and while they're face up.
//	901.5. Once all players have kept their opening hands and used the
//	       abilities of cards that allow them to take an action with
//	       those cards from their opening hands, the starting player
//	       moves the top card of their planar deck off that planar deck
//	       and turns it face up. If it's a phenomenon card, the player
//	       puts that card on the bottom of their planar deck and
//	       repeats this process until a plane card is turned face up.
//	       (See rule 103.7.) No abilities of any card turned face up
//	       this way trigger during this process. The face-up plane card
//	       becomes the starting plane.
//	901.6. The owner of a plane or phenomenon card is the player who
//	       started the game with it in their planar deck. The
//	       controller of a face-up plane or phenomenon card is the
//	       player designated as the planar controller. Normally, the
//	       planar controller is whoever the active player is. However,
//	       if the current planar controller would leave the game,
//	       instead the next player in turn order that wouldn't leave
//	       the game becomes the planar controller, then the old planar
//	       controller leaves the game. The new planar controller
//	       retains that designation until they leave the game or a
//	       different player becomes the active player, whichever comes
//	       first.
//	901.11. After the game has started, if a player moves the top card
//	       of their planar deck off that planar deck and turns it face
//	       up, that player has "planeswalked." [...]
//	901.15a As an alternative option, a Planechase game may be played
//	       with just a single communal planar deck. In that case, the
//	       number of cards in the planar deck must be at least forty or
//	       at least ten times the number of players in the game,
//	       whichever is smaller. The planar deck can't contain more
//	       phenomenon cards than twice the number of players in the
//	       game. Each card in the planar deck must have a different
//	       English name.
//	901.15b In a Planechase game using the single planar deck option,
//	       the planar controller is considered to be the owner of all
//	       cards in the planar deck.
//	901.15c If any rule or ability refers to a player's planar deck,
//	       the communal planar deck is used.
//
// What the text settles that ADR 0216 left to this slice (§3):
//
//   - The starting plane is NOT a planeswalk. CR 901.11 begins "After
//     the game has started", and 901.5 says no ability of a card turned
//     face up this way triggers. So turnUpStartingPlaneLocked emits no
//     planeswalk event, and "whenever you planeswalk to ~" (slice 2)
//     will not see it.
//   - A phenomenon on top goes to the bottom and the next card is
//     turned up, until a plane is. Slice 1's deck holds planes only,
//     but the loop is written as 901.5 says so slice 5 need not touch
//     it.
//   - The minimum is min(40, 10 × players), at most 2 × players
//     phenomena, and every name different (901.15a). One card per
//     oracle ID already makes every name different; the check is kept
//     so a broken source cannot slip a duplicate in.
//
// What slice 1 does NOT do: the planar die, planeswalking, chaos, plane
// abilities, phenomena and CR 901.6's leave-the-game hand-on (ADR 0216
// §§4–8, slice 2 onwards). A face-up plane is a card the table reads;
// nothing it says happens automatically yet.
//
// One ordering note on CR 901.5. The opening-hand actions (Leylines,
// ADR 0133) are prompts the mulligan close queues and the table answers
// while the first turn begins. The starting plane is turned up when the
// mulligan window closes, right after those prompts are queued, not
// after they are answered. Nothing in slice 1 can tell the difference:
// no plane ability functions yet, and 901.5 says none triggers.

// PlanarDeckMode says how a Planechase game holds its planar cards.
type PlanarDeckMode string

const (
	// PlanarDeckCommunal is CR 901.15's single communal planar deck,
	// the only mode ADR 0216 writes. CR 901.3's per-player decks would
	// add "per_player"; every reader goes through planarDeckForLocked.
	PlanarDeckCommunal PlanarDeckMode = "communal"
)

// PlanarKinds selects which planar cards PlanarCardSource returns.
type PlanarKinds int

const (
	// PlanarPlanes are plane cards (CR 311), type line "Plane — …".
	PlanarPlanes PlanarKinds = 1 << iota
	// PlanarPhenomena are phenomenon cards (CR 312). Not asked for
	// until ADR 0216 slice 5.
	PlanarPhenomena
)

// planarKindsThisSlice is what setUpPlanechaseLocked asks the source
// for: planes only (owner decision 3, ADR 0216).
const planarKindsThisSlice = PlanarPlanes

// PlanarCardSource is the card database's planar cards (ADR 0216 §2).
// main.go installs it from deck.NewCollection(cardIdx). internal/game
// cannot import internal/deck (OutsideCollectionSource explains the
// cycle). Nil means no index: a Planechase table then refuses to start
// (ErrNoPlanarCards).
type PlanarCardSource interface {
	// PlanarCards returns one game card per planar oracle ID of the
	// asked kinds, built by deck.toGameCard, in oracle-ID order. Each
	// is a TEMPLATE with no owner; its instance ID means nothing. The
	// game mints from it.
	PlanarCards(kinds PlanarKinds) []Card
}

// PlanarCards is the installed planar card source, or nil when the
// server runs with no card index.
var PlanarCards PlanarCardSource

// The errors a Planechase table can be refused with.
var (
	// ErrPlanechaseLocked is returned by UpdateSettings when a patch
	// changes Planechase once the game has left the lobby (ADR 0216
	// §1): the planar deck is built at Start. Nothing in the patch is
	// applied.
	ErrPlanechaseLocked = errors.New("game: planechase cannot change once the game has started")
	// ErrNoPlanarCards refuses Start for a Planechase table when the
	// server has no planar card source (no card index).
	ErrNoPlanarCards = errors.New("game: planechase needs the card database, and this server has none loaded")
	// ErrPlanarDeckTooSmall refuses Start for a Planechase table whose
	// communal planar deck would break CR 901.15a. The server never
	// chooses cards to drop.
	ErrPlanarDeckTooSmall = errors.New("game: the communal planar deck is too small (CR 901.15a)")
	// ErrPlanarDeckIllegal refuses Start for a communal planar deck
	// with too many phenomena or a repeated name (CR 901.15a).
	ErrPlanarDeckIllegal = errors.New("game: the communal planar deck breaks CR 901.15a")
	// ErrPlanechaseNotOffered is the server's refusal to turn
	// Planechase on where this deployment does not offer it (ADR 0216,
	// owner answer 3: cmd-dev only until the plane card batches land).
	// The engine never returns it; the lobby and the WebSocket edge do,
	// from the one switch ws.RoomManager.PlanechaseOffered.
	ErrPlanechaseNotOffered = errors.New("game: planechase is not offered on this server yet")
)

// TurnsOnPlanechase reports whether the patch asks for Planechase on.
// Turning it off is always allowed, so a table restored on a server
// that no longer offers the option can still drop it in the lobby.
func (p SettingsPatch) TurnsOnPlanechase() bool {
	return p.Planechase != nil && *p.Planechase
}

// PlanechaseState is a Planechase game's planar designation and die
// count (CR 901, ADR 0216 §3). Nil on Game when the option is off.
//
// The planar cards themselves are Game.PlanarDeck and
// Game.PlanarFaceUp, not fields here: this struct is plain data, so the
// closure census sees exactly the two routes to cards the owner
// approved (ADR 0216's 2026-11-04 amendment).
type PlanechaseState struct {
	// Mode is PlanarDeckCommunal, the only value ADR 0216 writes.
	Mode PlanarDeckMode
	// Controller is the planar controller (CR 901.6). Written only by
	// setPlanarControllerLocked. uuid.Nil until the starting player is
	// known.
	Controller uuid.UUID
	// DieTurnSeq and DieRolls count planar die special actions this
	// turn (CR 901.9's cost, slice 2). A count whose DieTurnSeq is not
	// Turn.Seq is 0, so no turn-start hook resets it. Slice 1 writes
	// neither.
	DieTurnSeq int
	DieRolls   int
}

// clonePlanechase deep-copies the state. Plain data, so a value copy.
func clonePlanechase(s *PlanechaseState) *PlanechaseState {
	if s == nil {
		return nil
	}
	out := *s
	return &out
}

// PlanarDeckCards is the communal planar deck, bottom first, top last,
// or nil when the table is not playing Planechase. The deck is face
// down: the wire carries only its size.
func (g *Game) PlanarDeckCards() []Card {
	if g == nil || g.PlanarDeck == nil {
		return nil
	}
	return g.PlanarDeck.Cards
}

// PlanarFaceUpCards is the face-up planar cards in the order they were
// turned up, or nil.
func (g *Game) PlanarFaceUpCards() []Card {
	if g == nil || g.PlanarFaceUp == nil {
		return nil
	}
	return g.PlanarFaceUp.Cards
}

// planarDeckForLocked is the planar deck CR 901.15c says a rule about
// owner's planar deck reads: the communal one, whoever owner is. It is
// the one accessor CR 901.3's per-player decks would change. Nil when
// the table is not playing Planechase.
func (g *Game) planarDeckForLocked(_ uuid.UUID) *Zone {
	if g.Planechase == nil {
		return nil
	}
	if g.PlanarDeck == nil {
		g.PlanarDeck = newZone(ZoneCommand, g.Planechase.Controller)
	}
	return g.PlanarDeck
}

// planarFaceUpLocked is the face-up planar cards' container, made on
// first use. Nil when the table is not playing Planechase.
func (g *Game) planarFaceUpLocked() *Zone {
	if g.Planechase == nil {
		return nil
	}
	if g.PlanarFaceUp == nil {
		g.PlanarFaceUp = newZone(ZoneCommand, g.Planechase.Controller)
	}
	return g.PlanarFaceUp
}

// MinCommunalPlanarDeck is CR 901.15a's minimum for a single communal
// planar deck: at least forty, or at least ten times the number of
// players, whichever is smaller.
func MinCommunalPlanarDeck(players int) int {
	return min(40, 10*players)
}

// MaxCommunalPlanarPhenomena is CR 901.15a's cap: no more phenomenon
// cards than twice the number of players.
func MaxCommunalPlanarPhenomena(players int) int {
	return 2 * players
}

// isPhenomenonCard reports whether a planar card is a phenomenon (CR
// 312) rather than a plane, by its printed type line.
func isPhenomenonCard(c *Card) bool {
	return strings.HasPrefix(c.TypeLine, "Phenomenon")
}

// validateCommunalPlanarDeck is CR 901.15a, for a deck of these cards
// at a table of this many players.
func validateCommunalPlanarDeck(deck []Card, players int) error {
	if want := MinCommunalPlanarDeck(players); len(deck) < want {
		return fmt.Errorf("%w: %d cards, and %d players need at least %d",
			ErrPlanarDeckTooSmall, len(deck), players, want)
	}
	names := make(map[string]bool, len(deck))
	phenomena := 0
	for i := range deck {
		c := &deck[i]
		if isPhenomenonCard(c) {
			phenomena++
		}
		key := strings.ToLower(c.Name)
		if names[key] {
			return fmt.Errorf("%w: %q is in it twice", ErrPlanarDeckIllegal, c.Name)
		}
		names[key] = true
	}
	if limit := MaxCommunalPlanarPhenomena(players); phenomena > limit {
		return fmt.Errorf("%w: %d phenomena, and %d players allow at most %d",
			ErrPlanarDeckIllegal, phenomena, players, limit)
	}
	return nil
}

// planarTemplates fetches and checks the communal planar deck's
// cards for a game of this many players, without changing the game.
// Start calls it before anything else moves, so a refused Start leaves
// the table exactly as it was.
func planarTemplates(players int) ([]Card, error) {
	if PlanarCards == nil {
		return nil, ErrNoPlanarCards
	}
	templates := PlanarCards.PlanarCards(planarKindsThisSlice)
	if err := validateCommunalPlanarDeck(templates, players); err != nil {
		return nil, err
	}
	return templates, nil
}

// installPlanarDeckLocked makes the game a Planechase game with this
// communal planar deck: each card minted fresh (a new instance ID and a
// class-1 object ordinal, mintFromCollectionLocked's shape), face down
// in the command zone, then shuffled once on its own stream so a seeded
// game's planar deck is reproducible and no player's library stream
// moves. The planar controller, and so every card's owner, is set once
// the starting player is known (dealAndBeginLocked).
//
// Caller must hold g.mu (write).
func (g *Game) installPlanarDeckLocked(templates []Card) {
	g.Planechase = &PlanechaseState{Mode: PlanarDeckCommunal}
	g.PlanarDeck = newZone(ZoneCommand, uuid.Nil)
	g.PlanarFaceUp = newZone(ZoneCommand, uuid.Nil)
	for _, t := range templates {
		c := t
		c.InstanceID = uuid.New()
		g.noteCreatedSourceLocked(c.InstanceID)
		c.Owner = uuid.Nil
		c.Controller = uuid.Nil
		c.IsSideboard = false
		c.IsCommander = false
		c.IsCompanion = false
		c.Counters = nil
		c.Stickers = nil
		c.CounterStampedAt = nil
		c.LostLastCounter = false
		c.Tapped = false
		c.ClearKnown()
		g.PlanarDeck.PushTop(c)
	}
	if len(g.PlanarDeck.Cards) > 1 {
		g.PlanarDeck.Shuffle(g.randForLocked(rngStream{kind: rngStreamPlanarShuffle, player: uuid.Nil}))
	}
}

// setUpPlanechaseLocked builds the communal planar deck for a table
// whose settings ask for Planechase (ADR 0216 §§1–3). A restart (CR
// 727) calls it for the new game; Start checks the deck first
// (planarTemplates) so a refusal changes nothing.
//
// Caller must hold g.mu (write).
func (g *Game) setUpPlanechaseLocked() error {
	templates, err := planarTemplates(g.playersInGameLocked())
	if err != nil {
		return err
	}
	g.installPlanarDeckLocked(templates)
	return nil
}

// playersInGameLocked is "the number of players in the game" CR
// 901.15a counts: every seat that has not left.
func (g *Game) playersInGameLocked() int {
	n := 0
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			n++
		}
	}
	return n
}

// setPlanarControllerLocked makes id the planar controller (CR 901.6)
// and, because the planar controller is considered the owner of every
// card in the communal planar deck (CR 901.15b), stamps Owner and
// Controller on every planar card in the same step. So "you" on a
// plane, ownership tests and the CR 800.4a sweep all read ordinary card
// fields. It is the only writer of PlanechaseState.Controller.
//
// Caller must hold g.mu (write).
func (g *Game) setPlanarControllerLocked(id uuid.UUID) {
	if g.Planechase == nil {
		return
	}
	g.Planechase.Controller = id
	for _, z := range []*Zone{g.PlanarDeck, g.PlanarFaceUp} {
		if z == nil {
			continue
		}
		z.Owner = id
		for i := range z.Cards {
			z.Cards[i].Owner = id
			z.Cards[i].Controller = id
		}
	}
}

// followActivePlayerAsPlanarControllerLocked is CR 901.6's "normally,
// the planar controller is whoever the active player is", run as each
// turn begins. A turn of the same player (an extra turn) changes
// nothing.
//
// Caller must hold g.mu (write).
func (g *Game) followActivePlayerAsPlanarControllerLocked() {
	if g.Planechase == nil {
		return
	}
	seat := g.Turn.ActiveSeat
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil {
		return
	}
	if id := g.Seats[seat].ID; id != g.Planechase.Controller {
		g.setPlanarControllerLocked(id)
	}
}

// turnUpStartingPlaneLocked is CR 901.5: once every opening hand is
// kept, the starting player (the planar controller by now) turns the
// top card of the planar deck face up; a phenomenon goes to the bottom
// and the next card is turned up, until a plane is. That plane is the
// starting plane. It is not a planeswalk (CR 901.11 is "after the game
// has started") and nothing triggers, so no event is emitted. A table
// that already has a face-up card (a restore, a second settle) is left
// alone.
//
// Caller must hold g.mu (write).
func (g *Game) turnUpStartingPlaneLocked() {
	if g.Planechase == nil {
		return
	}
	deck := g.planarDeckForLocked(g.Planechase.Controller)
	up := g.planarFaceUpLocked()
	if len(up.Cards) > 0 {
		return
	}
	viewers := make([]uuid.UUID, 0, len(g.Seats))
	for _, p := range g.Seats {
		if p != nil {
			viewers = append(viewers, p.ID)
		}
	}
	// Each phenomenon is looked at once at most, so the walk is
	// bounded even by a deck of nothing but phenomena.
	for range len(deck.Cards) {
		c, err := deck.PopTop()
		if err != nil {
			return
		}
		if isPhenomenonCard(&c) {
			deck.PushBottom(c)
			continue
		}
		// Face up in the command zone, which is public (CR 400.2).
		c.AddKnowersAll(viewers)
		c.EnteredBattlefieldAt = timeNowUnixNano()
		up.PushTop(c)
		return
	}
}

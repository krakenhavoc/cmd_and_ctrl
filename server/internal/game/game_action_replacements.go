package game

import (
	"reflect"
	"strconv"

	"github.com/google/uuid"
)

// game_action_replacements.go — replacement effects over two game actions
// that had none: losing the game, and planeswalking (#6096, S82 GAP2).
//
//	CR 614.1a "Effects that use the word “instead” are replacement effects."
//	CR 614.5  "A replacement effect doesn’t invoke itself repeatedly; it
//	          gets only one opportunity to affect an event or any modified
//	          events that may replace that event."
//	CR 614.6  "If an event is replaced, it never happens. A modified event
//	          occurs instead, which may in turn trigger abilities."
//	CR 704.7  "If multiple state-based actions would have the same result at
//	          the same time, a single replacement effect will replace all
//	          of them." (the Lich's Mirror example: a player who would lose
//	          to 704.5a and 704.5b at once is saved by the one Mirror)
//	CR 119.5  "If an effect sets a player’s life total to a specific
//	          number, the player gains or loses the necessary amount of
//	          life to end up with the new total."
//	CR 104.3a "A player can concede the game at any time. A player who
//	          concedes leaves the game immediately." — a concession is not
//	          an event anything replaces, as it is never gated.
//	CR 701.31b "To planeswalk is to put each face-up plane card and
//	          phenomenon card on the bottom of its owner’s planar deck face
//	          down, then move the top card of your planar deck off that
//	          planar deck and turn it face up."
//	CR 901.15c "If any rule or ability refers to a player’s planar deck,
//	          the communal planar deck is used."
//
// Both are statics of a permanent, declared as plain data on its catalog
// row (effects.Spec.GameActionReplacements) and read off the battlefield
// through CatalogAbilityKey, the way a "can't lose the game" gate is
// (game_end_gates.go): a permanent that has lost its abilities replaces
// nothing, and one that changed controller replaces for its new
// controller. Nothing is stored on the player, and no func is reachable
// from Game.
//
// The losing half is read at the two doors a loss goes through: the state-
// based loss pass (each loser, before it leaves) and loseGameLocked (every
// effect loss). The gate is read first: a player who can't lose is not a
// player who would lose, so nothing is replaced. With several replacements
// for one player (two Golden Thrones), CR 616.1 lets that player choose;
// every kind here does the same thing to the same player, so the earliest
// on the battlefield is applied and the choice is not asked.
//
// The planeswalking half is read in PlaneswalkForEffect, which every
// planeswalk goes through (the planeswalking ability, and a card that says
// "planeswalk"). Each source applies once (CR 614.5), in turn, as a pick of
// the planeswalking player's; the planeswalk itself follows.

// GameActionReplacementKind names what a GameActionReplacement replaces and
// with what. On-disk identity of nothing (the row is catalog data, never
// written), but the strings are stable for the logs.
type GameActionReplacementKind string

const (
	// ReplaceLoseGameExileThisLifeBecomes is "If you would lose the game,
	// instead exile [this permanent] and your life total becomes N" (The
	// Golden Throne's Arcane Life-support, N = 1). Life is the N.
	ReplaceLoseGameExileThisLifeBecomes GameActionReplacementKind = "lose-game/exile-this-life-becomes"

	// ReplacePlaneswalkOrderTopTwo is "If you would planeswalk, instead
	// look at the top two cards of your planar deck, put one on the bottom
	// of your planar deck and the other on top, then planeswalk" (Susan
	// Foreman).
	ReplacePlaneswalkOrderTopTwo GameActionReplacementKind = "planeswalk/order-top-two"
)

// GameActionReplacement is one printed "if you would <game action>,
// instead …" static. "You" is the permanent's controller. Plain data.
type GameActionReplacement struct {
	Kind GameActionReplacementKind
	// Life is the total ReplaceLoseGameExileThisLifeBecomes sets.
	Life int
	// StartingLife makes the total "equal to your starting life total"
	// (Exquisite Archangel, S82 OLD batch 255, #5536): the table's starting
	// life (g.Settings.StartingLife), read as the replacement applies, in
	// place of Life.
	StartingLife bool
}

// CatalogGameActionReplacements returns the game-action replacements a
// battlefield permanent with this catalog ability key prints, or nil. Set
// at init from CardDef.GameActionReplacements; nil (no catalog) replaces
// nothing.
var CatalogGameActionReplacements func(key string) []GameActionReplacement

func init() {
	CatalogGameActionReplacements = func(key string) []GameActionReplacement {
		if d := catalogDef(key); d != nil {
			return d.GameActionReplacements
		}
		return nil
	}
}

// gameActionReplacementSourcesLocked lists the battlefield permanents p
// controls that print a replacement of `kind`, in battlefield order, with
// the row.
//
// Caller must hold g.mu.
func (g *Game) gameActionReplacementSourcesLocked(p *Player, kind GameActionReplacementKind) ([]uuid.UUID, []GameActionReplacement) {
	if p == nil || p.Eliminated || g.Battlefield == nil || CatalogGameActionReplacements == nil {
		return nil, nil
	}
	var ids []uuid.UUID
	var rows []GameActionReplacement
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != p.ID {
			continue
		}
		key := catalogAbilityKeyOf(c)
		if key == "" {
			continue
		}
		for _, r := range CatalogGameActionReplacements(key) {
			if r.Kind == kind {
				ids = append(ids, c.InstanceID)
				rows = append(rows, r)
				break
			}
		}
	}
	return ids, rows
}

// replaceLossLocked is the losing half: when p would lose the game for
// `cause` and a permanent p controls replaces that, the replacement is
// applied and the loss never happens (CR 614.6). Reports whether it was
// replaced. A concession is never replaced (CR 104.3a).
//
// Caller must hold g.mu (write).
func (g *Game) replaceLossLocked(p *Player, cause LossCause) bool {
	if p == nil || p.Eliminated || cause == LossConcede || g.State != StateActive {
		return false
	}
	// #6244: a resolution's one-shot "the next time you would lose the
	// game this turn, instead …" (next_loss_replacement.go).
	if g.replaceLossByRecordLocked(p, cause) {
		return true
	}
	ids, rows := g.gameActionReplacementSourcesLocked(p, ReplaceLoseGameExileThisLifeBecomes)
	if len(ids) == 0 {
		return false
	}
	source, row := ids[0], rows[0]
	g.EmitEvent(Event{
		Kind:   EventGameLossReplaced,
		Actor:  p.ID,
		Source: source,
		CardID: source,
		Label:  string(cause),
	})
	// "exile [this]": the ordinary exile route, so a replacement on the
	// move (a commander's owner offered the command zone, CR 903.9a) sees
	// it.
	if err := g.ExileCardForEffect(source); err != nil {
		return true
	}
	// "your life total becomes N": a gain or a loss of the difference
	// (CR 119.5), through the life window like every other.
	total := row.Life
	if row.StartingLife {
		total = g.Settings.StartingLife
	}
	if delta := total - p.Life; delta != 0 {
		_ = g.ChangePlayerLifeForEffect(source, p.ID, delta)
	}
	return true
}

// EventGameLossReplaced — a player who would have lost the game did not,
// because a replacement effect applied instead (The Golden Throne). Actor
// is the player, Source/CardID the permanent whose replacement applied,
// Label the LossCause that was replaced.
const EventGameLossReplaced EventKind = "game_loss_replaced"

// The planeswalk replacement's key. On-disk identity (a table waiting on
// the pick is a restore point): never renamed, never reused.
const planeswalkOrderTopTwoPickKey = "planechase/order-top-two"

var planeswalkOrderTopTwoThen OptionPickThen

func init() {
	planeswalkOrderTopTwoThen = RegisterOptionPickThen(planeswalkOrderTopTwoPickKey, planeswalkOrderTopTwoAnswered)
}

// replacePlaneswalkLocked is the planeswalking half: when `player` would
// planeswalk and permanents they control replace that, the first source's
// look-and-order is asked now and the planeswalk waits for the answer.
// Reports whether it paused. With fewer than two cards in the planar deck
// there is nothing to order, and the planeswalk goes on.
//
// Caller must hold g.mu (write).
func (g *Game) replacePlaneswalkLocked(player uuid.UUID) bool {
	ids, _ := g.gameActionReplacementSourcesLocked(g.playerByIDLocked(player), ReplacePlaneswalkOrderTopTwo)
	if len(ids) == 0 {
		return false
	}
	return g.askPlaneswalkOrderTopTwoLocked(player, ids)
}

// askPlaneswalkOrderTopTwoLocked asks `player` which of the top two cards
// of the planar deck goes on the bottom, for the first of `sources`; the
// rest ride the Carry, each applying once in turn (CR 614.5). The player
// looks at both: they become knowers of the two cards.
//
// Caller must hold g.mu (write).
func (g *Game) askPlaneswalkOrderTopTwoLocked(player uuid.UUID, sources []uuid.UUID) bool {
	deck := g.planarDeckForLocked(player)
	if deck == nil || len(deck.Cards) < 2 || len(sources) == 0 {
		return false
	}
	n := len(deck.Cards)
	top, second := &deck.Cards[n-1], &deck.Cards[n-2]
	top.AddKnower(player)
	second.AddKnower(player)
	carry := make([]uuid.UUID, 0, len(sources)+1)
	carry = append(carry, player)
	carry = append(carry, sources[1:]...)
	g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  player,
		Source:   sources[0],
		Question: "Before you planeswalk: the top two cards of your planar deck. Which goes on the bottom? The other stays on top, and you planeswalk to it.",
		Options: []ChoiceOption{
			{Label: top.Name + " (now on top) goes to the bottom"},
			{Label: second.Name + " (second) goes to the bottom"},
		},
		ThenKey: planeswalkOrderTopTwoThen,
		Carry:   carry,
	})
	return true
}

// planeswalkOrderTopTwoAnswered puts the chosen card on the bottom of the
// planar deck, leaves the other on top, and goes on: the next source's
// look-and-order, or the planeswalk. Unanswered (the player left), the
// deck is left as it is and the planeswalk goes on.
func planeswalkOrderTopTwoAnswered(g *Game, r OptionPicked) error {
	if g.Planechase == nil || len(r.Carry) < 1 {
		return nil
	}
	player := r.Carry[0]
	deck := g.planarDeckForLocked(player)
	if deck != nil && len(deck.Cards) >= 2 && (r.Index == 0 || r.Index == 1) {
		n := len(deck.Cards)
		id := deck.Cards[n-1-r.Index].InstanceID
		if c, err := deck.Remove(id); err == nil {
			deck.PushBottom(c)
		}
	}
	if rest := r.Carry[1:]; len(rest) > 0 && g.askPlaneswalkOrderTopTwoLocked(player, rest) {
		return nil
	}
	if !g.playerInGameLocked(player) {
		return nil
	}
	g.continuePlaneswalkLocked(player)
	return nil
}

// gameActionReplacementRows words a GameActionReplacements slot for the
// tile's ability rows (#2219): one sentence per entry.
func gameActionReplacementRows(v reflect.Value) []string {
	rows, _ := v.Interface().([]GameActionReplacement)
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		switch r.Kind {
		case ReplaceLoseGameExileThisLifeBecomes:
			if r.StartingLife {
				out = append(out, "If you would lose the game, instead exile this and your life total becomes equal to your starting life total.")
				break
			}
			out = append(out, "If you would lose the game, instead exile this and your life total becomes "+strconv.Itoa(r.Life)+".")
		case ReplacePlaneswalkOrderTopTwo:
			out = append(out, "If you would planeswalk, instead look at the top two cards of your planar deck, put one on the bottom and the other on top, then planeswalk.")
		default:
			out = append(out, "Replaces a game action.")
		}
	}
	return out
}

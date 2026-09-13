package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// proliferate.go — the Proliferate primitive (CR 701.27):
//
//	"Choose any number of permanents and/or players with counters on
//	 them, then give each another counter of each kind already
//	 there."
//
// The rule is one choice and one application. game.ProliferateForEffect
// is the application and takes the chosen lists verbatim; this file is
// the choice.
//
// Sandbox simplification — the choice is made FOR the proliferating
// player, by a deterministic beneficial pick (BeneficialProliferateChoice
// below), rather than prompted. "Any number of permanents and/or
// players" is a free-form multi-select over the whole board, which is a
// picker the client does not have and nothing else in the catalog needs
// yet; the S20 target picker is per-slot and legality-checked, which is
// the wrong shape. The pick below is what a player takes essentially
// every time — everything of yours that a counter helps, everything of
// theirs that a counter hurts — and the two halves of the rule are
// already split so a real picker only has to supply the lists.
//
// Where the auto-pick is observably different from paper: a player who
// wants to DECLINE a counter cannot (proliferating a Stun counter onto
// your own permanent to untap it more slowly is not a thing anyone
// wants, but "put another -1/-1 counter on my own persist creature to
// stop it coming back" is a real, rare line, and the auto-pick will not
// take it). Declared rather than silently wrong.

// Proliferate gives each chosen permanent and player one more counter
// of each kind already on it.
//
// With Cards and Players both nil the primitive picks for the
// proliferating player via BeneficialProliferateChoice. A card that
// knows exactly what it wants (or a test) sets them explicitly; an
// explicitly empty, non-nil slice means "choose nothing", which is a
// legal proliferate.
type Proliferate struct {
	// Controller is the player proliferating. Zero value means the
	// effect's controller, which is what every printed proliferate
	// wants.
	Controller uuid.UUID

	// Cards / Players are an explicit choice. Nil for both means
	// "pick for me".
	Cards   []uuid.UUID
	Players []uuid.UUID
}

func (p Proliferate) Apply(ctx *Context) error {
	controller := p.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	cards, players := p.Cards, p.Players
	if cards == nil && players == nil {
		cards, players = BeneficialProliferateChoice(ctx.Game, controller)
	}
	return ctx.Game.ProliferateForEffect(cards, players)
}

// harmfulCardCounters are the counter kinds a permanent's controller
// does not want more of. Everything else — +1/+1, loyalty, charge,
// shield, lore, defense on a battle you are not fighting — is either
// wanted or harmless, and proliferate gives one of EACH kind a
// permanent has, so a single unwanted kind rules the permanent out
// entirely rather than being skipped.
//
// Lore is a judgement call: another lore counter advances a Saga
// towards its final chapter and its sacrifice, which is usually the
// point (you want the chapter abilities) — it is treated as wanted.
var harmfulCardCounters = map[string]bool{
	game.CounterMinusOne: true,
	game.CounterStun:     true,
}

// harmfulPlayerCounters are the player-level counters you do not want
// on yourself and do want on an opponent.
var harmfulPlayerCounters = map[string]bool{
	game.CounterPoison: true,
	game.CounterRad:    true,
}

// BeneficialProliferateChoice picks the permanents and players a
// proliferate should choose on `controller`'s behalf: everything they
// control whose counters all help, plus every other permanent whose
// counters all hurt, and the same test applied to each player's own
// counters.
//
// Only things that already have at least one counter can be chosen
// (CR 701.27a), so an empty board or a board with no counters on it
// returns two empty lists and the proliferate is a legal no-op.
//
// Deterministic: battlefield order then seat order, so the same board
// always produces the same event stream.
func BeneficialProliferateChoice(g *game.Game, controller uuid.UUID) (cards []uuid.UUID, players []uuid.UUID) {
	for _, c := range g.BattlefieldCardsForEffect() {
		if len(c.Counters) == 0 {
			continue
		}
		if wantsMoreCounters(c.Counters, harmfulCardCounters, c.Controller == controller) {
			cards = append(cards, c.InstanceID)
		}
	}
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || len(p.Counters) == 0 {
			continue
		}
		if wantsMoreCounters(p.Counters, harmfulPlayerCounters, p.ID == controller) {
			players = append(players, p.ID)
		}
	}
	return cards, players
}

// wantsMoreCounters is the shared test both halves of the pick use.
// For something of yours (`mine`): choose it only when NO kind on it
// is harmful. For something that isn't: choose it only when EVERY
// kind on it is harmful — an opponent's creature with a -1/-1 counter
// is a fine choice; the same creature also carrying a +1/+1 counter
// is not, because proliferate would hand them both.
func wantsMoreCounters(counters map[string]int, harmful map[string]bool, mine bool) bool {
	seen := false
	for name, n := range counters {
		if n <= 0 {
			continue
		}
		seen = true
		if wanted := !harmful[name]; wanted != mine {
			// Mine and harmful, or theirs and helpful.
			return false
		}
	}
	return seen
}

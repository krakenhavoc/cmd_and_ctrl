package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jmp_b7_helpers.go — shared reads for the seventh Jumpstart deck-request
// batch (#4894, Jumpstart Azorius). Each helper is used by more than one
// card of the batch, or is the keyed continuation of one.

// jmpB7CreaturesYouControlWithFlying counts the creatures `player`
// controls that have flying right now, after the layers: Aerial Assault's
// and Aven Gagglemaster's "for each creature you control with flying",
// read as the spell or ability resolves.
func jmpB7CreaturesYouControlWithFlying(g *game.Game, player uuid.UUID) int {
	return len(permanentsControlledByMatching(g, player, And(Creature(), HasKeyword("flying"))))
}

// jmpB7AttackingFlyersYouControl counts the attacking creatures with
// flying that `player` controls: Inniaz's "three or more creatures you
// control with flying attack".
func jmpB7AttackingFlyersYouControl(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, id := range b13AttackingCreaturesYouControl(g, player) {
		if c, ok := g.LookupCardForEffect(id); ok && game.HasKeyword(&c, "flying") {
			n++
		}
	}
	return n
}

// --- Inniaz, the Gale Force: "each player gains control of a nonland
// permanent of your choice controlled by the player to their right" -----

// jmpB7InniazLabel names the control effects in the log.
const jmpB7InniazLabel = "Inniaz, the Gale Force — each player gains control of a nonland permanent controlled by the player to their right"

// jmpB7InniazOrder is the players the effect walks, in turn order from
// the ability's controller (CR 101.4's APNAP order), with each one's
// neighbour to the right: the previous player still in the game in turn
// order (game.AttackDirectionRight). A player with no one to their right
// (alone at the table) gains nothing.
func jmpB7InniazOrder(g *game.Game, controller uuid.UUID) (players, rights []uuid.UUID) {
	n := len(g.Seats)
	at := -1
	for i, p := range g.Seats {
		if p != nil && p.ID == controller {
			at = i
			break
		}
	}
	if at < 0 {
		return nil, nil
	}
	for k := 0; k < n; k++ {
		p := g.Seats[(at+k)%n]
		if p == nil || p.Eliminated {
			continue
		}
		right := g.NearestOpponentForEffect(p.ID, game.AttackDirectionRight)
		if right == uuid.Nil {
			continue
		}
		players = append(players, p.ID)
		rights = append(rights, right)
	}
	return players, rights
}

// jmpB7InniazAsk asks the controller for the next player's permanent,
// starting at index `from` of jmpB7InniazOrder. A player whose right-hand
// neighbour controls no nonland permanent is passed over: there is nothing
// to choose. `picked` rides the prompt as pairs of (player, permanent);
// once every player has been asked, every gain of control is registered
// together (jmpB7InniazApply).
//
// Not a target: the choices are made as the ability resolves (CR 608.2),
// so hexproof and shroud do not stop them, and nothing re-checks them
// against a later board, because nothing can happen in between.
func jmpB7InniazAsk(g *game.Game, source, controller uuid.UUID, from int, picked []uuid.UUID) error {
	players, rights := jmpB7InniazOrder(g, controller)
	for i := from; i < len(players); i++ {
		pool := NonlandPermanentsControlledBy(g, rights[i])
		if len(pool) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:   controller,
			Source:    source,
			Question:  "Inniaz, the Gale Force — choose a nonland permanent " + attackTargetName(g, rights[i]) + " controls; " + attackTargetName(g, players[i]) + " gains control of it",
			Cards:     pool,
			Min:       1,
			Max:       1,
			ThenKey:   jmpB7InniazPicked,
			LiveZone:  game.ZoneBattlefield,
			Carry:     append([]uuid.UUID(nil), picked...),
			CarryNums: []int{i},
		})
		return nil
	}
	jmpB7InniazApply(g, source, picked)
	return nil
}

// jmpB7InniazPicked records one answer and asks the next player's
// permanent (#4253: keyed, so a table waiting on the choice is a restore
// point). The answer must still be a nonland permanent controlled by that
// player's right-hand neighbour; a prompt that ended unanswered (its
// chooser left the game, and with them the ability, CR 800.4a) does
// nothing more.
var jmpB7InniazPicked game.ChooseCardsThen

func init() {
	jmpB7InniazPicked = game.RegisterChooseCardsThen("choose-cards/jmp-inniaz-gale-force-control", jmpB7InniazPickedBody)
}

func jmpB7InniazPickedBody(g *game.Game, r game.CardsChosen) error {
	if !r.Answered || len(r.Nums) != 1 {
		return nil
	}
	i := r.Nums[0]
	players, rights := jmpB7InniazOrder(g, r.Chooser)
	if i < 0 || i >= len(players) {
		return nil
	}
	picked := append([]uuid.UUID(nil), r.Carry...)
	if len(r.Picked) == 1 {
		if c, ok := g.LookupCardForEffect(r.Picked[0]); ok && onBattlefield(g, c.InstanceID) &&
			c.Controller == rights[i] && !c.IsLand() {
			picked = append(picked, players[i], c.InstanceID)
		}
	}
	return jmpB7InniazAsk(g, r.Source, r.Chooser, i+1, picked)
}

// jmpB7InniazApply registers every gain of control the choices named, in
// one go, after the last choice: the players gain control simultaneously,
// so a permanent chosen for one player is still where it was while the
// next player's is chosen. Each is indefinite (CR 611.2a: no duration is
// stated) and pinned to the permanent as it is now.
func jmpB7InniazApply(g *game.Game, source uuid.UUID, picked []uuid.UUID) {
	for k := 0; k+1 < len(picked); k += 2 {
		g.GainControlForEffect(source, picked[k+1], picked[k], game.IndefiniteDuration(), jmpB7InniazLabel)
	}
}

// jmpB7FirstLegalTargetGains is "<target> gains <keyword> until end of
// turn" for the first target still legal as the ability resolves (CR
// 608.2b): Alabaster Mage's lifelink, Duelist's Heritage's double strike.
func jmpB7FirstLegalTargetGains(label string, keywords ...string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return GrantKeywordUntilEOT{Target: id, Keywords: keywords, Label: label}.Apply(ctx)
	}
}

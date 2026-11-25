package game

import (
	"errors"
	"strconv"

	"github.com/google/uuid"
)

// blocks_by_piles.go — Camouflage (#6200): "This turn, instead of
// declaring blockers, each defending player chooses any number of
// creatures they control and divides them into a number of piles equal
// to the number of attacking creatures for whom that player is the
// defending player. Creatures those players control that can block
// additional creatures may likewise be put into additional piles.
// Assign each pile to a different one of those attacking creatures at
// random. Each creature in a pile that can block the creature that pile
// is assigned to does so. (Piles can be empty.)"
//
//	509.1. First, the defending player declares blockers. This
//	       turn-based action doesn't use the stack.
//	700.3a Each of the affected objects must be put into exactly one of
//	       those piles, unless the effect specifies otherwise.
//	700.3d A pile can contain zero or more objects.
//	101.4. If multiple players would make choices and/or take actions at
//	       the same time, the active player … makes any choices required,
//	       then the next player in turn order …
//
// THE RECORD. ModBlocksByPiles, a ScopeGame rule record for the rest of
// the turn (CR 514.2 sweeps it). While one is live, the declare-blockers
// turn-based action is replaced for every defending player: the block
// verbs are refused (ErrBlocksDecidedByPiles) and nobody is completed as
// "declared, none" as the step begins. Instead, in APNAP order
// (CR 101.4), each defending player is asked for their piles, one keyed
// choose_cards prompt per pile over the creatures they control: a
// creature already in as many piles as it can block attackers
// (BlockCapacity, CR 509.1a) is not offered again, so an ordinary
// creature is in at most one pile (CR 700.3a) and one that can block an
// additional creature may be in two (the card's own exception).
//
// THE ASSIGNMENT. Once a player's last pile is chosen, their attackers
// (every creature for whom they are the defending player) are shuffled
// with the table's seeded pick stream, and pile i is assigned to the
// i-th of them. Each creature in a pile that can block its pile's
// attacker — untapped, and passing the per-pair restrictions and
// evasion (BlockPairRefusalLocked) — blocks it. A whole-declaration
// bound the piles break (menace's two blockers, "can't be blocked by
// more than one creature", a whole-combat limit) is not something the
// creatures in that pile can do, so that pile's blocks are taken out and
// the rest are judged again. Then the player's declaration completes
// exactly as finish_blocks completes one, and the next defending player
// is asked. The last one's completion hands the active player priority
// (CR 509.2).
//
// Every leg is a registered continuation (blocksByPilesPile) carrying the
// piles so far as plain data, so a table waiting on a pile is a restore
// point. The prompts are ordinary choose_cards picks: the client and the
// bot already answer them.

// ModBlocksByPiles is Camouflage's "this turn, instead of declaring
// blockers, …": a ScopeGame rule record for the rest of the turn. Reads
// nothing. Read by the declare-blockers step (beginBlockDeclarationLocked)
// and the block verbs. An on-disk identity: never renamed, never reused.
const ModBlocksByPiles ModKind = "blocksByPiles"

// ErrBlocksDecidedByPiles refuses a block verb while a ModBlocksByPiles
// record is live: the defending players don't declare blockers this
// turn, their piles do.
var ErrBlocksDecidedByPiles = errors.New("game: this turn, blocks are decided by piles instead of being declared")

// blocksByPilesPile is the keyed answer to one pile. An on-disk key:
// never renamed, never reused.
var blocksByPilesPile ChooseCardsThen

func init() {
	blocksByPilesPile = RegisterChooseCardsThen("choose-cards/blocks-by-piles", blocksByPilesPileAnswered)
}

// BlocksByPilesThisTurnForEffect registers Camouflage's record for the
// rest of the turn. Reports whether a record was written.
//
// Caller must hold g.mu (write).
func (g *Game) BlocksByPilesThisTurnForEffect(sourceID, controller uuid.UUID, label string) bool {
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, controller,
		[]Mod{{Kind: ModBlocksByPiles}}, g.UntilEndOfTurnDuration(), label)
}

// blocksByPilesLocked reports whether a live record replaces this turn's
// block declaration, and the card that made it.
//
// Caller must hold g.mu (read or write).
func (g *Game) blocksByPilesLocked() (bool, uuid.UUID) {
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if scopedEffectHasMod(e, ModBlocksByPiles) {
			return true, e.Source.ID
		}
	}
	return false, uuid.Nil
}

// BlocksByPilesForEffect is the exported reader, for card files, the
// view and tests. Caller must hold g.mu.
func (g *Game) BlocksByPilesForEffect() bool {
	on, _ := g.blocksByPilesLocked()
	return on
}

// pileAttackersLocked is every creature for whom `seat` is the defending
// player, in battlefield order.
//
// Caller must hold g.mu (read or write).
func (g *Game) pileAttackersLocked(seat uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget != uuid.Nil && g.defendingPlayerForAttackerLocked(c) == seat {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// pileCandidatesLocked is the creatures `seat` controls that may still go
// into another pile: those in fewer piles than the attackers they can
// block. `used` counts the piles each is in so far.
//
// Caller must hold g.mu (read or write), with fresh layers.
func (g *Game) pileCandidatesLocked(seat uuid.UUID, used map[uuid.UUID]int) []uuid.UUID {
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != seat || !c.IsCreature() {
			continue
		}
		if capacity := BlockCapacity(c); capacity != 0 && used[c.InstanceID] >= capacity {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// beginBlocksByPilesLocked starts the piles as the declare-blockers step
// begins: the first defending player, in APNAP order, who has not
// declared. Reports whether a record is live (the caller then skips the
// ordinary "no legal block" completion: no defender declares this turn).
//
// Caller must hold g.mu (write), with fresh layers.
func (g *Game) beginBlocksByPilesLocked() bool {
	on, source := g.blocksByPilesLocked()
	if !on {
		return false
	}
	g.askNextPilesLocked(uuid.Nil, source)
	return true
}

// askNextPilesLocked asks the next defending player after `after` (from
// the first, when `after` is uuid.Nil) for their first pile, completing
// in turn each one who has nothing to divide. With nobody left to ask,
// the declaration as a whole is closed and the active player receives
// priority.
//
// Caller must hold g.mu (write).
func (g *Game) askNextPilesLocked(after, source uuid.UUID) {
	g.RecomputeLayersIfStaleLocked()
	seats := g.defendingSeatsAPNAPLocked()
	start := 0
	if after != uuid.Nil {
		start = len(seats)
		for i, s := range seats {
			if s == after {
				start = i + 1
				break
			}
		}
	}
	for _, seat := range seats[start:] {
		if g.blockDeclarationDoneLocked(seat) {
			continue
		}
		n := len(g.pileAttackersLocked(seat))
		if n > 0 && g.askPileLocked(seat, source, n, nil, nil) {
			return
		}
		// Nothing to divide: every pile is empty and nothing blocks.
		g.completeBlockDeclarationLocked(seat)
	}
	if g.blockPriorityParkedLocked() {
		g.closeBlockDeclarationIfCompleteLocked()
	}
}

// askPileLocked asks `seat` for pile len(sizes)+1 of n. `members` are the
// piles chosen so far, flattened, and `sizes` their lengths. Reports
// whether a prompt was queued: false when the player controls no
// creature that can go into another pile, which leaves the rest of the
// piles empty.
//
// Caller must hold g.mu (write), with fresh layers.
func (g *Game) askPileLocked(seat, source uuid.UUID, n int, members []uuid.UUID, sizes []int) bool {
	used := map[uuid.UUID]int{}
	for _, id := range members {
		used[id]++
	}
	candidates := g.pileCandidatesLocked(seat, used)
	if len(candidates) == 0 {
		return false
	}
	nums := append([]int{n}, sizes...)
	return g.QueueChooseCardsForEffect(ChooseCardsPrompt{
		Chooser:    seat,
		FromPlayer: seat,
		Source:     source,
		Question: "Camouflage: choose the creatures for pile " + strconv.Itoa(len(sizes)+1) + " of " +
			strconv.Itoa(n) + ". Each pile is assigned to a different attacking creature at random.",
		Cards:     candidates,
		Min:       0,
		Max:       len(candidates),
		Zone:      ZoneBattlefield,
		ThenKey:   blocksByPilesPile,
		Carry:     members,
		CarryNums: nums,
	}) != uuid.Nil
}

// blocksByPilesPileAnswered records one pile and asks for the next, or,
// after the last, assigns the piles and completes the player's
// declaration. A prompt that ended unanswered (its chooser left) asks
// nothing more of them: their declaration does not happen (CR 800.4a).
func blocksByPilesPileAnswered(g *Game, r CardsChosen) error {
	seat := r.Chooser
	if len(r.Nums) == 0 {
		return nil
	}
	n, sizes := r.Nums[0], copyInts(r.Nums[1:])
	members := append(copyUUIDs(r.Carry), r.Picked...)
	sizes = append(sizes, len(r.Picked))
	if r.Answered && len(sizes) < n && g.playerInGameLocked(seat) {
		g.RecomputeLayersIfStaleLocked()
		if g.askPileLocked(seat, r.Source, n, members, sizes) {
			return nil
		}
	}
	if g.playerInGameLocked(seat) {
		g.assignPilesLocked(seat, r.Source, members, sizes)
		g.completeBlockDeclarationLocked(seat)
	}
	g.askNextPilesLocked(seat, r.Source)
	return nil
}

// assignPilesLocked assigns `seat`'s piles to their attackers at random
// and stages the blocks the piles make.
//
// Caller must hold g.mu (write).
func (g *Game) assignPilesLocked(seat, source uuid.UUID, members []uuid.UUID, sizes []int) {
	if g.State != StateActive || g.Turn.Step != StepDeclareBlockers {
		return
	}
	g.RecomputeLayersIfStaleLocked()
	attackers := g.pickAtRandomLocked(rngStream{kind: rngStreamPick, player: seat, source: source},
		g.pileAttackersLocked(seat), len(g.pileAttackersLocked(seat)))
	// piles[i] are the blocks pile i makes against its attacker.
	var piles [][]BlockDeclaration
	at := 0
	for i, size := range sizes {
		end := at + size
		if end > len(members) {
			end = len(members)
		}
		pile := members[at:end]
		at = end
		if i >= len(attackers) {
			break
		}
		atk := findBattlefieldCard(g, attackers[i])
		var decls []BlockDeclaration
		for _, id := range pile {
			b := findBattlefieldCard(g, id)
			if b == nil || b.Controller != seat || !b.IsCreature() || b.Tapped || atk == nil {
				continue
			}
			if !g.BlockPairRefusalLocked(atk, b).Legal() {
				continue
			}
			decls = append(decls, BlockDeclaration{Blocker: id, Attacker: atk.InstanceID})
		}
		if len(decls) > 0 {
			piles = append(piles, decls)
		}
	}
	// Judge the piles' blocks as one declaration; a pile whose blocks
	// break a whole-declaration bound is taken out and the rest judged
	// again.
	for len(piles) > 0 {
		var all []BlockDeclaration
		for _, p := range piles {
			all = append(all, p...)
		}
		_, after, err := g.checkBlockRestrictionsLocked(g.currentBlockAssignmentLocked(), all)
		if err == nil {
			for _, d := range all {
				if b := findBattlefieldCard(g, d.Blocker); b != nil {
					b.setBlockingSet(after[d.Blocker])
					b.AttackingTarget = uuid.Nil
				}
			}
			return
		}
		drop := len(piles) - 1
		var refused *BlockRefusedError
		if errors.As(err, &refused) {
			for i, p := range piles {
				for _, d := range p {
					if d.Attacker == refused.Attacker || d.Blocker == refused.Blocker {
						drop = i
					}
				}
			}
		}
		piles = append(piles[:drop], piles[drop+1:]...)
	}
}

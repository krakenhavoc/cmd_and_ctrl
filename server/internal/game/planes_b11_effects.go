package game

import "github.com/google/uuid"

// planes_b11_effects.go — the planar die count a gold-bordered plane reads
// (ADR 0216 slice 4, #4870; Li'l Giri Saves the Day, sAnS mERcY).

// PlanarDieRollsThisTurnForEffect is "the number of times you've rolled the
// planar die this turn" (CR 901.9, 706.7): the planar die special actions
// taken this turn, the roll that made chaos ensue included. Only the active
// player takes the special action (CR 901.9), who is the planar controller.
//
// Caller must hold g.mu (an effect resolving does).
func (g *Game) PlanarDieRollsThisTurnForEffect() int { return g.planarDieTakenThisTurnLocked() }

// StartEnginesForEffect is the keyword action "start your engines!" done by
// an effect (CR 702.179b): a player with no speed gets speed 1, a player who
// has speed is left alone. Ghirapur Grand Prix's "all players start their
// engines" calls it for every seat. A player who has left the game is
// skipped (setSpeedLocked refuses them).
//
// Caller must hold g.mu in write mode (an effect resolving does).
func (g *Game) StartEnginesForEffect(player uuid.UUID) {
	if g.SpeedOf(player) > 0 {
		return
	}
	g.setSpeedLocked(player, 1)
}

// PlayersWithTheHighestSpeedForEffect lists the players still in the game
// whose speed is the highest among them, seat order (ties all count). A
// player with no speed has speed 0 (CR 702.179f).
//
// Caller must hold g.mu.
func (g *Game) PlayersWithTheHighestSpeedForEffect() []uuid.UUID {
	best := -1
	var out []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		switch s := g.SpeedOf(p.ID); {
		case s > best:
			best, out = s, []uuid.UUID{p.ID}
		case s == best:
			out = append(out, p.ID)
		}
	}
	return out
}

// GrantCantAttackPlayerOnlyNextTurnForEffect is "`attacker` can't attack
// `protected` during their next turn" for a clause that names the player and
// not planeswalkers (City Hall's "can't attack you", CR 508.1c): the
// restriction covers the player alone, starts at the attacker's next turn and
// ends with it. Same no-ops and caller contract as
// GrantCantAttackPlayerForEffect (caller holds g.mu in write mode).
func (g *Game) GrantCantAttackPlayerOnlyNextTurnForEffect(attacker, protected uuid.UUID, label string, source uuid.UUID) {
	g.grantCantAttackLocked(attacker, protected, CantAttackScope{PlayerOnly: true}, g.turnsBegunForLocked(attacker)+1,
		g.UntilEndOfYourNextTurnDuration(attacker), label, source)
}

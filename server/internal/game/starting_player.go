package game

const startingPlayerDieSides = 20

// rollStartingSeatLocked has every seated player roll a d20. Only players
// tied for the highest result reroll, repeatedly, until one winner remains.
// The ordinary random-effect path emits every roll into the public event log
// and gives seeded games the same winner on replay.
//
// Caller must hold g.mu.
func (g *Game) rollStartingSeatLocked() int {
	return chooseStartingSeat(len(g.Seats), func(seat int) int {
		rolls, _ := g.RollDiceForEffect(RandomDraw{Player: g.Seats[seat].ID}, startingPlayerDieSides, 1)
		return rolls[0]
	})
}

// chooseStartingSeat is the tie-handling half kept separate from randomness
// so the rule can be proved without searching for a seed that produces each
// sequence. roll is called once for every contender in each round.
func chooseStartingSeat(seats int, roll func(seat int) int) int {
	contenders := make([]int, seats)
	for seat := range contenders {
		contenders[seat] = seat
	}
	for len(contenders) > 1 {
		highest := 0
		leaders := make([]int, 0, len(contenders))
		for _, seat := range contenders {
			result := roll(seat)
			switch {
			case result > highest:
				highest = result
				leaders = append(leaders[:0], seat)
			case result == highest:
				leaders = append(leaders, seat)
			}
		}
		contenders = leaders
	}
	return contenders[0]
}

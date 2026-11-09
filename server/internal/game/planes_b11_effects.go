package game

// planes_b11_effects.go — the planar die count a gold-bordered plane reads
// (ADR 0216 slice 4, #4870; Li'l Giri Saves the Day, sAnS mERcY).

// PlanarDieRollsThisTurnForEffect is "the number of times you've rolled the
// planar die this turn" (CR 901.9, 706.7): the planar die special actions
// taken this turn, the roll that made chaos ensue included. Only the active
// player takes the special action (CR 901.9), who is the planar controller.
//
// Caller must hold g.mu (an effect resolving does).
func (g *Game) PlanarDieRollsThisTurnForEffect() int { return g.planarDieTakenThisTurnLocked() }

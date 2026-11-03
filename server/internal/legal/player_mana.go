package legal

import "strconv"

// player_mana.go — #4885: the moves for a mana ability a PLAYER has,
// bound to no card (Channel's "pay 1 life: add {C}" until end of turn).
//
// One move per live grant, on the same activate_mana_ability verb a
// card's mana ability uses, naming the grant by its source in
// `player_ability`. The life is the move's Cost, as Mana Confluence's
// is. The move has no Source: it belongs to no card, so the per-card
// digest skips it and a per-card expansion never asks for it.

// playerManaParams is activate_mana_ability's payload for a player
// mana ability.
type playerManaParams struct {
	PlayerAbility string `json:"player_ability"`
}

// playerManaMoves offers each live player mana ability the seat can pay
// for. Called from manaMoves, so it is offered exactly when a card's
// mana abilities are.
func (e *enumerator) playerManaMoves() {
	e.leave()
	for _, o := range e.g.PlayerManaAbilitiesLocked(e.seat) {
		if !o.Payable {
			continue
		}
		label := o.Label + ": pay " + lifeWord(o.LifeCost) + ", add " + o.Produced
		e.add(Move{
			Type:   TypeActivateManaAbility,
			Player: e.seat,
			Kind:   KindMana,
			Label:  label,
			Cost:   moveCost(o.LifeCost, 0),
			Params: mustJSON(playerManaParams{PlayerAbility: o.Source.String()}),
		})
	}
}

// lifeWord is "1 life" / "N life".
func lifeWord(n int) string {
	return strconv.Itoa(n) + " life"
}

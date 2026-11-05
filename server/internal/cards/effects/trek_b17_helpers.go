package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// trek_b17_helpers.go — shared helpers for Star Trek batch 17 (#4973).

// trekB17NoncreatureNonlandCardsInYourGraveyard is Ravenous Cosmozoan's
// discount count: "for each noncreature, nonland card in your graveyard".
// The spell being cast is on the stack (CR 601.2a) and is never counted.
func trekB17NoncreatureNonlandCardsInYourGraveyard(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	p := q.Game.PlayerByIDForEffect(q.Controller)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID != q.Card.InstanceID && !c.IsCreature() && !c.IsLand() {
			n++
		}
	}
	return n
}

// trekB17TreasureTokensYouControl is "Treasure tokens" as Quark, Humble
// Proprietor's controller can sacrifice them: Treasure tokens that player
// controls (CR 701.21a). The subtype is read through the layered view.
func trekB17TreasureTokensYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.IsToken() && target.HasSubtype("Treasure")
}

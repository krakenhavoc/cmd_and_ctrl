package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// soc_b37_helpers.go — shared pieces of Secrets of Strixhaven batch 37
// (#4456).

// socB37PutTargetGraveyardCardOnBottom is "put target card from a graveyard
// on the bottom of its owner's library": the first still-legal target goes to
// the bottom of its owner's library (Grazing Kelpie, Sundering Archaic).
func socB37PutTargetGraveyardCardOnBottom(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return PutIntoLibrary{Card: t.ID, ToBottom: true}.Apply(ctx)
	}
	return nil
}

// socB37InstantAndSorceryCardsInYourGraveyard is the amount behind "for each
// instant and sorcery card in your graveyard", never counting the card being
// cast.
func socB37InstantAndSorceryCardsInYourGraveyard(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	p := q.Game.PlayerByIDForEffect(q.Controller)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID != q.Card.InstanceID && (c.IsInstant() || c.IsSorcery()) {
			n++
		}
	}
	return n
}

// socB37ExileTheLegalTarget exiles the first still-legal card target (Kaya,
// Bane of the Dead's −3, Sundering Archaic's converge trigger).
func socB37ExileTheLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return ExileTarget{Target: t.ID}.Apply(ctx)
		}
	}
	return nil
}

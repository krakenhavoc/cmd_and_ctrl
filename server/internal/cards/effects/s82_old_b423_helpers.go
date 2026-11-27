package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b423_helpers.go — S82 OLD batch 423 (Return to Ravnica, #5566): the
// detain cards' shared bodies and Tablet of the Guilds' two-colour reads.

// OldB423DetainTargets is "detain [each] target <permanent>" as an ability's
// whole effect: every target that is still legal as it resolves (CR 608.2b) is
// detained together, until the controller's next turn (CR 701.35a). With no
// legal target nothing happens.
func OldB423DetainTargets(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		ids := map[uuid.UUID]bool{}
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				ids[t.ID] = true
			}
		}
		if len(ids) == 0 {
			return nil
		}
		return Detain{
			Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return ids[c.InstanceID] },
			Label: label,
		}.Apply(ctx)
	}
}

// OldB423SpellChosenColors is how many of the colours chosen for source (the
// concatenated letters of Card.ChosenColor) the spell the cast event names is.
func OldB423SpellChosenColors(ev game.Event, source *game.Card, g *game.Game) int {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller || source.ChosenColor == "" {
		return 0
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return 0
	}
	n := 0
	for _, letter := range source.ChosenColor {
		if spell.HasColor(string(letter)) {
			n++
		}
	}
	return n
}

// OldB423PlayerLostLifeThisTurn is the target clause "target player who lost
// life this turn": the turn tally counts every point lost as it happens.
func OldB423PlayerLostLifeThisTurn() PlayerPredicate {
	return func(g *game.Game, _ uuid.UUID, p *game.Player) bool {
		return g.TurnTallyFor(p.ID).LifeLost > 0
	}
}

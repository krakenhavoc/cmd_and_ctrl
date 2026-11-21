package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b11_helpers.go — shared pieces for the S82 GAP batch 11 cards
// (Innistrad Remastered creatures, #5948).

// instantAndSorceryCardsInYourGraveyard is the amount behind "for each
// instant and sorcery card in your graveyard" (Bedlam Reveler): printed
// card types (a card in a graveyard has no layers), never counting the
// card being cast.
func InstantAndSorceryCardsInYourGraveyard(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	p := q.Game.PlayerByIDForEffect(q.Controller)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID != q.Card.InstanceID && (c.HasCardType("instant") || c.HasCardType("sorcery")) {
			n++
		}
	}
	return n
}

// scheduleManaAtYourNextMainPhase is "add <mana> at the beginning of your
// next main phase this turn" (Conduit of Storms, Conduit of Emrakul): a
// delayed trigger bound to the next main phase still to begin this turn
// (CR 603.7). A turn with no main phase left adds nothing. The mana is
// the body's, keyed per card.
func ScheduleManaAtYourNextMainPhase(g *game.Game, item *game.StackItem, label string, body game.BodyRef) error {
	step, phase, ok := g.NextMainPhaseThisTurnForEffect()
	if !ok {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:      step,
		OnPhase: phase,
		Label:   label,
		Body:    body,
	}.Apply(NewContext(g, item))
}

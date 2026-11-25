package game

import "github.com/google/uuid"

// chosen_name_set.go — a card name chosen by a keyed pick, stored on the
// permanent that asked (CR 614.12). The free-text prompt (choose_card_name.go)
// stores its answer itself; a name chosen from a LIST the card built, as
// Alhammarret, High Arbiter's "the name of a nonland card revealed this way"
// is, comes back through an option pick and lands here.

// SetChosenNameForEffect stores `name` as the card name chosen for the
// permanent `sourceID` on the battlefield, as the free-text answer does, and
// wakes the layer cache for a static that reads it. A permanent that has left
// takes nothing: the choice has nowhere to land (CR 608.2). Caller must hold
// g.mu.
func (g *Game) SetChosenNameForEffect(sourceID uuid.UUID, name string) {
	i := findCardOnBattlefield(g, sourceID)
	if i < 0 || name == "" {
		return
	}
	g.Battlefield.Cards[i].ChosenName = name
	g.layerVersion.Add(1)
	g.EmitEvent(Event{
		Kind:   EventCardNameChosen,
		Actor:  g.Battlefield.Cards[i].Controller,
		CardID: sourceID,
		Label:  name,
	})
}

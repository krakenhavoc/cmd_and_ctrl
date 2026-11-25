package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b296_helpers.go — shared pieces of Oath of the Gatewatch (devoid),
// S82 OLD batch 296.

// OldB296ControlsAnotherColorlessCreature reports whether `source`'s
// controller controls a creature other than `source` that is colorless as the
// layers have it (Eldrazi Aggressor, Flayer Drone).
func OldB296ControlsAnotherColorlessCreature(g *game.Game, source *game.Card) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != source.InstanceID && c.Controller == source.Controller && c.IsCreature() && c.IsColorless() {
			return true
		}
	}
	return false
}

// OldB296ColorlessManaRestriction is Cultivator Drone's "Spend this mana only to
// cast a colorless spell, activate an ability of a colorless permanent, or pay a
// cost that contains {C}" (CR 106.6, #6259): one tag, an OR of the three
// clauses.
func OldB296ColorlessManaRestriction() string {
	return game.ManaRestrictAnyOf(
		[]string{game.ManaRestrictCast, game.ManaRestrictColorless},
		[]string{game.ManaRestrictActivate, game.ManaRestrictColorless},
		[]string{game.ManaRestrictColorlessCost},
	)
}

// OldB296TargetOpponentExilesFromHand is "Target opponent exiles a card from
// their hand" (Mindmelter): the targeted opponent picks which (CR 406.3).
// With no card in hand nothing happens.
func OldB296TargetOpponentExilesFromHand(g *game.Game, item *game.StackItem) error {
	id := TargetedPlayerOf(NewContext(g, item))
	p := g.PlayerByIDForEffect(id)
	if id == uuid.Nil || p == nil || p.Hand == nil || p.Hand.Size() == 0 {
		return nil
	}
	cards := make([]uuid.UUID, 0, p.Hand.Size())
	for _, c := range p.Hand.Cards {
		cards = append(cards, c.InstanceID)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    id,
		FromPlayer: id,
		Source:     item.SourceCardID,
		Question:   "Mindmelter — exile a card from your hand",
		Cards:      cards,
		Min:        1,
		Max:        1,
		ThenKey:    stdB40ExileFromHand,
		LiveZone:   game.ZoneHand,
	})
	return nil
}

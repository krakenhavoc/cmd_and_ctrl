package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b295_helpers.go — shared pieces of Oath of the Gatewatch (devoid),
// S82 OLD batch 295.

// OldB295TargetOpponentExilesFromHand is "Target opponent exiles N cards from
// their hand": the targeted player picks which (CR 406.3), as many as they
// hold up to N. Witness the End's first half; the pick is the plain exile
// answer Mindleech Ghoul uses.
func OldB295TargetOpponentExilesFromHand(g *game.Game, item *game.StackItem, n int) error {
	ctx := NewContext(g, item)
	id := TargetedPlayerOf(ctx)
	p := g.PlayerByIDForEffect(id)
	if id == uuid.Nil || p == nil || p.Hand == nil || p.Hand.Size() == 0 {
		return nil
	}
	cards := make([]uuid.UUID, 0, p.Hand.Size())
	for _, c := range p.Hand.Cards {
		cards = append(cards, c.InstanceID)
	}
	k := min(n, len(cards))
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    id,
		FromPlayer: id,
		Source:     item.SourceCardID,
		Question:   "Witness the End — exile cards from your hand",
		Cards:      cards,
		Min:        k,
		Max:        k,
		ThenKey:    stdB40ExileFromHand,
		LiveZone:   game.ZoneHand,
	})
	return nil
}

// OldB295DamagedCreatureControllerLosesLife is Visions of Brutality's "Whenever
// enchanted creature deals damage, its controller loses that much life": the
// controller of the creature that dealt it, as it was when it did (CR 608.2h,
// last-known information), loses life equal to the damage dealt.
func OldB295DamagedCreatureControllerLosesLife(g *game.Game, item *game.StackItem) error {
	ev := item.Trigger.Event
	if ev.Amount <= 0 {
		return nil
	}
	var controller uuid.UUID
	if c, ok := g.LookupCardForEffect(ev.Source); ok && g.Battlefield != nil && g.Battlefield.Contains(ev.Source) {
		controller = c.Controller
	} else if info, ok := g.LastKnownPermanentForEffect(ev.Source); ok {
		controller = info.Controller
	} else if c, ok := g.LookupCardForEffect(ev.Source); ok {
		controller = c.Controller
	}
	if controller == uuid.Nil {
		return nil
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, controller, -ev.Amount)
}

// OldB295YouControlAnotherColorlessCreature is "you control another
// colorless creature": a creature other than the source, read through the
// layers, with no colour.
func OldB295YouControlAnotherColorlessCreature(g *game.Game, source *game.Card) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != source.InstanceID && c.Controller == source.Controller && c.IsCreature() && c.IsColorless() {
			return true
		}
	}
	return false
}

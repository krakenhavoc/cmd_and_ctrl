package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Key to the Vault — Legendary Artifact — Equipment {1}{U}:
//
//	"Whenever equipped creature deals combat damage to a player, look
//	 at that many cards from the top of your library. You may exile a
//	 nonland card from among them. Put the rest on the bottom of your
//	 library in a random order. You may cast the exiled card without
//	 paying its mana cost.
//	 Equip {2}{U}"
//
// Sword of Feast and Famine's equipped-creature damage trigger, with
// Sunbird's Invocation's look-choose-exile-bottom-cast sequence behind
// it. The count is the damage dealt, read off the triggering event. The
// chosen card is exiled first, the rest go to the bottom in a random
// order, and then the free cast is offered as part of the resolution
// (CR 608.2g): the permission lapses on the controller's next priority
// pass and a card that is not cast stays in exile, as printed.
//
// No simplification.
func init() {
	const label = "The Key to the Vault — look at that many cards, exile a nonland card, cast it free"
	Register(Spec{
		OracleID:     "7a75d565-b0a7-46bf-94a0-c9255c8abd6d",
		Name:         "The Key to the Vault",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key:    label,
			Effect: keyToTheVaultLook,
		}},
		Activated: []ActivatedAbility{EquipAbility("{2}{U}")},
	})
}

func keyToTheVaultLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	controller, source := item.Controller, item.SourceCardID
	n := ctx.Trigger().Event.Amount
	if n <= 0 {
		return nil
	}
	looked := g.LookAtTopOfLibraryForEffect(controller, n)
	if len(looked) == 0 {
		return nil
	}
	var eligible []uuid.UUID
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
			eligible = append(eligible, id)
		}
	}
	finish := func(g *game.Game, picked []uuid.UUID) error {
		var chosen uuid.UUID
		if len(picked) > 0 {
			chosen = picked[0]
		}
		var rest []uuid.UUID
		for _, id := range cardsStillInALibrary(g, looked) {
			if id != chosen {
				rest = append(rest, id)
			}
		}
		bottomRestThenCast := func(g *game.Game, exiled bool) error {
			if err := g.PutOnBottomInRandomOrderForEffect(controller, game.ZoneLibrary, rest); err != nil {
				return err
			}
			if exiled {
				grantFreeCasts(g, controller, source, "The Key to the Vault", game.TimingFlash, game.LapseStaysInExile, []uuid.UUID{chosen})
			}
			return nil
		}
		if chosen == uuid.Nil {
			return bottomRestThenCast(g, false)
		}
		return g.ExileCardThenForEffect(chosen, bottomRestThenCast)
	}
	if len(eligible) == 0 {
		return finish(g, nil)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   source,
		Question: "The Key to the Vault — exile a nonland card and cast it without paying its mana cost?",
		Cards:    eligible,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneLibrary,
		Then:     finish,
	})
	return nil
}

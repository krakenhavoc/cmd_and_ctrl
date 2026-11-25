package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kheru Mind-Eater — Creature — Vampire {2}{B}, 1/3 (Commander 2017, #5533):
//
//	"Menace
//	 Whenever this creature deals combat damage to a player, that player
//	 exiles a card from their hand face down.
//	 You may look at cards exiled with this creature, and you may play lands
//	 and cast spells from among those cards."
//
// Fiend of the Shadows's choose-and-exile, with the cards exiled face down and
// linked to this object (CR 607.2a): the damaged player picks the card (an
// empty hand exiles nothing, CR 609.3), and it lands FaceDownHidden with
// Card.ExiledWith naming the Mind-Eater, so only its controller may look at it
// (CR 406.3). Intellect Devourer's standing permission over exile, narrowed to
// the cards exiled with this source, lets its controller play a land (using
// the land drop) or cast a spell from among them for its own cost. The link
// is to the object, so a Mind-Eater that left and came back has no cards, and
// a new controller of the same Mind-Eater inherits the permission.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dfea918b-5702-4c12-ba36-e6aa4ace09c7",
		Name:            "Kheru Mind-Eater",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Kheru Mind-Eater — that player exiles a card from their hand face down",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					victim := item.Trigger.Event.Target
					p := g.PlayerByIDForEffect(victim)
					if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
						return nil
					}
					cards := make([]uuid.UUID, 0, p.Hand.Size())
					for _, c := range p.Hand.Cards {
						cards = append(cards, c.InstanceID)
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:    victim,
						FromPlayer: victim,
						Source:     item.SourceCardID,
						Question:   "Kheru Mind-Eater — exile a card from your hand face down",
						Cards:      cards,
						Min:        1,
						Max:        1,
						ThenKey:    kheruMindEaterPick,
						LiveZone:   game.ZoneHand,
						Asked:      game.NewAskedItem(item),
					})
					return nil
				}),
		},
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneExile,
			Filter: game.PermissionFilter{ExiledWithSource: true},
			Label:  "You may play lands and cast spells from among cards exiled with Kheru Mind-Eater",
		}},
	})
}

// kheruMindEaterPick exiles the chosen card face down, linked to the Mind-Eater
// as the object it is now. An on-disk key: never renamed, never reused.
var kheruMindEaterPick = effects.AskedPick("choose-cards/kheru-mind-eater-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	ref, ok := ctx.SourceRef()
	if !ok {
		// The Mind-Eater has left: the card is exiled with no link (CR 607.2a).
		for _, id := range r.Picked {
			if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
	for _, id := range r.Picked {
		if _, err := ctx.Game.ExileHiddenLinkedForEffect(game.PermissionCardRef(ref), ctx.Item.Trigger.Event.Target, id); err != nil {
			return err
		}
	}
	return nil
})

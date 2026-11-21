package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Owlbear Cub — Creature — Bird Bear {2}{G}, 3/3 (Commander Legends: Battle for
// Baldur's Gate):
//
//	"Mama's Coming — Whenever this creature attacks a player who controls eight
//	 or more lands, look at the top eight cards of your library. You may put a
//	 creature card from among them onto the battlefield tapped and attacking that
//	 player. Put the rest on the bottom of your library in a random order."
//
// The trigger fires only for an attack on a player (CR 508.3a: an attack on that
// player's planeswalker or battle is not one), and only if that player controls
// eight or more lands as it is declared. The look-and-put is Arthur, Marigold
// Knight's with the attack target fixed by the text: the creature enters tapped
// and attacking the attacked player, so there is no choice to ask (CR 508.4; it
// was never declared as an attacker, CR 506.3c). The attacked player rides the
// prompt's carry beside the cards looked at. "Mama's Coming" is an ability word
// with no rules meaning (CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c92432d-ac76-47d2-af7a-fdffd4b546a2",
		Name:         "Owlbear Cub",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclared(ev, source) {
					return false
				}
				p := effects.AttackedPlayer(g, ev)
				return p != uuid.Nil && effects.GapB217LandsControlledBy(g, p) >= 8
			}, "Owlbear Cub — look at the top eight cards; you may put a creature card onto the battlefield tapped and attacking that player",
				owlbearCubLook),
		},
	})
}

func owlbearCubLook(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Controller()
	if item.Trigger == nil {
		return nil
	}
	attacked := effects.AttackedPlayer(g, item.Trigger.Event)
	looked := g.LookAtTopOfLibraryForEffect(player, 8)
	if len(looked) == 0 {
		return nil
	}
	var creatures []uuid.UUID
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() {
			creatures = append(creatures, id)
		}
	}
	if len(creatures) == 0 || attacked == uuid.Nil {
		return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Owlbear Cub — you may put a creature card from among them onto the battlefield tapped and attacking that player (the rest go to the bottom in a random order)",
		Cards:    creatures,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  owlbearCubPicked,
		Carry:    append([]uuid.UUID{attacked}, looked...),
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

var owlbearCubPicked = effects.AskedPick("choose-cards/gap-b217-owlbear-cub-creature", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	attacked, looked := r.Carry[0], r.Carry[1:]
	var picked uuid.UUID
	if len(r.Picked) > 0 {
		picked = r.Picked[0]
	}
	var rest []uuid.UUID
	for _, id := range looked {
		if id != picked {
			rest = append(rest, id)
		}
	}
	if len(rest) > 0 {
		if err := ctx.Game.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, rest); err != nil {
			return err
		}
	}
	if picked == uuid.Nil {
		return nil
	}
	return effects.FinB11PutAttacking(ctx.Game, ctx.Controller(), picked, game.ZoneLibrary, attacked, game.KeyedTail{})
})

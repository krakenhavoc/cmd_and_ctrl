package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b09_helpers.go — shared pieces of Kamigawa: Neon Dynasty's
// "Samurai or Warrior attacks alone" cards (S82, batch 9).

// neoB09SamuraiOrWarriorAttacksAlone is "whenever a Samurai or Warrior you
// control attacks alone": a creature its controller controls, declared as
// an attacker, that is a Samurai or a Warrior as it is declared, and the
// only creature declared (CR 506.5, game.AttackedAlone, read when the
// attack trigger is harvested, after the whole declaration). Both
// subtype words are read through the source's text (CR 612.2).
func NeoB09SamuraiOrWarriorAttacksAlone(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !AttackDeclaredByYou(ev, source.Controller) || !game.AttackedAlone(g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Controller != source.Controller {
		return false
	}
	return c.HasSubtype(source.TextSubtype("Samurai")) || c.HasSubtype(source.TextSubtype("Warrior"))
}

// neoB09SamuraiOrWarrior matches a Samurai or a Warrior, the words read
// through the resolving ability's source text.
func NeoB09SamuraiOrWarrior(ctx *Context) CardPredicate {
	return Or(HasSubtype(ctx.TextSubtype("Samurai")), HasSubtype(ctx.TextSubtype("Warrior")))
}

// neoB09LoneAttacker is the creature the trigger's event names, when it is
// still on the battlefield as the same object it was.
func NeoB09LoneAttacker(g *game.Game, item *game.StackItem) (game.Card, bool) {
	id := item.Trigger.Event.CardID
	if !OnBattlefield(g, id) {
		return game.Card{}, false
	}
	return g.LookupCardForEffect(id)
}

// neoB09MayCastFromGraveyardThisTurn is the resolution of "you may cast
// target <card> from your graveyard this turn" (Zul Ashur's permission):
// a cast-only grant on the one card, for this turn, under the controller.
func NeoB09MayCastFromGraveyardThisTurn(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			card, ok := g.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			g.GrantCastPermissionToCardsForEffect(game.CastPermission{
				Player:     item.Controller,
				Zone:       game.ZoneGraveyard,
				CastOnly:   true,
				Source:     item.SourceCardID,
				SourceName: name,
				Label:      name + " — cast " + card.Name + " this turn",
			}, []game.Card{card})
			return nil
		}
		return nil
	}
}

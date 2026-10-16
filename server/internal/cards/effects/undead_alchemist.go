package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Undead Alchemist — Creature — Zombie {3}{U}, 4/2:
//
//	"If a Zombie you control would deal combat damage to a player,
//	 instead that player mills that many cards.
//	 Whenever a creature card is put into an opponent's graveyard from
//	 their library, exile that card and create a 2/2 black Zombie
//	 creature token."
//
// The first line is a damage replacement (Force Bubble's shape): the
// combat damage event is cancelled and the player mills that many
// cards. No damage is dealt, so lifelink and "deals combat damage"
// triggers see nothing. Whether the source is a Zombie you control is
// read off the damage's last-known source characteristics.
//
// The second line watches both library-to-graveyard routes (a mill and
// any other move, Narcomoeba's pair) for a creature card an opponent
// owns. The token is made whether or not the card is still there to
// exile (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3807e6fe-0555-4b09-aef2-efd58abaf669",
		Name:         "Undead Alchemist",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{undeadAlchemistMillInstead()},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventMill, game.EventZoneMove}, creatureCardMilledFromAnOpponent,
				"Undead Alchemist — exile that card and create a 2/2 Zombie",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					card := ctx.Trigger().Event.CardID
					if z := g.FindCardZoneForEffect(card); z != nil && z.Kind == game.ZoneGraveyard {
						if err := (ExileTarget{Target: card}).Apply(ctx); err != nil {
							return err
						}
					}
					return CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1}.Apply(ctx)
				}),
		},
	})
}

func undeadAlchemistMillInstead() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || !ev.IsCombatDamage || ev.DamageAmount <= 0 {
				return false
			}
			if g.PlayerByIDForEffect(ev.DamageTarget) == nil {
				return false
			}
			// "Zombie" is a word of the source's text (CR 612.2, #3984).
			return damageSourceIsZombieControlledBy(ev, g, src.Controller, src.TextSubtype("Zombie"))
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
			n, player := ev.DamageAmount, ev.DamageTarget
			ev.Cancel()
			return g.MillNForEffect(player, n)
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		// Not a prevention (CR 615): the damage is replaced by a mill.
		Prevention: false,
		Label:      "Undead Alchemist — that player mills that many cards instead",
	}
}

// damageSourceIsZombieControlledBy reports whether a damage event's
// source is a Zombie controlled by `controller`, as the source was when
// the event was created.
func damageSourceIsZombieControlledBy(ev *game.ReplacementEvent, g *game.Game, controller uuid.UUID, zombie string) bool {
	if lki := ev.SourceLKI; lki != nil && lki.Controller != uuid.Nil {
		if lki.Controller != controller {
			return false
		}
		if lki.AllCreatureTypes {
			return true
		}
		for _, s := range lki.Subtypes {
			if strings.EqualFold(s, zombie) {
				return true
			}
		}
		return false
	}
	c, ok := g.LookupCardForEffect(ev.DamageSource)
	return ok && c.Controller == controller && c.HasSubtype(zombie)
}

// creatureCardMilledFromAnOpponent is "a creature card is put into an
// opponent's graveyard from their library".
func creatureCardMilledFromAnOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.OldZone != game.ZoneLibrary || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Owner != source.Controller && g.PlayerByIDForEffect(c.Owner) != nil
}

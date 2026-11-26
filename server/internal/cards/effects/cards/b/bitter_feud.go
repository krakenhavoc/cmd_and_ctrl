package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bitter Feud — Enchantment {4}{R} (Commander 2014):
//
//	"As this enchantment enters, choose two players.
//	 If a source controlled by one of the chosen players would deal damage
//	 to the other chosen player or a permanent that player controls, that
//	 source deals double that damage to that player or permanent instead."
//
// Sower of Discord's two-player as-enters choice; the replacement reads
// the pair off the Feud. The source's controller is read as the source last
// existed (CR 608.2h). Damage between a chosen player and an unchosen one,
// or to a player's own permanents, is left alone.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4add901-3f2b-4f92-bb26-7f599e288802",
		Name:         "Bitter Feud",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChoosePlayersAsEnters("Bitter Feud", effects.Players),
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil || src == nil {
					return false
				}
				a, b := src.ChosenPlayer, src.ChosenSecondPlayer
				if a == uuid.Nil || b == uuid.Nil {
					return false
				}
				ch, ok := effects.DamageSourceCharacteristics(ev, g)
				if !ok {
					return false
				}
				victim := feudVictimController(ev, g)
				return (ch.Controller == a && victim == b) || (ch.Controller == b && victim == a)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Bitter Feud: damage between the chosen players is doubled",
		}},
	})
}

// feudVictimController is the player a damage event hurts: the player
// itself, or the controller of the permanent hit.
func feudVictimController(ev *game.ReplacementEvent, g *game.Game) uuid.UUID {
	if ev.DamageTarget == uuid.Nil {
		return uuid.Nil
	}
	if p := g.PlayerByIDForEffect(ev.DamageTarget); p != nil {
		return p.ID
	}
	if c, ok := g.LookupCardForEffect(ev.DamageTarget); ok && effects.OnBattlefield(g, ev.DamageTarget) {
		return c.Controller
	}
	return uuid.Nil
}

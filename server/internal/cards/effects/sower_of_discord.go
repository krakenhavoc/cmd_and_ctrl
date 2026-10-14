package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sower of Discord — Creature — Demon {4}{B}{B}, 6/6:
//
//	"Flying
//	 As this creature enters, choose two players.
//	 Whenever damage is dealt to one of the chosen players, the other
//	 chosen player also loses that much life."
//
// The as-enters pair is stored on the permanent (ChoosePlayersAsEnters
// → Card.ChosenPlayer and Card.ChosenSecondPlayer, #3389). The trigger
// watches every damage event, combat or not, from any source: damage to
// either chosen player makes the other lose that much life, and damage
// to anyone else does nothing. The engine emits one damage event per
// source, so two sources are two triggers, as printed (CR 120.3). The
// loss is life loss, not damage, so it never triggers Sower again.
//
// The other player is fixed when the ability triggers and carried on
// the item's payload, so a Sower that dies in response still drains the
// player it named (CR 603.10 last-known information). A chosen player
// who has left the game is nobody to lose life: the drain does nothing.
//
// No simplification beyond the engine-wide as-enters one: the prompts
// are queued from the AsEnters hook, so the Sower is briefly on the
// battlefield with nobody chosen, and nothing can act in that window.
func init() {
	t := On(game.EventDealDamage, damageToAChosenPlayer,
		"Sower of Discord — the other chosen player loses that much life",
		func(g *game.Game, item *game.StackItem) error {
			if len(item.Payload) == 0 || item.Payload[0].Kind != game.TargetPlayer {
				return nil
			}
			other := item.Payload[0].ID
			amount := NewContext(g, item).Trigger().Event.Amount
			if p := g.PlayerByIDForEffect(other); p == nil || p.Eliminated || amount <= 0 {
				return nil
			}
			return g.ChangePlayerLifeForEffect(item.SourceCardID, other, -amount)
		})
	// A fill-in Build: it records who the OTHER chosen player is as the
	// ability triggers, and leaves item.Effect for the engine to install.
	key := t.Key
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, key)
		if other := otherChosenPlayer(ev.Target, source); other != uuid.Nil {
			item.Payload = append(item.Payload, game.TargetRef{Kind: game.TargetPlayer, ID: other})
		}
		return item
	}
	Register(Spec{
		OracleID:        "c6aadaf0-eb47-4228-95e2-48e3ff7b6e2f",
		Name:            "Sower of Discord",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AsEnters:        ChoosePlayersAsEnters("Sower of Discord", Players),
		Triggered:       []game.TriggeredAbility{t},
	})
}

// otherChosenPlayer is the member of source's chosen pair that is not
// `hit`, or uuid.Nil when hit is neither of them or the pair is not
// complete.
func otherChosenPlayer(hit uuid.UUID, source *game.Card) uuid.UUID {
	a, b := source.ChosenPlayer, source.ChosenSecondPlayer
	if hit == uuid.Nil || a == uuid.Nil || b == uuid.Nil {
		return uuid.Nil
	}
	switch hit {
	case a:
		return b
	case b:
		return a
	}
	return uuid.Nil
}

// damageToAChosenPlayer is "damage is dealt to one of the chosen
// players": any positive damage event, from any source, whose target is
// a member of the source's complete chosen pair.
func damageToAChosenPlayer(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 {
		return false
	}
	return otherChosenPlayer(ev.Target, source) != uuid.Nil
}

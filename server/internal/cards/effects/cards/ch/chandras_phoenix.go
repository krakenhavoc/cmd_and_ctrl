package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra's Phoenix — Creature — Phoenix {1}{R}{R}, 2/2 (Magic 2014, S82 OLD
// batch 401, #5563):
//
//	"Flying
//	 Haste
//	 Whenever an opponent is dealt damage by a red instant or sorcery spell you
//	 control or by a red planeswalker you control, return this card from your
//	 graveyard to your hand."
//
// A graveyard trigger (Aurora Phoenix's return) on damage to an opponent, read
// while the spell is mid-resolution (Bloodfeather Phoenix's shape): the source is
// either a red instant or sorcery spell its controller controls, or a red
// planeswalker they control on the battlefield. Colour is read from the source's
// last-known values.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "65ddd04c-7207-4f6c-ae1b-c5e02a520a51",
		Name:            "Chandra's Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventDealDamage, chandrasPhoenixDamage,
				"Chandra's Phoenix — return this card from your graveyard to your hand",
				effects.ReturnThisCardFromYourGraveyardToHand)),
		},
	})
}

// chandrasPhoenixDamage is the trigger event: damage dealt to an opponent of the
// Phoenix's controller by a red instant or sorcery spell, or a red planeswalker,
// that controller controls.
func chandrasPhoenixDamage(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source == uuid.Nil || ev.Target == uuid.Nil ||
		ev.Target == source.Controller || g.PlayerByIDForEffect(ev.Target) == nil {
		return false
	}
	if spell, item, ok := g.ResolvingSpellForEffect(ev.Source); ok {
		return item.Kind == game.StackItemSpell && item.Controller == source.Controller &&
			(spell.IsInstant() || spell.IsSorcery()) && spell.HasColorWord("R")
	}
	pw, ok := g.LookupCardForEffect(ev.Source)
	return ok && pw.IsPlaneswalker() && pw.Controller == source.Controller && pw.HasColorWord("R")
}

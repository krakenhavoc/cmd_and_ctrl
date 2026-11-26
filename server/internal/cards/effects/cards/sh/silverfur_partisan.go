package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silverfur Partisan — Creature — Wolf Warrior {2}{G}, 2/2 (Shadows over
// Innistrad, #5543):
//
//	"Trample
//	 Whenever a Wolf or Werewolf you control becomes the target of an
//	 instant or sorcery spell, create a 2/2 green Wolf creature token."
//
// EventBecomesTarget fires once per target slot, with the targeted object
// in CardID and the spell in Source, so a spell that names two of your
// Wolves puts two triggers on the stack, as printed. The spell is judged
// on its own types as it was cast; an ability never triggers it. Wolf and
// Werewolf are words of the card's text (CR 612.2), and the target is read
// through the layers, so a changeling counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dfeb239b-8904-4ef0-9a91-edc901b0c49a",
		Name:            "Silverfur Partisan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, silverfurPartisanTrigger,
				"Silverfur Partisan — create a 2/2 green Wolf creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 1})),
		},
	})
}

func silverfurPartisanTrigger(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || ev.Source == uuid.Nil {
		return false
	}
	item := g.StackItemForEffect(ev.Source)
	if item == nil || item.Kind != game.StackItemSpell {
		return false
	}
	spell, ok := g.LookupCardForEffect(item.ID)
	if !ok || !(spell.IsInstant() || spell.IsSorcery()) {
		return false
	}
	if z := g.FindCardZoneForEffect(ev.CardID); z == nil || z.Kind != game.ZoneBattlefield {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || target.Controller != source.Controller {
		return false
	}
	return target.HasSubtype(source.TextSubtype("Wolf")) || target.HasSubtype(source.TextSubtype("Werewolf"))
}

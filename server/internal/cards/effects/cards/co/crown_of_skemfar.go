package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crown of Skemfar — Enchantment — Aura {2}{G}{G} (Kaldheim Commander,
// #5506):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 for each Elf you control and has reach.
//	 {2}{G}: Return this card from your graveyard to your hand."
//
// The pump is Blanchwood Armor's per-count attached modify (layer 7c) over
// the Elves the Crown's controller controls, read live through the layered
// view; the reach is the attached-creature grant. The recursion is an
// ability of the card in the graveyard (Bestial Bloodline's shape), so it can
// be activated only while the card is there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db961314-fe78-4362-902f-aa0dbf6c487d",
		Name:         "Crown of Skemfar",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
				return effects.S82OldB075ElvesYouControl(g, source.Controller, source)
			}),
			effects.GrantToAttached("reach"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}

package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Void Maw — Creature — Horror {4}{B}{B}, 4/5 (Coldsnap, S82 OLD b560):
//
//	"Trample
//	 If another creature would die, exile it instead.
//	 Put a card exiled with this creature into its owner's graveyard: This
//	 creature gets +2/+2 until end of turn."
//
// Gisa, Glorious Resurrector's exile-instead-of-dying replacement over every
// creature but the Maw, whoever controls it; each card it exiles is "exiled
// with" the Maw (CR 607.2a), and the activated ability's cost puts one of them
// into its owner's graveyard (Shelob's processing cost, with no type bound).
// Tokens are exiled too and cease to exist, so they are never a card to put
// back.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad2a1b07-58f5-44c2-92e0-464ea2d45e7e",
		Name:            "Void Maw",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.OldZone != game.ZoneBattlefield || ev.NewZone != game.ZoneGraveyard || ev.CardID == src.InstanceID {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature()
			},
			Replace: effects.OldB560ExileInsteadLinkedToSource,
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Void Maw: another creature is exiled instead of dying",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "Put a card exiled with this creature into its owner's graveyard: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:    effects.ProcessCardExiledWithThis("a card exiled with Void Maw", nil),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Void Maw — +2/+2 until end of turn"),
		}},
	})
}

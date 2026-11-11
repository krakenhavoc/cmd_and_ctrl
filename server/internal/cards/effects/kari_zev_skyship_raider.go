package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kari Zev, Skyship Raider — Legendary Creature — Human Pirate {1}{R}, 1/3
// (Lost Caverns of Ixalan Commander):
//
//	"First strike, menace
//	 Whenever Kari Zev attacks, create Ragavan, a legendary 2/1 red Monkey
//	 creature token. Ragavan enters tapped and attacking. Exile that token
//	 at end of combat."
//
// Kavaron Harrier's shape: the token is put onto the battlefield tapped and
// attacking without being declared (CR 508.4), its controller choosing what
// it attacks (a bot sends it where Kari Zev went), and a delayed trigger at
// the beginning of this combat's end of combat step exiles exactly the
// token that landed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "786baa29-afd4-40f0-95c8-920a972b9175",
		Name:            "Kari Zev, Skyship Raider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Kari Zev — create Ragavan, tapped and attacking",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokensAttackingYourChoice{
						Controller: item.Controller,
						Template:   TokenCard("2/1 red Legendary Monkey named Ragavan"),
						N:          1,
						Tapped:     true,
						Prefer:     triggerAttackTarget(item),
						Label:      "Kari Zev",
						Tail:       game.KeyedTail{Ref: kariZevExileRagavan},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// kariZevExileRagavan is "Exile that token at end of combat". An on-disk
// key.
var kariZevExileRagavan = AttackingTokensThen("attacking-tokens/kari-zev-exile-ragavan",
	func(ctx *Context, created []uuid.UUID, _ game.KeyedTail) error {
		if len(created) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{At: game.StepEndCombat, Controller: ctx.Item.Controller,
			Label: "Kari Zev — exile Ragavan", Cards: created, Body: exileListedCardsBody}.Apply(ctx)
	})

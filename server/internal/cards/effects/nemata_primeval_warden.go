package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nemata, Primeval Warden — Legendary Creature — Treefolk {2}{B}{G}, 3/4
// (Dominaria United):
//
//	"Reach
//	 If a creature an opponent controls would die, exile it instead. When you
//	 do, create a 1/1 green Saproling creature token.
//	 {G}, Sacrifice a Saproling: Nemata gets +2/+2 until end of turn.
//	 {1}{B}, Sacrifice two Saprolings: Draw a card."
//
// The replacement is Head of the Hunt's, word for word but for the token:
// "exile it instead" (rfOpponentCreatureWouldDie; tokens are creatures too,
// and the creature never died, CR 700.4) and the CR 603.12 reflexive trigger
// "when you do", declared on the move and created once the creature has
// landed in exile, under Nemata's controller. Each creature exiled this way
// makes its own Saproling, and none when another replacement sends the card
// elsewhere. The two abilities pay with Saprolings as a cost, so they can't
// be responded to by sacrificing the same one twice; Nemata gets +2/+2 as
// the first resolves, or nothing if she has left (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f72558f5-ac5c-4efa-b01b-439bc0bbf18d",
		Name:            "Nemata, Primeval Warden",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: rfOpponentCreatureWouldDie,
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				return g.QueueReflexiveTriggerAfterReplacedMoveForEffect(ev, src, game.ReflexiveTrigger{
					Label: "Nemata, Primeval Warden — create a 1/1 green Saproling creature token",
					Body:  nemataSaprolingBody,
				})
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Nemata, Primeval Warden: exile instead of dying",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{G}, Sacrifice a Saproling: Nemata gets +2/+2 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
				Cost:    Plus(ManaCost("{G}"), SacrificeN(1, "a Saproling", HasSubtype("Saproling"))),
				Effect:  thisGetsUntilEndOfTurn(2, 2, "Nemata, Primeval Warden — +2/+2 until end of turn"),
			},
			{
				Label:   "{1}{B}, Sacrifice two Saprolings: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Draws: 1},
				Cost:    Plus(ManaCost("{1}{B}"), SacrificeN(2, "two Saprolings", HasSubtype("Saproling"))),
				Effect:  Do(DrawCards{N: 1}),
			},
		},
	})
}

// nemataSaprolingBody is the reflexive trigger: one 1/1 green Saproling for
// the trigger's controller. An on-disk key: never renamed, never reused.
var nemataSaprolingBody = game.ReflexiveBody("nemata-primeval-warden/saproling", simpleBody(func(g *game.Game, item *game.StackItem) error {
	return CreateToken{Template: TokenCard("1/1 green Saproling"), N: 1}.Apply(NewContext(g, item))
}), nil)

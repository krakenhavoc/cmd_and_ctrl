package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kels, Fight Fixer — Legendary Creature — Azra Warlock {2}{B}{B}, 4/3:
//
//	"Menace
//	 Whenever you sacrifice a creature, you may pay {U/B}. If you do,
//	 draw a card. ({U/B} can be paid with either {U} or {B}.)
//	 {1}, Sacrifice a creature: Kels gains indestructible until end of
//	 turn."
//
// The sacrifice trigger watches EventSacrifice (Captain Lannery Storm's
// shape): "you sacrifice" is the event's actor being Kels's controller,
// and the creature is read before it leaves, so a sacrifice as a cost
// counts as much as one to an edict, and two creatures sacrificed at once
// are two triggers. Kels sacrificing itself sees its own sacrifice (CR
// 603.10a: a leaves-the-battlefield trigger looks back in time). The
// {U/B} is paid as the trigger resolves (MayPay, extort's hybrid cost);
// declining draws nothing.
//
// The activated ability's "a creature" includes Kels itself, which is a
// legal (if pointless) sacrifice, and still triggers the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fe2e27a9-437e-4d22-aa22-982978793cff",
		Name:            "Kels, Fight Fixer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, jmpB1YouSacrificedACreature, "Kels, Fight Fixer — pay {U/B} to draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{U/B}",
						Question: "Kels, Fight Fixer — pay {U/B} to draw a card?",
						Then:     mayPayDrawACard,
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: Kels gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Kels, Fight Fixer — indestructible until end of turn",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Proctor of Potential — Creature — Human Cleric {W}{U}, 3/1:
//
//	"Whenever this creature or another creature you control enters,
//	 surveil 1.
//	 {W}{U}: Return this card from your graveyard to the battlefield
//	 with a finality counter on it. Activate only if you've scried or
//	 surveilled this turn. (If a creature with a finality counter on it
//	 would die, exile it instead.)"
//
// Grim Repriser's graveyard activation with a different gate: this
// turn's scry or surveil (rfCreatureBScriedOrSurveilledThisTurn). The
// trigger is "a creature you control enters", which includes the
// Proctor itself (Impact Tremors' reading), so it surveils as it comes
// back. The finality counter is two self-replacements — it enters with
// one only when it comes from a graveyard, and a Proctor carrying one
// is exiled instead of dying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c612ba0a-5134-45fa-a762-f4621bda4f83",
		Name:         "Proctor of Potential",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			rfCreatureBEntersWithCounterFromGraveyard(rfCreatureBFinalityCounter, 1,
				"Proctor of Potential: enters with a finality counter"),
			rfMiscAFinalityExile("Proctor of Potential: exile instead of dying (finality counter)"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, CreatureEnteredUnderYourControl,
				"Proctor of Potential — surveil 1", Do(Surveil{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "{W}{U}: Return this card from your graveyard to the battlefield with a finality counter on it. Activate only if you've scried or surveilled this turn.",
			Cost:  ManaCost("{W}{U}"),
			Zones: []game.ZoneKind{game.ZoneGraveyard},
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return rfCreatureBScriedOrSurveilledThisTurn(g, controller)
			},
			Effect: returnThisFromGraveyardToBattlefield,
		}},
	})
}

package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Broodheart Engine — Artifact {B}{G} (Aetherdrift, #4464):
//
//	"At the beginning of your upkeep, surveil 1.
//	 {2}{B}{G}, {T}, Sacrifice this artifact: Return target creature or
//	 Vehicle card from your graveyard to the battlefield. Activate only
//	 as a sorcery."
//
// The upkeep trigger is the plain surveil every catalog surveil card
// uses. The activated ability is Back on Track's return ("creature or
// Vehicle card", your graveyard) on Codex Shredder's sacrifice-and-tap
// cost, at sorcery speed. The target is chosen before the Engine is
// sacrificed, so the Engine is never a card in the graveyard to name;
// a target that has left the graveyard by resolution does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b759831-abe8-4996-a897-3c6d6d30a1c7",
		Name:         "Broodheart Engine",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Broodheart Engine — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}{B}{G}, {T}, Sacrifice this artifact: Return target creature or Vehicle card from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost:  effects.Plus(effects.ManaCost("{2}{B}{G}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCardInGraveyard("target creature or Vehicle card from your graveyard",
				effects.Or(effects.Creature(), effects.HasSubtype("Vehicle")), effects.YouOwn()),
			SorcerySpeed: true,
			Effect:       effects.ReturnFirstLegalGraveyardTargetToBattlefield,
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gonti's Aether Heart — Legendary Artifact {6}:
//
//	"Whenever Gonti's Aether Heart or another artifact you control
//	 enters, you get {E}{E} (two energy counters).
//	 Pay eight {E}, Exile Gonti's Aether Heart: Take an extra turn after
//	 this one."
//
// ADR 0129 §2 (#1995, S58 deck requests #2077): "Pay eight {E}" is the
// energy cost component and "Exile Gonti's Aether Heart" is the
// exile-this-permanent cost (#1404), both paid as the ability is
// activated, so a Heart answered in response still gives the turn. The
// extra turn is ADR 0059's (CR 500.7).
//
// The trigger reads "Gonti's Aether Heart or another artifact you
// control": the Heart itself entering counts, as does each artifact
// (tokens too) entering under its controller's control.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "69428825-3c40-486d-b051-14e97a598ce6",
		Name:         "Gonti's Aether Heart",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsArtifact()
			}, "Gonti's Aether Heart — you get {E}{E}", Do(GetEnergy{N: 2})), game.Purpose{Energy: 2}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Pay eight {E}, Exile Gonti's Aether Heart: Take an extra turn after this one.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(PayEnergy(8), ExileThis()),
			Effect:  youTakeAnExtraTurnEffect,
		}},
	})
}

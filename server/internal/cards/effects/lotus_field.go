package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lotus Field — Land (#332):
//
//	"Hexproof
//	 This land enters tapped.
//	 When this land enters, sacrifice two lands.
//	 {T}: Add three mana of any one color."
//
// Hexproof is a printed keyword the deck importer stamps. The tapped
// entry is SelfEntersTapped. The mana ability is #742's one-pick,
// three-token "any one color" (see Gilded Lotus).
//
// The enters trigger is ONE choice of two lands, asked on resolution
// and sacrificed together: the own_permanents prompt (ChoosePermanents,
// #1214) with a floor and a ceiling of two, then SacrificeAllThenFor
// Effect, so both lands leave in a single CR 701.21 event. Lotus Field
// itself is a legal choice, as the printed card allows, and a
// controller with fewer than two lands sacrifices what they have (CR
// 609.3 — the prompt's bounds clamp to the candidates). Nothing is a
// target, so a hexproof land — a second Lotus Field — can be chosen.
//
// This used to be two single-land sacrifice prompts answered one after
// the other (PlayerSacrificesNForEffect), a declared caveat: the same
// two lands died, but in two events rather than one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "134d5b82-7940-4b33-a922-7f9d1f403e50",
		Name:         "Lotus Field",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lotus Field — sacrifice two lands", Do(ChoosePermanents{
				Question:   "Lotus Field — sacrifice two lands",
				Candidates: lotusFieldLands,
				Then:       lotusFieldSacrifice,
			})),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: OneColorOfAmount(3),
			Label:    "Add three mana of any one color",
		}},
	})
}

// lotusFieldLands offers every land the chooser controls, exactly two
// to be picked (clamped by the engine when there are fewer).
//
// Caller holds g.mu.
func lotusFieldLands(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == of && c.IsLand() {
			out = append(out, c.InstanceID)
		}
	}
	return out, 2, 2
}

// lotusFieldSacrifice sacrifices the chosen lands as one simultaneous
// exit, stamped with Lotus Field as the source that asked.
//
// Caller holds g.mu.
func lotusFieldSacrifice(ctx *Context, picked game.PromptedPicks) error {
	ids := picked.Cards()
	if len(ids) == 0 {
		return nil
	}
	return ctx.Game.SacrificeAllThenForEffect(ctx.Source(), ids, nil)
}

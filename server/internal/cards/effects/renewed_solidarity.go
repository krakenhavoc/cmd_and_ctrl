package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Renewed Solidarity — Enchantment {2}{W}:
//
//	"As this enchantment enters, choose a creature type.
//	 Creatures you control of the chosen type get +1/+0.
//	 At the beginning of your end step, for each token you control of
//	 the chosen type that entered this turn, create a token that's a
//	 copy of it."
//
// The choice is the CR 614.12 "as enters" prompt and lands on the
// permanent (Card.NamedTribe); until it is answered the anthem applies
// to nothing. The end-step copy is counted as the trigger resolves,
// from the set of tokens that had entered this turn BEFORE any copy is
// made, so the copies — which also entered this turn — are not copied
// in turn. "Of the chosen type" reads effective subtypes, so a
// changeling token counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bea3ff6e-7649-4e51-b2ad-763f9ac2d4b8",
		Name:         "Renewed Solidarity",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Renewed Solidarity"),
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Chosen: true, YoursOnly: true}, 1, 0),
		},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Renewed Solidarity — for each token you control of the chosen type that entered this turn, create a copy",
				renewedSolidarityCopyTokens),
		},
	})
}

// renewedSolidarityCopyTokens copies every token of the chosen type
// this controller controls that entered this turn.
func renewedSolidarityCopyTokens(g *game.Game, item *game.StackItem) error {
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || src.NamedTribe == "" {
		return nil
	}
	var tokens []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsToken() && c.HasSubtype(src.NamedTribe) && g.EnteredThisTurn(c.InstanceID) {
			tokens = append(tokens, c.InstanceID)
		}
	}
	ctx := NewContext(g, item)
	for _, id := range tokens {
		if err := (CreateTokenCopy{Controller: item.Controller, Copy: id, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eddymurk Crab — Creature — Elemental Crab {5}{U}{U}, 5/5 (Bloomburrow,
// #4467):
//
//	"Flash
//	 This spell costs {1} less to cast for each instant and sorcery card
//	 in your graveyard.
//	 This creature enters tapped if it's not your turn.
//	 When this creature enters, tap up to two target creatures."
//
// Flash is the printed keyword. The reduction is The Dawning Archaic's
// self cost modifier, counted as the spell is cast (CR 601.2f); it
// lowers only the generic part, so {U}{U} is always paid. "Enters tapped
// if it's not your turn" is EntersTappedUnless read as the Crab enters,
// a replacement (CR 614.1c), so it never enters untapped and then taps.
// The entry trigger is Elder Deep-Fiend's tap over up to two creatures;
// a target gone by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	enters := WhenThisEnters("Eddymurk Crab — tap up to two target creatures", tapEachLegalTarget)
	enters.Targets = TargetCreature("up to two target creatures").WithCount(0, 2)
	Register(Spec{
		OracleID:        "3c345e24-1726-45ac-91ca-e348d1e7acfe",
		Name:            "Eddymurk Crab",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(socB37InstantAndSorceryCardsInYourGraveyard,
				"This spell costs {1} less to cast for each instant and sorcery card in your graveyard."),
		},
		Replacements: []game.ReplacementEffect{
			EntersTappedUnless(func(g *game.Game, src *game.Card) bool { return IsYourTurn(g, src.Controller) }),
		},
		Triggered: []game.TriggeredAbility{enters},
	})
}

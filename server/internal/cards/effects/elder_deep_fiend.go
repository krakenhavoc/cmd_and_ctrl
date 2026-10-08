package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elder Deep-Fiend — Creature — Eldrazi Octopus {8}, 5/6:
//
//	"Flash
//	 Emerge {5}{U}{U} (You may cast this spell by sacrificing a creature
//	 and paying the emerge cost reduced by that creature's mana value.)
//	 When you cast this spell, tap up to four target permanents."
//
// Emerge is the shared alternative cost (ADR 0135 §4): {5}{U}{U} over a
// four-drop is {1}{U}{U}, over a seven-drop {U}{U}. Flash lets it come
// down in an opponent's upkeep or beginning of combat, and the cast
// trigger taps up to four permanents above the spell, so it resolves
// even if the Deep-Fiend is countered. A target gone by then is skipped
// (CR 608.2b).
//
// No simplification.
func init() {
	cast := WhenYouCastThisSpell("Elder Deep-Fiend — tap up to four target permanents", tapEachLegalTarget)
	cast.Targets = TargetPermanent("up to four target permanents").WithCount(0, 4)
	Register(Spec{
		OracleID:         "4eafe717-4ba4-4901-8c67-11757230eb54",
		Name:             "Elder Deep-Fiend",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flash"},
		AlternativeCosts: []game.AlternativeCost{Emerge("{5}{U}{U}")},
		Triggered:        []game.TriggeredAbility{cast},
	})
}

// tapEachLegalTarget taps each of the item's still-legal card targets —
// "tap up to N target permanents".
func tapEachLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range legalTargetCards(item, g) {
		if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

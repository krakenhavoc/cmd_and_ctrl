package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Black Widow, Double Agent — Legendary Creature — Human Hero Villain
// {1}{W}{B}, 3/2 (Marvel Super Heroes, #4455):
//
//	"Deathtouch
//	 Whenever a creature you control attacks alone, it gains first strike
//	 and menace until end of turn."
//
// Squall's Rough Divide shape: "alone" is game.AttackedAlone read when the
// attack trigger is harvested, and the creature that gains the two keywords
// is the one that attacked, read off the event. A creature removed in
// response gets nothing. Menace is granted at declaration, before blockers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e6c181f1-f4f7-4a96-b50c-717cf2a02098",
		Name:            "Black Widow, Double Agent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, mshB1YouAttackAlone,
				"Black Widow, Double Agent — the lone attacker gains first strike and menace",
				func(g *game.Game, item *game.StackItem) error {
					attacker := item.Trigger.Event.CardID
					if !onBattlefield(g, attacker) {
						return nil
					}
					return GrantKeywordUntilEOT{
						Target:   attacker,
						Keywords: []string{"first strike", "menace"},
						Label:    "Black Widow, Double Agent",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

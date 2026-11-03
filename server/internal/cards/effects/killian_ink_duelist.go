package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Killian, Ink Duelist — Legendary Creature — Human Warlock {W}{B}, 2/2:
//
//	"Lifelink
//	 Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 Spells you cast that target a creature cost {2} less to cast."
//
// The reduction reads the spell's announced targets (CR 601.2c comes
// before the total cost is determined in 601.2f), so the engine prices
// each target on its own: a creature on the battlefield takes {2} of
// generic mana off, a player or a creature card in a graveyard does not.
// It spends against generic mana only, never the coloured requirement,
// and applies to an alternative cost too, after any additional cost is
// added (the card's rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e233cb5b-7237-4271-8157-9179cc897b6c",
		Name:            "Killian, Ink Duelist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink", "menace"},
		CostModifiers: []game.CostModifier{
			CostsLessIfItTargets(2, "Spells you cast that target a creature cost {2} less to cast.",
				socB7CreaturePermanent(), YourSpell()),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archmage of Echoes — Creature — Faerie Wizard {4}{U}, 4/4 (Wilds of
// Eldraine Commander):
//
//	"Flying, ward {2}
//	 Whenever you cast a Faerie or Wizard permanent spell, copy it. (The
//	 copy becomes a token.)"
//
// The cast trigger is Double Down's shape: the spell's creature types are
// read from the spell as it was cast (a changeling spell is both), the
// permanent-spell clause excludes instants and sorceries, and "copy it"
// names the spell without targeting it, so a spell countered before the
// trigger resolves is still copied from last-known information (CR 608.2h).
// The copy is not cast, so it neither triggers Archmage again nor offers new
// targets (CR 707.10), and becomes a token as it resolves (CR 707.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f1ecb3d7-5ea0-45b2-b1b0-d7ae304db2a1",
		Name:            "Archmage of Echoes",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "ward"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Archmage of Echoes — ward {2}"),
			WheneverYouCast(And(AnySubtype("Faerie", "Wizard"), Not(Or(Instant(), Sorcery()))),
				"Archmage of Echoes — copy that Faerie or Wizard permanent spell",
				func(g *game.Game, item *game.StackItem) error {
					return CopySpell{
						StackID:       item.Trigger.Event.CardID,
						Controller:    item.Controller,
						FromLastKnown: true,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

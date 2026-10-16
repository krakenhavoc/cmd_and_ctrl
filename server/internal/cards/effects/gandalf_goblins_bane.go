package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gandalf, Goblins' Bane // Flameshape — an adventure card (CR 715):
//
//	Gandalf, Goblins' Bane — Legendary Creature — Avatar Wizard {2}{R}, 2/3
//	  "Whenever you cast a noncreature spell, Gandalf gets +1/+1 until
//	   end of turn and deals 1 damage to each opponent."
//	Flameshape — Sorcery — Adventure {1}{R}
//	  "Look at the top two cards of your library and exile them face
//	   down. For as long as they remain exiled, you may play them if you
//	   control a Wizard."
//
// The creature's trigger is Black Waltz No. 3's cast condition with a
// pump: the +1/+1 needs Gandalf on the battlefield, the damage does not
// (a Gandalf that left in response deals it as it last existed, CR
// 608.2h).
//
// Flameshape is a conditional exile play (#3192, ADR 0066's second
// 2026-10-13 amendment): the two cards go to exile face down, where only
// their owner may look at them (FaceDownPermitted), and the grant lasts
// for as long as each stays exiled and is live only while its holder
// controls a Wizard (game.CondControlSubtype, read each time the card is
// asked about). Gandalf himself is a Wizard, so casting him from exile
// after the adventure opens the cards; losing every Wizard shuts them
// again without ending the grant. The adventure lifecycle (exile the
// card, cast the creature later) is the engine's (game/adventure.go).
//
// No simplification.
func init() {
	const oracle = "f48f2a9b-4ba5-4984-8180-3950889bba65"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Gandalf, Goblins' Bane",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Gandalf, Goblins' Bane — +1/+1 until end of turn and 1 damage to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if onBattlefield(g, item.SourceCardID) {
						if err := (BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1,
							Label: "Gandalf, Goblins' Bane — +1/+1 until end of turn"}).Apply(ctx); err != nil {
							return err
						}
					}
					return DamageEachOpponent{N: 1}.Apply(ctx)
				}),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Flameshape",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileTopWithPermission{
				From: item.Controller, GrantTo: item.Controller, N: 2,
				FaceDown:    true,
				WhileExiled: true,
				Condition:   &game.PermissionCondition{Key: game.CondControlSubtype, Subtype: "Wizard"},
			}.Apply(ctx)
		},
	})
}

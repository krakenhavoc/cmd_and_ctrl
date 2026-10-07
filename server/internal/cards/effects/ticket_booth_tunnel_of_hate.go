package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ticket Booth // Tunnel of Hate — Enchantment — Room (ADR 0103):
//
//	Ticket Booth {2}{R}: "When you unlock this door, manifest dread."
//	Tunnel of Hate {4}{R}{R}: "Whenever you attack, target attacking
//	 creature gains double strike until end of turn."
//
// Ticket Booth is NOT implemented: manifest dread (CR 701.62a) has no
// keyword action in the engine, so the door is empty. That is weaker
// than printed, and the card says so. Tunnel of Hate is one trigger per
// attack declaration (the OncePerBatch Hospital Room uses) whose
// attacking-creature target is chosen as it goes on the stack.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "4d01b62b-b924-4da5-8ff5-b2f29d7f19b2",
		Name:         "Ticket Booth // Tunnel of Hate",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Ticket Booth's manifest dread isn't implemented, so unlocking it does nothing."},
		Right: Door{Triggered: []game.TriggeredAbility{
			OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Tunnel of Hate — target attacking creature gains double strike until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Tunnel of Hate — double strike"}.Apply(ctx)
					}),
				TargetCreature("target attacking creature", AttackingCreature()))),
		}},
	}))
}

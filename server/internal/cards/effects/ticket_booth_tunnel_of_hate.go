package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ticket Booth // Tunnel of Hate — Enchantment — Room (ADR 0103):
//
//	Ticket Booth {2}{R}: "When you unlock this door, manifest dread."
//	Tunnel of Hate {4}{R}{R}: "Whenever you attack, target attacking
//	 creature gains double strike until end of turn."
//
// Ticket Booth is manifest dread (CR 701.62a, ADR 0082's 2026-10-07
// amendment). Tunnel of Hate is one trigger per attack declaration (the
// OncePerBatch Hospital Room uses) whose attacking-creature target is
// chosen as it goes on the stack.
//
// No simplification.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "4d01b62b-b924-4da5-8ff5-b2f29d7f19b2",
		Name:         "Ticket Booth // Tunnel of Hate",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorLeft, "Ticket Booth — manifest dread", Do(ManifestDread{})),
		}},
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

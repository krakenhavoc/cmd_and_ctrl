package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hedron Detonator — Creature — Goblin Artificer {2}{R}, 2/3 (EDHREC
// rank 2911):
//
//	"Whenever an artifact you control enters, this creature deals 1
//	 damage to target opponent.
//	 {T}, Sacrifice two artifacts: Exile the top card of your library.
//	 You may play that card this turn."
//
// The Treasure deck's pinger that turns spent artifacts into cards.
//
//   - The trigger is the Reckless Fireweaver condition
//     (artifactEnteredUnderYourControl) with a target opponent chosen
//     as it goes on the stack (Abraded Bluffs' shape). The engine
//     emits one ETB per artifact, and the printed text is "whenever
//     an artifact", not "one or more", so three Treasures ping three
//     times — as printed.
//   - The impulse draw's cost is a tap plus a sacrifice clause with a
//     count of two (#747, SacrificeN). The Detonator is not an
//     artifact, so it is never one of the two. The body is the shared
//     "play it this turn" exile (Laelia's), so a land can be played
//     as the turn's land drop.
//
// No simplification.
func init() {
	const ping = "Hedron Detonator — 1 damage to target opponent"
	Register(Spec{
		OracleID:     "9221b29b-9351-4f70-9af4-8b9a33b43138",
		Name:         "Hedron Detonator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return artifactEnteredUnderYourControl(ev, source, g)
			},
			Targets: TargetPlayer("target opponent", Opponent()),
			Key:     ping,
			Effect: func(g *game.Game, item *game.StackItem) error {
				// CR 608.2b: an opponent who left the game in
				// response is no longer a legal target.
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 1}.Apply(ctx)
					}
				}
				return nil
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{T}, Sacrifice two artifacts: Exile the top card of your library. You may play that card this turn.",
			Cost:  Plus(TapCost(), SacrificeN(2, "two artifacts", Artifact())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				_, err := b12ImpulseExileForTurn(g, item, 1)
				return err
			},
		}},
	})
}

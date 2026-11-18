package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oni-Cult Anvil — Artifact {B}{R} (Kamigawa: Neon Dynasty / Brawl, S82):
//
//	"Whenever one or more artifacts you control leave the battlefield during
//	 your turn, create a 1/1 colorless Construct artifact creature token.
//	 This ability triggers only once each turn.
//	 {T}, Sacrifice an artifact: This artifact deals 1 damage to each
//	 opponent. You gain 1 life."
//
// Aang, Airbending Master's batch trigger (OncePerBatch, CR 603.2c) over
// artifacts you control leaving, read off the departure's last-known types and
// controller (CR 603.10a: the Anvil sees itself and the others go together),
// held to your turn and to once each turn by Baron Bertram Graywater's
// TriggersOnlyOnceEachTurn. The sacrifice cost may be the Anvil itself. The
// Construct is an artifact entering, not leaving, so it cannot loop.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c018f072-8f6d-4835-a717-ed4ecba70d97",
		Name:         "Oni-Cult Anvil",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(OncePerBatch(On(game.EventLTB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind != game.EventLTB || !IsYourTurn(g, source.Controller) {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && leftAsType(ev, c, "artifact") && leftUnderControlOf(ev, c) == source.Controller
				},
				"Oni-Cult Anvil — create a 1/1 colorless Construct artifact creature token",
				Do(CreateToken{Template: TokenCard("1/1 colorless Construct artifact"), N: 1})))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: This artifact deals 1 damage to each opponent. You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(TapCost(), SacrificeN(1, "an artifact", Artifact())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := g.DamageInstanceForEffect(func() error {
					for _, opp := range ctx.Opponents() {
						if err := (DealDamage{Source: item.SourceCardID, Target: opp, Amount: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}); err != nil {
					return err
				}
				return GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
			},
		}},
	})
}

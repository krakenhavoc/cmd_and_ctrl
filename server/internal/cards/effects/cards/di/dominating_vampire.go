package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dominating Vampire — Creature — Vampire {1}{R}{R}, 3/3 (Innistrad:
// Crimson Vow, #5495):
//
//	"When this creature enters, gain control of target creature with mana
//	 value less than or equal to the number of Vampires you control until
//	 end of turn. Untap that creature. It gains haste until end of turn."
//
// Act of Treason's three clauses behind a target whose legality is the
// mana-value bound: the Vampires are counted for the caster when the
// target is chosen and again when the trigger resolves (CR 608.2b), the
// Dominating Vampire itself included while it is on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "145ef633-a2e0-4bd9-a866-5767821eb44f",
		Name:         "Dominating Vampire",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Dominating Vampire — gain control of target creature with mana value at most the number of Vampires you control until end of turn, untap it, it gains haste",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.GainControl{
						Target:   id,
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Dominating Vampire — gain control until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{
						Target:   id,
						Keywords: []string{"haste"},
						Label:    "Dominating Vampire — haste until end of turn",
					}.Apply(ctx)
				}),
				effects.TargetCreature("target creature with mana value less than or equal to the number of Vampires you control",
					func(g *game.Game, caster uuid.UUID, c game.Card) bool {
						return c.ManaValue() <= effects.VowB05VampiresControlledBy(g, caster, c.TextSubtype("Vampire"))
					})),
		},
	})
}

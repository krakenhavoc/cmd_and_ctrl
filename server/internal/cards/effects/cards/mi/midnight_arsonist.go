package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Midnight Arsonist — Creature — Vampire {3}{R}, 3/2 (Innistrad: Crimson
// Vow Commander, #5495):
//
//	"When this creature enters, destroy up to X target artifacts without mana
//	 abilities, where X is the number of Vampires you control."
//
// Galvanizing Tenor's "up to X" built as the trigger goes on the stack: X is
// the Vampires its controller controls then (CR 601.2c), the Arsonist
// itself included, and with none the trigger asks for nothing. "Without
// mana abilities" reads each artifact's own abilities through
// game.ManaAbilitiesOf, so a Treasure or a mana rock is not a legal target
// and a Blood token or a Clue is. Every target still legal as the trigger
// resolves is destroyed (CR 608.2b).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Midnight Arsonist — destroy up to X target artifacts without mana abilities", func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
	enters.TargetsFrom = func(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
		x := effects.VowB05VampiresControlledBy(g, source.Controller, source.TextSubtype("Vampire"))
		if x <= 0 {
			return nil
		}
		return effects.TargetPermanent("up to X target artifacts without mana abilities",
			effects.Artifact(),
			func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return len(game.ManaAbilitiesOf(&c)) == 0 },
		).WithCount(0, x)
	}
	effects.Register(effects.Spec{
		OracleID:     "d9ee7916-2c9d-4489-8dd8-86929b09c6ea",
		Name:         "Midnight Arsonist",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}

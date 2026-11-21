package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nyssa of Traken — Legendary Creature — Human Scientist {3}{U}, 3/4 (Doctor
// Who, #5948):
//
//	"You have no maximum hand size.
//	 Sonic Booster — Whenever Nyssa of Traken attacks, sacrifice any number
//	 of artifacts. When you sacrifice one or more artifacts this way, tap up
//	 to that many target creatures and draw that many cards.
//	 Doctor's companion"
//
// No maximum hand size is Spec.NoMaxHandSize (Reliquary Tower). The attack
// trigger asks which artifacts to sacrifice (any number, including none),
// sacrifices them together, and on one or more raises a CR 603.12 reflexive
// trigger carrying the count: its target clause is "up to that many target
// creatures", derived from the carried amount, and it taps each target that
// is still legal and then draws that many cards. Nyssa may sacrifice herself
// if she is an artifact. Doctor's companion is a deck-construction rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "07802bd6-29ca-40c5-beba-c2baf7a92555",
		Name:          "Nyssa of Traken",
		Completeness:  effects.CompletenessFull,
		NoMaxHandSize: true,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Nyssa of Traken — sacrifice any number of artifacts", nyssaAttack),
		},
	})
}

// nyssaAttack asks for the artifacts to sacrifice.
func nyssaAttack(g *game.Game, item *game.StackItem) error {
	return effects.ChoosePermanents{
		Question: "Nyssa of Traken — sacrifice any number of artifacts",
		Keyed:    nyssaSacrificed,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.IsArtifact() {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, 0
		},
	}.Apply(effects.NewContext(g, item))
}

var nyssaSacrificed = effects.RegisterPermanentsPick("prompt-run/nyssa-of-traken-sacrifice", effects.PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		n := len(picked.Cards())
		if n == 0 {
			return nil
		}
		return effects.ReflexiveTrigger{
			Label:  "Nyssa of Traken — tap up to that many target creatures and draw that many cards",
			Body:   nyssaBoosterBody,
			Params: game.EffectParams{Amount: n},
		}.Apply(ctx)
	},
})

var nyssaBoosterBody = game.ReflexiveBody("nyssa-of-traken/sonic-booster",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (effects.TapTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
		}
		return effects.DrawCards{Player: ctx.Controller(), N: p.Amount}.Apply(ctx)
	},
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return effects.TargetCreature("up to that many target creatures").WithCount(0, p.Amount)
	})

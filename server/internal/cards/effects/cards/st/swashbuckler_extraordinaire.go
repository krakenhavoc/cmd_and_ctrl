package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swashbuckler Extraordinaire — Creature — Dragon Rogue Warrior {2}{R}, 2/2
// (Commander Legends: Battle for Baldur's Gate):
//
//	"When this creature enters, create a Treasure token.
//	 Whenever you attack, you may sacrifice one or more Treasures. When you do, up to
//	 that many target creatures gain double strike until end of turn."
//
// "Whenever you attack" is once per declaration (OncePerBatch). Nyssa of Traken's shape:
// the trigger asks which Treasures to sacrifice (any number, none meaning no), sacrifices
// them together, and on one or more raises a CR 603.12 reflexive trigger carrying the
// count; its target clause is "up to that many target creatures", and each target still
// legal as it resolves gains double strike until end of turn (CR 702.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66f0f43c-829d-44d6-adda-d8885bfbfbaf",
		Name:         "Swashbuckler Extraordinaire",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Swashbuckler Extraordinaire — create a Treasure token", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Swashbuckler Extraordinaire — you may sacrifice one or more Treasures; when you do, up to that many target creatures gain double strike until end of turn",
				swashbucklerAsk)),
		},
	})
}

// swashbucklerAsk asks which Treasures to sacrifice.
func swashbucklerAsk(g *game.Game, item *game.StackItem) error {
	return effects.ChoosePermanents{
		Question: "Swashbuckler Extraordinaire — sacrifice one or more Treasures? When you do, up to that many target creatures gain double strike",
		Keyed:    swashbucklerSacrificed,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.HasSubtype("Treasure") {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, 0
		},
	}.Apply(effects.NewContext(g, item))
}

var swashbucklerSacrificed = effects.RegisterPermanentsPick("prompt-run/gap-b220-swashbuckler-extraordinaire-sacrifice", effects.PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		n := len(picked.Cards())
		if n == 0 {
			return nil
		}
		return effects.ReflexiveTrigger{
			Label:  "Swashbuckler Extraordinaire — up to that many target creatures gain double strike until end of turn",
			Body:   swashbucklerDoubleStrikeBody,
			Params: game.EffectParams{Amount: n},
		}.Apply(ctx)
	},
})

var swashbucklerDoubleStrikeBody = game.ReflexiveBody("gap-b220-swashbuckler-extraordinaire/double-strike",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"double strike"}, Label: "Swashbuckler Extraordinaire — double strike until end of turn"}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return effects.TargetCreature("up to that many target creatures").WithCount(0, p.Amount)
	})

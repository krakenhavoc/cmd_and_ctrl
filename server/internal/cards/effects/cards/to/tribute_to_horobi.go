package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tribute to Horobi // Echo of Death's Wail — a transforming Saga
// (Kamigawa: Neon Dynasty, #5493).
//
//	I, II — Each opponent creates a 1/1 black Rat Rogue creature token.
//	III   — Exile this Saga, then return it to the battlefield
//	        transformed under your control.
//
// Echo of Death's Wail, Enchantment Creature — Spirit 3/3: "Flying,
// haste. When this creature enters, gain control of all Rat tokens.
// Whenever this creature attacks, you may sacrifice another creature. If
// you do, draw a card."
//
// "Rat tokens" is every token with the creature type Rat on the
// battlefield, whoever controls it, taken for good (CR 611.2a, no
// duration). The attack trigger's "you may" is its own prompt and the
// draw is told what was sacrificed.
func init() {
	chapter := func(n int) game.TriggeredAbility {
		return effects.TriggerWithPurpose(effects.ChapterTrigger(n, effects.SagaChapterLabel("Tribute to Horobi", n, "each opponent creates a 1/1 black Rat Rogue creature token"),
			effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				for _, opp := range ctx.Opponents() {
					if err := (effects.CreateToken{Controller: opp, Template: effects.TokenCard("1/1 black Rat Rogue"), N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}))), game.Purpose{Tokens: 1})
	}
	effects.Register(effects.Spec{
		OracleID:     TributeToHorobiOracleID,
		Name:         "Tribute to Horobi",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			chapter(1),
			chapter(2),
			effects.ChapterTrigger(3, effects.SagaChapterLabel("Tribute to Horobi", 3, "exile it, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})

	effects.Register(effects.Spec{
		OracleID:        TributeToHorobiOracleID + "#1",
		Name:            "Echo of Death's Wail",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Echo of Death's Wail — gain control of all Rat tokens",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item).AsGroupMember()
					rat := ctx.TextSubtype("Rat")
					var rats []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsToken() && c.IsCreature() && c.HasSubtype(rat) && c.Controller != item.Controller {
							rats = append(rats, c.InstanceID)
						}
					}
					for _, id := range rats {
						if err := (effects.GainControl{
							Target:   id,
							Duration: game.IndefiniteDuration(),
							Label:    "Echo of Death's Wail — gain control of all Rat tokens",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			effects.WheneverThisAttacks("Echo of Death's Wail — you may sacrifice another creature; if you do, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Echo of Death's Wail — sacrifice another creature to draw a card?",
						effects.Creature(), true, echoOfDeathsWailSacrificed)
				}),
		},
	})
}

// tributeToHorobiOracleID is shared with the back face ("#1").
const TributeToHorobiOracleID = "81622cf6-3021-4ad4-933d-a7e3d8ac817f"

var echoOfDeathsWailSacrificed = effects.RegisterSacrificedOne("prompt-run/s82-neo-b18-echo-of-deaths-wail",
	func(ctx *effects.Context, _ uuid.UUID) error { return effects.DrawCards{N: 1}.Apply(ctx) })

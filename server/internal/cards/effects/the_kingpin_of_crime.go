package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Kingpin of Crime — Legendary Creature — Human Villain {1}{W}{B},
// 1/5 (Marvel Super Heroes, #4455):
//
//	"Extort (Whenever you cast a spell, you may pay {W/B}. If you do, each
//	 opponent loses 1 life and you gain that much life.)
//	 Whenever you attack, you may pay 2 life. If you do, until end of turn,
//	 creatures you control with toughness greater than their power assign
//	 combat damage equal to their toughness rather than their power."
//
// Extort is the shared keyword trigger. "Whenever you attack" is one
// trigger per declaration (OncePerBatch, Adeline's shape); the 2 life is
// paid as the ability resolves (Bringer of the Black Dawn's MayChoice),
// re-checked in the yes branch. The effect is a continuous effect from a
// resolving ability, so its set is locked at resolution (CR 611.2c): the
// creatures you control with toughness greater than power right then.
// Each is given Ghalta's flag (Characteristic.DamageByToughness) through a
// duration grant of the bundle below, until end of turn; the combat damage
// pass answers with the toughness whenever it is greater than the power.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3eeeb159-aa45-4a9b-9618-8542604b9830",
		Name:         "The Kingpin of Crime",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: mshA8KingpinGrant,
			Static: []game.StaticAbility{{
				Label: "This creature assigns combat damage equal to its toughness rather than its power.",
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.DamageByToughness = true
				},
			}},
			Text: "This creature assigns combat damage equal to its toughness rather than its power.",
		}},
		Triggered: []game.TriggeredAbility{
			Extort("The Kingpin of Crime"),
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "The Kingpin of Crime — you may pay 2 life; creatures with toughness greater than power assign combat damage by toughness",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "The Kingpin of Crime — pay 2 life so creatures assign combat damage equal to their toughness?",
						LifeCost: 2,
						Then:     mshA8KingpinAnswer,
					}.Apply(NewContext(g, item))
				})),
		},
	})
}

const mshA8KingpinGrant = "the-kingpin-of-crime/damage-by-toughness"

// mshA8KingpinAnswer is the "yes": pay 2 life and, if that was possible,
// give the creatures their flag until end of turn. Keyed on disk: never
// renamed, never reused.
var mshA8KingpinAnswer = MayYes("confirm/msh-a8-kingpin-pay-life", func(ctx *Context, _ game.ConfirmAnswer) error {
	controller := ctx.Controller()
	p := ctx.Game.PlayerByIDForEffect(controller)
	if p == nil || p.Eliminated || p.Life < 2 {
		return nil
	}
	if err := ctx.Game.PayLifeForEffect(ctx.Source(), controller, 2); err != nil {
		return err
	}
	return GrantAbilitiesFor{
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return c.IsCreature() && c.Controller == controller && c.CurrentToughness() > c.CurrentPower()
		},
		Keys:  []string{mshA8KingpinGrant},
		Label: "The Kingpin of Crime — assign combat damage by toughness",
	}.Apply(ctx)
})

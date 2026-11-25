package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cavalier of Night — Creature — Elemental Knight {2}{B}{B}{B}, 4/5 (Core Set
// 2020):
//
//	"Lifelink
//	 When this creature enters, you may sacrifice another creature. When you
//	 do, destroy target creature an opponent controls.
//	 When this creature dies, return target creature card with mana value 3 or
//	 less from your graveyard to the battlefield."
//
// The enters trigger is Anthropede's keyed sacrifice with a reflexive trigger
// (CR 603.12): the "may" asks for another creature, and the target is chosen
// only if one was sacrificed. The dies trigger is Balduvian Atrocity's return
// of a creature card with mana value 3 or less, with no further rider.
//
// No simplification.
func init() {
	const name = "Cavalier of Night"
	effects.Register(effects.Spec{
		OracleID:        "27c2b8ce-17b9-4f6c-8444-3bdc636dd413",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — you may sacrifice another creature; when you do, destroy target creature an opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, name+" — sacrifice another creature to destroy target creature an opponent controls?",
						effects.Creature(), true, cavalierOfNightSacrificed)
				}),
			effects.Targeting(
				effects.WhenThisDies(name+" — return target creature card with mana value 3 or less from your graveyard to the battlefield",
					effects.ReturnFirstLegalGraveyardTargetToBattlefield),
				effects.TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard",
					effects.YouOwn(), effects.Creature(), effects.ManaValueLE(3))),
		},
	})
}

// cavalierOfNightSacrificed is the "when you do" of the sacrifice. An on-disk
// key: never renamed, never reused.
var cavalierOfNightSacrificed = effects.RegisterSacrificedOne("prompt-run/cavalier-of-night-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.WhenYouDo("Cavalier of Night — destroy target creature an opponent controls", cavalierOfNightDestroyBody).Apply(ctx)
})

var cavalierOfNightDestroyBody = game.ReflexiveBody("cavalier-of-night/destroy", effects.SimpleBody(effects.DestroyFirstLegalCardTarget),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())
	}))

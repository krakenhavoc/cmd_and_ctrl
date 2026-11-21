package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bristlebud Farmer — Creature — Plant Druid {2}{G}{G}, 5/5 (The Big
// Score, #4468):
//
//	"Trample
//	 When this creature enters, create two Food tokens.
//	 Whenever this creature attacks, you may sacrifice a Food. If you do,
//	 mill three cards. You may put a permanent card from among them into
//	 your hand."
//
// The enters trigger is two Foods made as one creation (CR 111.2). The
// attack trigger is Cornered Crook's "you may sacrifice" (maySacrificeOne;
// any Food artifact, token or not, per the ruling), then Scout the City's
// mill-then-take: only the cards this mill put into the graveyard are
// offered, filtered to permanent cards (artifact, battle, creature,
// enchantment, land or planeswalker), and the pick is optional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3e44226b-51a2-4099-aff8-a5befab720cb",
		Name:            "Bristlebud Farmer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bristlebud Farmer — create two Food tokens",
				effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 2})),
			effects.WheneverThisAttacks("Bristlebud Farmer — you may sacrifice a Food; if you do, mill three cards",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Bristlebud Farmer — sacrifice a Food to mill three cards?",
						effects.HasSubtype("Food"), false, bristlebudFarmerSacrificed)
				}),
		},
	})
}

// bristlebudFarmerSacrificed is "If you do, mill three cards": keyed.
var bristlebudFarmerSacrificed = effects.RegisterSacrificedOne("prompt-run/bristlebud-farmer-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.MillToZone{N: 3, ThenTail: ctx.ItemTail(bristlebudFarmerMilled)}.Apply(ctx)
})

var bristlebudFarmerMilled = effects.CardTail("bristlebud-farmer/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.MayTakeOneFromAmongThem(ctx, o.IDs, effects.Permanent(),
		"Bristlebud Farmer — you may put a permanent card from among them into your hand")
})

package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "3e44226b-51a2-4099-aff8-a5befab720cb",
		Name:            "Bristlebud Farmer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bristlebud Farmer — create two Food tokens",
				Do(CreateToken{Template: FoodToken(), N: 2})),
			WheneverThisAttacks("Bristlebud Farmer — you may sacrifice a Food; if you do, mill three cards",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Bristlebud Farmer — sacrifice a Food to mill three cards?",
						HasSubtype("Food"), false, bristlebudFarmerSacrificed)
				}),
		},
	})
}

// bristlebudFarmerSacrificed is "If you do, mill three cards": keyed.
var bristlebudFarmerSacrificed = RegisterSacrificedOne("prompt-run/bristlebud-farmer-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return MillToZone{N: 3, ThenTail: ctx.ItemTail(bristlebudFarmerMilled)}.Apply(ctx)
})

var bristlebudFarmerMilled = CardTail("bristlebud-farmer/after-mill", func(ctx *Context, o game.TailOutcome) error {
	return mayTakeOneFromAmongThem(ctx, o.IDs, Permanent(),
		"Bristlebud Farmer — you may put a permanent card from among them into your hand")
})

package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Dalek Emperor — Legendary Artifact Creature — Dalek {5}{B}{R}, 6/6:
//
//	"Affinity for Daleks (This spell costs {1} less to cast for each
//	 Dalek you control.)
//	 Other Daleks you control have haste.
//	 At the beginning of combat on your turn, each opponent faces a
//	 villainous choice — That player sacrifices a creature of their
//	 choice, or you create a 3/3 black Dalek artifact creature token
//	 with menace."
//
// A villainous choice (CR 701.55, #3140) asked of each opponent in APNAP
// order, one at a time (CR 701.55d); a table with The Valeyard asks them
// again (CR 701.55c). An opponent with no creature may pick the
// sacrifice and lose nothing (CR 701.55b), which is printed. The token
// is the Dalek token card (menace).
//
// "Daleks" are the creature type, counted on the battlefield through
// Card.HasSubtype, so a changeling counts. The haste is a Layer 6 grant
// to the other Daleks the controller has, the Emperor not included.
//
// No simplifications.
var daleksEmperorChoice = NewVillainousChoice("the-dalek-emperor",
	"The Dalek Emperor — sacrifice a creature, or the controller creates a 3/3 Dalek?",
	VillainousOption{
		Label:    "Sacrifice a creature of your choice",
		Harm:     2,
		Possible: villainousCanSacrifice(Creature()),
		Body:     villainousSacrifice("a creature", Creature()),
	},
	VillainousOption{
		Label: "The controller of The Dalek Emperor creates a 3/3 black Dalek artifact creature token with menace",
		Harm:  2,
		Body: func(ctx *Context, _ Facing) error {
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("3/3 black Dalek artifact with menace"), N: 1}.Apply(ctx)
		},
	})

func init() {
	daleks := TribeFilter{Tribes: []string{"Dalek"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:     "07799012-7051-400e-adde-b54b41b80c70",
		Name:         "The Dalek Emperor",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for Daleks", OfCreatureType("Dalek")),
		},
		Static: []game.StaticAbility{TribalKeywordGrant(daleks, "haste")},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("The Dalek Emperor — each opponent faces a villainous choice",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return FaceVillainousChoice{Players: ctx.Opponents(), Choice: daleksEmperorChoice}.Apply(ctx)
				}),
		},
	})
}

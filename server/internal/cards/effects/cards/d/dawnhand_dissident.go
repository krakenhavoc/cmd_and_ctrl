package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// dawnhandDissidentAltCostKey is the offer a cast through Dawnhand
// Dissident's permission claims: the creature card's printed mana cost and
// three counters removed from among creatures you control. An on-disk key.
const DawnhandDissidentAltCostKey = "dawnhand_dissident"

// Dawnhand Dissident — Creature — Elf Warlock {B}, 1/2 (Lorwyn Eclipsed,
// #5406):
//
//	"{T}, Blight 1: Surveil 1.
//	 {T}, Blight 2: Exile target card from a graveyard.
//	 During your turn, you may cast creature spells from among cards you
//	 own exiled with this creature by removing three counters from among
//	 creatures you control in addition to paying their other costs."
//
// Both blights are part of the cost (BlightN, #4684): the activator names
// a creature they control, the Dissident included, and the -1/-1 counters
// all go on that one creature as the cost is paid (CR 701.68a); with no
// creature to blight neither ability can be activated (CR 701.68b).
//
// The exile is LINKED to this Dissident object (CR 607.2a,
// avB7ExileTargetLinked, Boiling Rock Rioter's), so "cards exiled with this creature" is exactly the cards its
// second ability exiled, and not those of an earlier incarnation
// (CR 400.7).
//
// The cast clause is Azula, Cunning Usurper's gated standing permission
// over exile (ExiledWithSource, gated on DuringYourTurn), narrowed to
// creature spells (CreatureOnly) and to cards the holder owns
// (OwnedByHolder). It is derived off the battlefield on every query, so it
// ends the moment the Dissident leaves; a spell already being cast is
// unaffected (the ruling). The spell's own timing is in force: a creature
// without flash is cast in your main phase with the stack empty. The price
// is the card's mana cost plus three counters of any kinds, split among
// your creatures however you choose (RemoveCountersInAddition, Quilled
// Greatwurm's offer, #5390), so a player whose creatures hold fewer than
// three counters in all cannot claim it (CR 118.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2cabf7a4-17af-45c7-b314-a84f5dc0d070",
		Name:         "Dawnhand Dissident",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Blight 1: Surveil 1.",
				Cost:    effects.Plus(effects.TapCost(), effects.BlightN(1)),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  dawnhandDissidentSurveil,
			},
			{
				Label:   "{T}, Blight 2: Exile target card from a graveyard.",
				Cost:    effects.Plus(effects.TapCost(), effects.BlightN(2)),
				Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
				Effect:  effects.AvB7ExileTargetLinked,
			},
		},
		GatedCastPermissions: []game.CastPermissionGate{{
			Condition: effects.DuringYourTurn(),
			Permission: game.CastPermission{
				Zone: game.ZoneExile,
				Filter: game.PermissionFilter{
					ExiledWithSource: true,
					OwnedByHolder:    true,
					CreatureOnly:     true,
				},
				CastOnly:                 true,
				AltCostKey:               DawnhandDissidentAltCostKey,
				RemoveCountersInAddition: 3,
				Label:                    "During your turn, cast a creature spell you own exiled with Dawnhand Dissident by removing three counters from among creatures you control (Dawnhand Dissident)",
			},
		}},
	})
}

// dawnhandDissidentSurveil is "Surveil 1."
func dawnhandDissidentSurveil(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return effects.Surveil{Player: ctx.Controller(), N: 1}.Apply(ctx)
}

package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glamdring, Foe-hammer // Gleam of Death — an Adventure card (CR 715, The
// Hobbit, #4454), oracle 2802069f…:
//
//	Glamdring, Foe-hammer — Legendary Artifact — Equipment {2}
//	  "Instant and sorcery spells you cast cost {X} less to cast, where X is
//	   equipped creature's power.
//	   Equip {2}"
//	Gleam of Death — Sorcery — Adventure {3}{U}
//	  "Mill six cards, then put all instant and sorcery cards from among
//	   them into your hand. (Then exile this card. You may cast the artifact
//	   later from exile.)"
//
// The discount is Archmage of Runes' cost modifier with an amount computed
// per cast (CostsLessEach, as Animar's counters are): the power of the
// creature the Equipment is attached to, read after the layers as the cost is
// determined (CR 601.2f). q.Source is a fresh copy of the Equipment, so its
// AttachedTo is live. An unattached Glamdring, or a host with power 0 or less,
// reduces nothing, and the reduction spends against generic mana only.
//
// Gleam of Death is Cantankerous Keepers' mill: the clause after it runs on
// the cards that really reached the graveyard, and takes every instant and
// sorcery among them with no prompt.
//
// No simplification.
const glamdringFoeHammerOracleID = "2802069f-201e-43a7-b5d3-43a95951a2ec"

func init() {
	effects.Register(effects.Spec{
		OracleID:     glamdringFoeHammerOracleID,
		Name:         "Glamdring, Foe-hammer",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(glamdringFoeHammerDiscount,
				"Instant and sorcery spells you cast cost {X} less to cast, where X is equipped creature's power.",
				effects.YourSpell(), effects.InstantOrSorcerySpell()),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
	effects.Register(effects.Spec{
		OracleID:     glamdringFoeHammerOracleID + "#1",
		Name:         "Gleam of Death",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.MillToZone{N: 6, ThenTail: ctx.ItemTail(gleamOfDeathMilled)}.Apply(ctx)
		},
	})
}

// glamdringFoeHammerDiscount is X: the equipped creature's power.
func glamdringFoeHammerDiscount(q game.CostQuery) int {
	if q.Game == nil || q.Source.AttachedTo.Kind != game.TargetCard {
		return 0
	}
	host, ok := q.Game.LookupCardForEffect(q.Source.AttachedTo.ID)
	if !ok || !host.IsCreature() {
		return 0
	}
	return max(0, host.PowerForComparison())
}

// gleamOfDeathMilled is the clause after the mill: every milled instant and
// sorcery card still in the graveyard goes to its owner's hand.
var gleamOfDeathMilled = effects.CardTail("glamdring-foe-hammer/gleam-of-death-after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.HobA11AllMilledToHand(ctx, o, effects.Or(effects.Instant(), effects.Sorcery()))
})

package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frostwalk Bastion — Snow Land (Modern Horizons):
//
//	"{T}: Add {C}.
//	 {1}{S}: Until end of turn, this land becomes a 2/3 Construct artifact
//	 creature. It's still a land. ({S} can be paid with one mana from a snow
//	 source.)
//	 Whenever this land deals combat damage to a creature, tap that creature and
//	 it doesn't untap during its controller's next untap step."
//
// The animation is a scoped effect on this permanent until end of turn: layer 4
// types (artifact, creature) and subtype (Construct), layer 7b base power and
// toughness 2/3, with the land types kept (CR 205.1b); it is colorless, as the
// land is. It is the same object only while it stays on the battlefield (CR
// 400.7). The trigger names the damaged creature as the object it was (CR
// 400.7) and taps it and keeps it tapped through its controller's next untap
// step (CR 502.3), if it survived.
//
// No simplification.
const frostwalkBastionFreezeLabel = "Frostwalk Bastion — tap that creature; it doesn't untap during its controller's next untap step"

func init() {
	effects.Register(effects.Spec{
		OracleID:      "ae4a18ec-70a3-4d21-b9e5-b13ab4901600",
		Name:          "Frostwalk Bastion",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.TapForColorless()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{S}: Until end of turn, this land becomes a 2/3 Construct artifact creature. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{1}{S}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Artifact", "Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Construct")...),
					}, game.SetBasePTMods(2, 3)...),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Frostwalk Bastion — a 2/3 Construct artifact creature until end of turn",
				}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: effects.ThisDealtCombatDamageToACreature,
			Key:       frostwalkBastionFreezeLabel,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, frostwalkBastionFreezeLabel)
				item.Params.Object = effects.DamagedCreatureRef(g, ev.Target)
				return item
			},
			Effect: effects.OldB166TapAndFreezeDamagedCreature(frostwalkBastionFreezeLabel),
		}},
	})
}

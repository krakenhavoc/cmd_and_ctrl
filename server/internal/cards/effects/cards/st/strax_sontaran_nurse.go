package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Strax, Sontaran Nurse — Legendary Creature — Alien Cleric {3}{R}{G}, 5/5
// (Doctor Who):
//
//	"Vigilance, trample
//	 Grenades! — {2}, {T}, Sacrifice an artifact: Choose a player at random. When
//	 you do, Strax fights another target creature that player controls.
//	 Glory of Battle — Whenever Strax deals damage to a creature, put a +1/+1
//	 counter on Strax."
//
// The random pick is one of the players still in the game (Strax's controller
// included) drawn from the keyed stream (ChooseAtRandomForEffect). It raises the
// printed "when you do" as a CR 603.12 reflexive trigger, carrying the chosen
// player, that chooses its own "another target creature that player controls" as
// it goes on the stack; with no such creature it does nothing. Strax fights it
// if both are still legal (CR 701.14b). Glory of Battle counts any damage Strax
// deals to a creature, combat or from a fight, and fires once per damage event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ffde2558-2168-4b40-a31a-86d160bdb289",
		Name:            "Strax, Sontaran Nurse",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Grenades! — {2}, {T}, Sacrifice an artifact: Choose a player at random. When you do, Strax fights another target creature that player controls.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Effect:  straxGrenades,
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtDamageToACreature,
				"Glory of Battle — put a +1/+1 counter on Strax",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

// straxGrenades chooses a player at random among those still in the game and
// raises the reflexive fight.
func straxGrenades(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var players []uuid.UUID
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			players = append(players, p.ID)
		}
	}
	pick := g.ChooseAtRandomForEffect(effects.RandomDraw(ctx), players, 1)
	if len(pick) != 1 {
		return nil
	}
	return effects.ReflexiveTrigger{
		Label:  "Grenades! — Strax fights another target creature that player controls",
		Body:   straxFightBody,
		Params: game.EffectParams{Player: pick[0]},
	}.Apply(ctx)
}

var straxFightBody = game.ReflexiveBody("strax-sontaran-nurse/fight",
	effects.SimpleBody(effects.B30SourceFightsFirstLegalTarget),
	func(sourceID uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return effects.TargetCreature("another target creature that player controls", effects.And(effects.OtherThan(sourceID), effects.ControlledBy(p.Player)))
	})

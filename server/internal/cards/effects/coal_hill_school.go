package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coal Hill School — Plane — Earth (Doctor Who, #4870):
//
//	"Whenever a player casts a historic spell, that player draws a card.
//	 (Artifacts, legendaries, and Sagas are historic.)
//	 Whenever chaos ensues, return target historic card from your
//	 graveyard to your hand."
//
// A plane (ADR 0216 §6): both abilities function only while it is face up
// in the command zone (CR 901.7), and the planar controller controls them
// (CR 901.6). The first watches every player's cast; "historic" is an
// artifact, a legendary or a Saga (CR 700.6), read off the spell on the
// stack, and the caster draws. The chaos ability (CR 311.7) targets a
// historic card in the planar controller's graveyard; with none it is
// removed without a prompt (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "51a53042-6baa-442a-84ec-2f16fdc12087",
		Name:         "Coal Hill School",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, aPlayerCastAHistoricSpell, "Coal Hill School — that player draws a card",
				func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: triggeringActor(item), N: 1}.Apply(NewContext(g, item))
				}),
			Targeting(WheneverChaosEnsues("Coal Hill School — return target historic card from your graveyard to your hand", returnTargetCardToHand),
				TargetCardInGraveyard("target historic card from your graveyard", b14Historic(), YouOwn())),
		},
	})
}

// aPlayerCastAHistoricSpell is "whenever a player casts a historic spell"
// (CR 700.6): any caster, the spell on the stack an artifact, a legendary
// or a Saga.
func aPlayerCastAHistoricSpell(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && b09IsHistoric(spell)
}

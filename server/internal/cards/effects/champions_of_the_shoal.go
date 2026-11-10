package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Champions of the Shoal — Creature — Merfolk Soldier {3}{U}, 4/6
// (Lorwyn Eclipsed, #4458):
//
//	"As an additional cost to cast this spell, behold a Merfolk and
//	 exile it. (Exile a Merfolk you control or a Merfolk card from your
//	 hand.)
//	 Whenever this creature enters or becomes tapped, tap up to one
//	 target creature and put a stun counter on it.
//	 When this creature leaves the battlefield, return the exiled card to
//	 its owner's hand."
//
// Champion of the Clachan's cost and leave trigger (#4687,
// BeholdAndExileCost): one Merfolk, a permanent you control or a card in
// your hand (a changeling counts), exiled with the spell on the stack and
// linked to the permanent the spell becomes (CR 607.2q); a countered
// Champions leaves it in exile (the ruling). The middle ability is one
// trigger with two conditions: entering, or becoming tapped (Attentive
// Sunscribe's ThisBecameTapped: the tap event, or an attack that taps
// it, so a vigilance attack is not one). It taps up to one target
// creature and puts one stun counter on it (Cryotheory Adept's
// tapTargetAndStunIt); with no target it does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "fc76ec95-f001-44df-89b2-b0f1b10c8e11",
		Name:           "Champions of the Shoal",
		Completeness:   CompletenessFull,
		AdditionalCost: BeholdAndExileCost("a", "Merfolk"),
		Triggered: []game.TriggeredAbility{
			Targeting(OnAny([]game.EventKind{game.EventETB, game.EventTapCard, game.EventAttack},
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return (ev.Kind == game.EventETB && ev.CardID == source.InstanceID) || ThisBecameTapped(ev, source, lki, g)
				}, "Champions of the Shoal — tap up to one target creature and put a stun counter on it",
				tapTargetAndStunIt),
				TargetCreature("up to one target creature").WithCount(0, 1)),
			ReturnTheExiledCardToHand("Champions of the Shoal"),
		},
	})
}

package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nether Traitor — Creature — Spirit {B}{B}, 1/1:
//
//		"Haste
//		 Shadow (This creature can block or be blocked by only creatures
//		 with shadow.)
//		 Whenever another creature is put into your graveyard from the
//		 battlefield, you may pay {B}. If you do, return this card from
//		 your graveyard to the battlefield."
//
//	  - Haste and shadow are printed keywords the engine already enforces.
//	  - The trigger watches from the GRAVEYARD only (InGraveyard, #925),
//	    the one zone where "return this card from your graveyard" can do
//	    anything. "You" is the card's owner (CR 108.4).
//	  - "Your graveyard" is the owner's: a creature goes to ITS owner's
//	    graveyard whoever controlled it, so a stolen creature that dies
//	    counts for its owner's Traitor and a creature of the Traitor's
//	    owner that died under an opponent's control still does. Any
//	    creature card or token counts, since a token is put into the
//	    graveyard before it ceases to exist.
//	  - The Traitor must already be in the graveyard when the other
//	    creature is put there. One that died in the same simultaneous
//	    exit, or after the creature, never saw it.
//	  - "You may pay {B}" is asked when the ability RESOLVES (MayPay), not
//	    when it triggers (#2494), and a payment is only offered while the
//	    Traitor is still the graveyard object that triggered the ability.
//	    The card is checked again after the payment, so one that left in
//	    the meantime stays gone (CR 400.7).
//
// "At the same time" is read from the actual simultaneous-exit set,
// not Event.Batch, which spans an entire resolution. A Traitor that
// entered its owner's graveyard in an earlier instruction sees a later
// creature arrive during that same resolution.
func init() {
	Register(Spec{
		OracleID:        "6f16c3ac-a9b8-47e6-b18b-ae37c74d44a0",
		Name:            "Nether Traitor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste", "shadow"},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventLTB, netherTraitorSeesACreatureArrive,
				"Nether Traitor — you may pay {B} to return it to the battlefield", netherTraitorMayPay)),
		},
	})
}

// netherTraitorSeesACreatureArrive is the trigger condition: another
// creature went from the battlefield to the Traitor's owner's
// graveyard, and the Traitor was in that graveyard first.
func netherTraitorSeesACreatureArrive(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := diedCreature(ev, g)
	if !ok || dead.Owner != source.Owner {
		return false
	}
	return netherTraitorWasThere(g, source, ev)
}

// netherTraitorWasThere reports whether the Traitor sat in the
// graveyard before the creature arrived. Only membership of both
// objects in the same simultaneous exit means they died together;
// Event.Batch may cover several sequential instructions.
func netherTraitorWasThere(g *game.Game, source *game.Card, ev game.Event) bool {
	return !g.InSameSimultaneousExitForEffect(source.InstanceID, ev.CardID)
}

func netherTraitorMayPay(g *game.Game, item *game.StackItem) error {
	return mayPayForGraveyardReturn(g, item, "{B}", "Nether Traitor — pay {B} to return it to the battlefield?")
}

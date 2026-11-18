package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tameshi, Reality Architect — Legendary Creature — Moonfolk Wizard {2}{U},
// 2/3 (Kamigawa: Neon Dynasty):
//
//	"Whenever one or more noncreature permanents are returned to hand, draw
//	 a card. This ability triggers only once each turn.
//	 {X}{W}, Return a land you control to its owner's hand: Return target
//	 artifact or enchantment card with mana value X or less from your
//	 graveyard to the battlefield. Activate only as a sorcery."
//
// The trigger is a batch (OncePerBatch, CR 603.2c): every noncreature
// permanent, whoever controls or owns it, that goes from the battlefield to a
// hand in one batch of events is one trigger, and the once-per-turn claim
// (TriggersOnlyOnceEachTurn) is spent only by a trigger that happened. A
// permanent is a noncreature permanent by the types it had as it left
// (EventLTB's last-known types, CR 603.10a), so a Vehicle crewed into a
// creature that is bounced does not count, and a bounced animated land does
// not. A creature bounced in the same batch does not stop the trigger. A token
// returned to hand counts: it was returned there before it ceased to exist.
// A Tameshi that is returned to hand with another permanent sees it
// (a leaves-the-battlefield look-back).
//
// The activation is a sorcery-speed ability whose {X} the activator announces;
// the returned land is a cost, so the land can be paid and still be the one
// returned (CR 602.2b). The target is bound to the announced X
// (WithManaValueAtMostX), judged at announce and again at resolution
// (CR 608.2b); an {X} card counts X as 0 (CR 202.3). The card returns under
// its owner's control, which is the activator's: "your graveyard".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b47b3cfe-7005-4b33-9d41-83b60808e12b",
		Name:         "Tameshi, Reality Architect",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(OncePerBatch(On(game.EventLTB, tameshiNoncreaturePermanentReturned,
				"Tameshi, Reality Architect — draw a card", Do(DrawCards{N: 1})))),
		},
		Activated: []ActivatedAbility{{
			Label:        "{X}{W}, Return a land you control to its owner's hand: Return target artifact or enchantment card with mana value X or less from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{X}{W}"), ReturnAPermanentToHand("a land you control", Land(), YouControl())),
			SorcerySpeed: true,
			Targets: TargetCardInGraveyard("target artifact or enchantment card with mana value X or less from your graveyard",
				YouOwn(), Or(Artifact(), Enchantment())).WithManaValueAtMostX(),
			Effect: returnFirstLegalGraveyardTargetToBattlefield,
		}},
	})
}

// tameshiNoncreaturePermanentReturned is "one or more noncreature permanents
// are returned to hand": a permanent that left the battlefield for a hand
// and was not a creature as it left. An event that carries no last-known
// types (one built by hand) is judged on the card as it now sits.
func tameshiNoncreaturePermanentReturned(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventLTB || ev.NewZone != game.ZoneHand {
		return false
	}
	if was, known := ev.WasType("creature"); known {
		return !was
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsCreature()
}

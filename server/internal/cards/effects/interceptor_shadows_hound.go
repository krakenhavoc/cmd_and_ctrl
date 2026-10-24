package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Interceptor, Shadow's Hound — Legendary Creature — Dog {2}{B}{B}, 4/3
// (Final Fantasy VI; deck request #3655, seam #3771):
//
//	"Menace
//	 Assassins you control have menace.
//	 Whenever you attack with one or more legendary creatures, you may
//	 pay {2}{B}. If you do, return this card from your graveyard to the
//	 battlefield tapped and attacking."
//
// Four pieces, all real:
//
//   - Its own menace is a printed keyword; "Assassins you control have
//     menace" is a layer-6 grant over the Assassin subtype. Interceptor
//     is a Dog, so the grant does not reach it and the printed keyword
//     is not redundant.
//   - "Whenever you attack with one or more legendary creatures" is one
//     trigger per attack declaration. The engine emits EventAttack per
//     creature, the predicate keeps only the legendary ones, and
//     OncePerBatch fires on the first of them (CR 603.2c) — a Commander
//     attack with three legends is one offer, not three.
//   - The trigger watches from the GRAVEYARD and only from there
//     (InGraveyard, #925): a copy on the battlefield has nothing to
//     return. "You" inside it is the card's owner (CR 108.4).
//   - The payment and the return. The question is asked in the declare
//     attackers step (InThisStep) so it cannot be answered after
//     blockers; the card is checked against the graveyard as the trigger
//     resolves and again when it moves (CR 400.7), and it enters through
//     the CR 506.3c door tapped and attacking whichever player,
//     planeswalker or battle its controller picks. It was never declared
//     as an attacker, so it triggers nothing that watches attacks and
//     does not prompt its own trigger again (CR 508.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8571bd94-66c0-4cca-b10f-2c34062d008e",
		Name:            "Interceptor, Shadow's Hound",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Assassin"}, YoursOnly: true}, "menace"),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(InGraveyard(On(game.EventAttack, attackedWithALegendaryCreature,
				"Interceptor, Shadow's Hound — you may pay {2}{B}: return it from your graveyard tapped and attacking",
				interceptorMayReturn))),
		},
	})
}

// attackedWithALegendaryCreature is "whenever you attack with one or
// more legendary creatures", read per attacker: the attacker is
// declared by you and is a legendary creature.
func attackedWithALegendaryCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.IsLegendary()
}

// interceptorMayReturn is the trigger's body: the offer, then the put.
func interceptorMayReturn(g *game.Game, item *game.StackItem) error {
	card := item.SourceCardID
	if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return MayPay{
		Chooser:    item.Controller,
		Cost:       "{2}{B}",
		Question:   "Interceptor, Shadow's Hound — pay {2}{B}: return it from your graveyard to the battlefield tapped and attacking?",
		InThisStep: true,
		Then:       interceptorPaid,
	}.Apply(NewContext(g, item))
}

// interceptorPaid is the paid half, keyed (#4264).
var interceptorPaid = PayUnlessBranches("pay-unless/interceptor-shadows-hound-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return putAttackingYourChoice(ctx, ctx.Source(), game.ZoneGraveyard, "Interceptor, Shadow's Hound")
	}, nil)

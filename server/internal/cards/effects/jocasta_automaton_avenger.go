package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jocasta, Automaton Avenger — {3} Legendary Artifact Creature — Robot Hero
// 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever your commander deals combat damage to a player, put a +1/+1
//	 counter on Jocasta.
//	 Whenever you attack with your commander, if this card is in your
//	 graveyard, you may return it to the battlefield tapped and attacking."
//
// The counter trigger is Kediss' commander-damage condition (a commander of
// yours dealt combat damage to a player other than you, which in a free-for-
// all is every player it can hurt). The return is Interceptor, Shadow's
// Hound's graveyard trigger without the payment: once per attack declaration,
// from the graveyard only (an intervening if, CR 603.4, checked again as it
// resolves), a yes/no for "may", and the shared tapped-and-attacking entry
// that asks which defender it attacks. It was never declared as an attacker
// (CR 508.4), so nothing watching attackers fires for it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "787e6a7b-64f9-48ba-bccd-68dda3489b0f",
		Name:            "Jocasta, Automaton Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10CommanderYouControlDealtCombatDamageToOpponent(ev, source, g)
			}, "Jocasta — put a +1/+1 counter on Jocasta", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			OncePerBatch(InGraveyard(On(game.EventAttack, msc13AttackedWithYourCommander,
				"Jocasta — you may return her from your graveyard to the battlefield tapped and attacking",
				msc13JocastaOffer))),
		},
	})
}

// msc13AttackedWithYourCommander is "whenever you attack with your
// commander": the attack declaration is yours and the attacker is a
// commander you control.
func msc13AttackedWithYourCommander(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCommander && c.Controller == source.Controller
}

func msc13JocastaOffer(g *game.Game, item *game.StackItem) error {
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return MayChoice{
		Question: "Jocasta — return her from your graveyard to the battlefield tapped and attacking?",
		YesLabel: "Return her",
		NoLabel:  "Leave her",
		Then:     msc13JocastaReturn,
	}.Apply(NewContext(g, item))
}

var msc13JocastaReturn = MayYes("confirm/jocasta-return", func(ctx *Context, _ game.ConfirmAnswer) error {
	return putAttackingYourChoice(ctx, ctx.Source(), game.ZoneGraveyard, "Jocasta, Automaton Avenger")
})

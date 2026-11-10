package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Persistent Marshstalker — Creature — Rat Berserker {1}{B}, 3/1 (Bloomburrow,
// #4467):
//
//	"This creature gets +1/+0 for each other Rat you control.
//	 Threshold — Whenever you attack with one or more Rats, if there are
//	 seven or more cards in your graveyard, you may pay {2}{B}. If you do,
//	 return this card from your graveyard to the battlefield tapped and
//	 attacking."
//
// The first is a layer 7c self-pump counting the other Rats through the
// layered view (a changeling counts, hobA9OtherSubtypeYouControl).
// The second is Interceptor, Shadow's Hound's shape: one trigger per attack
// declaration (OncePerBatch), working only from the graveyard (InGraveyard,
// CR 113.6), whose pay-then-return enters through the CR 506.3c door tapped
// and attacking whatever its controller picks (putAttackingYourChoice). The
// threshold is an intervening if (CR 603.4): asked as the attack is
// declared and again as the ability resolves, before the "you may" is
// asked. The card was never declared as an attacker, so nothing that
// watches attacks sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77eb76cf-a361-44e0-aeee-e8ef4b695799",
		Name:         "Persistent Marshstalker",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Gets +1/+0 for each other Rat you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += hobA9OtherSubtypeYouControl(g, source, source.TextSubtype("Rat"))
			},
		}},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondYourGraveyardAtLeast{N: 7},
				OncePerBatch(InGraveyard(On(game.EventAttack, stdB27AttackedWithARat,
					"Persistent Marshstalker — you may pay {2}{B}: return it from your graveyard tapped and attacking",
					persistentMarshstalkerMayReturn)))),
		},
	})
}

// stdB27AttackedWithARat is "whenever you attack with one or more Rats",
// read per attacker, and the threshold half of the condition as the
// attack is declared: the attacker is declared by you and is a Rat, and
// there are seven or more cards in your graveyard.
func stdB27AttackedWithARat(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || !attackDeclaredByYou(ev, source.Controller) || b31GraveyardSize(g, source.Controller) < 7 {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Rat"))
}

// persistentMarshstalkerMayReturn is the offer, then the put.
func persistentMarshstalkerMayReturn(g *game.Game, item *game.StackItem) error {
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return MayPay{
		Chooser:    item.Controller,
		Cost:       "{2}{B}",
		Question:   "Persistent Marshstalker — pay {2}{B}: return it from your graveyard to the battlefield tapped and attacking?",
		InThisStep: true,
		Then:       persistentMarshstalkerPaid,
	}.Apply(NewContext(g, item))
}

// persistentMarshstalkerPaid is the paid half, keyed. An on-disk key.
var persistentMarshstalkerPaid = PayUnlessBranches("pay-unless/persistent-marshstalker-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return putAttackingYourChoice(ctx, ctx.Source(), game.ZoneGraveyard, "Persistent Marshstalker")
	}, nil)

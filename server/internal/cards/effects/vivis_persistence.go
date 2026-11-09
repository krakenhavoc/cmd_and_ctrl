package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vivi's Persistence — Instant {1}{R} (Final Fantasy, #4462):
//
//	"Create a 0/1 black Wizard creature token with "Whenever you cast a
//	 noncreature spell, this token deals 1 damage to each opponent."
//	 Whenever your commander enters or attacks, you may pay {2}. If you do,
//	 return this card from your graveyard to your hand."
//
// The spell is the set's noncreature-cast Wizard (Cornered by Black
// Mages' token). The second ability works only from the graveyard, as
// Endless Ranks of HYDRA's does: it watches a commander its owner owns
// entering the battlefield or being declared as an attacker, offers the {2}
// as it resolves (a decline returns nothing), and checks the card is still
// in the graveyard when it returns it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd4a1f4b-674d-41f3-a9a6-c2b65bf5c1ab",
		Name:         "Vivi's Persistence",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: NoncreatureCastWizardToken(), N: 1}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{vivisPersistenceReturn()},
	})
}

// vivisPersistenceReturn is the graveyard half of Vivi's Persistence.
func vivisPersistenceReturn() game.TriggeredAbility {
	t := OnAny([]game.EventKind{game.EventETB, game.EventAttack}, msc10YourCommanderEntersOrAttacks,
		"Vivi's Persistence — you may pay {2} to return this card to your hand",
		func(g *game.Game, item *game.StackItem) error {
			return MayPay{
				Chooser:  item.Controller,
				Cost:     "{2}",
				Question: "Vivi's Persistence — pay {2} to return it from your graveyard to your hand?",
				Then:     vivisPersistencePaid,
			}.Apply(NewContext(g, item))
		})
	t.Zones = []game.ZoneKind{game.ZoneGraveyard}
	return t
}

// vivisPersistencePaid returns the card on a payment. An on-disk key:
// never renamed, never reused.
var vivisPersistencePaid = PayUnlessBranches("pay-unless/vivis-persistence-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return returnThisCardFromYourGraveyardToYourHand(ctx.Game, ctx.Item)
	}, nil)

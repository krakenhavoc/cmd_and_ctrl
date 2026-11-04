package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Neutrinos — Legendary Creature — Elf Rebel {2}{R}{W}, 2/4
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Flying
//	 Alliance — Whenever another creature you control enters, The
//	 Neutrinos get +1/+0 until end of turn.
//	 Whenever The Neutrinos attack, exile up to one target creature you
//	 own, then return it to the battlefield under your control tapped
//	 and attacking."
//
// Alliance is an ability word (CR 207.2c) over
// WheneverAnotherCreatureEntersUnderYourControl. The attack trigger is a
// flicker whose return is the shared tapped-and-attacking entry
// (putAttackingYourChoice, Jocasta's): the card comes back from exile
// only if it really got there, as a new object (CR 400.7) under the
// controller's control, attacking a player, planeswalker or battle its
// controller chooses (CR 508.4). It was never declared as an attacker,
// so its "whenever this attacks" does not trigger. "You own" is the
// owner, so a creature of yours an opponent controls is a legal target,
// and The Neutrinos may blink themselves. A token that is exiled ceases
// to exist and does not return.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b1462c82-7702-451b-b612-990353a79ac4",
		Name:            "The Neutrinos",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("The Neutrinos — Alliance: +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !onBattlefield(g, item.SourceCardID) {
						return nil
					}
					return BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "The Neutrinos — Alliance"}.Apply(NewContext(g, item))
				}),
			Targeting(
				WheneverThisAttacks("The Neutrinos — exile up to one target creature you own, then return it tapped and attacking",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return ExileTarget{Target: id, ThenTail: ctx.ItemTail(theNeutrinosReturnTail)}.Apply(ctx)
					}),
				TargetCreature("up to one target creature you own", YouOwn()).WithCount(0, 1),
			),
		},
	})
}

// theNeutrinosReturnTail is "then return it to the battlefield under
// your control tapped and attacking", run once the card has landed in
// exile. An on-disk key.
var theNeutrinosReturnTail = CardTail("then/the-neutrinos-return-attacking", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	return putAttackingYourChoice(ctx, o.IDs[0], game.ZoneExile, "The Neutrinos")
})

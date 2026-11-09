package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Godo, Bandit Warlord — Legendary Creature — Human Barbarian {5}{R}, 3/3
// (Final Fantasy, #4462):
//
//	"When Godo enters, you may search your library for an Equipment card,
//	 put it onto the battlefield, then shuffle.
//	 Whenever Godo attacks for the first time each turn, untap it and all
//	 Samurai you control. After this phase, there is an additional combat
//	 phase."
//
// The enters trigger is Captain America's search with no mana value cap:
// optional, to the battlefield, shuffled either way (CR 701.23b).
//
// "For the first time each turn" is Aurelia's per-object attack tally
// (attacksForTheFirstTimeThisTurn), so a Godo that left and came back is a
// new object that may trigger again (CR 400.7). The untap is Godo, if it is
// still the same object, and every Samurai you control (a changeling counts);
// the extra combat comes after this phase with no extra main phase (the
// ruling, CR 500.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f890d51a-dc90-4179-8423-4738a72a527c",
		Name:         "Godo, Bandit Warlord",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Godo, Bandit Warlord — you may search for an Equipment card and put it onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: b09IsEquipmentCard,
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Optional:  true,
						Shuffle:   true,
						Reason:    "Godo, Bandit Warlord — an Equipment card, onto the battlefield",
					}.Apply(NewContext(g, item))
				}),
			On(game.EventAttack, attacksForTheFirstTimeThisTurn,
				"Godo, Bandit Warlord — untap it and all Samurai you control; an additional combat phase after this phase",
				Do(Instruction(godoUntapItAndSamurai), ExtraCombatAfterThisPhase())),
		},
	})
}

// godoUntapItAndSamurai is "untap it and all Samurai you control".
func godoUntapItAndSamurai(ctx *Context) error {
	self := ctx.Source()
	if ctx.Item != nil && sourceIsNewObject(ctx.Game, ctx.Item) {
		self = uuid.Nil
	}
	return b16UntapAllYouControlMatching(ctx, ctx.Controller(), func(c game.Card) bool {
		return c.InstanceID == self || c.HasSubtype("Samurai")
	})
}

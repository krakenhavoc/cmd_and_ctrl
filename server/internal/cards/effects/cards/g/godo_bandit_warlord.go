package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "f890d51a-dc90-4179-8423-4738a72a527c",
		Name:         "Godo, Bandit Warlord",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Godo, Bandit Warlord — you may search for an Equipment card and put it onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.B09IsEquipmentCard,
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Optional:  true,
						Shuffle:   true,
						Reason:    "Godo, Bandit Warlord — an Equipment card, onto the battlefield",
					}.Apply(effects.NewContext(g, item))
				}),
			effects.On(game.EventAttack, effects.AttacksForTheFirstTimeThisTurn,
				"Godo, Bandit Warlord — untap it and all Samurai you control; an additional combat phase after this phase",
				effects.Do(effects.Instruction(godoUntapItAndSamurai), effects.ExtraCombatAfterThisPhase())),
		},
	})
}

// godoUntapItAndSamurai is "untap it and all Samurai you control".
func godoUntapItAndSamurai(ctx *effects.Context) error {
	self := ctx.Source()
	if ctx.Item != nil && effects.SourceIsNewObject(ctx.Game, ctx.Item) {
		self = uuid.Nil
	}
	samurai := ctx.TextSubtype("Samurai")
	return effects.B16UntapAllYouControlMatching(ctx, ctx.Controller(), func(c game.Card) bool {
		return c.InstanceID == self || c.HasSubtype(samurai)
	})
}

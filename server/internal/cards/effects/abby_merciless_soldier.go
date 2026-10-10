package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abby, Merciless Soldier — Legendary Creature — Human Survivor 4/4
// for {1}{R}{G}:
//
//	"When you cast this spell, create a number of 1/1 black Fungus
//	 Zombie creature tokens named Cordyceps Infected equal to the amount
//	 of mana spent to cast it.
//	 Abby enters under the control of an opponent of your choice.
//	 Partner—Survivors"
//
// Two halves that belong to two different players. The cast trigger is
// the CASTER's — a "when you cast this spell" ability fires from the
// stack (Desolation Twin's FromStack), so the tokens are made even if
// Abby is countered, and they are made under the caster. Abby herself
// then enters under an opponent of the caster's choice (ADR 0102),
// declared as HARM: a 4/4 is a gift, but the deck that plays her is
// sending the opponent a body that its own tokens were built to face.
//
// The amount is read as the spell is cast, into the item's params, off
// the payment record (#761) — not at resolution, when a countered Abby
// has already left the stack.
//
// Partner—Survivors is a deck-building rule (CR 702.124i) that
// internal/deck reads off the oracle text: Abby and another Survivors
// card are a legal pair of commanders (#2874).
//
// Declared simplification (weaker than printed): with strict mana off
// the engine does not know what was spent, and "unknown" is always the
// weaker answer (docs/adding-cards.md), so no tokens are made.
func init() {
	Register(Spec{
		OracleID:     "e05efd01-0202-4bd6-bac6-c6210819d463",
		Name:         "Abby, Merciless Soldier",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"With strict mana off, the game doesn't track what you spent, so Abby makes no Cordyceps Infected tokens — turn strict mana on for it to count.",
		},
		Replacements: []game.ReplacementEffect{
			EntersUnderTheControlOfAnOpponentOfYourChoice("Abby, Merciless Soldier", game.ControlForHarm),
		},
		Triggered: []game.TriggeredAbility{{
			FromStack: true,
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			// The controller is the CASTER (ev.Actor), and the amount is
			// what was spent to cast it — both facts of the moment of the
			// cast, filled in here beside the declared Effect.
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Abby, Merciless Soldier — create a Cordyceps Infected for each mana spent")
				item.Controller, item.Owner = ev.Actor, ev.Actor
				item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{
					Controller: item.Controller,
					Template:   TokenCard("1/1 black Fungus Zombie named Cordyceps Infected"),
					N:          item.Params.Amount,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

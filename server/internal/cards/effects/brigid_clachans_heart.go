package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brigid, Clachan's Heart // Brigid, Doun's Mind — {2}{W} Legendary
// Creature — Kithkin Warrior 3/2 // Legendary Creature — Kithkin Soldier
// 3/2 (Lorwyn Eclipsed, #4458):
//
//	Front: "Whenever this creature enters or transforms into Brigid,
//	        Clachan's Heart, create a 1/1 green and white Kithkin
//	        creature token.
//	        At the beginning of your first main phase, you may pay {G}.
//	        If you do, transform Brigid."
//	Back:  "{T}: Add X {G} or X {W}, where X is the number of other
//	        creatures you control.
//	        At the beginning of your first main phase, you may pay {W}.
//	        If you do, transform Brigid."
//
// Ashling, Rekindled's shape. The front's first trigger is one ability
// with two conditions, entering or turning over to the front face (CR
// 701.27e, Huntmaster of the Fells), making the printed tecl Kithkin
// row. Each face's "you may pay … If you do, transform Brigid" is a
// MayPay whose paid branch is Ashling's ashlingTransformPaid, so it
// does nothing if Brigid has turned over since the trigger went on the
// stack (CR 701.27f, the ruling).
//
// The back face's mana ability counts the other creatures its
// controller controls as it is activated, and offers all X as green or
// all X as white (a one-colour ManaAmount over {G} and {W}). It is a
// mana ability (CR 605.1a) and does not use the stack.
//
// No simplification.
func init() {
	const oracle = "60fc89a1-cc6f-4a32-84c3-09abf7a4a1df"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Brigid, Clachan's Heart",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventTransform}, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return (ev.Kind == game.EventETB && ev.CardID == source.InstanceID) || ThisTransformedInto(0)(ev, source, lki, g)
			}, "Brigid, Clachan's Heart — create a 1/1 green and white Kithkin creature token",
				Do(CreateToken{Template: TokenCard("1/1 green and white Kithkin"), N: 1})),
			AtYourPrecombatMain("Brigid, Clachan's Heart — you may pay {G} to transform Brigid", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{G}",
					Question: "Brigid, Clachan's Heart — pay {G} to transform Brigid?",
					Then:     ashlingTransformPaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Brigid, Doun's Mind",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: brigidDounsMindMana,
			Label:          "{T}: Add X {G} or X {W}, where X is the number of other creatures you control.",
		}},
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Brigid, Doun's Mind — you may pay {W} to transform Brigid", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{W}",
					Question: "Brigid, Doun's Mind — pay {W} to transform Brigid?",
					Then:     ashlingTransformPaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

// brigidDounsMindMana is "X {G} or X {W}, where X is the number of other
// creatures you control", counted as the ability is activated.
func brigidDounsMindMana(g *game.Game, controller, source uuid.UUID) game.ManaAmount {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.InstanceID != source && c.IsCreature() {
			n++
		}
	}
	return game.ManaAmount{Options: []string{"G", "W"}, N: n, OneColor: true}
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Water Crystal — Legendary Artifact {2}{U}{U}:
//
//	"Blue spells you cast cost {1} less to cast.
//	 If an opponent would mill one or more cards, they mill that many
//	 cards plus four instead.
//	 {4}{U}{U}, {T}: Each opponent mills cards equal to the number of
//	 cards in your hand."
//
// The second printing of the mill-amount replacement (#569), and the
// reason the kind is not shaped around Bruvac the Grandiloquent: the
// two say the same sentence with different arithmetic, ×2 against +4,
// which is exactly the pair CR 616.1's ordering question is about.
// With both on one battlefield an opponent told to mill three is asked
// which order they applied in, and the answers differ — Bruvac first
// is 10 cards, the Crystal first is 14 — so the prompt is real and it
// belongs to the OPPONENT, whoever controls the two artifacts (#982).
//
// "That many plus four" adds to the count the event carries rather
// than to the printed one, which is what lets the two compose.
//
// The cost clause is an ordinary CostsLess with a colour predicate —
// ColoredSpell is new here, and it reads EffectiveColors, so a
// multicolour spell with blue in it is a blue spell (CR 105.2b) and a
// layer-5 colour change is honoured.
//
// The activated ability's count is read when the ability RESOLVES,
// which is the printed card: "cards equal to the number of cards in
// your hand" is not locked in on activation, so a hand emptied in
// response mills nothing. Each opponent's mill is its own instruction,
// so each opens its own window and the Crystal's own clause applies to
// each of them — an opponent milled for three by the Crystal mills
// seven.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f8d2a94f-7be1-4b17-8556-398dde531360",
		Name:         "The Water Crystal",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Blue spells you cast cost {1} less to cast.",
				YourSpell(), ColoredSpell("U")),
		},
		Replacements: []game.ReplacementEffect{
			OpponentsMillPlus(4, "The Water Crystal — mill that many plus four"),
		},
		Activated: []ActivatedAbility{{
			Label: "{4}{U}{U}, {T}: Each opponent mills cards equal to the number of cards in your hand.",
			Cost:  Plus(ManaCost("{4}{U}{U}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				p := g.PlayerByIDForEffect(item.Controller)
				if p == nil {
					return nil
				}
				return b12EachOpponentMills(g, item, p.Hand.Size())
			},
		}},
	})
}

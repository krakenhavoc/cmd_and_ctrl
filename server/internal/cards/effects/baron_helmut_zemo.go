package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baron Helmut Zemo — Legendary Creature — Human Noble Villain {B}{B}{B} 3/3:
//
//	"Whenever you cast a black spell from your hand, Baron Helmut Zemo
//	 connives.
//	 Boast — Exile any number of black cards from your graveyard with
//	 fifteen or more black mana symbols among their mana costs: Copy those
//	 exiled cards. You may cast up to three of the copies without paying
//	 their mana costs. (Activate only if this creature attacked this turn
//	 and only once each turn.)"
//
// The trigger is YouCastFromYourHand over a black spell, and the connive is
// sourceConnives (CR 701.50): it does nothing more when Zemo has left.
//
// The boast is boast.go's constructor, so the engine owns the attack record
// and the once-a-turn tally (CR 702.142a). Its COST is the new part:
// game.ExileCost.AnyNumber with a game.ManaSymbolTotal floor. The activator
// names as many black cards from their own graveyard as they like at the
// announce (CR 602.2b), and the cards named must hold fifteen or more black
// mana symbols among their printed mana costs, or the activation is refused
// before anything is paid (CR 601.2h). A hybrid or Phyrexian symbol that
// offers black counts once ({B/G}, {B/P}, {2/B}); a card with no mana cost
// has none. The exile is a cost, so nobody can respond to it.
//
// The effect is CopyAndCast over ctx.Exiled(), the cards the cost really
// exiled (CR 707.12): a copy of each exists in exile, and the controller
// casts up to three of them without paying their mana costs, during the
// resolution (CR 608.2g), choosing individually which (CR 707.12a). The
// copies not cast cease to exist. The three-cast limit is one permission
// that spends a cast count (game.CastPermission.CastsLeft), so the third
// cast closes it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0fe6c042-1182-4392-8938-fee36149d74f",
		Name:         "Baron Helmut Zemo",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastFromYourHand(OfColor("B")),
				"Baron Helmut Zemo — it connives",
				Do(Instruction(sourceConnives))),
		},
		Activated: []ActivatedAbility{
			// What it answers is whatever the copies it casts do, so the row
			// declares the union of every answer a spell can give (ADR 0142's
			// ruling for a conditional row, as Isochron Scepter does).
			BoastAnswering(game.AnswerProtect|game.AnswerSurvive|game.AnswerPump|
				game.AnswerPrevent|game.AnswerRemove|game.AnswerSacOutlet|game.AnswerRestrict|
				game.AnswerGraveyard|game.AnswerCombatGrant|game.AnswerAnimate|game.AnswerMakesBlocker,
				"Exile any number of black cards from your graveyard with fifteen or more black mana symbols among their mana costs: Copy those exiled cards. You may cast up to three of the copies without paying their mana costs.",
				ExileAnyNumberFromGraveyardWithSymbols("any number of black cards", "B", 15, blackCard),
				func(g *game.Game, item *game.StackItem) error {
					return CopyAndCast{Cards: NewContext(g, item).Exiled(), Free: true, AtMostCasts: 3}.Apply(NewContext(g, item))
				}),
		},
	})
}

// blackCard is "black cards" in the boast cost. It reads the colour as a
// word of Zemo's text (HasColorWord, CR 612.2), through the lens the cost
// walk lends (game.lendCostText), so a text-changing effect on Zemo reaches it.
func blackCard(c game.Card) bool { return c.HasColorWord("B") }

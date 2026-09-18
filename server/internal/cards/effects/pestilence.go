package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pestilence — Enchantment {2}{B}{B} (EDHREC rank 4021):
//
//	"At the beginning of the end step, if no creatures are on the
//	 battlefield, sacrifice this enchantment.
//	 {B}: This enchantment deals 1 damage to each creature and each
//	 player."
//
// The original of the pair — Pyrohemia, already in the catalog, is
// the red reprint with {R} for {B} and nothing else changed. That is
// the whole reason Pestilence is in this batch: two cards with one
// text is how a shared body stays honest, and the body it shares
// (b23DamageEachCreatureAndEachPlayer, b23NoCreaturesOnBattlefield)
// gets its second caller here.
//
// It is also the better half of the pair in practice. Mono-black has
// more ways to make {B} than mono-red has to make {R} — Cabal Coffers,
// Nykthos, Crypt Ghast — so the repeatable sweep goes further, and a
// deck built around it kills the table by emptying the board and then
// pointing the last few activations at the players.
//
// # The order of the two clauses is the card
//
// The self-sacrifice is checked at the beginning of EACH end step —
// any player's, because the printed text says "the end step" with no
// possessive. By then the turn's activations have already happened,
// which is why emptying the board is a winning line rather than a
// suicide: with no creature left, Pestilence dies at the end step,
// but every point it dealt before that has already landed.
//
// The condition is an intervening-if (CR 603.4), checked twice — when
// the trigger would go on the stack, and again as it resolves — so
// someone who flashes in a creature in response keeps the
// enchantment. "No creatures ON THE BATTLEFIELD" is the whole
// battlefield, an opponent's creature included, and it reads
// post-layer types, so an animated land counts and a creature that
// lost its type does not.
//
// # The activation
//
// "{B}:" with no tap and no limit, so it can be activated any number
// of times while black mana lasts, at instant speed, by the
// controller only. Damage, not destruction: an X/2 survives one
// activation and dies to the second, indestructible does not stop the
// lethal-damage state-based action, and prevention shields work as
// printed. Every player is hit too, the controller included —
// Pestilence has never been a one-sided sweeper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dafe63ef-f3d6-45e7-877a-573da92ba85e",
		Name:         "Pestilence",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{B}: Pestilence deals 1 damage to each creature and each player.",
			Cost:  ManaCost("{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b23DamageEachCreatureAndEachPlayer(NewContext(g, item), 1)
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b23NoCreaturesOnBattlefield(g)
			}, "Pestilence — no creatures: sacrifice it", b23SacrificeSelfWhenNoCreatures),
		},
	})
}

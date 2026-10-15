package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primeval Spawn — Creature — Avatar {5}{W}{U}{B}{R}{G}, 10/10:
//
//	"If this creature would enter and it wasn't cast or no mana was
//	 spent to cast it, exile it instead.
//	 Vigilance, trample, lifelink
//	 When this creature leaves the battlefield, exile the top ten cards
//	 of your library. You may cast any number of spells with total mana
//	 value 10 or less from among them without paying their mana costs."
//
// # The entry clause
//
// A CR 614.1 self-replacement reading the entering spell's payment inside
// the entry window (ADR 0109 §11, game.EntrySpentForEffect), the same
// read Freestrider Commando uses. A Spawn that was reanimated, put onto
// the battlefield from a hand or flickered was not cast and is exiled
// instead of entering; one cast without paying its mana cost spent a
// known nothing and is exiled too; one cast for its eleven mana enters.
// Because the entry is rewritten rather than followed by a trigger, an
// exiled Spawn never entered, so its leaves-the-battlefield ability has
// nothing to trigger from.
//
// # The ability
//
// "When this creature leaves the battlefield" is every exit, not only a
// death: On(EventLTB, Self). The ten cards are exiled by the shared
// impulse-exile path, so a commander among them is offered the command
// zone (CR 903.9) and only the cards that really landed in exile are
// castable. The casts are ONE permission with a running total of 10
// (game.CastPermission.ManaValueBudget, ADR 0066's 2026-10-14 amendment
// and its 2026-10-15 follow-up): each cast spends its spell's mana value,
// X at 0 as a free cast locks it, and a card whose mana value is more
// than what is left is refused by the one gate the cast path, the
// client's offers and the bot enumerator share. The order is the
// caster's, and any set with a total of 10 or less can be cast in some
// order without going over, so the running total and the printed "total
// mana value 10 or less" accept the same casts. A land among the ten
// cannot be cast and stays in exile.
//
// As with every "you may cast it" a resolution offers (ADR 0066, ADR
// 0091), the casts are grants taken with ordinary cast_spell actions
// after the ability resolves, and the window closes when the caster next
// passes priority; the cards not cast stay in exile.
//
// One declared simplification, the paid-cost record's: with strict mana
// off the engine never saw what paid, so a Spawn cast while it is off is
// not known to have spent nothing, and is not exiled even when it was
// cast for free (ADR 0068 §3). A Spawn that was not cast at all is
// exiled either way.
func init() {
	Register(Spec{
		OracleID:        "a11c6601-7e91-4def-bfea-bb34023ef3c6",
		Name:            "Primeval Spawn",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"When “Charge mana costs” is off, or you choose Cast anyway, the game doesn't track which mana you spent, so a Primeval Spawn you cast that way is never exiled for being free — one put onto the battlefield still is."},
		PrintedKeywords: []string{"vigilance", "trample", "lifelink"},
		Replacements: []game.ReplacementEffect{
			SelfExiledInsteadIfNoManaSpent("Primeval Spawn"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, Self,
				"Primeval Spawn — exile the top ten cards of your library; you may cast spells with total mana value 10 or less from among them without paying their mana costs",
				primevalSpawnLeaves),
		},
	})
}

// primevalSpawnBudget is "total mana value 10 or less".
const primevalSpawnBudget = 10

// primevalSpawnLeaves exiles the top ten cards of the controller's library
// and grants the shared free cast over the ones that landed.
func primevalSpawnLeaves(g *game.Game, item *game.StackItem) error {
	budget := primevalSpawnBudget
	return ExileTopWithPermission{
		From:                 item.Controller,
		GrantTo:              item.Controller,
		N:                    10,
		TotalManaValueAtMost: &budget,
	}.Apply(NewContext(g, item))
}

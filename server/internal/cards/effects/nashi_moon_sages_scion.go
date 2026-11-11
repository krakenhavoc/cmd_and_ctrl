package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nashi, Moon Sage's Scion — Legendary Creature — Rat Ninja {1}{B}{B}, 3/2
// (Outlaws of Thunder Junction Commander):
//
//	"Ninjutsu {3}{B}
//	 Whenever Nashi deals combat damage to a player, exile the top card of each
//	 player's library. Until end of turn, you may play one of those cards. If
//	 you cast a spell this way, pay life equal to its mana value rather than
//	 paying its mana cost."
//
// Fire Lord Ozai's one-play budget over a card off every library (here every
// player's, yours included, in APNAP order): one permission per card tied
// into one budget, so playing any of them spends the rest, ending with the
// turn. The price is Bolas's Citadel's alternative cost, as Valgavoth, Terror
// Eater claims it: life equal to the mana value, no mana, a cost a player
// short of life cannot claim, and nothing for a land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d04b6c2c-50fc-464d-9991-230efec70d2b",
		Name:         "Nashi, Moon Sage's Scion",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"Nashi — exile the top card of each player's library; you may play one of them, paying life equal to its mana value",
				cmdr2S45NashiExile),
		},
	})
}

func cmdr2S45NashiExile(g *game.Game, item *game.StackItem) error {
	perm := ExileTopWithPermission{}.permission()
	perm.CastsLeft = 1
	perm.BudgetGroup = uuid.New()
	perm.Source = item.SourceCardID
	perm.SourceName = "Nashi, Moon Sage's Scion"
	perm.AltCostKey = "nashi_moon_sages_scion"
	perm.LifeEqualToManaValue = true
	perm.Label = "Play one of those cards, paying life equal to its mana value rather than its mana cost (Nashi, Moon Sage's Scion)"
	step := game.KeyedTail{Ref: exileTopEachStep, Source: NewContext(g, item).Source(), N: 1, Permission: &perm, Player: item.Controller}
	return exileTopEachFrom(g, step, apnapPlayers(g), nil)
}

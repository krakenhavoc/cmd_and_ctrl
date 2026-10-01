package effects

import (
	"slices"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serendib Djinn — Creature — Djinn {2}{U}{U}, 5/6:
//
//	"Flying
//	 At the beginning of your upkeep, sacrifice a land. If you sacrifice
//	 an Island this way, this creature deals 3 damage to you.
//	 When you control no lands, sacrifice this creature."
//
// ADR 0107 §1 (#1858). Flying is the printed keyword.
//
//   - The upkeep sacrifice is the controller's choice of land, made on
//     resolution. "An Island" is read off the land as it last existed on
//     the battlefield (CR 608.2h), so a land that was an Island only
//     through an effect counts. With no land to sacrifice nothing
//     happens (CR 609.3).
//   - "When you control no lands" is a CR 603.8 state trigger: the upkeep
//     that eats the last land triggers it straight away.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "683e7135-de54-49c8-a978-4f84628a6a91",
		Name:            "Serendib Djinn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Serendib Djinn — sacrifice a land", serendibDjinnUpkeep),
			WhenYouControlNo(QueryType("land"), "Serendib Djinn — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}

func serendibDjinnUpkeep(g *game.Game, item *game.StackItem) error {
	return g.PlayerSacrificesThenForEffect(item.SourceCardID, item.Controller,
		sacrificeSpec("a land", Land()), "Serendib Djinn — sacrifice a land", 1,
		func(g *game.Game, sacrificed game.PromptedSacrifices) error {
			for _, id := range sacrificed.By(item.Controller) {
				ref, ok := g.PermanentRefForEffect(id)
				if !ok {
					continue
				}
				info, ok := g.PermanentForEffect(ref)
				if !ok || !slices.ContainsFunc(info.Characteristic.Subtypes, func(s string) bool { return strings.EqualFold(s, "Island") }) {
					continue
				}
				return DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 3}.Apply(NewContext(g, item))
			}
			return nil
		})
}

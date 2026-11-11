package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dire Fleet Daredevil — Creature — Human Pirate {1}{R}, 2/1:
//
//	"First strike
//	 When this creature enters, exile target instant or sorcery card from
//	 an opponent's graveyard. You may cast it this turn, and mana of any
//	 type can be spent to cast that spell. If that spell would be put into
//	 a graveyard, exile it instead."
//
// The exile is part of the ability; the cast is NOT free and is not made
// during the resolution: a permission over the exiled card that lasts the
// turn and keeps the card's own timing (a sorcery waits for a main phase
// of yours with an empty stack), pays the printed cost with mana of any
// type (CR 609.4b), and sends the spell to exile instead of a graveyard
// however it leaves the stack (ExileInsteadOfGraveyard). A card that is not
// cast stays in exile.
func init() {
	enters := WhenThisEnters("Dire Fleet Daredevil — exile target instant or sorcery card from an opponent's graveyard; you may cast it this turn", direFleetDaredevilEffect)
	enters.Targets = TargetCardInGraveyard("target instant or sorcery card from an opponent's graveyard",
		Or(Instant(), Sorcery()), Not(YouOwn()))
	Register(Spec{
		OracleID:        "7a3e50a5-c163-4c41-b62d-d52c233c55b7",
		Name:            "Dire Fleet Daredevil",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}

func direFleetDaredevilEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return g.ExileCardWithPermissionForEffect(t.ID, game.CastPermission{
			Player:                  ctx.Controller(),
			Duration:                g.UntilEndOfTurnDuration(),
			CastOnly:                true,
			AltCostKey:              game.AltCostKeyGrantedCost,
			AnyColor:                true,
			AnyType:                 true,
			ExileInsteadOfGraveyard: true,
			Source:                  ctx.Source(),
			SourceName:              "Dire Fleet Daredevil",
			Label:                   "Dire Fleet Daredevil — you may cast it this turn, spending mana of any type",
		})
	}
	return nil
}
